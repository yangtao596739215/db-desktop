package app

import (
	"context"
	"db-desktop/backend/database"
	"db-desktop/backend/integration"
	"db-desktop/backend/logic"
	"db-desktop/backend/models"
	"db-desktop/backend/sqlite"
	"db-desktop/backend/utils"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// 自定义错误类型
var (
	ErrConversationIDRequired = errors.New("CONVERSATION_ID_REQUIRED")
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	// 设置全局实例
	logic.SetGlobalDatabase(&databaseAdapter{})

	// 设置SQLite管理器
	logic.SetGlobalSQLiteManager(&sqliteAdapter{})

	// 验证所有组件都已正确初始化
	if err := validateInitialization(); err != nil {
		utils.Errorf("Initialization validation failed: %v", err)
		panic(fmt.Sprintf("Failed to initialize application: %v", err))
	}

	utils.Infof("✅ Application initialized successfully")
	return &App{}
}

// validateInitialization 验证所有组件都已正确初始化
func validateInitialization() error {
	// 验证 logic 包中的全局变量
	if logic.GetAIService() == nil {
		return fmt.Errorf("AIService is not initialized")
	}

	if logic.GetCardManager() == nil {
		return fmt.Errorf("CardManager is not initialized")
	}

	if logic.GetGlobalDatabase() == nil {
		return fmt.Errorf("GlobalDatabase is not initialized")
	}

	if logic.GetGlobalSQLiteManager() == nil {
		return fmt.Errorf("GlobalSQLiteManager is not initialized")
	}

	return nil
}

// databaseAdapter 适配器，将database包的函数适配到DatabaseInterface接口
type databaseAdapter struct{}

func (d *databaseAdapter) ExecuteQuery(connectionID string, query string) (*database.QueryResult, error) {
	return database.ExecuteQuery(connectionID, query)
}

func (d *databaseAdapter) ListConnections() []*database.ConnectionConfig {
	return database.ListConnections()
}

func (d *databaseAdapter) GetConnectionStatus(id string) *database.ConnectionStatus {
	return database.GetConnectionStatus(id)
}

// sqliteAdapter 适配器，将sqlite包的函数适配到SQLiteManagerInterface接口
type sqliteAdapter struct{}

func (s *sqliteAdapter) AddMessageToConversation(conversationID string, message *models.Message) error {
	return sqlite.AddMessageToConversation(conversationID, message)
}

func (s *sqliteAdapter) GetMessagesForLLM(conversationID string) ([]*models.Message, error) {
	return sqlite.GetMessagesForLLM(conversationID)
}

// Startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	// Load saved connections
	if err := database.LoadConnections(); err != nil {
		utils.Errorf("Failed to load connections: %v", err)
	}

	// Load AI configuration
	if err := logic.LoadConfig(); err != nil {
		utils.Errorf("Failed to load AI config: %v", err)
	}

	// Initialize default connections if none exist
	a.initializeDefaultConnections()

	// Auto-connect to all saved connections
	a.autoConnectToDatabases()

	utils.Infof("Database Desktop application started")
}

// initializeDefaultConnections loads connections from config file and attempts to connect
func (a *App) initializeDefaultConnections() {
	connections := database.ListConnections()
	if len(connections) == 0 {
		utils.Infof("No saved connections found in config file")
		return
	}

	utils.Infof("📋 Found %d saved connections in config file", len(connections))

	// Log all available connections
	for _, conn := range connections {
		utils.Infof("  - %s (%s) - %s:%d", conn.Name, conn.Type, conn.Host, conn.Port)
	}
}

