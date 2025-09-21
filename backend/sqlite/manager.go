package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"db-desktop/backend/models"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

// 全局SQLite管理器实例
var (
	sqliteManager *Manager
	initOnce      sync.Once
)

// Manager handles SQLite database operations for conversations and messages
type Manager struct {
	db     *sql.DB
	logger *logrus.Logger
}

// init 初始化SQLite管理器
func init() {
	initOnce.Do(func() {
		// Get user home directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			panic(fmt.Sprintf("Failed to get home directory: %v", err))
		}

		// Create .db-desktop directory if it doesn't exist
		dbDir := filepath.Join(homeDir, ".db-desktop")
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			panic(fmt.Sprintf("Failed to create db directory: %v", err))
		}

		// Database file path
		dbPath := filepath.Join(dbDir, "conversations.db")

		logger := logrus.New()
		logger.SetLevel(logrus.InfoLevel)
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
		logger.SetOutput(os.Stdout)

		// Open database
		db, err := sql.Open("sqlite3", dbPath)
		if err != nil {
			panic(fmt.Sprintf("Failed to open database: %v", err))
		}

		sqliteManager = &Manager{
			db:     db,
			logger: logger,
		}

		// Initialize tables
		if err := sqliteManager.initTables(); err != nil {
			panic(fmt.Sprintf("Failed to initialize tables: %v", err))
		}

		logger.Info("SQLite manager initialized successfully")
	})
}

// initTables creates the necessary tables
func (m *Manager) initTables() error {
	// Create conversations table
	conversationsSQL := `
	CREATE TABLE IF NOT EXISTS conversations (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		message_count INTEGER DEFAULT 0
	);`

	// Create messages table
	messagesSQL := `
	CREATE TABLE IF NOT EXISTS messages (
		id TEXT PRIMARY KEY,
		conversation_id TEXT NOT NULL,
		role TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		tool_call_id TEXT,
		tool_calls TEXT,
		FOREIGN KEY (conversation_id) REFERENCES conversations (id) ON DELETE CASCADE
	);`

	// Create query_history table
	queryHistorySQL := `
	CREATE TABLE IF NOT EXISTS query_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		query TEXT NOT NULL,
		execution_time INTEGER NOT NULL,
		db_type TEXT NOT NULL,
		connection_id TEXT NOT NULL,
		connection_name TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		success BOOLEAN NOT NULL DEFAULT 1,
		error TEXT,
		result_rows INTEGER DEFAULT 0
	);`

	// Create indexes
	indexesSQL := []string{
		"CREATE INDEX IF NOT EXISTS idx_messages_conversation_id ON messages (conversation_id);",
		"CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages (created_at);",
		"CREATE INDEX IF NOT EXISTS idx_conversations_updated_at ON conversations (updated_at);",
		"CREATE INDEX IF NOT EXISTS idx_query_history_created_at ON query_history (created_at);",
		"CREATE INDEX IF NOT EXISTS idx_query_history_db_type ON query_history (db_type);",
		"CREATE INDEX IF NOT EXISTS idx_query_history_success ON query_history (success);",
		"CREATE INDEX IF NOT EXISTS idx_query_history_connection_id ON query_history (connection_id);",
		"CREATE INDEX IF NOT EXISTS idx_query_history_connection_name ON query_history (connection_name);",
	}

	// Execute table creation
	if _, err := m.db.Exec(conversationsSQL); err != nil {
		return fmt.Errorf("failed to create conversations table: %w", err)
	}

	if _, err := m.db.Exec(messagesSQL); err != nil {
		return fmt.Errorf("failed to create messages table: %w", err)
	}

	if _, err := m.db.Exec(queryHistorySQL); err != nil {
		return fmt.Errorf("failed to create query_history table: %w", err)
	}

	// Add tool_calls column to messages table if it doesn't exist
	if err := m.migrateMessagesTable(); err != nil {
		return fmt.Errorf("failed to migrate messages table: %w", err)
	}

	// Execute index creation
	for _, indexSQL := range indexesSQL {
		if _, err := m.db.Exec(indexSQL); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	m.logger.Info("Database tables initialized successfully")
	return nil
}

// migrateMessagesTable adds tool_calls column to messages table if it doesn't exist
func (m *Manager) migrateMessagesTable() error {
	// Check if tool_calls column exists
	var count int
	err := m.db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('messages') WHERE name='tool_calls'").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check column existence: %w", err)
	}

	// If column doesn't exist, add it
	if count == 0 {
		m.logger.Info("Adding tool_calls column to messages table")
		_, err := m.db.Exec("ALTER TABLE messages ADD COLUMN tool_calls TEXT")
		if err != nil {
			return fmt.Errorf("failed to add tool_calls column: %w", err)
		}
		m.logger.Info("Successfully added tool_calls column to messages table")
	}

	return nil
}

