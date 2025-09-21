package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"db-desktop/backend/config"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// 全局数据库管理器实例
var (
	dbManager *SimpleDatabaseManager
	initOnce  sync.Once
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

// init 初始化数据库管理器
func init() {
	initOnce.Do(func() {
		homeDir, _ := os.UserHomeDir()
		configFile := filepath.Join(homeDir, ".db-desktop", "connections.json")

		logger := logrus.New()
		logger.SetLevel(logrus.InfoLevel)
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})

		dbManager = &SimpleDatabaseManager{
			mysqlConnections:      make(map[string]*sql.DB),
			redisConnections:      make(map[string]*redis.Client),
			clickhouseConnections: make(map[string]*sql.DB),
			connections:           make(map[string]*ConnectionConfig),
			logger:                logger,
			configFile:            configFile,
			ctx:                   context.Background(),
		}

		// 加载连接配置
		if err := dbManager.LoadConnections(); err != nil {
			panic(fmt.Sprintf("Failed to initialize database manager: %v", err))
		}
	})
}

// LoadConnections 加载保存的连接
func (s *SimpleDatabaseManager) LoadConnections() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 使用统一配置管理器
	configManager := config.GetGlobalConfigManager()
	configConnections := configManager.GetConnections()

	// 清空现有连接
	s.connections = make(map[string]*ConnectionConfig)

	// 加载连接，转换类型
	for _, conn := range configConnections {
		dbConn := &ConnectionConfig{
			ID:       conn.ID,
			Name:     conn.Name,
			Type:     DatabaseType(conn.Type),
			Host:     conn.Host,
			Port:     conn.Port,
			Username: conn.Username,
			Password: conn.Password,
			Database: conn.Database,
			Status:   conn.Status,
		}
		s.connections[conn.ID] = dbConn
	}

	s.logger.Infof("Loaded %d connections from unified config", len(configConnections))
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

	// 转换为config包的类型
	configConnections := make([]*config.ConnectionConfig, 0, len(connections))
	for _, conn := range connections {
		configConn := &config.ConnectionConfig{
			ID:       conn.ID,
			Name:     conn.Name,
			Type:     string(conn.Type),
			Host:     conn.Host,
			Port:     conn.Port,
			Username: conn.Username,
			Password: conn.Password,
			Database: conn.Database,
			Status:   conn.Status,
		}
		configConnections = append(configConnections, configConn)
	}

	// 使用统一配置管理器保存
	configManager := config.GetGlobalConfigManager()
	if err := configManager.SaveConnections(configConnections); err != nil {
		return fmt.Errorf("failed to save connections: %w", err)
	}

	s.logger.Infof("Saved %d connections to unified config", len(connections))
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
		s.logger.WithField("connectionID", config.ID).Error("Connection not found in UpdateConnection")
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

// 包级函数，替代原来的对象方法调用

// LoadConnections 加载保存的连接
func LoadConnections() error {
	return dbManager.LoadConnections()
}

// SaveConnections 保存连接到文件
func SaveConnections() error {
	return dbManager.SaveConnections()
}

// AddConnection 添加新连接
func AddConnection(config *ConnectionConfig) error {
	return dbManager.AddConnection(config)
}

// UpdateConnection 更新连接
func UpdateConnection(config *ConnectionConfig) error {
	return dbManager.UpdateConnection(config)
}

// DeleteConnection 删除连接
func DeleteConnection(id string) error {
	return dbManager.DeleteConnection(id)
}

// ListConnections 列出所有连接
func ListConnections() []*ConnectionConfig {
	return dbManager.ListConnections()
}

// Connect 建立连接
func Connect(id string) error {
	return dbManager.Connect(id)
}

// Disconnect 断开连接
func Disconnect(id string) error {
	return dbManager.Disconnect(id)
}

// TestConnection 测试连接
func TestConnection(config *ConnectionConfig) error {
	return dbManager.TestConnection(config)
}

// ExecuteQuery 执行查询
func ExecuteQuery(connectionID string, query string) (*QueryResult, error) {
	return dbManager.ExecuteQuery(connectionID, query)
}

// ExecuteQueryWithLimit 执行带限制的查询
func ExecuteQueryWithLimit(connectionID string, query string, limit int) (*QueryResult, error) {
	return dbManager.ExecuteQueryWithLimit(connectionID, query, limit)
}

// GetDatabases 获取数据库列表
func GetDatabases(connectionID string) ([]string, error) {
	return dbManager.GetDatabases(connectionID)
}

// GetTables 获取表列表
func GetTables(connectionID string, database string) ([]TableInfo, error) {
	return dbManager.GetTables(connectionID, database)
}

// GetTableInfo 获取表信息
func GetTableInfo(connectionID string, database string, table string) (*TableInfo, error) {
	return dbManager.GetTableInfo(connectionID, database, table)
}

// GetTableData 获取表数据
func GetTableData(connectionID string, database string, table string, limit int, offset int) (*QueryResult, error) {
	return dbManager.GetTableData(connectionID, database, table, limit, offset)
}

// GetDatabaseInfo 获取数据库信息
func GetDatabaseInfo(connectionID string) (*DatabaseInfo, error) {
	return dbManager.GetDatabaseInfo(connectionID)
}

// FormatQuery 格式化查询
func FormatQuery(connectionID string, query string) string {
	return dbManager.FormatQuery(connectionID, query)
}

// ValidateQuery 验证查询
func ValidateQuery(connectionID string, query string) error {
	return dbManager.ValidateQuery(connectionID, query)
}

// GetConnection 获取连接配置
func GetConnection(id string) (*ConnectionConfig, error) {
	return dbManager.GetConnection(id)
}

// GetConnectionStatus 获取连接状态
func GetConnectionStatus(id string) *ConnectionStatus {
	return dbManager.GetConnectionStatus(id)
}

// 内部方法实现将在后续文件中提供