// autoConnectToDatabases automatically connects to all saved database connections
func (a *App) autoConnectToDatabases() {
	utils.Infof("🔄 Auto-connecting to saved database connections...")

	connections := database.ListConnections()
	if len(connections) == 0 {
		utils.Infof("No saved connections found, skipping auto-connect")
		return
	}

	connectedCount := 0
	for _, conn := range connections {
		utils.Infof("🔌 Attempting to connect to %s (%s)...", conn.Name, conn.Type)

		if err := database.Connect(conn.ID); err != nil {
			utils.Warnf("❌ Failed to connect to %s (%s): %v", conn.Name, conn.Type, err)
		} else {
			utils.Infof("✅ Successfully connected to %s (%s)", conn.Name, conn.Type)
			connectedCount++
		}
	}

	utils.Infof("🎯 Auto-connect completed: %d/%d connections successful", connectedCount, len(connections))
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// Database Connection Management

// AddConnection adds a new database connection
func (a *App) AddConnection(config map[string]interface{}) error {
	utils.Infof("AddConnection called with config: %+v", config)

	// Convert map to ConnectionConfig
	connConfig := &database.ConnectionConfig{
		Name:     getString(config, "name"),
		Type:     database.DatabaseType(getString(config, "type")),
		Host:     getString(config, "host"),
		Port:     getInt(config, "port"),
		Username: getString(config, "username"),
		Password: getString(config, "password"),
		Database: getString(config, "database"),
		SSLMode:  getString(config, "sslMode"),
		Timeout:  getInt(config, "timeout"), // Store as seconds
		MaxConns: getInt(config, "maxConns"),
	}

	utils.Infof("Converted config: %+v", connConfig)

	// Use a channel to handle the operation with timeout
	resultChan := make(chan error, 1)
	go func() {
		resultChan <- database.AddConnection(connConfig)
	}()

	// Wait for result with timeout
	select {
	case err := <-resultChan:
		if err != nil {
			utils.Errorf("AddConnection failed: %v", err)
			return err
		}
		utils.Infof("AddConnection successful")
		return nil
	case <-time.After(10 * time.Second):
		utils.Errorf("AddConnection timeout after 10 seconds")
		return fmt.Errorf("operation timeout")
	}
}

// UpdateConnection updates an existing database connection
func (a *App) UpdateConnection(config map[string]interface{}) error {
	// Convert map to ConnectionConfig
	connConfig := &database.ConnectionConfig{
		ID:       getString(config, "id"),
		Name:     getString(config, "name"),
		Type:     database.DatabaseType(getString(config, "type")),
		Host:     getString(config, "host"),
		Port:     getInt(config, "port"),
		Username: getString(config, "username"),
		Password: getString(config, "password"),
		Database: getString(config, "database"),
		SSLMode:  getString(config, "sslMode"),
		Timeout:  getInt(config, "timeout"), // Store as seconds
		MaxConns: getInt(config, "maxConns"),
	}

	return database.UpdateConnection(connConfig)
}

// DeleteConnection deletes a database connection
func (a *App) DeleteConnection(id string) error {
	return database.DeleteConnection(id)
}

// GetConnection returns a connection configuration by ID
func (a *App) GetConnection(id string) (*database.ConnectionConfig, error) {
	return database.GetConnection(id)
}

// ListConnections returns all connection configurations
func (a *App) ListConnections() []*database.ConnectionConfig {
	return database.ListConnections()
}

// Connect establishes a connection to a database
func (a *App) Connect(id string) error {
	return database.Connect(id)
}

// Disconnect closes a database connection
func (a *App) Disconnect(id string) error {
	return database.Disconnect(id)
}

// TestConnection tests a database connection
func (a *App) TestConnection(config map[string]interface{}) error {
	utils.Infof("TestConnection called with config: %+v", config)

	// Convert map to ConnectionConfig
	connConfig := &database.ConnectionConfig{
		Name:     getString(config, "name"),
		Type:     database.DatabaseType(getString(config, "type")),
		Host:     getString(config, "host"),
		Port:     getInt(config, "port"),
		Username: getString(config, "username"),
		Password: getString(config, "password"),
		Database: getString(config, "database"),
		SSLMode:  getString(config, "sslMode"),
		Timeout:  getInt(config, "timeout"), // Store as seconds
		MaxConns: getInt(config, "maxConns"),
	}

	utils.Infof("Converted config for test: %+v", connConfig)

	err := database.TestConnection(connConfig)
	if err != nil {
		utils.Errorf("TestConnection failed: %v", err)
		return err
	}

	utils.Infof("TestConnection successful")
	return nil
}

// GetConnectionStatus returns the status of a database connection
func (a *App) GetConnectionStatus(id string) *database.ConnectionStatus {
	return database.GetConnectionStatus(id)
}

// Query Operations

// ExecuteQuery executes a query on a database
func (a *App) ExecuteQuery(connectionID string, query string) (*database.QueryResult, error) {
	utils.Infof("ExecuteQuery called - ConnectionID: %s, Query: %s", connectionID, query)

	result, err := database.ExecuteQuery(connectionID, query)

	// 获取连接信息以确定数据库类型
	connection, err2 := database.GetConnection(connectionID)
	var dbType string
	if err2 == nil && connection != nil {
		dbType = string(connection.Type)
		utils.Infof("Database type detected: %s for connection: %s", dbType, connectionID)
	} else {
		// 如果无法获取连接信息，尝试从连接ID推断
		if strings.HasPrefix(connectionID, "mysql_") {
			dbType = "mysql"
		} else if strings.HasPrefix(connectionID, "redis_") {
			dbType = "redis"
		} else if strings.HasPrefix(connectionID, "clickhouse_") {
			dbType = "clickhouse"
		} else {
			dbType = "unknown"
		}
		utils.Warnf("Failed to get connection info for %s, inferred type: %s", connectionID, dbType)
	}

	// 保存查询历史
	success := err == nil
	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
	}

	// 确保result不为nil
	var executionTime int64
	var resultCount int
	if result != nil {
		executionTime = int64(result.Time)
		resultCount = result.Count
	}

	// 获取连接名称
	var connectionName string
	if connection != nil {
		connectionName = connection.Name
	} else {
		connectionName = "Unknown Connection"
	}

	utils.Infof("Saving query history - Query: %s, DBType: %s, Connection: %s, Success: %v", query, dbType, connectionName, success)

	_, historyErr := sqlite.AddQueryHistory(
		query,
		dbType,
		connectionID,
		connectionName,
		executionTime,
		success,
		errorMsg,
		resultCount,
	)
	if historyErr != nil {
		utils.Errorf("Failed to save query history: %v", historyErr)
	} else {
		utils.Infof("Query history saved successfully")
	}

	if err != nil {
		utils.Errorf("ExecuteQuery failed - ConnectionID: %s, Error: %v", connectionID, err)
		return result, err
	}

	utils.Infof("ExecuteQuery success - ConnectionID: %s, Rows: %d, Time: %dms",
		connectionID, result.Count, result.Time)

	return result, nil
}

