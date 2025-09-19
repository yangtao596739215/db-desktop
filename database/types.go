package database

import (
	"time"
)

// DatabaseType represents the type of database
type DatabaseType string

const (
	MySQL      DatabaseType = "mysql"
	Redis      DatabaseType = "redis"
	ClickHouse DatabaseType = "clickhouse"
)

// ConnectionConfig represents database connection configuration
type ConnectionConfig struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Type      DatabaseType `json:"type"`
	Host      string       `json:"host"`
	Port      int          `json:"port"`
	Username  string       `json:"username"`
	Password  string       `json:"password"`
	Database  string       `json:"database"`
	SSLMode   string       `json:"ssl_mode"`
	Timeout   int          `json:"timeout"` // Store as seconds for JSON compatibility
	MaxConns  int          `json:"max_conns"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	// 运行时状态字段
	Status   string    `json:"status,omitempty"`
	LastPing time.Time `json:"last_ping,omitempty"`
	Message  string    `json:"message,omitempty"`
}

// ConnectionStatus represents the status of a database connection
type ConnectionStatus struct {
	ID       string    `json:"id"`
	Status   string    `json:"status"` // "connected", "disconnected", "error"
	Message  string    `json:"message"`
	LastPing time.Time `json:"last_ping"`
}

// QueryResult represents the result of a database query
type QueryResult struct {
	Columns []string        `json:"columns"`
	Rows    [][]interface{} `json:"rows"`
	Count   int             `json:"count"`
	Error   string          `json:"error,omitempty"`
	Time    int64           `json:"time"` // 存储为毫秒数
}

// TableInfo represents information about a database table
type TableInfo struct {
	Name    string            `json:"name"`
	Schema  string            `json:"schema"`
	Comment string            `json:"comment"`
	Columns []ColumnInfo      `json:"columns"`
	Indexes []IndexInfo       `json:"indexes"`
	Stats   map[string]string `json:"stats"`
}

// ColumnInfo represents information about a table column
type ColumnInfo struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Nullable     bool   `json:"nullable"`
	DefaultValue string `json:"default_value"`
	Key          string `json:"key"`
	Extra        string `json:"extra"`
	Comment      string `json:"comment"`
}

// IndexInfo represents information about a table index
type IndexInfo struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
	Unique  bool     `json:"unique"`
	Type    string   `json:"type"`
}

// DatabaseInfo represents general database information
type DatabaseInfo struct {
	Name      string      `json:"name"`
	Version   string      `json:"version"`
	Charset   string      `json:"charset"`
	Collation string      `json:"collation"`
	Tables    []TableInfo `json:"tables"`
}

// DatabaseManager interface defines methods for database operations
type DatabaseManager interface {
	// Connection management
	Connect(config *ConnectionConfig) error
	Disconnect(id string) error
	TestConnection(config *ConnectionConfig) error
	GetConnectionStatus(id string) *ConnectionStatus
	ListConnections() []*ConnectionConfig

	// Query operations
	ExecuteQuery(connectionID string, query string) (*QueryResult, error)
	ExecuteQueryWithLimit(connectionID string, query string, limit int) (*QueryResult, error)

	// Schema operations
	GetDatabases(connectionID string) ([]string, error)
	GetTables(connectionID string, database string) ([]TableInfo, error)
	GetTableInfo(connectionID string, database string, table string) (*TableInfo, error)
	GetTableData(connectionID string, database string, table string, limit int, offset int) (*QueryResult, error)

	// Utility operations
	GetDatabaseInfo(connectionID string) (*DatabaseInfo, error)
	FormatQuery(query string) string
	ValidateQuery(query string) error
}
