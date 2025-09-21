# 会话管理功能说明

## 新增功能

### 1. 后端重构
- 创建了 `backend/` 目录，将Go后端代码重新组织
- 在 `backend/sqlite/` 目录下实现了SQLite数据库管理
- 支持会话和消息的持久化存储

### 2. 数据库结构
- **conversations表**: 存储会话信息
  - `id`: 会话唯一标识
  - `title`: 会话标题
  - `created_at`: 创建时间
  - `updated_at`: 更新时间
  - `message_count`: 消息数量

- **messages表**: 存储消息信息
  - `id`: 消息唯一标识
  - `conversation_id`: 所属会话ID
  - `role`: 角色（user/assistant）
  - `content`: 消息内容
  - `created_at`: 创建时间
  - `is_tool_result`: 是否为工具执行结果

### 3. 前端功能
- **历史会话管理**: 可以查看、创建、删除历史会话
- **会话切换**: 可以在不同会话间切换
- **消息持久化**: 所有对话消息都会保存到数据库
- **会话标题编辑**: 可以修改会话标题
- **新建会话**: 支持创建新的对话会话

### 4. API接口
- `CreateConversation(title)`: 创建新会话
- `ListConversations()`: 获取会话列表
- `GetConversation(id)`: 获取指定会话
- `DeleteConversation(id)`: 删除会话
- `UpdateConversation(conversation)`: 更新会话
- `GetConversationWithMessages(id)`: 获取会话及消息
- `AddMessageToConversation(conversationId, role, content, isToolResult)`: 添加消息
- `GetMessages(conversationId)`: 获取会话消息
- `SendMessageToConversation(conversationId, message)`: 发送消息到指定会话

## 使用方法

### 1. 创建新会话
- 点击"新建会话"按钮
- 输入会话标题
- 开始对话

### 2. 查看历史会话
- 点击"历史会话"按钮
- 在左侧抽屉中查看所有会话
- 点击会话标题切换到该会话

### 3. 管理会话
- 编辑会话标题：点击编辑按钮
- 删除会话：点击删除按钮并确认
- 查看会话详情：点击会话标题

### 4. 对话功能
- 所有对话都会自动保存到当前会话
- 支持工具调用确认
- 支持流式响应

## 技术实现

### 后端
- 使用SQLite作为本地数据库
- 数据库文件存储在 `~/.db-desktop/conversations.db`
- 支持事务和索引优化

### 前端
- 使用Zustand进行状态管理
- 支持实时更新会话列表
- 响应式设计，支持移动端

## 注意事项

1. 数据库文件会自动创建在用户主目录的 `.db-desktop` 文件夹中
2. 删除会话会同时删除该会话下的所有消息
3. 会话标题支持中文和特殊字符
4. 消息内容支持Markdown格式
5. 工具执行结果会标记为特殊消息类型

## 故障排除

如果遇到数据库相关错误：
1. 检查 `~/.db-desktop` 目录权限
2. 确保SQLite驱动已正确安装
3. 查看应用日志获取详细错误信息

如果前端显示异常：
1. 检查浏览器控制台错误
2. 确认Wails绑定文件已正确生成
3. 重启应用重新加载绑定
