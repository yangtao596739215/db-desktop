package models

import (
	"time"
)

// Message represents a single message in the conversation
type Message struct {
	Role       string         `json:"role"`
	Content    string         `json:"content,omitempty"`
	ToolCalls  []*MCPToolCall `json:"tool_calls,omitempty"`   //assistant role的才有，大模型返回的
	ToolCallID string         `json:"tool_call_id,omitempty"` //role为tool的才有，自己封装的带着tool执行结果再请求大模型
}

// Conversation represents a chat conversation
type Conversation struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	MessageCount int       `json:"messageCount"`
}

// ConversationWithMessages represents a conversation with its messages
type ConversationWithMessages struct {
	Conversation
	Messages []*Message `json:"messages"`
}

// QueryHistory represents a database query execution history
type QueryHistory struct {
	ID             int       `json:"id"`
	Query          string    `json:"query"`
	ExecutionTime  int64     `json:"executionTime"`  // 执行时间（毫秒）
	DBType         string    `json:"dbType"`         // mysql, redis, clickhouse
	ConnectionID   string    `json:"connectionId"`   // 连接ID
	ConnectionName string    `json:"connectionName"` // 连接名称
	CreatedAt      time.Time `json:"createdAt"`
	Success        bool      `json:"success"`    // 是否执行成功
	Error          string    `json:"error"`      // 错误信息，如果有的话
	ResultRows     int       `json:"resultRows"` // 结果行数
}

type MsgType string

const (
	MsgTypeText     MsgType = "text"
	MsgTypeCard     MsgType = "card"
	MsgTypeComplete MsgType = "complete"
)

type MsgVo struct {
	ConversationID string  `json:"conversation_id"`
	Type           MsgType `json:"type"`
	Content        string  `json:"content"`
}
