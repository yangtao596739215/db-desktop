# 流式响应MCP处理逻辑

## 概述

本文档描述了优化后的流式响应处理逻辑，该逻辑能够智能检测流式响应中是否包含MCP（Model Context Protocol）工具调用，并相应地调整处理策略。

## 处理流程

### 1. 流式响应检测

当AI模型返回流式响应时，系统会实时检测响应中是否包含工具调用：

```go
// 在流式响应中检测MCP调用
if len(choice.Delta.ToolCalls) > 0 {
    hasMCP = true
    // 检测到工具调用，标记为MCP响应
}
```

### 2. 分支处理逻辑

#### 2.1 无MCP调用的情况

- **实时返回**: 流式响应内容实时返回给前端展示
- **消息存储**: 流式响应结束后，将完整的用户消息和AI回复保存到SQLite数据库

```go
if !hasMCP {
    // 如果没有MCP调用，直接保存消息到数据库
    s.saveStreamMessageToDatabase(conversationID, message, fullContent)
}
```

#### 2.2 包含MCP调用的情况

- **实时返回**: 流式响应内容仍然实时返回给前端展示
- **等待完整响应**: 等待流式响应完全结束后，再执行MCP逻辑
- **重新请求**: 由于流式响应中的工具调用信息可能不完整，系统会重新发送非流式请求以获取完整的工具调用信息
- **MCP处理**: 获取完整工具调用信息后，创建确认卡片并执行相应的数据库操作

```go
if hasMCP {
    // 如果包含MCP调用，等待流式响应结束后再处理MCP逻辑
    s.handleStreamCompleteWithMCP(fullContent, conversationID)
}
```

### 3. 核心组件

#### 3.1 StreamResponseCallback

```go
type StreamResponseCallback struct {
    OnContent     func(string)                    // 每个内容块的实时回调
    OnComplete    func(string, bool)              // 流式响应完成回调 (内容, 是否有MCP)
    OnError       func(error)                    // 错误回调
}
```

#### 3.2 主要方法

- `SendMessageStreamWithCallback`: 支持MCP检测的流式响应发送
- `handleStreamCompleteWithMCP`: 处理包含MCP的流式响应完成
- `handleMCPAfterStream`: 流式响应结束后的MCP处理
- `saveStreamMessageToDatabase`: 保存流式响应消息到数据库

## 优势

1. **用户体验优化**: 无论是否包含MCP调用，用户都能实时看到AI的回复内容
2. **MCP处理完整性**: 确保MCP工具调用信息的完整性，避免流式响应中信息不完整的问题
3. **性能优化**: 只有在检测到MCP调用时才进行额外的非流式请求
4. **数据一致性**: 统一的消息存储逻辑，确保数据库中的数据完整性

## 使用示例

### 前端调用

```javascript
// 前端调用流式响应
await sendMessageToCurrentConversation(userMessage)
```

### 后端处理

```go
// 后端使用新的流式响应处理
err := a.aiService.SendMessageStreamWithMCPDetection(message, conversation, conversationID, callback)
```

## 日志输出

系统会输出详细的日志信息来跟踪处理过程：

```
📤 Streaming request details: url=..., method=POST, bodySize=..., stream=true
Stream completed - hasMCP: true, content length: 150
Stream response contains MCP calls, processing after stream completion
Handling stream completion with MCP: conversationID=...
MCP tool calls detected after stream completion
Stream message saved to database: conversationID=...
```

## 测试

使用提供的测试脚本验证功能：

```bash
./test_streaming_mcp.sh
```

该脚本会测试：
1. 普通流式响应（无MCP）
2. 包含MCP的流式响应（Redis）
3. 包含MCP的流式响应（MySQL）

## 注意事项

1. 流式响应中的工具调用信息可能不完整，因此需要重新发送非流式请求
2. 只有在检测到MCP调用时才会进行额外的处理
3. 消息存储统一在流式响应结束后进行，确保数据一致性
4. 前端无需修改，现有的流式响应处理逻辑仍然有效
