package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// SimpleDatabaseManager 简化的数据库管理器，直接管理三种数据库类型
type SimpleDatabaseManager struct {
	// MySQL 连接
	mysqlConnections map[string]*sql.DB

	// Redis 连接
	redisConnections map[string]*redis.Client

	// ClickHouse 连接
	clickhouseConnections map[string]*sql.DB

	// 连接配置
	connections map[string]*ConnectionConfig

	mu         sync.RWMutex
	logger     *logrus.Logger
	configFile string
	ctx        context.Context
}

// NewSimpleDatabaseManager 创建简化的数据库管理器
func NewSimpleDatabaseManager() *SimpleDatabaseManager {
	homeDir, _ := os.UserHomeDir()
	configFile := filepath.Join(homeDir, ".db-desktop", "connections.json")

	return &SimpleDatabaseManager{
		mysqlConnections:      make(map[string]*sql.DB),
		redisConnections:      make(map[string]*redis.Client),
		clickhouseConnections: make(map[string]*sql.DB),
		connections:           make(map[string]*ConnectionConfig),
		logger:                logrus.New(),
		configFile:            configFile,
		ctx:                   context.Background(),
	}
}

// LoadConnections 加载保存的连接
func (s *SimpleDatabaseManager) LoadConnections() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 创建配置目录
	configDir := filepath.Dir(s.configFile)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// 检查配置文件是否存在
	if _, err := os.Stat(s.configFile); os.IsNotExist(err) {
		return nil // 没有配置文件，从空连接开始
	}

	data, err := os.ReadFile(s.configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var connections []*ConnectionConfig
	if err := json.Unmarshal(data, &connections); err != nil {
		return fmt.Errorf("failed to unmarshal connections: %w", err)
	}

	for _, conn := range connections {
		s.connections[conn.ID] = conn
	}

	s.logger.Infof("Loaded %d connections from config file", len(connections))
	return nil
}

// SaveConnections 保存连接到文件
func (s *SimpleDatabaseManager) SaveConnections() error {
	s.mu.RLock()
	connections := make([]*ConnectionConfig, 0, len(s.connections))
	for _, conn := range s.connections {
		connections = append(connections, conn)
	}
	s.mu.RUnlock()

	// 创建配置目录
	configDir := filepath.Dir(s.configFile)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(connections, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal connections: %w", err)
	}

	if err := os.WriteFile(s.configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	s.logger.Infof("Successfully saved %d connections to config file", len(connections))
	return nil
}

// AddConnection 添加新连接
func (s *SimpleDatabaseManager) AddConnection(config *ConnectionConfig) error {
	s.mu.Lock()

	// 生成ID
	if config.ID == "" {
		config.ID = fmt.Sprintf("%s_%d", config.Type, time.Now().Unix())
	}

	// 设置时间戳
	now := time.Now()
	if config.CreatedAt.IsZero() {
		config.CreatedAt = now
	}
	config.UpdatedAt = now

	s.connections[config.ID] = config
	s.mu.Unlock()

	// 保存到文件
	if err := s.SaveConnections(); err != nil {
		// 如果保存失败，从内存中删除连接
		s.mu.Lock()
		delete(s.connections, config.ID)
		s.mu.Unlock()
		return err
	}

	s.logger.Infof("Added connection: %s (%s)", config.Name, config.Type)
	return nil
}

// UpdateConnection 更新连接
func (s *SimpleDatabaseManager) UpdateConnection(config *ConnectionConfig) error {
	s.mu.Lock()

	if _, exists := s.connections[config.ID]; !exists {
		s.mu.Unlock()
		return fmt.Errorf("connection not found: %s", config.ID)
	}

	config.UpdatedAt = time.Now()
	s.connections[config.ID] = config
	s.mu.Unlock()

	// 保存到文件
	if err := s.SaveConnections(); err != nil {
		return err
	}

	s.logger.Infof("Updated connection: %s (%s)", config.Name, config.Type)
	return nil
}

// DeleteConnection 删除连接
func (s *SimpleDatabaseManager) DeleteConnection(id string) error {
	s.mu.Lock()

	conn, exists := s.connections[id]
	if !exists {
		s.mu.Unlock()
		return fmt.Errorf("connection not found: %s", id)
	}

	// 断开连接
	s.disconnectInternal(id, conn.Type)

	delete(s.connections, id)
	s.mu.Unlock()

	// 保存到文件
	if err := s.SaveConnections(); err != nil {
		return err
	}

	s.logger.Infof("Deleted connection: %s", id)
	return nil
}

// ListConnections 列出所有连接
func (s *SimpleDatabaseManager) ListConnections() []*ConnectionConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	connections := make([]*ConnectionConfig, 0, len(s.connections))
	for _, conn := range s.connections {
		// 检查连接状态
		status := s.getConnectionStatusInternal(conn)
		conn.Status = status.Status
		conn.LastPing = status.LastPing
		conn.Message = status.Message
		connections = append(connections, conn)
	}

	return connections
}

