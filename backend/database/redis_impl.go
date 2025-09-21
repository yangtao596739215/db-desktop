package database

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// Redis 连接相关方法

func (s *SimpleDatabaseManager) connectRedis(config *ConnectionConfig) error {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       0, // Redis 不使用数据库名，我们使用 DB 编号
	})

	// 测试连接
	timeout := time.Duration(config.Timeout) * time.Second
	// 如果超时时间为0，设置默认30秒超时
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		s.logger.Errorf("Failed to ping Redis database: host=%s, port=%s, database=%s, error=%s", config.Host, config.Port, config.Database, err)
		return fmt.Errorf("failed to ping Redis database: %w", err)
	}

	s.redisConnections[config.ID] = rdb
	s.logger.Infof("Connected to Redis database: %s", config.Name)
	return nil
}

func (s *SimpleDatabaseManager) testRedisConnection(config *ConnectionConfig) error {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       0,
	})
	defer rdb.Close()

	// 测试连接
	timeout := time.Duration(config.Timeout) * time.Second
	// 如果超时时间为0，设置默认30秒超时
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		s.logger.Errorf("Failed to ping Redis database in test: host=%s, port=%s, database=%s, error=%s", config.Host, config.Port, config.Database, err)
		return fmt.Errorf("failed to ping Redis database: %w", err)
	}

	return nil
}

func (s *SimpleDatabaseManager) executeRedisQuery(connectionID string, query string) (*QueryResult, error) {
	client, exists := s.redisConnections[connectionID]
	if !exists {
		s.logger.WithField("connectionID", connectionID).Error("Redis connection not found")
		return nil, fmt.Errorf("connection not found: %s", connectionID)
	}

	start := time.Now()

	// 解析 Redis 命令
	parts := strings.Fields(query)
	if len(parts) == 0 {
		s.logger.WithField("connectionID", connectionID).Error("Empty Redis command")
		return &QueryResult{Error: "empty command", Time: time.Since(start).Milliseconds()}, fmt.Errorf("empty command")
	}

	cmd := parts[0]
	args := make([]interface{}, len(parts)-1)
	for i, part := range parts[1:] {
		args[i] = part
	}

	// 执行 Redis 命令
	cmdArgs := make([]interface{}, 0, len(args)+1)
	cmdArgs = append(cmdArgs, cmd)
	cmdArgs = append(cmdArgs, args...)

	result, err := client.Do(s.ctx, cmdArgs...).Result()
	if err != nil {
		s.logger.Errorf("Redis command execution failed: connectionID=%s, command=%s, error=%s", connectionID, cmd, err)
		return &QueryResult{Error: err.Error(), Time: time.Since(start).Milliseconds()}, err
	}

	// 转换结果为行格式
	columns := []string{"Result"}
	rows := [][]interface{}{}

	switch v := result.(type) {
	case string:
		rows = append(rows, []interface{}{v})
	case int64:
		rows = append(rows, []interface{}{v})
	case []interface{}:
		for _, item := range v {
			rows = append(rows, []interface{}{item})
		}
	case map[string]interface{}:
		columns = []string{"Key", "Value"}
		for k, v := range v {
			rows = append(rows, []interface{}{k, v})
		}
	default:
		rows = append(rows, []interface{}{fmt.Sprintf("%v", v)})
	}

	return &QueryResult{
		Columns: columns,
		Rows:    rows,
		Count:   len(rows),
		Time:    time.Since(start).Milliseconds(),
	}, nil
}

func (s *SimpleDatabaseManager) executeRedisQueryWithLimit(connectionID string, query string, limit int) (*QueryResult, error) {
	// 对于 Redis，如果是 SCAN 命令，添加 COUNT 参数
	if strings.HasPrefix(strings.ToUpper(query), "SCAN") {
		query = fmt.Sprintf("%s COUNT %d", query, limit)
	}
	return s.executeRedisQuery(connectionID, query)
}

func (s *SimpleDatabaseManager) getRedisDatabases(connectionID string) ([]string, error) {
	_, exists := s.redisConnections[connectionID]
	if !exists {
		s.logger.WithField("connectionID", connectionID).Error("Redis connection not found in getRedisDatabases")
		return nil, fmt.Errorf("connection not found: %s", connectionID)
	}

	// Redis 有 16 个数据库 (0-15)
	databases := make([]string, 16)
	for i := 0; i < 16; i++ {
		databases[i] = fmt.Sprintf("db%d", i)
	}

	return databases, nil
}

