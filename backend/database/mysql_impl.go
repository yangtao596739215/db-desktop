package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// MySQL 连接相关方法

func (s *SimpleDatabaseManager) connectInternal(config *ConnectionConfig) error {
	switch config.Type {
	case MySQL:
		return s.connectMySQL(config)
	case Redis:
		return s.connectRedis(config)
	case ClickHouse:
		return s.connectClickHouse(config)
	default:
		return fmt.Errorf("unsupported database type: %s", config.Type)
	}
}

func (s *SimpleDatabaseManager) connectMySQL(config *ConnectionConfig) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Username, config.Password, config.Host, config.Port, config.Database)

	if config.SSLMode != "" {
		dsn += "&tls=" + config.SSLMode
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open MySQL connection: %w", err)
	}

	// 设置连接池
	db.SetMaxOpenConns(config.MaxConns)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	// 测试连接
	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping MySQL database: %w", err)
	}

	s.mysqlConnections[config.ID] = db
	s.logger.Infof("Connected to MySQL database: %s", config.Name)
	return nil
}

func (s *SimpleDatabaseManager) disconnectInternal(id string, dbType DatabaseType) error {
	switch dbType {
	case MySQL:
		if db, exists := s.mysqlConnections[id]; exists {
			err := db.Close()
			delete(s.mysqlConnections, id)
			s.logger.Infof("Disconnected from MySQL database: %s", id)
			return err
		}
	case Redis:
		if client, exists := s.redisConnections[id]; exists {
			err := client.Close()
			delete(s.redisConnections, id)
			s.logger.Infof("Disconnected from Redis database: %s", id)
			return err
		}
	case ClickHouse:
		if db, exists := s.clickhouseConnections[id]; exists {
			err := db.Close()
			delete(s.clickhouseConnections, id)
			s.logger.Infof("Disconnected from ClickHouse database: %s", id)
			return err
		}
	}
	return fmt.Errorf("connection not found: %s", id)
}

func (s *SimpleDatabaseManager) testMySQLConnection(config *ConnectionConfig) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Username, config.Password, config.Host, config.Port, config.Database)

	if config.SSLMode != "" {
		dsn += "&tls=" + config.SSLMode
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open MySQL connection: %w", err)
	}
	defer db.Close()

	// 测试连接
	timeout := time.Duration(config.Timeout) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping MySQL database: %w", err)
	}

	return nil
}

func (s *SimpleDatabaseManager) getConnectionStatusInternal(config *ConnectionConfig) *ConnectionStatus {
	status := &ConnectionStatus{
		ID:       config.ID,
		Status:   "disconnected",
		LastPing: time.Now(),
	}

	switch config.Type {
	case MySQL:
		if db, exists := s.mysqlConnections[config.ID]; exists {
			if err := db.Ping(); err != nil {
				status.Status = "error"
				status.Message = err.Error()
			} else {
				status.Status = "connected"
			}
		}
	case Redis:
		if client, exists := s.redisConnections[config.ID]; exists {
			if _, err := client.Ping(s.ctx).Result(); err != nil {
				status.Status = "error"
				status.Message = err.Error()
			} else {
				status.Status = "connected"
			}
		}
	case ClickHouse:
		if db, exists := s.clickhouseConnections[config.ID]; exists {
			if err := db.Ping(); err != nil {
				status.Status = "error"
				status.Message = err.Error()
			} else {
				status.Status = "connected"
			}
		}
	}

	return status
}

func (s *SimpleDatabaseManager) executeMySQLQuery(connectionID string, query string) (*QueryResult, error) {
	s.logger.Infof("ExecuteQuery called: connectionID=%s, query=%s", connectionID, query)

	db, exists := s.mysqlConnections[connectionID]
	if !exists {
		s.logger.WithField("connectionID", connectionID).Error("Connection not found")
		return nil, fmt.Errorf("connection not found: %s", connectionID)
	}

	start := time.Now()
	rows, err := db.Query(query)
	if err != nil {
		s.logger.Errorf("Query failed: connectionID=%s, error=%s", connectionID, err)
		return &QueryResult{Error: err.Error(), Time: time.Since(start).Milliseconds()}, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		s.logger.Errorf("Get columns failed: connectionID=%s, error=%s", connectionID, err)
		return &QueryResult{Error: err.Error(), Time: time.Since(start).Milliseconds()}, err
	}

	s.logger.WithField("columns", columns).Debug("Query columns")

	var resultRows [][]interface{}
	rowCount := 0
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			s.logger.Errorf("Scan row failed: connectionID=%s, error=%s", connectionID, err)
			return &QueryResult{Error: err.Error(), Time: time.Since(start).Milliseconds()}, err
		}

		// 转换 []byte 为 string
		for i, v := range values {
			if b, ok := v.([]byte); ok {
				values[i] = string(b)
			}
		}

		resultRows = append(resultRows, values)
		rowCount++
	}

	s.logger.Infof("Query completed: connectionID=%s, rows=%s, timeMs=%s", connectionID, rowCount, time.Since(start).Milliseconds())

	return &QueryResult{
		Columns: columns,
		Rows:    resultRows,
		Count:   len(resultRows),
		Time:    time.Since(start).Milliseconds(),
	}, nil
}

func (s *SimpleDatabaseManager) executeMySQLQueryWithLimit(connectionID string, query string, limit int) (*QueryResult, error) {
	limitedQuery := query
	if !strings.Contains(strings.ToLower(query), "limit") {
		limitedQuery = fmt.Sprintf("%s LIMIT %d", query, limit)
	}
	return s.executeMySQLQuery(connectionID, limitedQuery)
}

