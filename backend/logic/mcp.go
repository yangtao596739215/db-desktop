package logic

import (
	"db-desktop/backend/database"
	"db-desktop/backend/models"
	"db-desktop/backend/sqlite"
	"db-desktop/backend/utils"
	"encoding/json"
	"fmt"
)

// 注意：globalSQLiteManager现在在ai.go中定义

// GetMCPTools returns the available MCP tools
func GetMCPTools() []models.MCPTool {
	return []models.MCPTool{
		{
			Type: "function",
			Function: models.MCPFunction{
				Name:        "execute_redis_command",
				Description: "Execute Redis commands. Use this to interact with Redis databases. The system will automatically use the currently connected Redis database.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"command": map[string]interface{}{
							"type":        "string",
							"description": "The Redis command to execute (e.g., 'GET key', 'SET key value', 'KEYS *')",
						},
					},
					"required": []string{"command"},
				},
			},
		},
		{
			Type: "function",
			Function: models.MCPFunction{
				Name:        "execute_mysql_query",
				Description: "Execute MySQL/SQL queries. Use this to query MySQL databases. The system will automatically use the currently connected MySQL database.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"query": map[string]interface{}{
							"type":        "string",
							"description": "The SQL query to execute",
						},
					},
					"required": []string{"query"},
				},
			},
		},
		{
			Type: "function",
			Function: models.MCPFunction{
				Name:        "execute_clickhouse_query",
				Description: "Execute ClickHouse queries. Use this to query ClickHouse databases. The system will automatically use the currently connected ClickHouse database.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"query": map[string]interface{}{
							"type":        "string",
							"description": "The ClickHouse query to execute",
						},
					},
					"required": []string{"query"},
				},
			},
		},
	}
}

type McpRunner func(parameters map[string]interface{}) (string, error)

// mcp执行失败把errorMsg放content里，让大模型回答
func ExecuteMcp(tool *models.MCPToolCall) *models.Message {
	runner, ok := mcpNameToRunner[tool.Function.Name]
	if !ok {
		utils.Errorf("mcp runner not found: %s", tool.Function.Name)
		return &models.Message{
			Role:       "tool",
			ToolCallID: tool.ID,
			Content:    "mcp runner not found",
		}
	}

	var args map[string]interface{}
	if err := json.Unmarshal([]byte(tool.Function.Arguments), &args); err != nil {
		utils.Errorf("failed to unmarshal function arguments: %v", err)
		return &models.Message{
			Role:       "tool",
			ToolCallID: tool.ID,
			Content:    "failed to unmarshal function arguments",
		}
	}

	result, err := runner(args)
	if err != nil {
		utils.Errorf("failed to execute mcp: %v", err)
		return &models.Message{
			Role:       "tool",
			ToolCallID: tool.ID,
			Content:    err.Error(),
		}
	}
	return &models.Message{
		Role:       "tool",
		ToolCallID: tool.ID,
		Content:    result,
	}
}

var mcpNameToRunner = map[string]McpRunner{
	"execute_redis_command":    RedisMcpRunner,
	"execute_mysql_query":      MysqlMcpRunner,
	"execute_clickhouse_query": ClickhouseMcpRunner,
}