// ExecuteQueryWithLimit executes a query with limit
func (a *App) ExecuteQueryWithLimit(connectionID string, query string, limit int) (*database.QueryResult, error) {
	utils.Infof("ExecuteQueryWithLimit called - ConnectionID: %s, Query: %s, Limit: %d", connectionID, query, limit)

	result, err := database.ExecuteQueryWithLimit(connectionID, query, limit)
	if err != nil {
		utils.Errorf("ExecuteQueryWithLimit failed - ConnectionID: %s, Error: %v", connectionID, err)
		return result, err
	}

	utils.Infof("ExecuteQueryWithLimit success - ConnectionID: %s, Rows: %d, Time: %dms",
		connectionID, result.Count, result.Time)

	return result, nil
}

// GetDatabases returns list of databases
func (a *App) GetDatabases(connectionID string) ([]string, error) {
	return database.GetDatabases(connectionID)
}

// GetTables returns list of tables in a database
func (a *App) GetTables(connectionID string, dbName string) ([]database.TableInfo, error) {
	return database.GetTables(connectionID, dbName)
}

// GetTableInfo returns detailed information about a table
func (a *App) GetTableInfo(connectionID string, dbName string, table string) (*database.TableInfo, error) {
	return database.GetTableInfo(connectionID, dbName, table)
}

// GetTableData returns data from a table with pagination
func (a *App) GetTableData(connectionID string, dbName string, table string, limit int, offset int) (*database.QueryResult, error) {
	return database.GetTableData(connectionID, dbName, table, limit, offset)
}

// GetDatabaseInfo returns general database information
func (a *App) GetDatabaseInfo(connectionID string) (*database.DatabaseInfo, error) {
	return database.GetDatabaseInfo(connectionID)
}

// Utility Operations

// FormatQuery formats a query
func (a *App) FormatQuery(connectionID string, query string) string {
	return database.FormatQuery(connectionID, query)
}

// ValidateQuery validates a query
func (a *App) ValidateQuery(connectionID string, query string) error {
	return database.ValidateQuery(connectionID, query)
}

