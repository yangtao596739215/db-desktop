package sqlite

import (
	"encoding/json"
	"time"

	"db-desktop/backend/models"
)

// MessageDo represents a message in the database
type MessageDo struct {
	ID             string    `json:"id" db:"id"`
	ConversationID string    `json:"conversationId" db:"conversation_id"`
	Role           string    `json:"role" db:"role"` // "user", "assistant", or "tool"
	Content        string    `json:"content" db:"content"`
	CreatedAt      time.Time `json:"createdAt" db:"created_at"`
	ToolCallID     string    `json:"toolCallId" db:"tool_call_id"` // 工具调用ID，如果有值则说明是MCP工具消息
	ToolCalls      string    `json:"toolCalls" db:"tool_calls"`    // JSON序列化的ToolCalls
}

// ConversationDo represents a conversation in the database
type ConversationDo struct {
	ID           string    `json:"id" db:"id"`
	Title        string    `json:"title" db:"title"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time `json:"updatedAt" db:"updated_at"`
	MessageCount int       `json:"messageCount" db:"message_count"`
}

// ConversationWithMessagesDo represents a conversation with its messages in the database
type ConversationWithMessagesDo struct {
	ConversationDo
	Messages []MessageDo `json:"messages"`
}

// QueryHistoryDo represents a query history in the database
type QueryHistoryDo struct {
	ID             int       `json:"id" db:"id"`
	Query          string    `json:"query" db:"query"`
	ExecutionTime  int64     `json:"executionTime" db:"execution_time"`   // 执行时间（毫秒）
	DBType         string    `json:"dbType" db:"db_type"`                 // mysql, redis, clickhouse
	ConnectionID   string    `json:"connectionId" db:"connection_id"`     // 连接ID
	ConnectionName string    `json:"connectionName" db:"connection_name"` // 连接名称
	CreatedAt      time.Time `json:"createdAt" db:"created_at"`
	Success        bool      `json:"success" db:"success"`        // 是否执行成功
	Error          string    `json:"error" db:"error"`            // 错误信息，如果有的话
	ResultRows     int       `json:"resultRows" db:"result_rows"` // 结果行数
}

// ToMessageDo converts Message to MessageDo
func ToMessageDo(m *models.Message, conversationID string) *MessageDo {
	toolCallsJSON := ""
	if len(m.ToolCalls) > 0 {
		if data, err := json.Marshal(m.ToolCalls); err == nil {
			toolCallsJSON = string(data)
		}
	}

	return &MessageDo{
		ID:             "", // Will be set by database
		ConversationID: conversationID,
		Role:           m.Role,
		Content:        m.Content,
		CreatedAt:      time.Now(),
		ToolCallID:     m.ToolCallID,
		ToolCalls:      toolCallsJSON,
	}
}

// ToMessage converts MessageDo to Message
func (md *MessageDo) ToMessage() *models.Message {
	var toolCalls []models.MCPToolCall
	if md.ToolCalls != "" {
		json.Unmarshal([]byte(md.ToolCalls), &toolCalls)
	}

	// Convert to pointer slice
	pointerToolCalls := make([]*models.MCPToolCall, len(toolCalls))
	for i := range toolCalls {
		pointerToolCalls[i] = &toolCalls[i]
	}

	return &models.Message{
		Role:       md.Role,
		Content:    md.Content,
		ToolCalls:  pointerToolCalls,
		ToolCallID: md.ToolCallID,
	}
}

// ToConversationDo converts Conversation to ConversationDo
func ToConversationDo(c *models.Conversation) *ConversationDo {
	return &ConversationDo{
		ID:           c.ID,
		Title:        c.Title,
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
		MessageCount: c.MessageCount,
	}
}

// ToConversation converts ConversationDo to Conversation
func (cd *ConversationDo) ToConversation() *models.Conversation {
	return &models.Conversation{
		ID:           cd.ID,
		Title:        cd.Title,
		CreatedAt:    cd.CreatedAt,
		UpdatedAt:    cd.UpdatedAt,
		MessageCount: cd.MessageCount,
	}
}

// ToQueryHistoryDo converts QueryHistory to QueryHistoryDo
func ToQueryHistoryDo(qh *models.QueryHistory) *QueryHistoryDo {
	return &QueryHistoryDo{
		ID:             qh.ID,
		Query:          qh.Query,
		ExecutionTime:  qh.ExecutionTime,
		DBType:         qh.DBType,
		ConnectionID:   qh.ConnectionID,
		ConnectionName: qh.ConnectionName,
		CreatedAt:      qh.CreatedAt,
		Success:        qh.Success,
		Error:          qh.Error,
		ResultRows:     qh.ResultRows,
	}
}

// ToQueryHistory converts QueryHistoryDo to QueryHistory
func (qhd *QueryHistoryDo) ToQueryHistory() *models.QueryHistory {
	return &models.QueryHistory{
		ID:             qhd.ID,
		Query:          qhd.Query,
		ExecutionTime:  qhd.ExecutionTime,
		DBType:         qhd.DBType,
		ConnectionID:   qhd.ConnectionID,
		ConnectionName: qhd.ConnectionName,
		CreatedAt:      qhd.CreatedAt,
		Success:        qhd.Success,
		Error:          qhd.Error,
		ResultRows:     qhd.ResultRows,
	}
}
