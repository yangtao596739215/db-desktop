package integration

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"db-desktop/backend/models"
	"db-desktop/backend/utils"
)

// AIRequest represents the request structure for AI API
type AIRequest struct {
	Model       string            `json:"model"`
	Messages    []*models.Message `json:"messages"`
	Temperature float64           `json:"temperature,omitempty"`
	Stream      bool              `json:"stream,omitempty"`
	Tools       []models.MCPTool  `json:"tools,omitempty"`
	ToolChoice  string            `json:"tool_choice,omitempty"`
	MaxTokens   int               `json:"max_tokens,omitempty"`
}

// AIResponse represents the complete response structure from AI API
type AIResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
	Error   *AIError `json:"error,omitempty"`
}

// Choice represents a choice in the AI response
type Choice struct {
	Index        int            `json:"index"`
	Message      models.Message `json:"message"`
	Delta        models.Message `json:"delta,omitempty"` // For streaming responses
	FinishReason string         `json:"finish_reason"`
}

// Usage represents token usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// AIError represents an error from the AI API
type AIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code,omitempty"`
}

// AIConfig represents AI model configuration
type AIConfig struct {
	APIKey      string  `json:"apiKey"`
	BaseURL     string  `json:"baseURL"`
	Temperature float64 `json:"temperature"`
	Stream      bool    `json:"stream"`
}

// ChatCompletionChunk represents a streaming response chunk
type ChatCompletionChunk struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Delta struct {
			Role      string               `json:"role,omitempty"`
			Content   string               `json:"content,omitempty"`
			ToolCalls []models.MCPToolCall `json:"tool_calls,omitempty"`
		} `json:"delta"`
		Index        int    `json:"index"`
		FinishReason string `json:"finish_reason,omitempty"`
	} `json:"choices"`
}

// CompleteResponse represents the complete response after processing stream
type CompleteResponse struct {
	Content      string
	ToolCalls    []*models.MCPToolCall
	FinishReason string
}

// AIClient handles AI API calls
type AIClient struct {
	config     AIConfig
	client     *http.Client
	configFile string
}