// Close closes the database connection
func (m *Manager) Close() error {
	return m.db.Close()
}

// CreateConversation creates a new conversation
func (m *Manager) CreateConversation(title string) (*models.Conversation, error) {
	id := uuid.New().String()
	now := time.Now()

	conversation := &models.Conversation{
		ID:           id,
		Title:        title,
		CreatedAt:    now,
		UpdatedAt:    now,
		MessageCount: 0,
	}

	m.logger.Infof("💾 Creating new conversation in database: conversationID=%s, title=%s, createdAt=%s", id, title, now)

	// Convert to DO for database operations
	conversationDo := ToConversationDo(conversation)

	query := `
		INSERT INTO conversations (id, title, created_at, updated_at, message_count)
		VALUES (?, ?, ?, ?, ?)`

	_, err := m.db.Exec(query, conversationDo.ID, conversationDo.Title,
		conversationDo.CreatedAt, conversationDo.UpdatedAt, conversationDo.MessageCount)
	if err != nil {
		m.logger.Errorf("Failed to create conversation in database: conversationID=%s, title=%s, error=%s", id, title, err)
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	m.logger.Infof("✅ New conversation created successfully: conversationID=%s, title=%s", id, title)
	return conversation, nil
}

// GetConversation retrieves a conversation by ID
func (m *Manager) GetConversation(id string) (*models.Conversation, error) {
	query := `
		SELECT id, title, created_at, updated_at, message_count
		FROM conversations
		WHERE id = ?`

	var conversationDo ConversationDo
	err := m.db.QueryRow(query, id).Scan(
		&conversationDo.ID,
		&conversationDo.Title,
		&conversationDo.CreatedAt,
		&conversationDo.UpdatedAt,
		&conversationDo.MessageCount,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("conversation not found")
		}
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	return conversationDo.ToConversation(), nil
}

// ListConversations retrieves all conversations ordered by updated_at DESC
func (m *Manager) ListConversations() ([]*models.Conversation, error) {
	query := `
		SELECT id, title, created_at, updated_at, message_count
		FROM conversations
		ORDER BY updated_at DESC`

	rows, err := m.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list conversations: %w", err)
	}
	defer rows.Close()

	var conversations []*models.Conversation
	for rows.Next() {
		var conversationDo ConversationDo
		err := rows.Scan(
			&conversationDo.ID,
			&conversationDo.Title,
			&conversationDo.CreatedAt,
			&conversationDo.UpdatedAt,
			&conversationDo.MessageCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan conversation: %w", err)
		}
		conversations = append(conversations, conversationDo.ToConversation())
	}

	return conversations, nil
}

// UpdateConversation updates a conversation
func (m *Manager) UpdateConversation(conversation *models.Conversation) error {
	// Convert to DO for database operations
	conversationDo := ToConversationDo(conversation)

	query := `
		UPDATE conversations
		SET title = ?, updated_at = ?, message_count = ?
		WHERE id = ?`

	_, err := m.db.Exec(query, conversationDo.Title, conversationDo.UpdatedAt,
		conversationDo.MessageCount, conversationDo.ID)
	if err != nil {
		return fmt.Errorf("failed to update conversation: %w", err)
	}

	return nil
}

// DeleteConversation deletes a conversation and all its messages
func (m *Manager) DeleteConversation(id string) error {
	query := `DELETE FROM conversations WHERE id = ?`
	_, err := m.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete conversation: %w", err)
	}

	m.logger.WithField("conversationId", id).Info("Deleted conversation")
	return nil
}

