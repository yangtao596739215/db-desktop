# DB Desktop

一个现代化的数据库管理工具，基于 Wails3 + Vue3 + Element Plus 构建，支持 MySQL、Redis 和 ClickHouse 数据库。

## 功能特性

### 🗄️ 多数据库支持
- **MySQL**: 完整的 SQL 查询支持，表结构浏览，数据管理
- **Redis**: 键值对操作，数据类型支持（String、List、Set、ZSet、Hash）
- **ClickHouse**: 高性能分析数据库支持

### 🔗 连接管理
- 可视化连接配置界面
- 连接测试和状态监控
- 连接配置的导入/导出
- 支持 SSL 连接配置

### 💻 SQL 查询
- 现代化的 SQL 编辑器
- 语法高亮和自动补全
- 查询历史记录
- 结果集分页显示
- 查询性能统计

### 📊 数据浏览
- 数据库结构树形展示
- 表结构详细信息
- 数据预览和筛选
- 索引信息查看

### 🎨 用户界面
- 现代化的 Material Design 风格
- 深色/浅色主题切换
- 响应式设计，支持不同屏幕尺寸
- 直观的操作流程

### ⚙️ 开发辅助功能
- SQL 查询格式化
- 查询历史管理
- 连接配置备份
- 自定义设置选项

## 技术栈

### 后端
- **Go**: 主要编程语言
- **Wails3**: 跨平台桌面应用框架
- **MySQL Driver**: go-sql-driver/mysql
- **Redis Client**: go-redis/v9
- **ClickHouse Driver**: ClickHouse/clickhouse-go/v2

### 前端
- **Vue 3**: 渐进式 JavaScript 框架
- **Element Plus**: Vue 3 UI 组件库
- **Pinia**: Vue 状态管理
- **Vue Router**: 路由管理
- **Vite**: 构建工具

## 安装和使用

### 系统要求
- macOS 10.15+ (Intel/Apple Silicon)
- Windows 10+ (计划支持)
- Linux (计划支持)

### 下载安装
1. 从 [Releases](https://github.com/your-username/db-desktop/releases) 页面下载最新版本
2. 解压并运行应用程序

### 开发环境搭建
```bash
# 克隆项目
git clone https://github.com/your-username/db-desktop.git
cd db-desktop

# 安装 Go 依赖
go mod tidy

# 安装前端依赖
cd frontend
npm install

# 开发模式运行
cd ..
wails dev

# 构建应用
wails build
```

## 使用指南

### 1. 添加数据库连接
1. 点击侧边栏的"连接管理"
2. 点击"添加连接"按钮
3. 填写连接信息：
   - 连接名称：便于识别的名称
   - 数据库类型：选择 MySQL、Redis 或 ClickHouse
   - 主机地址：数据库服务器地址
   - 端口：数据库端口
   - 用户名/密码：认证信息
   - 数据库名：目标数据库（Redis 不需要）
4. 点击"测试连接"验证配置
5. 保存连接

### 2. 执行 SQL 查询
1. 点击侧边栏的"SQL 查询"
2. 选择已连接的数据库
3. 在编辑器中输入 SQL 语句
4. 点击"执行"按钮或使用 Ctrl+Enter 快捷键
5. 查看查询结果

### 3. 浏览数据库结构
1. 在查询页面左侧的数据库浏览器中
2. 选择数据库查看表列表
3. 点击表名查看表结构和数据
4. 支持分页浏览大量数据

### 4. 管理查询历史
1. 所有执行的查询都会自动保存到历史记录
2. 在设置页面可以清空或导出查询历史
3. 支持查询历史的重用和分享

## 配置说明

### 连接配置
- **MySQL**: 支持标准连接参数，包括 SSL 模式配置
- **Redis**: 支持密码认证，可选择数据库编号 (0-15)
- **ClickHouse**: 支持 HTTP 和 Native 协议

### 应用设置
- **主题**: 浅色/深色/跟随系统
- **语言**: 简体中文/English
- **查询超时**: 5-300 秒
- **结果限制**: 默认查询结果行数限制
- **编辑器**: 自动补全、语法高亮、行号显示等

## 安全说明

- 所有连接密码都经过加密存储
- 支持本地配置文件的导入/导出
- 查询历史可以随时清空
- 不收集任何用户数据

## 开发计划

### 即将推出
- [ ] 数据导入/导出功能
- [ ] 数据库备份和恢复
- [ ] 查询性能分析
- [ ] 多标签页支持
- [ ] 插件系统

### 长期计划
- [ ] 团队协作功能
- [ ] 云端配置同步
- [ ] 更多数据库支持 (PostgreSQL, MongoDB)
- [ ] 数据可视化图表

## 贡献指南

欢迎提交 Issue 和 Pull Request！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 致谢

- [Wails](https://wails.io/) - 优秀的 Go 桌面应用框架
- [Vue.js](https://vuejs.org/) - 渐进式 JavaScript 框架
- [Element Plus](https://element-plus.org/) - Vue 3 UI 组件库
- [Go Database Drivers](https://github.com/golang/go/wiki/SQLDrivers) - 各种数据库驱动

---

**DB Desktop** - 让数据库管理更简单、更高效！ 🚀