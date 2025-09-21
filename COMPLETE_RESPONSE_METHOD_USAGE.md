# SendMessageStreamWithCompleteResponse 方法使用说明

## 概述

`SendMessageStreamWithCompleteResponse` 是 logic 层新增的方法，它调用 integration 层的同名方法，返回后将 message 写入 SQLite，然后传入流式响应 thunk 的回调函数。

## 方法签名

```go
func (s *AIService) SendMessageStreamWithCompleteResponse(message string, conversation []*models.Message, conversationID string, callback func(string)) error
```

## 参数说明

- `message`: 用户输入的消息内容
- `conversation`: 当前对话的历史消息列表
- `conversationID`: 对话ID，用于保存消息到数据库
- `callback`: 流式响应回调函数，用于实时返回内容给前端

## 功能特性

1. **流式响应**: 实时返回 AI 生成的内容给前端
2. **完整响应保存**: 流式响应完成后，将完整的响应保存到 SQLite 数据库
3. **MCP 工具调用支持**: 自动检测和处理 MCP 工具调用
4. **工具确认流程**: 支持工具确认卡片显示和用户交互

## 工作流程

1. 检查 API Key 配置
2. 构建完整的消息列表（包含系统消息、历史消息、用户消息）
3. 创建流式响应回调函数
4. 调用 integration 层的 `SendMessageStreamWithCompleteResponse` 方法
5. 将完整响应保存到 SQLite 数据库
6. 如果有 MCP 工具调用，处理工具调用逻辑

## 使用示例

```go
// 创建 AI 服务实例
aiService := logic.NewAIService(config)

// 设置必要的管理器
aiService.SetSQLiteManager(sqliteManager)
aiService.SetDatabaseManager(dbManager)
aiService.SetCardManager(cardManager)

// 定义流式响应回调
callback := func(chunk string) {
    // 实时处理流式内容
    fmt.Printf("收到流式内容: %s", chunk)
}

// 调用方法
err := aiService.SendMessageStreamWithCompleteResponse(
    "请查询数据库中的用户表",
    conversation,
    conversationID,
    callback,
)

if err != nil {
    log.Printf("发送消息失败: %v", err)
}
```

## 与现有方法的区别

| 方法 | 流式响应 | 完整响应保存 | MCP 支持 | 工具确认 |
|------|----------|--------------|----------|----------|
| `SendMessageStream` | ✅ | ❌ | ❌ | ❌ |
| `SendMessageStreamWithMCPDetection` | ✅ | ✅ | ✅ | ✅ |
| `SendMessageStreamWithCompleteResponse` | ✅ | ✅ | ✅ | ✅ |

## 优势

1. **更好的性能**: 直接调用 integration 层方法，减少中间层处理
2. **完整的数据保存**: 确保所有响应都被正确保存到数据库
3. **统一的接口**: 提供统一的流式响应和完整响应处理接口
4. **更好的错误处理**: 统一的错误处理机制

## 注意事项

1. 确保在调用前设置了必要的管理器（SQLite、Database、Card）
2. 流式响应回调函数应该能够处理空字符串和特殊字符
3. 对话ID必须是有效的，否则无法保存到数据库
4. 方法会自动处理系统消息的添加，无需手动添加

## 测试

使用提供的测试脚本进行功能验证：

```bash
./test_complete_response_method.sh
```

测试将验证：
- 流式响应是否正常工作
- 完整响应是否正确保存到 SQLite
- MCP 工具调用是否正确处理
- 工具确认卡片是否正常显示