// GetSupportedDatabaseTypes returns list of supported database types
func (a *App) GetSupportedDatabaseTypes() []string {
	return []string{"mysql", "redis", "clickhouse"}
}

// Helper functions for map conversion
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getInt(m map[string]interface{}, key string) int {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				return i
			}
		}
	}
	return 0
}

// AI Assistant Methods

// GetAIConfig gets the current AI configuration
func (a *App) GetAIConfig() integration.AIConfig {
	return logic.GetConfig()
}

// UpdateAIConfig updates the AI configuration
func (a *App) UpdateAIConfig(config integration.AIConfig) error {
	utils.Infof("Updating AI config: %+v", config)

	// 更新AI服务配置
	if err := logic.UpdateConfig(config); err != nil {
		utils.Errorf("Failed to update AI config: %v", err)
		return err
	}

	// 重新设置数据库管理器
	logic.SetGlobalDatabase(&databaseAdapter{})
	return nil
}

func (a *App) SendMessage(conversationID, message string) error {
	return a.SendMessageWithEvents(conversationID, message)
}

// SendMessageWithEvents sends a message to the AI model using Wails events for streaming
// This method uses Wails events instead of callbacks for real-time communication
func (a *App) SendMessageWithEvents(conversationID, message string) error {
	utils.Infof("SendMessageWithEvents called - conversationID: %s, message: %s", conversationID, message)

	// 创建一个事件回调函数，通过Wails事件系统发送MsgVo对象到前端
	eventCallback := func(msgVo *models.MsgVo) {
		// 通过Wails事件系统发送MsgVo对象到前端
		// runtime.EventsEmit(a.ctx, "ai-message-chunk", msgVo)
		utils.Infof("Sent MsgVo event to frontend: Type=%s, Content=%s", msgVo.Type, msgVo.Content)
	}

	// 使用流式响应处理，支持MCP工具调用
	err := logic.SendMessageStreamWithCompleteResponse(message, conversationID, eventCallback)
	if err != nil {
		utils.Errorf("AI service streaming error: %v", err)
		// 发送错误事件到前端
		// errorMsgVo := &models.MsgVo{
		// 	ConversationID: conversationID,
		// 	Type:           models.MsgTypeText,
		// 	Content:        fmt.Sprintf("AI服务错误：%v", err),
		// }
		// runtime.EventsEmit(a.ctx, "ai-message-error", errorMsgVo)
		return err
	}

	return nil
}

// ConfirmToolCall confirms or rejects a tool call by cardID
func (a *App) ConfirmToolCall(cardID string, confirmed bool) error {
	utils.Infof("Tool call confirmation: cardID=%s, confirmed=%v", cardID, confirmed)

	// 根据cardID找到对应的卡片并执行相应的回调函数
	if confirmed {
		return logic.ConfirmCardByID(cardID)
	} else {
		return logic.RejectCardByID(cardID)
	}
}

// Conversation Management Methods

// CreateConversation creates a new conversation and adds system message
func (a *App) CreateConversation(title string) (*models.Conversation, error) {
	// 创建会话
	conversation, err := sqlite.CreateConversation(title)
	if err != nil {
		utils.Errorf("Failed to create conversation: %v", err)
		return nil, err
	}

	// 添加系统消息到新创建的会话
	systemMessage := &models.Message{
		Role:    "system",
		Content: "You are a helpful database assistant. You can execute Redis commands, MySQL queries, and ClickHouse queries to help users interact with their databases.",
	}

	err = sqlite.AddMessageToConversation(conversation.ID, systemMessage)
	if err != nil {
		utils.Errorf("Failed to add system message to conversation: %v", err)
		return nil, err
	}
	utils.Infof("System message added to conversation: %s", conversation.ID)
	return conversation, nil
}

// GetConversation retrieves a conversation by ID
func (a *App) GetConversation(id string) (*models.Conversation, error) {
	return sqlite.GetConversation(id)
}

// ListConversations retrieves all conversations
func (a *App) ListConversations() ([]*models.Conversation, error) {
	return sqlite.ListConversations()
}

// UpdateConversation updates a conversation
func (a *App) UpdateConversation(conversation *models.Conversation) error {
	return sqlite.UpdateConversation(conversation)
}