func RedisMcpRunner(parameters map[string]interface{}) (string, error) {
	command, ok := parameters["command"].(string)
	if !ok {
		utils.Errorf("Missing or invalid Redis command in arguments: args=%s", parameters)
		return "Missing or invalid command", nil
	}

	utils.Infof("🔍 Looking for connected Redis database: command=%s", command)

	// Find connected Redis database
	connectionID, err := findConnectedDatabase(database.Redis)
	if err != nil {
		utils.Errorf("No connected Redis database found for MCP tool: command=%s, error=%s", command, err)
		return fmt.Sprintf("No connected Redis database found: %v", err), nil
	}

	utils.Infof("✅ Found connected Redis database, executing command: connectionID=%s, command=%s", connectionID, command)

	// Execute Redis command using the global database
	result, err := globalDB.ExecuteQuery(connectionID, command)
	if err != nil {
		utils.Errorf("Redis command execution failed: connectionID=%s, command=%s, error=%s", connectionID, command, err)
		return fmt.Sprintf("Redis command failed: %v", err), nil
	}

	utils.Infof("✅ Redis command executed successfully: connectionID=%s, command=%s, resultRows=%d, executionTime=%d, hasError=%t",
		connectionID, command, len(result.Rows), result.Time, result.Error != "")

	// Record query history
	success := result.Error == ""
	errorMsg := result.Error
	_, err = sqlite.AddQueryHistory(command, "redis", "ai_redis", "AI Redis Command",
		int64(result.Time), success, errorMsg, len(result.Rows))
	if err != nil {
		utils.Errorf("Failed to add Redis command to query history: command=%s, error=%s", command, err)
	}

	// Format the result
	content := "🔧 **Redis Command Executed**\n"
	content += "**Command:** `" + command + "`\n"
	content += "**Connection ID:** " + connectionID + "\n"

	if result.Error != "" {
		content += fmt.Sprintf("**❌ Error:** %s\n", result.Error)
	} else {
		if len(result.Rows) > 0 {
			content += "**✅ Result:**\n"
			for i, row := range result.Rows {
				if len(row) > 0 {
					content += fmt.Sprintf("  %d. %v\n", i+1, row[0])
				}
			}
		} else {
			content += "**✅ Result:** (empty)\n"
		}
		content += fmt.Sprintf("**⏱️ Execution time:** %dms\n", result.Time)
	}

	return content, nil
}

func MysqlMcpRunner(parameters map[string]interface{}) (string, error) {
	query, ok := parameters["query"].(string)
	if !ok {
		utils.Errorf("Missing or invalid MySQL query in arguments: args=%s", parameters)
		return "Missing or invalid query", nil
	}

	utils.Infof("🔍 Looking for connected MySQL database: query=%s", query)

	// Find connected MySQL database
	connectionID, err := findConnectedDatabase(database.MySQL)
	if err != nil {
		utils.Errorf("No connected MySQL database found for MCP tool: query=%s, error=%s", query, err)
		return fmt.Sprintf("No connected MySQL database found: %v", err), nil
	}

	utils.Infof("✅ Found connected MySQL database, executing query: connectionID=%s, query=%s", connectionID, query)

	// Execute MySQL query using the global database
	result, err := globalDB.ExecuteQuery(connectionID, query)
	if err != nil {
		utils.Errorf("MySQL query execution failed: connectionID=%s, query=%s, error=%s", connectionID, query, err)
		return fmt.Sprintf("MySQL query failed: %v", err), nil
	}

	utils.Infof("✅ MySQL query executed successfully: connectionID=%s, query=%s, resultRows=%d, resultColumns=%d, executionTime=%d, hasError=%t",
		connectionID, query, len(result.Rows), len(result.Columns), result.Time, result.Error != "")

	// Record query history
	success := result.Error == ""
	errorMsg := result.Error
	_, err = sqlite.AddQueryHistory(query, "mysql", "ai_mysql", "AI MySQL Query",
		int64(result.Time), success, errorMsg, len(result.Rows))
	if err != nil {
		utils.Errorf("Failed to add MySQL query to query history: query=%s, error=%s", query, err)
	}

	// Format the result
	content := "🔧 **MySQL Query Executed**\n"
	content += "**Query:** `" + query + "`\n"
	content += "**Connection ID:** " + connectionID + "\n"

	if result.Error != "" {
		content += fmt.Sprintf("**❌ Error:** %s\n", result.Error)
	} else {
		content += "**✅ Result:**\n"
		content += fmt.Sprintf("**Columns:** %v\n", result.Columns)
		content += fmt.Sprintf("**Rows count:** %d\n", result.Count)

		if len(result.Rows) > 0 {
			content += "**Data:**\n"
			// 显示前几行数据
			maxRows := 10
			if len(result.Rows) < maxRows {
				maxRows = len(result.Rows)
			}
			for i := 0; i < maxRows; i++ {
				content += fmt.Sprintf("  %d. %v\n", i+1, result.Rows[i])
			}
			if len(result.Rows) > maxRows {
				content += fmt.Sprintf("  ... and %d more rows\n", len(result.Rows)-maxRows)
			}
		}
		content += fmt.Sprintf("**⏱️ Execution time:** %dms\n", result.Time)
	}

	return content, nil
}