// NewAIClient creates a new AI client
func NewAIClient(config AIConfig) *AIClient {
	homeDir, _ := os.UserHomeDir()
	configFile := filepath.Join(homeDir, ".db-desktop", "ai_config.json")

	return &AIClient{
		config:     config,
		configFile: configFile,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// LoadConfig loads AI configuration from file
func (c *AIClient) LoadConfig() error {
	// 创建配置目录
	configDir := filepath.Dir(c.configFile)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		utils.Errorf("Failed to create AI config directory: %v", err)
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// 检查配置文件是否存在
	if _, err := os.Stat(c.configFile); os.IsNotExist(err) {
		utils.Infof("No AI config file found, using default config")
		return nil // 没有配置文件，使用默认配置
	}

	data, err := os.ReadFile(c.configFile)
	if err != nil {
		utils.Errorf("Failed to read AI config file: %v", err)
		return fmt.Errorf("failed to read AI config file: %w", err)
	}

	var config AIConfig
	if err := json.Unmarshal(data, &config); err != nil {
		utils.Errorf("Failed to unmarshal AI config: %v", err)
		return fmt.Errorf("failed to unmarshal AI config: %w", err)
	}

	c.config = config
	apiKeyDisplay := config.APIKey
	if len(apiKeyDisplay) > 10 {
		apiKeyDisplay = apiKeyDisplay[:10] + "..."
	}
	utils.Infof("Loaded AI config from file: apiKey=%s, temperature=%f, stream=%t",
		apiKeyDisplay, config.Temperature, config.Stream)
	return nil
}

// SaveConfig saves AI configuration to file
func (c *AIClient) SaveConfig() error {
	// 创建配置目录
	configDir := filepath.Dir(c.configFile)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal AI config: %w", err)
	}

	if err := os.WriteFile(c.configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write AI config file: %w", err)
	}

	utils.Infof("Successfully saved AI config to file")
	return nil
}

// UpdateConfig updates AI configuration and saves to file
func (c *AIClient) UpdateConfig(newConfig AIConfig) error {
	c.config = newConfig
	return c.SaveConfig()
}

// GetConfig returns the current AI configuration
func (c *AIClient) GetConfig() AIConfig {
	return c.config
}

// ProcessStreamResponse processes streaming response and returns CompleteResponse
func ProcessStreamResponse(stream io.Reader, callback func(ChatCompletionChunk)) (*CompleteResponse, error) {
	scanner := bufio.NewScanner(stream)
	completeResponse := &CompleteResponse{
		ToolCalls: make([]*models.MCPToolCall, 0),
	}

	// 用于累积工具调用参数，使用索引作为key（因为流式响应中ID可能为空）
	toolCallArgs := make(map[int]*strings.Builder)
	toolCallMap := make(map[int]*models.MCPToolCall)
	toolCallIndex := 0

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")

			// 检查是否是结束标记
			if data == "[DONE]" {
				break
			}

			// 解析JSON数据
			var chunk ChatCompletionChunk
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				utils.Warnf("Failed to parse JSON chunk: %v, data: %s", err, data)
				continue // 跳过解析失败的chunk
			}

			// 调试日志：记录每个chunk的tool_calls信息
			if len(chunk.Choices) > 0 && len(chunk.Choices[0].Delta.ToolCalls) > 0 {
				utils.Infof("Received tool_calls chunk: count=%d", len(chunk.Choices[0].Delta.ToolCalls))
				for i, tc := range chunk.Choices[0].Delta.ToolCalls {
					utils.Infof("  ToolCall[%d]: ID='%s', Name='%s', Args='%s'",
						i, tc.ID, tc.Function.Name, tc.Function.Arguments)
				}
			}

			// 调用回调函数
			if callback != nil {
				callback(chunk)
			}

			// 处理每个选择
			for _, choice := range chunk.Choices {
				// 累积内容
				if choice.Delta.Content != "" {
					completeResponse.Content += choice.Delta.Content
				}

				// 处理工具调用
				for _, toolCall := range choice.Delta.ToolCalls {
					// 检查是否是新工具调用的开始（有ID表示新工具调用）
					if toolCall.ID != "" {
						// 检查是否已经存在这个ID的工具调用
						var existingIndex int = -1
						for i, existingToolCall := range toolCallMap {
							if existingToolCall.ID == toolCall.ID {
								existingIndex = i
								break
							}
						}

						if existingIndex == -1 {
							// 新工具调用，创建新的条目
							toolCallMap[toolCallIndex] = &models.MCPToolCall{
								ID:   toolCall.ID,
								Type: "function",
								Function: models.MCPFunctionCall{
									Name:      toolCall.Function.Name,
									Arguments: "",
								},
							}
							toolCallArgs[toolCallIndex] = &strings.Builder{}
							toolCallIndex++
						}
					}

					// 找到对应的工具调用进行更新
					var targetIndex int = -1
					if toolCall.ID != "" {
						// 通过ID查找
						for i, existingToolCall := range toolCallMap {
							if existingToolCall.ID == toolCall.ID {
								targetIndex = i
								break
							}
						}
					} else if toolCallIndex > 0 {
						// 如果没有ID，使用最新的工具调用
						targetIndex = toolCallIndex - 1
					}

					if targetIndex >= 0 {
						// 更新工具调用ID（如果提供了）
						if toolCall.ID != "" {
							toolCallMap[targetIndex].ID = toolCall.ID
						}

						// 更新函数名（如果提供了）
						if toolCall.Function.Name != "" {
							toolCallMap[targetIndex].Function.Name = toolCall.Function.Name
						}

						// 累积参数
						if toolCall.Function.Arguments != "" {
							if toolCallArgs[targetIndex] == nil {
								toolCallArgs[targetIndex] = &strings.Builder{}
							}
							toolCallArgs[targetIndex].WriteString(toolCall.Function.Arguments)
						}
					}
				}

				// 记录完成原因
				if choice.FinishReason != "" {
					completeResponse.FinishReason = choice.FinishReason
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取流失败: %v", err)
	}

	// 设置累积的参数并添加到结果中
	for i := 0; i < toolCallIndex; i++ {
		if toolCall, exists := toolCallMap[i]; exists {
			if builder, exists := toolCallArgs[i]; exists {
				toolCall.Function.Arguments = builder.String()
			}
			completeResponse.ToolCalls = append(completeResponse.ToolCalls, toolCall)

			// 调试日志：记录最终的工具调用信息
			utils.Infof("Final tool call[%d]: ID='%s', Type='%s', Name='%s', Args='%s'",
				i, toolCall.ID, toolCall.Type, toolCall.Function.Name, toolCall.Function.Arguments)
		}
	}

	utils.Infof("Processed streaming response: content length=%d, tool calls count=%d",
		len(completeResponse.Content), len(completeResponse.ToolCalls))

	return completeResponse, nil
}

// SendMessageStreamWithCompleteResponse sends a message and returns CompleteResponse directly
func (c *AIClient) SendMessageStreamWithCompleteResponse(messages []*models.Message, callback func(ChatCompletionChunk), tools []models.MCPTool) (*CompleteResponse, error) {
	// 设置默认的BaseURL
	baseURL := c.config.BaseURL
	if baseURL == "" {
		baseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"
	}

	// Create the request with tools enabled for MCP detection
	req := AIRequest{
		Model:       "qwen-plus", // 固定使用千问模型
		Messages:    messages,
		Temperature: c.config.Temperature,
		Stream:      true,
		Tools:       tools, // 启用工具以检测MCP调用
		ToolChoice:  "auto",
		MaxTokens:   2000,
	}

	// Marshal the request
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Print request details for debugging (streaming)
	utils.Infof("📤 Streaming request details: url=%s, method=%s, bodySize=%d, stream=%t", baseURL, "POST", len(reqBody), true)

	// Print request body (formatted JSON)
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, reqBody, "", "  "); err == nil {
		utils.Infof("📤 Streaming request body: %s", prettyJSON.String())
	} else {
		utils.Infof("📤 Streaming request body (raw): %s", string(reqBody))
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(
		context.Background(),
		"POST",
		baseURL,
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	httpReq.Header.Set("Accept", "text/event-stream")

	// Print headers
	apiKeyDisplay := c.config.APIKey
	if len(apiKeyDisplay) > 10 {
		apiKeyDisplay = apiKeyDisplay[:10] + "..."
	}
	utils.Infof("📤 Streaming request headers: Content-Type=%s, Authorization=%s, Accept=%s, User-Agent=%s",
		httpReq.Header.Get("Content-Type"),
		"Bearer "+apiKeyDisplay,
		httpReq.Header.Get("Accept"),
		httpReq.Header.Get("User-Agent"))

	// Send the request
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, string(respBody))
	}

	// 使用新的流式响应处理
	completeResponse, err := ProcessStreamResponse(resp.Body, callback)
	if err != nil {
		return nil, err
	}

	utils.Infof("📤 SendMessageStreamWithCompleteResponse request %v, completeResponse: %v", req, completeResponse)
	return completeResponse, nil
}

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