// Connect 建立连接
func (s *SimpleDatabaseManager) Connect(id string) error {
	s.mu.RLock()
	config, exists := s.connections[id]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("connection not found: %s", id)
	}

	return s.connectInternal(config)
}

// Disconnect 断开连接
func (s *SimpleDatabaseManager) Disconnect(id string) error {
	s.mu.RLock()
	config, exists := s.connections[id]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("connection not found: %s", id)
	}

	return s.disconnectInternal(id, config.Type)
}

// TestConnection 测试连接
func (s *SimpleDatabaseManager) TestConnection(config *ConnectionConfig) error {
	switch config.Type {
	case MySQL:
		return s.testMySQLConnection(config)
	case Redis:
		return s.testRedisConnection(config)
	case ClickHouse:
		return s.testClickHouseConnection(config)
	default:
		return fmt.Errorf("unsupported database type: %s", config.Type)
	}
}

// ExecuteQuery 执行查询
func (s *SimpleDatabaseManager) ExecuteQuery(connectionID string, query string) (*QueryResult, error) {
	s.mu.RLock()
	config, exists := s.connections[connectionID]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("connection not found: %s", connectionID)
	}

	switch config.Type {
	case MySQL:
		return s.executeMySQLQuery(connectionID, query)
	case Redis:
		return s.executeRedisQuery(connectionID, query)
	case ClickHouse:
		return s.executeClickHouseQuery(connectionID, query)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.Type)
	}
}

// ExecuteQueryWithLimit 执行带限制的查询
func (s *SimpleDatabaseManager) ExecuteQueryWithLimit(connectionID string, query string, limit int) (*QueryResult, error) {
	s.mu.RLock()
	config, exists := s.connections[connectionID]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("connection not found: %s", connectionID)
	}

	switch config.Type {
	case MySQL:
		return s.executeMySQLQueryWithLimit(connectionID, query, limit)
	case Redis:
		return s.executeRedisQueryWithLimit(connectionID, query, limit)
	case ClickHouse:
		return s.executeClickHouseQueryWithLimit(connectionID, query, limit)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.Type)
	}
}

// GetDatabases 获取数据库列表
func (s *SimpleDatabaseManager) GetDatabases(connectionID string) ([]string, error) {
	s.mu.RLock()
	config, exists := s.connections[connectionID]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("connection not found: %s", connectionID)
	}

	switch config.Type {
	case MySQL:
		return s.getMySQLDatabases(connectionID)
	case Redis:
		return s.getRedisDatabases(connectionID)
	case ClickHouse:
		return s.getClickHouseDatabases(connectionID)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.Type)
	}
}

