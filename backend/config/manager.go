package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"db-desktop/backend/integration"
	"db-desktop/backend/utils"
)

// ConnectionConfig 数据库连接配置
type ConnectionConfig struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
	Status   string `json:"status"`
}

// UnifiedConfig 统一配置结构
type UnifiedConfig struct {
	AI          integration.AIConfig `json:"ai"`
	Connections []*ConnectionConfig  `json:"connections"`
	Version     string               `json:"version"`
	UpdatedAt   string               `json:"updatedAt"`
}

// ConfigManager 统一配置管理器
type ConfigManager struct {
	configFile string
	config     *UnifiedConfig
	mu         sync.RWMutex
}

var (
	globalConfigManager *ConfigManager
	initOnce            sync.Once
)

// GetGlobalConfigManager 获取全局配置管理器实例
func GetGlobalConfigManager() *ConfigManager {
	initOnce.Do(func() {
		homeDir, _ := os.UserHomeDir()
		configFile := filepath.Join(homeDir, ".db-desktop", "config.json")

		globalConfigManager = &ConfigManager{
			configFile: configFile,
			config: &UnifiedConfig{
				AI: integration.AIConfig{
					APIKey:      "",
					BaseURL:     "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions",
					Temperature: 0.7,
					Stream:      true,
				},
				Connections: make([]*ConnectionConfig, 0),
				Version:     "1.0.0",
			},
		}

		// 尝试加载配置
		if err := globalConfigManager.LoadConfig(); err != nil {
			utils.Warnf("Failed to load unified config, using defaults: %v", err)
		}
	})
	return globalConfigManager
}

// LoadConfig 加载统一配置文件
func (cm *ConfigManager) LoadConfig() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 创建配置目录
	configDir := filepath.Dir(cm.configFile)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// 检查统一配置文件是否存在
	if _, err := os.Stat(cm.configFile); os.IsNotExist(err) {
		// 尝试迁移旧配置文件
		if err := cm.migrateOldConfigs(); err != nil {
			utils.Warnf("Failed to migrate old configs: %v", err)
		}
		return cm.SaveConfig() // 保存默认配置或迁移后的配置
	}

	// 读取配置文件
	data, err := os.ReadFile(cm.configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config UnifiedConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	cm.config = &config
	utils.Infof("Loaded unified config: AI configured=%t, Connections=%d",
		config.AI.APIKey != "", len(config.Connections))
	return nil
}

// SaveConfig 保存统一配置文件
func (cm *ConfigManager) SaveConfig() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 更新时间戳
	cm.config.UpdatedAt = fmt.Sprintf("%d", os.Getpid()) // 简单的时间戳

	// 创建配置目录
	configDir := filepath.Dir(cm.configFile)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cm.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(cm.configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	utils.Infof("Successfully saved unified config")
	return nil
}

// migrateOldConfigs 迁移旧的配置文件
func (cm *ConfigManager) migrateOldConfigs() error {
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, ".db-desktop")

	migrated := false

	// 迁移AI配置
	aiConfigFile := filepath.Join(configDir, "ai_config.json")
	if _, err := os.Stat(aiConfigFile); err == nil {
		data, err := os.ReadFile(aiConfigFile)
		if err == nil {
			var aiConfig integration.AIConfig
			if err := json.Unmarshal(data, &aiConfig); err == nil {
				cm.config.AI = aiConfig
				migrated = true
				utils.Infof("Migrated AI config from %s", aiConfigFile)
			}
		}
	}

	// 迁移连接配置
	connectionsFile := filepath.Join(configDir, "connections.json")
	if _, err := os.Stat(connectionsFile); err == nil {
		data, err := os.ReadFile(connectionsFile)
		if err == nil {
			var connections []*ConnectionConfig
			if err := json.Unmarshal(data, &connections); err == nil {
				cm.config.Connections = connections
				migrated = true
				utils.Infof("Migrated %d connections from %s", len(connections), connectionsFile)
			}
		}
	}

	if migrated {
		utils.Infof("Successfully migrated old configuration files")
		// 备份旧配置文件
		if err := cm.backupOldConfigs(); err != nil {
			utils.Warnf("Failed to backup old config files: %v", err)
		}
	}

	return nil
}

// backupOldConfigs 备份旧配置文件
func (cm *ConfigManager) backupOldConfigs() error {
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, ".db-desktop")
	backupDir := filepath.Join(configDir, "backup")

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return err
	}

	// 备份AI配置
	aiConfigFile := filepath.Join(configDir, "ai_config.json")
	if _, err := os.Stat(aiConfigFile); err == nil {
		backupFile := filepath.Join(backupDir, "ai_config.json.bak")
		if err := copyFile(aiConfigFile, backupFile); err == nil {
			os.Remove(aiConfigFile) // 删除原文件
		}
	}

	// 备份连接配置
	connectionsFile := filepath.Join(configDir, "connections.json")
	if _, err := os.Stat(connectionsFile); err == nil {
		backupFile := filepath.Join(backupDir, "connections.json.bak")
		if err := copyFile(connectionsFile, backupFile); err == nil {
			os.Remove(connectionsFile) // 删除原文件
		}
	}

	return nil
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

// GetAIConfig 获取AI配置
func (cm *ConfigManager) GetAIConfig() integration.AIConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config.AI
}

// UpdateAIConfig 更新AI配置
func (cm *ConfigManager) UpdateAIConfig(config integration.AIConfig) error {
	cm.mu.Lock()
	cm.config.AI = config
	cm.mu.Unlock()
	return cm.SaveConfig()
}

// GetConnections 获取所有连接配置
func (cm *ConfigManager) GetConnections() []*ConnectionConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// 返回副本以避免并发修改
	connections := make([]*ConnectionConfig, len(cm.config.Connections))
	copy(connections, cm.config.Connections)
	return connections
}

// SaveConnections 保存连接配置
func (cm *ConfigManager) SaveConnections(connections []*ConnectionConfig) error {
	cm.mu.Lock()
	cm.config.Connections = connections
	cm.mu.Unlock()
	return cm.SaveConfig()
}

// AddConnection 添加连接配置
func (cm *ConfigManager) AddConnection(conn *ConnectionConfig) error {
	cm.mu.Lock()
	cm.config.Connections = append(cm.config.Connections, conn)
	cm.mu.Unlock()
	return cm.SaveConfig()
}

// UpdateConnection 更新连接配置
func (cm *ConfigManager) UpdateConnection(conn *ConnectionConfig) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for i, existing := range cm.config.Connections {
		if existing.ID == conn.ID {
			cm.config.Connections[i] = conn
			break
		}
	}

	return cm.SaveConfig()
}

// DeleteConnection 删除连接配置
func (cm *ConfigManager) DeleteConnection(id string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for i, conn := range cm.config.Connections {
		if conn.ID == id {
			cm.config.Connections = append(cm.config.Connections[:i], cm.config.Connections[i+1:]...)
			break
		}
	}

	return cm.SaveConfig()
}
