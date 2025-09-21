# ConversationID 验证功能

## 功能概述

在发送消息接口中，如果 `conversationID` 为空，后端会返回特定的错误，前端收到这个错误后可以自动调用创建会话的接口。

## 实现细节

### 1. 错误定义

```go
var (
    ErrConversationIDRequired = errors.New("CONVERSATION_ID_REQUIRED")
)
```

### 2. 接口修改

#### SendMessageToConversation
```go
func (a *App) SendMessageToConversation(conversationID, message string) (string, error) {
    // 检查conversationID是否为空
    if conversationID == "" {
        utils.Warnf("ConversationID is required but not provided")
        return "", ErrConversationIDRequired
    }
    
    // ... 其他逻辑
}
```

#### SendMessageStreamToConversation
```go
func (a *App) SendMessageStreamToConversation(conversationID, message string, callback func(string)) error {
    // 检查conversationID是否为空
    if conversationID == "" {
        utils.Warnf("ConversationID is required but not provided for streaming")
        callback("错误：需要先创建会话")
        return ErrConversationIDRequired
    }
    
    // ... 其他逻辑
}
```

## 前端处理流程

### 1. 错误捕获
前端需要捕获 `CONVERSATION_ID_REQUIRED` 错误：

```javascript
try {
    const response = await sendMessage(conversationID, message);
    // 处理正常响应
} catch (error) {
    if (error.message === 'CONVERSATION_ID_REQUIRED') {
        // 需要创建新会话
        await handleConversationRequired(message);
    } else {
        // 处理其他错误
        console.error('发送消息失败:', error);
    }
}
```

### 2. 自动创建会话
```javascript
async function handleConversationRequired(message) {
    try {
        // 创建新会话
        const conversation = await createConversation('新对话');
        
        // 使用新会话ID重新发送消息
        const response = await sendMessage(conversation.id, message);
        
        // 处理响应
        return response;
    } catch (error) {
        console.error('创建会话失败:', error);
        throw error;
    }
}
```

### 3. 流式消息处理
```javascript
try {
    await sendMessageStream(conversationID, message, (chunk) => {
        // 处理流式响应
        console.log('收到消息片段:', chunk);
    });
} catch (error) {
    if (error.message === 'CONVERSATION_ID_REQUIRED') {
        // 创建新会话并重新发送
        const conversation = await createConversation('新对话');
        await sendMessageStream(conversation.id, message, (chunk) => {
            console.log('收到消息片段:', chunk);
        });
    }
}
```

## 错误码说明

| 错误码 | 说明 | 前端处理 |
|--------|------|----------|
| `CONVERSATION_ID_REQUIRED` | 需要提供conversationID | 调用创建会话接口 |

## 测试验证

运行测试脚本：
```bash
./test_conversation_id_validation.sh
```

## 优势

1. **自动化处理**: 前端可以自动处理会话创建，无需用户手动操作
2. **错误明确**: 提供明确的错误码，便于前端识别和处理
3. **用户体验**: 用户无需关心会话创建过程，直接发送消息即可
4. **向后兼容**: 不影响现有的正常流程

## 注意事项

1. 前端需要确保 `CreateConversation` 接口可用
2. 建议在前端添加重试机制，避免网络问题导致的失败
3. 可以考虑在前端缓存会话ID，避免重复创建
4. 流式消息的错误处理需要特别注意，确保callback被正确调用
