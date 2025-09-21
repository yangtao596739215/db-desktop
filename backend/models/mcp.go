package models

import "fmt"

// MCPTool represents a tool that can be called by the AI
type MCPTool struct {
	ID       string      `json:"id"` // 工具ID,result的tool_call_id
	Type     string      `json:"type"`
	Function MCPFunction `json:"function"`
}

// MCPFunction represents the function definition for MCP tools
type MCPFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// MCPToolCall represents a tool call from the AI
type MCPToolCall struct {
	ID       string          `json:"id"`
	Type     string          `json:"type"`
	Function MCPFunctionCall `json:"function"`
}

func (m *MCPToolCall) String() string {
	return fmt.Sprintf("ID=%s, Type=%s, Function=%s", m.ID, m.Type, m.Function)
}

// MCPFunctionCall represents the actual function call
type MCPFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}
