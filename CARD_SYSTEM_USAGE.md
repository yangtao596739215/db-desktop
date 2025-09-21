# 确认卡片系统使用说明

## 概述

确认卡片系统用于在AI助手执行MCP工具前，向用户展示确认弹窗。系统包含以下核心组件：

- `ConfirmCard`: 确认卡片结构体
- `CardManager`: 卡片管理器
- 集成到AI服务中的自动卡片创建

## 核心结构

### ConfirmCard 结构体

```go
type ConfirmCard struct {
    CardID          string    `json:"cardId"`          // UUID生成的卡片ID
    ShowContent     string    `json:"showContent"`     // 卡片中展示的内容
    ConfirmCallback func()    `json:"-"`               // 确认时执行的回调函数
    RejectCallback  func()    `json:"-"`               // 拒绝时执行的回调函数
    CreatedAt       time.Time `json:"createdAt"`       // 创建时间
    Status          string    `json:"status"`          // 状态：pending, confirmed, rejected, expired
    ExpiresAt       time.Time `json:"expiresAt"`       // 过期时间（5分钟后）
}
```

## 前端可调用的方法

### 1. 获取待确认的卡片列表

```javascript
// 获取所有待确认的卡片
const pendingCards = await window.go.main.App.GetPendingConfirmCards();
console.log('待确认卡片:', pendingCards);
```

### 2. 确认卡片

```javascript
// 确认指定ID的卡片
const cardId = "123e4567-e89b-12d3-a456-426614174000";
await window.go.main.App.ConfirmCard(cardId);
```

### 3. 拒绝卡片

```javascript
// 拒绝指定ID的卡片
const cardId = "123e4567-e89b-12d3-a456-426614174000";
await window.go.main.App.RejectCard(cardId);
```

### 4. 获取卡片统计信息

```javascript
// 获取卡片统计信息
const stats = await window.go.main.App.GetCardStats();
console.log('卡片统计:', stats);
// 输出示例: { total: 3, pending: 2, confirmed: 1, rejected: 0, expired: 0 }
```

### 5. 获取特定卡片

```javascript
// 根据ID获取特定卡片
const card = await window.go.main.App.GetConfirmCard(cardId);
if (card) {
    console.log('卡片详情:', card);
}
```

## 工作流程

1. **AI请求MCP工具**: 当AI助手需要执行数据库操作时，系统会自动创建确认卡片
2. **前端获取卡片**: 前端调用 `GetPendingConfirmCards()` 获取待确认的卡片列表
3. **用户确认**: 用户查看卡片内容，点击确认或拒绝按钮
4. **执行回调**: 系统根据用户选择执行相应的回调函数
5. **自动清理**: 卡片在5分钟后自动过期，系统会定期清理过期卡片

## 示例场景

### 场景1: 执行MySQL查询

当用户询问"查询用户表的所有数据"时：

1. AI服务检测到需要执行 `execute_mysql_query` 工具
2. 自动创建确认卡片，内容为："执行MySQL查询: `SELECT * FROM users`"
3. 前端显示确认弹窗
4. 用户点击确认后，系统执行查询并返回结果

### 场景2: 执行Redis命令

当用户询问"获取所有键"时：

1. AI服务检测到需要执行 `execute_redis_command` 工具
2. 自动创建确认卡片，内容为："执行Redis命令: `KEYS *`"
3. 前端显示确认弹窗
4. 用户可以选择确认或拒绝

## 卡片状态说明

- `pending`: 等待用户确认
- `confirmed`: 用户已确认，正在执行
- `rejected`: 用户已拒绝
- `expired`: 卡片已过期（5分钟后）

## 安全特性

1. **自动过期**: 卡片在5分钟后自动过期，防止内存泄漏
2. **线程安全**: 所有卡片操作都是线程安全的
3. **错误处理**: 回调函数执行时的panic会被捕获并记录
4. **状态管理**: 防止重复处理同一张卡片

## 集成说明

确认卡片系统已经集成到现有的AI服务中，无需额外配置。当AI服务检测到需要执行MCP工具时，会自动创建相应的确认卡片。

前端只需要调用相应的API方法来获取卡片列表和处理用户确认即可。