// AddMessageWithToolCallID adds a message with tool call ID
func (m *Manager) AddMessageToConversation(conversationID string, message *models.Message) error {
	id := uuid.New().String()
	now := time.Now()
	// Convert to DO for database operations
	messageDo := ToMessageDo(message, conversationID)
	messageDo.ID = id
	messageDo.CreatedAt = now

	m.logger.Infof("💾 Adding message to database: messageID=%s, conversationID=%s, role=%s, contentLength=%d, toolCallID=%s", id, conversationID, messageDo.Role, len(messageDo.Content), messageDo.ToolCallID)

	query := `
		INSERT INTO messages (id, conversation_id, role, content, created_at, tool_call_id, tool_calls)
		VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := m.db.Exec(query, messageDo.ID, messageDo.ConversationID,
		messageDo.Role, messageDo.Content, messageDo.CreatedAt, messageDo.ToolCallID, messageDo.ToolCalls)
	if err != nil {
		m.logger.Errorf("Failed to add message to database: messageID=%s, conversationID=%s, error=%s", id, conversationID, err)
		return fmt.Errorf("failed to add message: %w", err)
	}

	m.logger.Infof("✅ Message added to database successfully: messageID=%s, conversationID=%s", id, conversationID)

	// Update conversation's message count and updated_at
	updateQuery := `
		UPDATE conversations
		SET message_count = message_count + 1, updated_at = ?
		WHERE id = ?`

	m.logger.Infof("🔄 Updating conversation message count: conversationID=%s, updatedAt=%s", conversationID, now)

	_, err = m.db.Exec(updateQuery, now, conversationID)
	if err != nil {
		m.logger.Errorf("Failed to update conversation message count: conversationID=%s, error=%s", conversationID, err)
	} else {
		m.logger.Infof("✅ Conversation message count updated successfully: conversationID=%s", conversationID)
	}

	return nil
}

// GetMessages retrieves all messages for a conversation
func (m *Manager) GetMessagesForLLM(conversationID string) ([]*models.Message, error) {
	query := `
		SELECT id, conversation_id, role, content, created_at, tool_call_id, tool_calls
		FROM messages
		WHERE conversation_id = ? and role != 'card'
		ORDER BY created_at ASC`

	rows, err := m.db.Query(query, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		var messageDo MessageDo
		err := rows.Scan(
			&messageDo.ID,
			&messageDo.ConversationID,
			&messageDo.Role,
			&messageDo.Content,
			&messageDo.CreatedAt,
			&messageDo.ToolCallID,
			&messageDo.ToolCalls,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, messageDo.ToMessage())
	}

	return messages, nil
}

// GetMessages retrieves all messages for a conversation
func (m *Manager) GetMessages(conversationID string) ([]*models.Message, error) {
	query := `
		SELECT id, conversation_id, role, content, created_at, tool_call_id, tool_calls
		FROM messages
		WHERE conversation_id = ?
		ORDER BY created_at ASC`

	rows, err := m.db.Query(query, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		var messageDo MessageDo
		err := rows.Scan(
			&messageDo.ID,
			&messageDo.ConversationID,
			&messageDo.Role,
			&messageDo.Content,
			&messageDo.CreatedAt,
			&messageDo.ToolCallID,
			&messageDo.ToolCalls,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, messageDo.ToMessage())
	}

	return messages, nil
}

// GetConversationWithMessages retrieves a conversation with all its messages
func (m *Manager) GetConversationWithMessages(conversationID string) (*models.ConversationWithMessages, error) {
	// Get conversation
	conversation, err := m.GetConversation(conversationID)
	if err != nil {
		return nil, err
	}

	// Get messages
	messages, err := m.GetMessages(conversationID)
	if err != nil {
		return nil, err
	}

	return &models.ConversationWithMessages{
		Conversation: *conversation,
		Messages:     messages,
	}, nil
}

// AddQueryHistory adds a query execution record to the history
func (m *Manager) AddQueryHistory(query, dbType, connectionID, connectionName string, executionTime int64, success bool, errorMsg string, resultRows int) (*models.QueryHistory, error) {
	now := time.Now()

	queryHistory := &models.QueryHistory{
		Query:          query,
		ExecutionTime:  executionTime,
		DBType:         dbType,
		ConnectionID:   connectionID,
		ConnectionName: connectionName,
		CreatedAt:      now,
		Success:        success,
		Error:          errorMsg,
		ResultRows:     resultRows,
	}

	m.logger.Infof("💾 Adding query to execution history: query=%s, dbType=%s, executionTime=%d, success=%t, resultRows=%d", query, dbType, executionTime, success, resultRows)

	// Convert to DO for database operations
	queryHistoryDo := ToQueryHistoryDo(queryHistory)

	querySQL := `
		INSERT INTO query_history (query, execution_time, db_type, connection_id, connection_name, created_at, success, error, result_rows)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := m.db.Exec(querySQL, queryHistoryDo.Query, queryHistoryDo.ExecutionTime,
		queryHistoryDo.DBType, queryHistoryDo.ConnectionID, queryHistoryDo.ConnectionName,
		queryHistoryDo.CreatedAt, queryHistoryDo.Success, queryHistoryDo.Error, queryHistoryDo.ResultRows)
	if err != nil {
		m.logger.Errorf("Failed to add query to history: query=%s, dbType=%s, error=%s", query, dbType, err)
		return nil, fmt.Errorf("failed to add query history: %w", err)
	}

	// Get the inserted ID
	id, err := result.LastInsertId()
	if err != nil {
		m.logger.Errorf("Failed to get last insert ID: query=%s, dbType=%s, error=%s", query, dbType, err)
		return nil, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	queryHistory.ID = int(id)

	m.logger.Infof("✅ Query added to execution history successfully: queryHistoryID=%d, query=%s, dbType=%s", queryHistory.ID, query, dbType)

	return queryHistory, nil
}

// GetQueryHistory retrieves query execution history with pagination
func (m *Manager) GetQueryHistory(limit, offset int) ([]*models.QueryHistory, error) {
	query := `
		SELECT id, query, execution_time, db_type, connection_id, connection_name, created_at, success, error, result_rows
		FROM query_history
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`

	rows, err := m.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get query history: %w", err)
	}
	defer rows.Close()

	var history []*models.QueryHistory
	for rows.Next() {
		var qhDo QueryHistoryDo
		err := rows.Scan(
			&qhDo.ID,
			&qhDo.Query,
			&qhDo.ExecutionTime,
			&qhDo.DBType,
			&qhDo.ConnectionID,
			&qhDo.ConnectionName,
			&qhDo.CreatedAt,
			&qhDo.Success,
			&qhDo.Error,
			&qhDo.ResultRows,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan query history: %w", err)
		}
		history = append(history, qhDo.ToQueryHistory())
	}

	m.logger.Infof("📋 Retrieved query history: limit=%d, offset=%d, count=%d", limit, offset, len(history))

	return history, nil
}