func (s *SimpleDatabaseManager) getMySQLDatabases(connectionID string) ([]string, error) {
	db, exists := s.mysqlConnections[connectionID]
	if !exists {
		return nil, fmt.Errorf("connection not found: %s", connectionID)
	}

	rows, err := db.Query("SHOW DATABASES")
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

func (s *SimpleDatabaseManager) getMySQLTables(connectionID string, database string) ([]TableInfo, error) {
	s.logger.Infof("getMySQLTables called: connectionID=%s, database=%s", connectionID, database)

	db, exists := s.mysqlConnections[connectionID]
	if !exists {
		s.logger.WithField("connectionID", connectionID).Error("Connection not found")
		return nil, fmt.Errorf("connection not found: %s", connectionID)
	}

	query := "SELECT TABLE_NAME, TABLE_COMMENT FROM information_schema.TABLES WHERE TABLE_SCHEMA = ?"
	s.logger.Debugf("Executing query: query=%s, database=%s", query, database)

	rows, err := db.Query(query, database)
	if err != nil {
		s.logger.WithError(err).Error("Query failed")
		return nil, err
	}
	defer rows.Close()

	var tables []TableInfo
	for rows.Next() {
		var table TableInfo
		var comment sql.NullString
		if err := rows.Scan(&table.Name, &comment); err != nil {
			s.logger.WithError(err).Error("Scan row failed")
			return nil, err
		}
		table.Schema = database
		if comment.Valid {
			table.Comment = comment.String
		}
		tables = append(tables, table)
		s.logger.WithField("tableName", table.Name).Debug("Found table")
	}

	s.logger.WithField("tableCount", len(tables)).Info("Found tables")
	return tables, nil
}

func (s *SimpleDatabaseManager) getMySQLTableInfo(connectionID string, database string, table string) (*TableInfo, error) {
	db, exists := s.mysqlConnections[connectionID]
	if !exists {
		return nil, fmt.Errorf("connection not found: %s", connectionID)
	}

	// 获取列信息
	query := `
		SELECT 
			COLUMN_NAME, 
			DATA_TYPE, 
			IS_NULLABLE, 
			COLUMN_DEFAULT, 
			COLUMN_KEY, 
			EXTRA, 
			COLUMN_COMMENT
		FROM information_schema.COLUMNS 
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION
	`

	rows, err := db.Query(query, database, table)
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
		var nullable, key, extra, comment sql.NullString
		var defaultValue sql.NullString

		if err := rows.Scan(&col.Name, &col.Type, &nullable, &defaultValue, &key, &extra, &comment); err != nil {
			return nil, err
		}

		col.Nullable = nullable.String == "YES"
		if defaultValue.Valid {
			col.DefaultValue = defaultValue.String
		}
		if key.Valid {
			col.Key = key.String
		}
		if extra.Valid {
			col.Extra = extra.String
		}
		if comment.Valid {
			col.Comment = comment.String
		}

		tableInfo.Columns = append(tableInfo.Columns, col)
	}

	// 获取表统计信息
	statsQuery := "SELECT TABLE_ROWS, DATA_LENGTH, INDEX_LENGTH FROM information_schema.TABLES WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?"
	var rowsCount, dataLength, indexLength sql.NullInt64
	err = db.QueryRow(statsQuery, database, table).Scan(&rowsCount, &dataLength, &indexLength)
	if err == nil {
		if rowsCount.Valid {
			tableInfo.Stats["rows"] = fmt.Sprintf("%d", rowsCount.Int64)
		}
		if dataLength.Valid {
			tableInfo.Stats["data_size"] = fmt.Sprintf("%d bytes", dataLength.Int64)
		}
		if indexLength.Valid {
			tableInfo.Stats["index_size"] = fmt.Sprintf("%d bytes", indexLength.Int64)
		}
	}

	return tableInfo, nil
}

func (s *SimpleDatabaseManager) getMySQLTableData(connectionID string, database string, table string, limit int, offset int) (*QueryResult, error) {
	query := fmt.Sprintf("SELECT * FROM `%s`.`%s` LIMIT %d OFFSET %d", database, table, limit, offset)
	return s.executeMySQLQuery(connectionID, query)
}

func (s *SimpleDatabaseManager) getMySQLDatabaseInfo(connectionID string) (*DatabaseInfo, error) {
	db, exists := s.mysqlConnections[connectionID]
	if !exists {
		return nil, fmt.Errorf("connection not found: %s", connectionID)
	}

	info := &DatabaseInfo{}

	// 获取数据库版本
	var version string
	err := db.QueryRow("SELECT VERSION()").Scan(&version)
	if err != nil {
		return nil, err
	}
	info.Version = version

	// 获取当前数据库名
	var dbName string
	err = db.QueryRow("SELECT DATABASE()").Scan(&dbName)
	if err != nil {
		return nil, err
	}
	info.Name = dbName

	// 获取字符集和排序规则
	var charset, collation string
	err = db.QueryRow("SELECT @@character_set_database, @@collation_database").Scan(&charset, &collation)
	if err == nil {
		info.Charset = charset
		info.Collation = collation
	}

	return info, nil
}

func (s *SimpleDatabaseManager) validateMySQLQuery(query string) error {
	query = strings.ToLower(strings.TrimSpace(query))

	dangerousOps := []string{"drop database", "drop table", "truncate", "delete from", "update"}
	for _, op := range dangerousOps {
		if strings.Contains(query, op) {
			return fmt.Errorf("potentially dangerous operation detected: %s", op)
		}
	}

	return nil
}