func (s *SimpleDatabaseManager) getRedisTables(connectionID string, database string) ([]TableInfo, error) {
	client, exists := s.redisConnections[connectionID]
	if !exists {
		s.logger.WithField("connectionID", connectionID).Error("Redis connection not found in getRedisTables")
		return nil, fmt.Errorf("connection not found: %s", connectionID)
	}

	// 选择数据库
	dbNum, err := strconv.Atoi(strings.TrimPrefix(database, "db"))
	if err != nil {
		s.logger.Errorf("Invalid Redis database number: connectionID=%s, database=%s, error=%s", connectionID, database, err)
		return nil, fmt.Errorf("invalid database number: %s", database)
	}

	// 创建指定数据库的新客户端
	client = redis.NewClient(&redis.Options{
		Addr:     client.Options().Addr,
		Password: client.Options().Password,
		DB:       dbNum,
	})
	defer client.Close()

	// 获取所有键
	keys, err := client.Keys(s.ctx, "*").Result()
	if err != nil {
		s.logger.Errorf("Failed to get Redis keys: connectionID=%s, database=%s, error=%s", connectionID, database, err)
		return nil, err
	}

	var tables []TableInfo
	for _, key := range keys {
		// 获取键类型
		keyType, err := client.Type(s.ctx, key).Result()
		if err != nil {
			continue
		}

		// 获取 TTL
		ttl, err := client.TTL(s.ctx, key).Result()
		if err != nil {
			continue
		}

		table := TableInfo{
			Name:   key,
			Schema: database,
			Stats: map[string]string{
				"type": keyType,
				"ttl":  ttl.String(),
			},
		}
		tables = append(tables, table)
	}

	return tables, nil
}

func (s *SimpleDatabaseManager) getRedisTableInfo(connectionID string, database string, table string) (*TableInfo, error) {
	client, exists := s.redisConnections[connectionID]
	if !exists {
		return nil, fmt.Errorf("connection not found: %s", connectionID)
	}

	// 选择数据库
	dbNum, err := strconv.Atoi(strings.TrimPrefix(database, "db"))
	if err != nil {
		s.logger.Errorf("Invalid Redis database number: connectionID=%s, database=%s, error=%s", connectionID, database, err)
		return nil, fmt.Errorf("invalid database number: %s", database)
	}

	// 创建指定数据库的新客户端
	client = redis.NewClient(&redis.Options{
		Addr:     client.Options().Addr,
		Password: client.Options().Password,
		DB:       dbNum,
	})
	defer client.Close()

	tableInfo := &TableInfo{
		Name:   table,
		Schema: database,
		Stats:  make(map[string]string),
	}

	// 获取键类型
	keyType, err := client.Type(s.ctx, table).Result()
	if err != nil {
		return nil, err
	}
	tableInfo.Stats["type"] = keyType

	// 获取 TTL
	ttl, err := client.TTL(s.ctx, table).Result()
	if err != nil {
		return nil, err
	}
	tableInfo.Stats["ttl"] = ttl.String()

	// 获取内存使用
	memory, err := client.MemoryUsage(s.ctx, table).Result()
	if err == nil {
		tableInfo.Stats["memory"] = fmt.Sprintf("%d bytes", memory)
	}

	// 根据类型获取键信息
	switch keyType {
	case "string":
		value, err := client.Get(s.ctx, table).Result()
		if err == nil {
			tableInfo.Stats["value"] = value
		}
	case "list":
		length, err := client.LLen(s.ctx, table).Result()
		if err == nil {
			tableInfo.Stats["length"] = fmt.Sprintf("%d", length)
		}
	case "set":
		length, err := client.SCard(s.ctx, table).Result()
		if err == nil {
			tableInfo.Stats["length"] = fmt.Sprintf("%d", length)
		}
	case "zset":
		length, err := client.ZCard(s.ctx, table).Result()
		if err == nil {
			tableInfo.Stats["length"] = fmt.Sprintf("%d", length)
		}
	case "hash":
		length, err := client.HLen(s.ctx, table).Result()
		if err == nil {
			tableInfo.Stats["length"] = fmt.Sprintf("%d", length)
		}
	}

	return tableInfo, nil
}