// GetQueryHistoryByDBType retrieves query history filtered by database type
func (m *Manager) GetQueryHistoryByDBType(dbType string, limit, offset int) ([]*models.QueryHistory, error) {
	query := `
		SELECT id, query, execution_time, db_type, connection_id, connection_name, created_at, success, error, result_rows
		FROM query_history
		WHERE db_type = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`

	rows, err := m.db.Query(query, dbType, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get query history by db type: %w", err)
	}
	defer rows.Close()

	var history []*models.QueryHistory
	for rows.Next() {
		var qhDo QueryHistoryDo
		err := rows.Scan(
			&qhDo.ID,
			&qhDo.Query,
			&qhDo.ExecutionTime,
			&qhDo.DBType,
			&qhDo.ConnectionID,
			&qhDo.ConnectionName,
			&qhDo.CreatedAt,
			&qhDo.Success,
			&qhDo.Error,
			&qhDo.ResultRows,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan query history: %w", err)
		}
		history = append(history, qhDo.ToQueryHistory())
	}

	m.logger.Infof("📋 Retrieved query history by database type: dbType=%s, limit=%d, offset=%d, count=%d", dbType, limit, offset, len(history))

	return history, nil
}

// GetQueryHistoryByID retrieves a single query history record by ID
func (m *Manager) GetQueryHistoryByID(id int) (*models.QueryHistory, error) {
	query := `
		SELECT id, query, execution_time, db_type, connection_id, connection_name, created_at, success, error, result_rows
		FROM query_history
		WHERE id = ?`

	var qhDo QueryHistoryDo
	err := m.db.QueryRow(query, id).Scan(
		&qhDo.ID,
		&qhDo.Query,
		&qhDo.ExecutionTime,
		&qhDo.DBType,
		&qhDo.ConnectionID,
		&qhDo.ConnectionName,
		&qhDo.CreatedAt,
		&qhDo.Success,
		&qhDo.Error,
		&qhDo.ResultRows,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 记录不存在
		}
		return nil, fmt.Errorf("failed to get query history by ID: %w", err)
	}

	return qhDo.ToQueryHistory(), nil
}

