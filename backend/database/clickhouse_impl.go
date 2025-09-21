package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
)

// ClickHouse 连接相关方法

func (s *SimpleDatabaseManager) connectClickHouse(config *ConnectionConfig) error {
	// 构建 ClickHouse 连接选项
	options := &clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", config.Host, config.Port)},
		Auth: clickhouse.Auth{
			Database: config.Database,
			Username: config.Username,
			Password: config.Password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout: func() time.Duration {
			if config.Timeout > 0 {
				return time.Duration(config.Timeout) * time.Second
			}
			return 30 * time.Second // 默认30秒超时
		}(),
		// 先尝试原生协议，如果失败再尝试HTTP协议
		// Protocol: clickhouse.HTTP,
	}

	// 创建连接
	conn := clickhouse.OpenDB(options)
	if conn == nil {
		s.logger.Errorf("Failed to create ClickHouse connection - OpenDB returned nil: host=%s, port=%s, database=%s", config.Host, config.Port, config.Database)
		return fmt.Errorf("failed to create ClickHouse connection")
	}

	// 设置连接池
	conn.SetMaxOpenConns(config.MaxConns)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(time.Hour)

	// 测试连接
	if err := conn.Ping(); err != nil {
		s.logger.Errorf("Failed to ping ClickHouse database: host=%s, port=%s, database=%s, error=%s", config.Host, config.Port, config.Database, err)
		conn.Close()
		return fmt.Errorf("failed to ping ClickHouse database: %w", err)
	}

	s.clickhouseConnections[config.ID] = conn
	s.logger.Infof("Connected to ClickHouse database: %s", config.Name)
	return nil
}

func (s *SimpleDatabaseManager) testClickHouseConnection(config *ConnectionConfig) error {
	s.logger.Infof("Testing ClickHouse connection to %s:%d", config.Host, config.Port)

	// 构建 ClickHouse 连接选项
	options := &clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", config.Host, config.Port)},
		Auth: clickhouse.Auth{
			Database: config.Database,
			Username: config.Username,
			Password: config.Password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout: func() time.Duration {
			if config.Timeout > 0 {
				return time.Duration(config.Timeout) * time.Second
			}
			return 30 * time.Second // 默认30秒超时
		}(),
		// 先尝试原生协议，如果失败再尝试HTTP协议
		// Protocol: clickhouse.HTTP,
	}

	s.logger.Infof("ClickHouse connection options: %+v", options)

	// 创建连接
	conn := clickhouse.OpenDB(options)
	if conn == nil {
		s.logger.Errorf("Failed to create ClickHouse connection - OpenDB returned nil")
		return fmt.Errorf("failed to create ClickHouse connection")
	}
	defer conn.Close()

	s.logger.Infof("ClickHouse connection created successfully")

	// 添加连接测试
	s.logger.Infof("Testing ClickHouse connection with simple query...")

	// 测试连接
	timeout := time.Duration(config.Timeout) * time.Second
	// 如果超时时间为0，设置默认30秒超时
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	s.logger.Infof("Testing ClickHouse ping with timeout: %v", timeout)
	if err := conn.PingContext(ctx); err != nil {
		s.logger.Errorf("ClickHouse ping failed: %v", err)
		s.logger.Errorf("Connection details - Host: %s, Port: %d, Database: %s, Username: %s",
			config.Host, config.Port, config.Database, config.Username)
		return fmt.Errorf("failed to ping ClickHouse database: %w", err)
	}

	s.logger.Infof("ClickHouse connection test successful")
	return nil
}

func (s *SimpleDatabaseManager) executeClickHouseQuery(connectionID string, query string) (*QueryResult, error) {
	conn, exists := s.clickhouseConnections[connectionID]
	if !exists {
		return nil, fmt.Errorf("connection not found: %s", connectionID)
	}

	start := time.Now()
	rows, err := conn.Query(query)
	if err != nil {
		return &QueryResult{Error: err.Error(), Time: time.Since(start).Milliseconds()}, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return &QueryResult{Error: err.Error(), Time: time.Since(start).Milliseconds()}, err
	}

	var resultRows [][]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return &QueryResult{Error: err.Error(), Time: time.Since(start).Milliseconds()}, err
		}

		// 转换 []byte 为 string
		for i, v := range values {
			if b, ok := v.([]byte); ok {
				values[i] = string(b)
			}
		}

		resultRows = append(resultRows, values)
	}

	return &QueryResult{
		Columns: columns,
		Rows:    resultRows,
		Count:   len(resultRows),
		Time:    time.Since(start).Milliseconds(),
	}, nil
}

func (s *SimpleDatabaseManager) executeClickHouseQueryWithLimit(connectionID string, query string, limit int) (*QueryResult, error) {
	limitedQuery := query
	if !strings.Contains(strings.ToLower(query), "limit") {
		limitedQuery = fmt.Sprintf("%s LIMIT %d", query, limit)
	}
	return s.executeClickHouseQuery(connectionID, limitedQuery)
}

