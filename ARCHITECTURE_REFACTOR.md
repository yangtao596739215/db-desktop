# 架构重构说明

## 重构概述

根据您的要求，我们对后端代码进行了架构重构，将代码按照分层架构重新组织：

## 新的目录结构

```
backend/
├── handler/          # Handler层 - 提供给前端调用的方法声明和定义
│   └── app.go       # 主要的API接口，只包含方法声明和调用
├── integration/      # Integration层 - 封装调用第三方服务的逻辑
│   └── ai.go        # 大模型API调用，包含path、req、resp定义
├── logic/           # Logic层 - 核心业务逻辑
│   ├── ai.go        # AI服务核心逻辑
│   └── card.go      # 卡片管理逻辑
├── database/        # 数据库相关
├── models/          # 数据模型
├── sqlite/          # SQLite相关
└── utils/           # 工具函数
```

## 各层职责

### 1. Handler层 (`backend/handler/`)
- **职责**: 提供给前端调用的方法声明和定义
- **特点**: 
  - 只包含API接口方法
  - 不包含业务逻辑
  - 直接调用Logic层的方法
  - 处理前端请求和响应

### 2. Integration层 (`backend/integration/`)
- **职责**: 封装调用第三方服务的逻辑
- **特点**:
  - 包含大模型的path、req、resp定义
  - 定义复合结构体，不拆分成子结构体
  - 只做API调用，不包含业务逻辑
  - 目前只有AI大模型调用

### 3. Logic层 (`backend/logic/`)
- **职责**: 核心业务逻辑
- **特点**:
  - 调用模型前，基于会话ID查出所有历史消息
  - 将历史消息放到上下文里进行调用
  - 处理工具调用和确认卡片
  - 协调Integration层和Handler层

## 主要改进

### 1. 清晰的职责分离
- Handler层只负责API接口
- Integration层只负责第三方服务调用
- Logic层负责核心业务逻辑

### 2. 复合结构体设计
- 在Integration层定义了完整的AIResponse结构体
- 包含所有必要的字段，便于处理
- 避免了过度拆分导致的复杂性

### 3. 历史消息处理
- Logic层在调用模型前自动获取历史消息
- 将历史消息作为上下文传递给模型
- 支持对话的连续性

### 4. 代码复用
- 各层职责明确，便于维护
- 接口设计清晰，便于测试
- 减少了代码重复

## 文件变更

### 新增文件
- `backend/integration/ai.go` - AI API调用封装
- `backend/logic/ai.go` - AI服务核心逻辑

### 修改文件
- `backend/handler/app.go` - 简化为只包含API接口
- `backend/logic/card.go` - 更新包名
- `main.go` - 更新import路径

### 删除文件
- `backend/logic/ai_service.go` - 逻辑已迁移到ai.go

## 使用方式

重构后的代码使用方式保持不变，前端调用方式没有变化：

```go
// 创建App实例
app := handler.NewApp()

// 发送消息
response, err := app.SendMessageToConversation(conversationID, message)

// 更新AI配置
err := app.UpdateAIConfig(config)
```

## 优势

1. **可维护性**: 各层职责清晰，便于维护和扩展
2. **可测试性**: 每层都可以独立测试
3. **可扩展性**: 新增第三方服务只需在Integration层添加
4. **代码复用**: 逻辑层可以被多个Handler复用
5. **清晰的数据流**: 请求 → Handler → Logic → Integration → 第三方服务

这种架构设计符合分层架构的最佳实践，为后续功能扩展和维护提供了良好的基础。
