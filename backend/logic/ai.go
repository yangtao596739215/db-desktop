package logic

import (
	"encoding/json"
	"fmt"
	"sync"

	"db-desktop/backend/database"
	"db-desktop/backend/integration"
	"db-desktop/backend/models"
	"db-desktop/backend/utils"
)

// DatabaseInterface 定义数据库操作接口
type DatabaseInterface interface {
	ExecuteQuery(connectionID string, query string) (*database.QueryResult, error)
	ListConnections() []*database.ConnectionConfig
	GetConnectionStatus(id string) *database.ConnectionStatus
}

// 全局实例
var (
	globalDB            DatabaseInterface
	globalSQLiteManager SQLiteManagerInterface
	aiService           *AIService
	cardManager         *CardManager
	initOnce            sync.Once
)

// init 初始化logic包
func init() {
	initOnce.Do(func() {
		// 初始化卡片管理器
		cardManager = &CardManager{
			cards: make(map[string]*ConfirmCard),
		}

		// 初始化AI服务
		aiService = &AIService{
			cardManager: cardManager,
		}

		// 设置全局数据库接口
		globalDB = &databaseAdapter{}

		utils.Infof("Logic package initialized successfully")
	})
}

// SetGlobalDatabase sets the global database interface
func SetGlobalDatabase(db DatabaseInterface) {
	globalDB = db
}

// SetGlobalSQLiteManager sets the global SQLite manager
func SetGlobalSQLiteManager(manager SQLiteManagerInterface) {
	globalSQLiteManager = manager
}

// AIService handles AI logic and coordinates with integration layer
type AIService struct {
	cardManager *CardManager
	mu          sync.RWMutex
}