func (s *SimpleDatabaseManager) getRedisTableData(connectionID string, database string, table string, limit int, offset int) (*QueryResult, error) {
	client, exists := s.redisConnections[connectionID]
	if !exists {
		return nil, fmt.Errorf("connection not found: %s", connectionID)
	}

	// 选择数据库
	dbNum, err := strconv.Atoi(strings.TrimPrefix(database, "db"))
	if err != nil {
		s.logger.Errorf("Invalid Redis database number: connectionID=%s, database=%s, error=%s", connectionID, database, err)
		return nil, fmt.Errorf("invalid database number: %s", database)
	}

	// 创建指定数据库的新客户端
	client = redis.NewClient(&redis.Options{
		Addr:     client.Options().Addr,
		Password: client.Options().Password,
		DB:       dbNum,
	})
	defer client.Close()

	start := time.Now()

	// 获取键类型
	keyType, err := client.Type(s.ctx, table).Result()
	if err != nil {
		return &QueryResult{Error: err.Error(), Time: time.Since(start).Milliseconds()}, err
	}

	var columns []string
	var rows [][]interface{}

	switch keyType {
	case "string":
		columns = []string{"Value"}
		value, err := client.Get(s.ctx, table).Result()
		if err != nil {
			return &QueryResult{Error: err.Error(), Time: time.Since(start).Milliseconds()}, err
		}
		rows = append(rows, []interface{}{value})

	case "list":
		columns = []string{"Index", "Value"}
		values, err := client.LRange(s.ctx, table, int64(offset), int64(offset+limit-1)).Result()
		if err != nil {
			return &QueryResult{Error: err.Error(), Time: time.Since(start).Milliseconds()}, err
		}
		for i, value := range values {
			rows = append(rows, []interface{}{offset + i, value})
		}

	case "set":
		columns = []string{"Value"}
		values, err := client.SMembers(s.ctx, table).Result()
		if err != nil {
			return &QueryResult{Error: err.Error(), Time: time.Since(start).Milliseconds()}, err
		}
		// 应用分页
		if offset < len(values) {
			end := offset + limit
			if end > len(values) {
				end = len(values)
			}
			for _, value := range values[offset:end] {
				rows = append(rows, []interface{}{value})
			}
		}

	case "zset":
		columns = []string{"Score", "Value"}
		values, err := client.ZRangeWithScores(s.ctx, table, int64(offset), int64(offset+limit-1)).Result()
		if err != nil {
			return &QueryResult{Error: err.Error(), Time: time.Since(start).Milliseconds()}, err
		}
		for _, z := range values {
			rows = append(rows, []interface{}{z.Score, z.Member})
		}

	case "hash":
		columns = []string{"Field", "Value"}
		values, err := client.HGetAll(s.ctx, table).Result()
		if err != nil {
			return &QueryResult{Error: err.Error(), Time: time.Since(start).Milliseconds()}, err
		}
		// 应用分页
		keys := make([]string, 0, len(values))
		for k := range values {
			keys = append(keys, k)
		}
		if offset < len(keys) {
			end := offset + limit
			if end > len(keys) {
				end = len(keys)
			}
			for _, k := range keys[offset:end] {
				rows = append(rows, []interface{}{k, values[k]})
			}
		}
	}

	return &QueryResult{
		Columns: columns,
		Rows:    rows,
		Count:   len(rows),
		Time:    time.Since(start).Milliseconds(),
	}, nil
}

func (s *SimpleDatabaseManager) getRedisDatabaseInfo(connectionID string) (*DatabaseInfo, error) {
	client, exists := s.redisConnections[connectionID]
	if !exists {
		return nil, fmt.Errorf("connection not found: %s", connectionID)
	}

	info := &DatabaseInfo{}

	// 获取 Redis 版本
	version, err := client.Info(s.ctx, "server").Result()
	if err != nil {
		return nil, err
	}

	// 从信息中解析版本
	lines := strings.Split(version, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "redis_version:") {
			info.Version = strings.TrimSpace(strings.TrimPrefix(line, "redis_version:"))
			break
		}
	}

	// 获取当前数据库
	dbSize, err := client.DBSize(s.ctx).Result()
	if err != nil {
		return nil, err
	}

	info.Name = fmt.Sprintf("Redis (db0) - %d keys", dbSize)

	return info, nil
}

func (s *SimpleDatabaseManager) validateRedisQuery(query string) error {
	query = strings.ToLower(strings.TrimSpace(query))

	dangerousOps := []string{"flushdb", "flushall", "shutdown", "debug"}
	for _, op := range dangerousOps {
		if strings.HasPrefix(query, op) {
			return fmt.Errorf("potentially dangerous operation detected: %s", op)
		}
	}

	return nil
}