func ClickhouseMcpRunner(parameters map[string]interface{}) (string, error) {
	query, ok := parameters["query"].(string)
	if !ok {
		utils.Errorf("Missing or invalid ClickHouse query in arguments: args=%s", parameters)
		return "Missing or invalid query", nil
	}

	utils.Infof("🔍 Looking for connected ClickHouse database: query=%s", query)

	// Find connected ClickHouse database
	connectionID, err := findConnectedDatabase(database.ClickHouse)
	if err != nil {
		utils.Errorf("No connected ClickHouse database found for MCP tool: query=%s, error=%s", query, err)
		return fmt.Sprintf("No connected ClickHouse database found: %v", err), nil
	}

	utils.Infof("✅ Found connected ClickHouse database, executing query: connectionID=%s, query=%s", connectionID, query)

	// Execute ClickHouse query using the global database
	result, err := globalDB.ExecuteQuery(connectionID, query)
	if err != nil {
		utils.Errorf("ClickHouse query execution failed: connectionID=%s, query=%s, error=%s", connectionID, query, err)
		return fmt.Sprintf("ClickHouse query failed: %v", err), nil
	}

	utils.Infof("✅ ClickHouse query executed successfully: connectionID=%s, query=%s, resultRows=%d, resultColumns=%d, executionTime=%d, hasError=%t",
		connectionID, query, len(result.Rows), len(result.Columns), result.Time, result.Error != "")

	// Record query history
	success := result.Error == ""
	errorMsg := result.Error
	_, err = sqlite.AddQueryHistory(query, "clickhouse", "ai_clickhouse", "AI ClickHouse Query",
		int64(result.Time), success, errorMsg, len(result.Rows))
	if err != nil {
		utils.Errorf("Failed to add ClickHouse query to query history: query=%s, error=%s", query, err)
	}

	// Format the result
	content := "🔧 **ClickHouse Query Executed**\n"
	content += "**Query:** `" + query + "`\n"
	content += "**Connection ID:** " + connectionID + "\n"

	if result.Error != "" {
		content += fmt.Sprintf("**❌ Error:** %s\n", result.Error)
	} else {
		content += "**✅ Result:**\n"
		content += fmt.Sprintf("**Columns:** %v\n", result.Columns)
		content += fmt.Sprintf("**Rows count:** %d\n", result.Count)

		if len(result.Rows) > 0 {
			content += "**Data:**\n"
			// 显示前几行数据
			maxRows := 10
			if len(result.Rows) < maxRows {
				maxRows = len(result.Rows)
			}
			for i := 0; i < maxRows; i++ {
				content += fmt.Sprintf("  %d. %v\n", i+1, result.Rows[i])
			}
			if len(result.Rows) > maxRows {
				content += fmt.Sprintf("  ... and %d more rows\n", len(result.Rows)-maxRows)
			}
		}
		content += fmt.Sprintf("**⏱️ Execution time:** %dms\n", result.Time)
	}

	return content, nil
}

// findConnectedDatabase finds a connected database of the specified type
func findConnectedDatabase(dbType database.DatabaseType) (string, error) {
	if globalDB == nil {
		return "", fmt.Errorf("global database not available")
	}

	connections := globalDB.ListConnections()
	for _, conn := range connections {
		if conn.Type == dbType {
			status := globalDB.GetConnectionStatus(conn.ID)
			if status != nil && status.Status == "connected" {
				return conn.ID, nil
			}
		}
	}

	utils.Warnf("No connected database found for MCP tool: dbType=%s", dbType)
	return "", fmt.Errorf("no connected %s database found", dbType)
}