// GetQueryHistoryStats returns statistics about query history
func (m *Manager) GetQueryHistoryStats() (map[string]interface{}, error) {
	// Total queries
	var totalQueries int
	err := m.db.QueryRow("SELECT COUNT(*) FROM query_history").Scan(&totalQueries)
	if err != nil {
		return nil, fmt.Errorf("failed to get total queries: %w", err)
	}

	// Queries by database type
	query := `
		SELECT db_type, COUNT(*) as count, AVG(execution_time) as avg_time, 
		       SUM(CASE WHEN success = 1 THEN 1 ELSE 0 END) as success_count
		FROM query_history
		GROUP BY db_type`

	rows, err := m.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get query stats by db type: %w", err)
	}
	defer rows.Close()

	statsByType := make(map[string]map[string]interface{})
	for rows.Next() {
		var dbType string
		var count, successCount int
		var avgTime float64
		err := rows.Scan(&dbType, &count, &avgTime, &successCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan query stats: %w", err)
		}
		statsByType[dbType] = map[string]interface{}{
			"count":        count,
			"avgTime":      avgTime,
			"successCount": successCount,
			"errorCount":   count - successCount,
		}
	}

	stats := map[string]interface{}{
		"totalQueries": totalQueries,
		"statsByType":  statsByType,
	}

	m.logger.Infof("📊 Retrieved query history statistics: totalQueries=%d, typesCount=%d", totalQueries, len(statsByType))

	return stats, nil
}

// ClearQueryHistory clears all query history
func (m *Manager) ClearQueryHistory() error {
	query := "DELETE FROM query_history"
	_, err := m.db.Exec(query)
	if err != nil {
		m.logger.WithError(err).Error("Failed to clear query history")
		return fmt.Errorf("failed to clear query history: %w", err)
	}

	m.logger.Info("🗑️ Cleared all query history")
	return nil
}

// 包级函数，替代原来的对象方法调用

// CreateConversation creates a new conversation
func CreateConversation(title string) (*models.Conversation, error) {
	return sqliteManager.CreateConversation(title)
}

// GetConversation retrieves a conversation by ID
func GetConversation(id string) (*models.Conversation, error) {
	return sqliteManager.GetConversation(id)
}

// ListConversations retrieves all conversations ordered by updated_at DESC
func ListConversations() ([]*models.Conversation, error) {
	return sqliteManager.ListConversations()
}

// UpdateConversation updates a conversation
func UpdateConversation(conversation *models.Conversation) error {
	return sqliteManager.UpdateConversation(conversation)
}

// DeleteConversation deletes a conversation and all its messages
func DeleteConversation(id string) error {
	return sqliteManager.DeleteConversation(id)
}

// AddMessageToConversation adds a message to a conversation
func AddMessageToConversation(conversationID string, message *models.Message) error {
	return sqliteManager.AddMessageToConversation(conversationID, message)
}

// GetMessagesForLLM retrieves all messages for a conversation (excluding card messages)
func GetMessagesForLLM(conversationID string) ([]*models.Message, error) {
	return sqliteManager.GetMessagesForLLM(conversationID)
}

// GetMessages retrieves all messages for a conversation
func GetMessages(conversationID string) ([]*models.Message, error) {
	return sqliteManager.GetMessages(conversationID)
}

// GetConversationWithMessages retrieves a conversation with all its messages
func GetConversationWithMessages(conversationID string) (*models.ConversationWithMessages, error) {
	return sqliteManager.GetConversationWithMessages(conversationID)
}

// AddQueryHistory adds a query execution record to the history
func AddQueryHistory(query, dbType, connectionID, connectionName string, executionTime int64, success bool, errorMsg string, resultRows int) (*models.QueryHistory, error) {
	return sqliteManager.AddQueryHistory(query, dbType, connectionID, connectionName, executionTime, success, errorMsg, resultRows)
}

// GetQueryHistory retrieves query execution history with pagination
func GetQueryHistory(limit, offset int) ([]*models.QueryHistory, error) {
	return sqliteManager.GetQueryHistory(limit, offset)
}

// GetQueryHistoryByDBType retrieves query history filtered by database type
func GetQueryHistoryByDBType(dbType string, limit, offset int) ([]*models.QueryHistory, error) {
	return sqliteManager.GetQueryHistoryByDBType(dbType, limit, offset)
}

// GetQueryHistoryByID retrieves a single query history record by ID
func GetQueryHistoryByID(id int) (*models.QueryHistory, error) {
	return sqliteManager.GetQueryHistoryByID(id)
}

// GetQueryHistoryStats returns statistics about query history
func GetQueryHistoryStats() (map[string]interface{}, error) {
	return sqliteManager.GetQueryHistoryStats()
}

// ClearQueryHistory clears all query history
func ClearQueryHistory() error {
	return sqliteManager.ClearQueryHistory()
}

// Close closes the database connection
func Close() error {
	return sqliteManager.Close()
}