// GetTables 获取表列表
func (s *SimpleDatabaseManager) GetTables(connectionID string, database string) ([]TableInfo, error) {
	s.mu.RLock()
	config, exists := s.connections[connectionID]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("connection not found: %s", connectionID)
	}

	switch config.Type {
	case MySQL:
		return s.getMySQLTables(connectionID, database)
	case Redis:
		return s.getRedisTables(connectionID, database)
	case ClickHouse:
		return s.getClickHouseTables(connectionID, database)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.Type)
	}
}

// GetTableInfo 获取表信息
func (s *SimpleDatabaseManager) GetTableInfo(connectionID string, database string, table string) (*TableInfo, error) {
	s.mu.RLock()
	config, exists := s.connections[connectionID]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("connection not found: %s", connectionID)
	}

	switch config.Type {
	case MySQL:
		return s.getMySQLTableInfo(connectionID, database, table)
	case Redis:
		return s.getRedisTableInfo(connectionID, database, table)
	case ClickHouse:
		return s.getClickHouseTableInfo(connectionID, database, table)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.Type)
	}
}

// GetTableData 获取表数据
func (s *SimpleDatabaseManager) GetTableData(connectionID string, database string, table string, limit int, offset int) (*QueryResult, error) {
	s.mu.RLock()
	config, exists := s.connections[connectionID]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("connection not found: %s", connectionID)
	}

	switch config.Type {
	case MySQL:
		return s.getMySQLTableData(connectionID, database, table, limit, offset)
	case Redis:
		return s.getRedisTableData(connectionID, database, table, limit, offset)
	case ClickHouse:
		return s.getClickHouseTableData(connectionID, database, table, limit, offset)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.Type)
	}
}

// GetDatabaseInfo 获取数据库信息
func (s *SimpleDatabaseManager) GetDatabaseInfo(connectionID string) (*DatabaseInfo, error) {
	s.mu.RLock()
	config, exists := s.connections[connectionID]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("connection not found: %s", connectionID)
	}

	switch config.Type {
	case MySQL:
		return s.getMySQLDatabaseInfo(connectionID)
	case Redis:
		return s.getRedisDatabaseInfo(connectionID)
	case ClickHouse:
		return s.getClickHouseDatabaseInfo(connectionID)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.Type)
	}
}

// FormatQuery 格式化查询
func (s *SimpleDatabaseManager) FormatQuery(connectionID string, query string) string {
	s.mu.RLock()
	config, exists := s.connections[connectionID]
	s.mu.RUnlock()

	if !exists {
		return query
	}

	switch config.Type {
	case MySQL, ClickHouse:
		return strings.TrimSpace(query)
	case Redis:
		return strings.TrimSpace(query)
	default:
		return query
	}
}

// ValidateQuery 验证查询
func (s *SimpleDatabaseManager) ValidateQuery(connectionID string, query string) error {
	s.mu.RLock()
	config, exists := s.connections[connectionID]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("connection not found: %s", connectionID)
	}

	switch config.Type {
	case MySQL:
		return s.validateMySQLQuery(query)
	case Redis:
		return s.validateRedisQuery(query)
	case ClickHouse:
		return s.validateClickHouseQuery(query)
	default:
		return fmt.Errorf("unsupported database type: %s", config.Type)
	}
}

// GetConnection 获取连接配置
func (s *SimpleDatabaseManager) GetConnection(id string) (*ConnectionConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	conn, exists := s.connections[id]
	if !exists {
		return nil, fmt.Errorf("connection not found: %s", id)
	}

	return conn, nil
}

// GetConnectionStatus 获取连接状态
func (s *SimpleDatabaseManager) GetConnectionStatus(id string) *ConnectionStatus {
	s.mu.RLock()
	config, exists := s.connections[id]
	s.mu.RUnlock()

	if !exists {
		return &ConnectionStatus{
			ID:      id,
			Status:  "error",
			Message: "connection not found",
		}
	}

	return s.getConnectionStatusInternal(config)
}

// 内部方法实现将在后续文件中提供