// DeleteConversation deletes a conversation
func (a *App) DeleteConversation(id string) error {
	return sqlite.DeleteConversation(id)
}

// GetConversationHistory retrieves all messages for a conversation including card messages
func (a *App) GetConversationHistory(conversationID string) ([]*models.Message, error) {
	return sqlite.GetMessages(conversationID)
}

// ConfirmCard confirms a card and executes the confirm callback
func (a *App) ConfirmCard(cardID string) error {
	return logic.ConfirmCardByID(cardID)
}

// RejectCard rejects a card and executes the reject callback
func (a *App) RejectCard(cardID string) error {
	return logic.RejectCardByID(cardID)
}

// GetQueryHistory retrieves query execution history with pagination
func (a *App) GetQueryHistory(limit, offset int) ([]*models.QueryHistory, error) {
	utils.Infof("📋 Getting query history: limit=%d, offset=%d", limit, offset)

	history, err := sqlite.GetQueryHistory(limit, offset)
	if err != nil {
		utils.Errorf("Failed to get query history: %v", err)
		return nil, err
	}

	utils.Infof("✅ Retrieved query history successfully: count=%d", len(history))

	return history, nil
}

// GetQueryHistoryByDBType retrieves query history filtered by database type
func (a *App) GetQueryHistoryByDBType(dbType string, limit, offset int) ([]*models.QueryHistory, error) {
	utils.Infof("📋 Getting query history by database type: dbType=%s, limit=%d, offset=%d", dbType, limit, offset)

	history, err := sqlite.GetQueryHistoryByDBType(dbType, limit, offset)
	if err != nil {
		utils.Errorf("Failed to get query history by database type: %v", err)
		return nil, err
	}

	utils.Infof("✅ Retrieved query history by database type successfully: dbType=%s, count=%d", dbType, len(history))

	return history, nil
}

// GetQueryHistoryStats returns statistics about query history
func (a *App) GetQueryHistoryStats() (map[string]interface{}, error) {
	utils.Infof("📊 Getting query history statistics")

	stats, err := sqlite.GetQueryHistoryStats()
	if err != nil {
		utils.Errorf("Failed to get query history statistics: %v", err)
		return nil, err
	}

	utils.Infof("✅ Retrieved query history statistics successfully")

	return stats, nil
}

// ClearQueryHistory clears all query history
func (a *App) ClearQueryHistory() error {
	utils.Infof("🗑️ Clearing query history")

	err := sqlite.ClearQueryHistory()
	if err != nil {
		utils.Errorf("Failed to clear query history: %v", err)
		return err
	}

	utils.Infof("✅ Query history cleared successfully")

	return nil
}

// RetryQuery retries a query from history by ID
func (a *App) RetryQuery(historyID int) (*database.QueryResult, error) {
	utils.Infof("🔄 Retrying query with history ID: %d", historyID)

	// 获取查询历史记录
	history, err := sqlite.GetQueryHistoryByID(historyID)
	if err != nil {
		utils.Errorf("Failed to get query history by ID: %v", err)
		return nil, err
	}

	if history == nil {
		return nil, fmt.Errorf("query history not found with ID: %d", historyID)
	}

	// 检查连接是否仍然存在
	_, err = database.GetConnection(history.ConnectionID)
	if err != nil {
		utils.Errorf("Connection not found for retry: %v", err)
		return nil, fmt.Errorf("connection not found: %s", history.ConnectionID)
	}

	// 重新执行查询
	utils.Infof("Retrying query: %s on connection: %s", history.Query, history.ConnectionName)
	result, err := database.ExecuteQuery(history.ConnectionID, history.Query)

	// 更新查询历史记录
	success := err == nil
	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
	}

	var executionTime int64
	var resultCount int
	if result != nil {
		executionTime = int64(result.Time)
		resultCount = result.Count
	}

	utils.Infof("Updating query history after retry - Success: %v", success)

	_, historyErr := sqlite.AddQueryHistory(
		history.Query,
		history.DBType,
		history.ConnectionID,
		history.ConnectionName,
		executionTime,
		success,
		errorMsg,
		resultCount,
	)
	if historyErr != nil {
		utils.Errorf("Failed to update query history after retry: %v", historyErr)
	} else {
		utils.Infof("Query history updated after retry successfully")
	}

	return result, err
}