// SQLiteManagerInterface 定义SQLite管理器接口
type SQLiteManagerInterface interface {
	AddMessageToConversation(conversationID string, message *models.Message) error
	GetMessagesForLLM(conversationID string) ([]*models.Message, error)
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

// SetCardManager sets the card manager
func (s *AIService) SetCardManager(cardManager *CardManager) {
	s.cardManager = cardManager
}

// LoadConfig loads AI configuration
func (s *AIService) LoadConfig() error {
	aiClient := integration.NewAIClient(integration.AIConfig{})
	return aiClient.LoadConfig()
}

// SaveConfig saves AI configuration
func (s *AIService) SaveConfig() error {
	aiClient := integration.NewAIClient(integration.AIConfig{})
	return aiClient.SaveConfig()
}

// UpdateConfig updates AI configuration
func (s *AIService) UpdateConfig(config integration.AIConfig) error {
	aiClient := integration.NewAIClient(config)
	return aiClient.UpdateConfig(config)
}

// GetConfig returns the current AI configuration
func (s *AIService) GetConfig() integration.AIConfig {
	aiClient := integration.NewAIClient(integration.AIConfig{})
	return aiClient.GetConfig()
}

// SendMessageStreamWithCompleteResponse sends a message and returns CompleteResponse directly
func (s *AIService) SendMessageStreamWithCompleteResponse(message string, conversationID string, callback func(*models.MsgVo)) error {
	utils.Infof("Sending message to AI with streaming: conversationID=%s, message=%s", conversationID, message)
	// 检查conversationID是否为空
	if conversationID == "" {
		utils.Warnf("ConversationID is required but not provided for streaming")
		callback(&models.MsgVo{
			ConversationID: conversationID,
			Type:           models.MsgTypeText,
			Content:        "错误：需要先创建会话",
		})
		return fmt.Errorf("conversationID is required")
	}

	// 立即保存用户消息到数据库
	if globalSQLiteManager != nil && message != "" {
		err := globalSQLiteManager.AddMessageToConversation(conversationID, &models.Message{
			Role:    "user",
			Content: message,
		})
		if err != nil {
			utils.Errorf("Failed to save user message: %v", err)
		} else {
			utils.Infof("User message saved to database: conversationID=%s", conversationID)
		}
	}
	return s.SendHistoryToAI(conversationID, callback)
}

// 所有需要发送的消息，先写db，然后调用这个方法，组合所有的系统上下文，系统提示词会在创建会话的时候就写入db
func (s *AIService) SendHistoryToAI(conversationID string, callback func(*models.MsgVo)) error {
	utils.Infof("Sending history to AI with streaming: conversationID=%s", conversationID)
	// 获取对话历史
	var history []*models.Message
	if globalSQLiteManager != nil {
		messages, err := globalSQLiteManager.GetMessagesForLLM(conversationID)
		if err != nil {
			utils.Errorf("Failed to get conversation messages: %v", err)
			return err
		}
		history = messages
	}

	// Check if API key is configured
	aiClient := integration.NewAIClient(integration.AIConfig{})
	// Load configuration from file
	if err := aiClient.LoadConfig(); err != nil {
		utils.Errorf("Failed to load AI config: %v", err)
		callback(&models.MsgVo{
			ConversationID: conversationID,
			Type:           models.MsgTypeText,
			Content:        "请先配置API Key和Base URL",
		})
		return nil
	}
	config := aiClient.GetConfig()
	if config.APIKey == "" {
		callback(&models.MsgVo{
			ConversationID: conversationID,
			Type:           models.MsgTypeText,
			Content:        "请先配置API Key和Base URL",
		})
		return nil
	}

	// 创建流式响应回调
	streamCallback := func(chunk integration.ChatCompletionChunk) {
		// 实时返回内容给前端
		if len(chunk.Choices) > 0 {
			delta := chunk.Choices[0].Delta
			if delta.Content != "" {
				callback(&models.MsgVo{
					ConversationID: conversationID,
					Type:           models.MsgTypeText,
					Content:        delta.Content,
				})
			}
		}
	}

	// 调用integration层的方法获取完整响应
	completeResponse, err := aiClient.SendMessageStreamWithCompleteResponse(history, streamCallback, GetMCPTools())
	if err != nil {
		utils.Errorf("Failed to send message with complete response: %v", err)
		return err
	}

	// 将完整响应保存到数据库
	// 创建assistant消息
	assistantMessage := &models.Message{
		Role:      "assistant",
		Content:   completeResponse.Content,
		ToolCalls: completeResponse.ToolCalls,
	}

	// 保存到数据库 - 使用全局SQLite管理器
	if globalSQLiteManager != nil {
		err = globalSQLiteManager.AddMessageToConversation(
			conversationID,
			assistantMessage,
		)
		if err != nil {
			utils.Errorf("Failed to save assistant message to database: %v", err)
		}
	}

	// 如果有工具调用，处理它们
	if len(completeResponse.ToolCalls) > 0 {
		utils.Infof("Stream response contains MCP calls, processing tool calls")
		s.handleToolCalls(completeResponse.ToolCalls, conversationID, callback)
	} else {
		// 如果没有工具调用，发送完成消息
		callback(&models.MsgVo{
			ConversationID: conversationID,
			Type:           models.MsgTypeComplete,
			Content:        "Stream complete",
		})
	}
	return nil
}

// handleToolCalls handles tool calls from AI response
func (s *AIService) handleToolCalls(toolCalls []*models.MCPToolCall, conversationID string, callback func(*models.MsgVo)) (string, error) {
	utils.Infof("AI wants to call tools, creating confirmation cards,toolCalls=%v", toolCalls)

	// Create confirmation cards for tool calls
	for _, toolCall := range toolCalls {
		utils.Infof("Processing tool call: toolCallID=%s, function=%s, arguments=%s",
			toolCall.ID, toolCall.Function.Name, toolCall.Function.Arguments)

		// 添加调试信息
		utils.Infof("Tool call details: ID='%s', Type='%s', Function.Name='%s', Function.Arguments='%s'",
			toolCall.ID, toolCall.Type, toolCall.Function.Name, toolCall.Function.Arguments)

		// Parse the function arguments
		var args map[string]interface{}
		if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
			continue // Skip invalid tool calls
		}

		// Create confirmation card if card manager is available
		if s.cardManager != nil {
			showContent := s.formatToolConfirmationMessage(toolCall.Function.Name, args)

			// Create callbacks for confirm and reject
			confirmCallback := func() {
				utils.Infof("Tool call confirmed via card: toolCallID=%s", toolCall.ID)
				// Execute the tool call
				mcpMsg := ExecuteMcp(toolCall)
				utils.Infof("✅ Tool call executed successfully: toolCallID=%s", toolCall.ID)

				// 保存工具执行结果到sqlite
				if globalSQLiteManager != nil {
					err := globalSQLiteManager.AddMessageToConversation(conversationID, mcpMsg)
					if err != nil {
						utils.Errorf("Failed to save tool result message to database: %v", err)
					}
				}
				callback(&models.MsgVo{
					ConversationID: conversationID,
					Type:           models.MsgTypeText,
					Content:        "工具执行成功",
				})
				//继续请求大模型，继续生成后续内容
				s.SendHistoryToAI(conversationID, callback)
			}

			rejectCallback := func() {
				utils.Infof("❌ Tool call rejected via card: toolCallID=%s, function=%s", toolCall.ID, toolCall.Function.Name)

				// 创建拒绝结果
				rejectResult := &models.Message{
					Role:       "tool",
					ToolCallID: toolCall.ID,
					Content:    "user reject the tool call",
				}

				// 保存工具执行结果到sqlite
				if globalSQLiteManager != nil {
					err := globalSQLiteManager.AddMessageToConversation(conversationID, rejectResult)
					if err != nil {
						utils.Errorf("Failed to save tool result message to database: %v", err)
					}
				}

				// 工具被拒绝，结果已保存到数据库
				callback(&models.MsgVo{
					ConversationID: conversationID,
					Type:           models.MsgTypeText,
					Content:        "工具被拒绝执行",
				})
				s.SendHistoryToAI(conversationID, callback)
			}

			// Create the confirmation card with metadata
			card := s.cardManager.CreateCardWithMetadata(showContent, confirmCallback, rejectCallback, conversationID, toolCall.ID)
			utils.Infof("Created confirmation card for tool call: cardID=%s, toolCallID=%s, showContent=%s",
				card.CardID, toolCall.ID, showContent)

			// 通过回调函数通知前端有新的卡片
			callback(&models.MsgVo{
				ConversationID: conversationID,
				Type:           models.MsgTypeCard,
				Content:        fmt.Sprintf("%s|%s|%s", card.CardID, toolCall.ID, showContent),
			})

			//saveCardMsg
			cardMsg := &models.Message{
				Role:    "card",
				Content: showContent,
			}
			if globalSQLiteManager != nil {
				err := globalSQLiteManager.AddMessageToConversation(conversationID, cardMsg)
				if err != nil {
					utils.Errorf("Failed to save card message to database: %v", err)
				}
			}
		}
	}

	// Return confirmation request message
	return "⚠️ 请查看下方的工具确认卡片，点击确认或拒绝按钮来继续。", nil
}