func (s *SimpleDatabaseManager) getClickHouseDatabases(connectionID string) ([]string, error) {
	conn, exists := s.clickhouseConnections[connectionID]
	if !exists {
		return nil, fmt.Errorf("connection not found: %s", connectionID)
	}

	rows, err := conn.Query("SHOW DATABASES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			return nil, err
		}
		databases = append(databases, dbName)
	}

	return databases, nil
}

func (s *SimpleDatabaseManager) getClickHouseTables(connectionID string, database string) ([]TableInfo, error) {
	conn, exists := s.clickhouseConnections[connectionID]
	if !exists {
		return nil, fmt.Errorf("connection not found: %s", connectionID)
	}

	query := "SELECT name, comment FROM system.tables WHERE database = ?"
	rows, err := conn.Query(query, database)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []TableInfo
	for rows.Next() {
		var table TableInfo
		var comment sql.NullString
		if err := rows.Scan(&table.Name, &comment); err != nil {
			return nil, err
		}
		table.Schema = database
		if comment.Valid {
			table.Comment = comment.String
		}
		tables = append(tables, table)
	}

	return tables, nil
}

func (s *SimpleDatabaseManager) getClickHouseTableInfo(connectionID string, database string, table string) (*TableInfo, error) {
	conn, exists := s.clickhouseConnections[connectionID]
	if !exists {
		return nil, fmt.Errorf("connection not found: %s", connectionID)
	}

	// 获取列信息
	query := `
		SELECT 
			name, 
			type, 
			position, 
			default_kind, 
			default_expression,
			comment
		FROM system.columns 
		WHERE database = ? AND table = ?
		ORDER BY position
	`

	rows, err := conn.Query(query, database, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tableInfo := &TableInfo{
		Name:    table,
		Schema:  database,
		Columns: []ColumnInfo{},
		Stats:   make(map[string]string),
	}

	for rows.Next() {
		var col ColumnInfo
		var position int
		var defaultKind, defaultExpr, comment sql.NullString

		if err := rows.Scan(&col.Name, &col.Type, &position, &defaultKind, &defaultExpr, &comment); err != nil {
			return nil, err
		}

		col.Nullable = !strings.Contains(strings.ToLower(col.Type), "not null")
		if defaultKind.Valid && defaultExpr.Valid {
			col.DefaultValue = defaultExpr.String
		}
		if comment.Valid {
			col.Comment = comment.String
		}

		tableInfo.Columns = append(tableInfo.Columns, col)
	}

	// 获取表统计信息
	statsQuery := `
		SELECT 
			rows, 
			bytes, 
			compressed_bytes,
			uncompressed_bytes
		FROM system.tables 
		WHERE database = ? AND name = ?
	`
	var rowsCount, bytes, compressedBytes, uncompressedBytes sql.NullInt64
	err = conn.QueryRow(statsQuery, database, table).Scan(&rowsCount, &bytes, &compressedBytes, &uncompressedBytes)
	if err == nil {
		if rowsCount.Valid {
			tableInfo.Stats["rows"] = fmt.Sprintf("%d", rowsCount.Int64)
		}
		if bytes.Valid {
			tableInfo.Stats["size"] = fmt.Sprintf("%d bytes", bytes.Int64)
		}
		if compressedBytes.Valid {
			tableInfo.Stats["compressed_size"] = fmt.Sprintf("%d bytes", compressedBytes.Int64)
		}
		if uncompressedBytes.Valid {
			tableInfo.Stats["uncompressed_size"] = fmt.Sprintf("%d bytes", uncompressedBytes.Int64)
		}
	}

	return tableInfo, nil
}

func (s *SimpleDatabaseManager) getClickHouseTableData(connectionID string, database string, table string, limit int, offset int) (*QueryResult, error) {
	query := fmt.Sprintf("SELECT * FROM `%s`.`%s` LIMIT %d OFFSET %d", database, table, limit, offset)
	return s.executeClickHouseQuery(connectionID, query)
}

func (s *SimpleDatabaseManager) getClickHouseDatabaseInfo(connectionID string) (*DatabaseInfo, error) {
	conn, exists := s.clickhouseConnections[connectionID]
	if !exists {
		return nil, fmt.Errorf("connection not found: %s", connectionID)
	}

	info := &DatabaseInfo{}

	// 获取 ClickHouse 版本
	var version string
	err := conn.QueryRow("SELECT version()").Scan(&version)
	if err != nil {
		return nil, err
	}
	info.Version = version

	// 获取当前数据库名
	var dbName string
	err = conn.QueryRow("SELECT currentDatabase()").Scan(&dbName)
	if err != nil {
		return nil, err
	}
	info.Name = dbName

	// 获取数据库引擎
	var engine string
	err = conn.QueryRow("SELECT engine FROM system.databases WHERE name = ?", dbName).Scan(&engine)
	if err == nil {
		info.Charset = engine
	}

	return info, nil
}

func (s *SimpleDatabaseManager) validateClickHouseQuery(query string) error {
	query = strings.ToLower(strings.TrimSpace(query))

	dangerousOps := []string{"drop database", "drop table", "truncate", "delete from", "alter table"}
	for _, op := range dangerousOps {
		if strings.Contains(query, op) {
			return fmt.Errorf("potentially dangerous operation detected: %s", op)
		}
	}

	return nil
}
