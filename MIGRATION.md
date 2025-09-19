# 项目迁移说明

## 主要变化

### 前端框架迁移
- **从 Vue 3 迁移到 React 18**
- **从 Element Plus 迁移到 Ant Design**
- **从 Pinia 迁移到 Zustand**
- **从 Vue Router 迁移到 React Router**

### 后端架构简化
- **移除了抽象层**：不再使用 `DatabaseManager` 接口和 `UnifiedDatabaseManager`
- **直接实现**：创建了 `SimpleDatabaseManager` 直接管理三种数据库类型
- **代码分离**：将 MySQL、Redis、ClickHouse 的具体实现分别放在独立文件中

### 左侧菜单更新
- **分类显示**：左侧菜单现在分别显示 MySQL、ClickHouse 和 Redis 三个分类
- **独立管理**：每种数据库类型都有独立的连接管理页面

## 文件结构变化

### 新增文件
- `frontend/src/App.jsx` - React 主应用组件
- `frontend/src/main.jsx` - React 入口文件
- `frontend/src/views/Home.jsx` - 首页组件
- `frontend/src/views/Connections.jsx` - 连接管理组件
- `frontend/src/views/Query.jsx` - 查询组件
- `frontend/src/views/Settings.jsx` - 设置组件
- `frontend/src/stores/connection.js` - Zustand 连接状态管理
- `frontend/src/stores/query.js` - Zustand 查询状态管理
- `database/simple_manager.go` - 简化的数据库管理器
- `database/mysql_impl.go` - MySQL 具体实现
- `database/redis_impl.go` - Redis 具体实现
- `database/clickhouse_impl.go` - ClickHouse 具体实现

### 删除文件
- `frontend/src/App.vue`
- `frontend/src/main.js`
- `frontend/src/views/*.vue`
- `database/manager.go`
- `database/mysql.go`
- `database/redis.go`
- `database/clickhouse.go`

## 技术栈对比

| 组件 | 之前 | 现在 |
|------|------|------|
| 前端框架 | Vue 3 | React 18 |
| UI 组件库 | Element Plus | Ant Design |
| 状态管理 | Pinia | Zustand |
| 路由 | Vue Router | React Router |
| 构建工具 | Vite | Vite |
| 后端架构 | 抽象层 + 统一管理器 | 直接实现 + 简化管理器 |

## 功能保持
- 所有原有功能都保持不变
- 支持 MySQL、Redis、ClickHouse 三种数据库
- 连接管理、查询执行、数据浏览等功能完全一致
- 用户界面和交互体验基本一致

## 开发说明
- 前端使用 React 18 + Ant Design 开发
- 后端使用 Go 直接实现，无抽象层
- 状态管理使用 Zustand，更轻量级
- 代码结构更清晰，维护性更好