// formatToolConfirmationMessage formats a confirmation message for a tool call
func (s *AIService) formatToolConfirmationMessage(functionName string, args map[string]interface{}) string {
	switch functionName {
	case "execute_redis_command":
		command, _ := args["command"].(string)
		return fmt.Sprintf("执行Redis命令: `%s`", command)
	case "execute_mysql_query":
		query, _ := args["query"].(string)
		return fmt.Sprintf("执行MySQL查询: `%s`", query)
	case "execute_clickhouse_query":
		query, _ := args["query"].(string)
		return fmt.Sprintf("执行ClickHouse查询: `%s`", query)
	default:
		return fmt.Sprintf("执行工具: %s", functionName)
	}
}

// 包级函数，替代原来的对象方法调用

// GetAIService 获取全局AIService实例
func GetAIService() *AIService {
	return aiService
}

// GetCardManager 获取全局CardManager实例
func GetCardManager() *CardManager {
	return cardManager
}

// GetGlobalDatabase 获取全局数据库接口
func GetGlobalDatabase() DatabaseInterface {
	return globalDB
}

// GetGlobalSQLiteManager 获取全局SQLite管理器
func GetGlobalSQLiteManager() SQLiteManagerInterface {
	return globalSQLiteManager
}

// LoadConfig loads AI configuration
func LoadConfig() error {
	return aiService.LoadConfig()
}

// SaveConfig saves AI configuration
func SaveConfig() error {
	return aiService.SaveConfig()
}

// UpdateConfig updates AI configuration
func UpdateConfig(config integration.AIConfig) error {
	return aiService.UpdateConfig(config)
}

// GetConfig returns the current AI configuration
func GetConfig() integration.AIConfig {
	return aiService.GetConfig()
}

// SendMessageStreamWithCompleteResponse sends a message and returns CompleteResponse directly
func SendMessageStreamWithCompleteResponse(message string, conversationID string, callback func(*models.MsgVo)) error {
	return aiService.SendMessageStreamWithCompleteResponse(message, conversationID, callback)
}

// ConfirmCardByID confirms a card and executes the confirm callback
func ConfirmCardByID(cardID string) error {
	return cardManager.ConfirmCard(cardID)
}

// RejectCardByID rejects a card and executes the reject callback
func RejectCardByID(cardID string) error {
	return cardManager.RejectCard(cardID)
}
