# 项目结构：
```
server
├── README.md
├── controller       // 控制器层，处理请求和响应
│   ├── configs.go
│   ├── system_log.go
│   ├── task.go
│   ├── task_log.go
│   ├── file_history.go
│   └── user.go
├── database
│   └── database.go  // 数据库连接初始化
├── config
│   └── config.go    // 读取&获取 evn 配置 服务器配置|日志配置|数据库配置|JWT配置|初始用户配置 数据来源环境变量
├── go.mod
├── go.sum
├── main.go
├── middleware        // 中间件层，处理请求的预处理和后处理
│   └── jwt.go
├── model             // 模型层，定义数据结构
│   ├── configs
│   │   ├── request   // 请求结构体
│   │   │   └── configs.go
│   │   ├── response  // 响应结构体
│   │   │   └── configs.go
│   │   └── configs.go
│   ├── systemlog
│   │   ├── request
│   │   │   └── system_log.go
│   │   ├── response
│   │   │   └── system_log.go
│   │   └── system_log.go
│   ├── task
│   │   ├── request
│   │   │   └── task.go
│   │   ├── response
│   │   │   └── task.go
│   │   └── task.go
│   ├── tasklog
│   │   ├── request
│   │   │   └── task_log.go
│   │   ├── response
│   │   │   └── task_log.go
│   │   └── task_log.go
│   ├── filehistory
│   │   ├── request
│   │   │   └── file_history.go
│   │   ├── response
│   │   │   └── file_history.go
│   │   └── file_history.go
│   ├── user
│   │   ├── request
│   │   │   └── user.go
│   │   ├── response
│   │   │   └── user.go
│   │   └── user.go
│   ├── common      // 公共模型
│   │   ├── request
│   │   │   └── common.go
│   │   └── response
│   │       ├── common.go
│   │       └── response.go
├── service         // 服务层，处理业务逻辑
│   ├── configs_service.go
│   ├── file_history_service.go
│   ├── system_log_service.go
│   ├── task_log_service.go
│   ├── task_service.go
│   └── user_service.go
├── repository        // 仓库层，处理数据存取
│   ├── configs_repository.go
│   ├── file_history_repository.go
│   ├── system_log_repository.go
│   ├── task_log_repository.go
│   ├── task_repository.go
│   └── user_repository.go
├── router.go   // 路由配置
└── utils            // 工具层，提供通用功能
    ├── hash.go
    ├── jwt.go
    ├── validator.go
    └── verify.go
```


# Go Server 配置说明

## 环境变量配置

本项目使用 `.env` 文件进行配置管理。你可以从 `.env.example` 复制并创建自己的 `.env` 文件：

```bash
cp .env.example .env
```

### 配置项说明

#### 服务器配置
- `PORT`: 服务器监听端口（默认：8080）

#### 日志配置
- `LOG_BASE_DIR`: 日志文件基础目录（默认：./data/logs）
- `LOG_APP_NAME`: 应用名称，用于日志文件夹名称（默认：alist-strm）
- `LOG_LEVEL`: 日志级别（debug/info/warn/error）（默认：info）
- `LOG_MAX_DAYS`: 日志文件保留天数（默认：7）
- `LOG_MAX_FILE_SIZE`: 单个日志文件最大大小，单位 MB（默认：10）

#### 数据库配置
- `DB_BASE_DIR`: 数据库文件基础目录（默认：./data/db）
- `DB_NAME`: 数据库文件名（默认：database.sqlite）

#### JWT 配置
- `JWT_SECRET_KEY`: JWT生成密钥
- `JWT_EXPIRES_IN`: JWT过期时间

#### 用户配置
- `USER_NAME`: 默认用户名称
- `USER_PASSWORD`: 默认用户密码（留空随机生成,请在日志文件查看）

### 开发环境

开发环境可以使用 `.env.development` 配置文件：

```bash
cp .env.development .env
```

### 日志目录结构

```
${LOG_BASE_DIR}/${LOG_APP_NAME}/
├── info/
│   └── app.log     # 普通信息日志
└── error/
    └── error.log   # 错误日志
```

### 数据库文件位置

数据库文件将位于：`${DB_BASE_DIR}/${DB_NAME}`

例如：`./data/db/database.sqlite`

## 数据库设计变更

### 新表结构设计（2025-06-08）

为支持 Emby 通知、Telegram 通知和失效检测功能，对数据库结构进行了重大升级。新的表结构文件位于：`sql/new_table.sql`

#### 主要变更内容

##### 1. `file_histories` 表增强
**原有字段保持不变，新增以下功能字段：**

**失效检测相关：**
- `sourceUrl` - 源文件的完整访问URL（包含签名）
- `lastCheckedAt` - 最后检查时间（用于失效检测）
- `isValid` - 文件是否有效（失效检测结果）
- `validationMessage` - 验证失败的原因
- `validationRetryCount` - 验证重试次数
- `nextCheckAt` - 下次检查时间

**处理状态管理：**
- `processingStatus` - 处理状态：success/failed/pending/skipped
- `processingMessage` - 处理消息或错误信息
- `retryCount` - 重试次数
- `lastProcessedAt` - 最后处理时间

**通知相关：**
- `notificationStatus` - 通知状态：0=未发送,1=成功通知,2=失败通知,3=失效通知
- `embyNotified` - Emby 通知状态
- `telegramNotified` - Telegram 通知状态
- `notificationSentAt` - 通知发送时间
- `notificationMessage` - 通知消息内容

**扩展字段：**
- `mediaInfo` - 媒体信息（JSON格式）
- `hash` - 文件哈希值（用于变更检测）
- `metadata` - 扩展元数据（JSON格式）
- `tags` - 标签（逗号分隔）

##### 2. 新增 `notification_queue` 表
异步处理通知任务，避免阻塞主流程：
- 支持 Emby 和 Telegram 通知类型
- 重试机制和优先级管理
- 事件类型：task_completed/task_failed/file_invalid

##### 3. 新增 `validation_tasks` 表
管理失效检测任务的执行状态：
- 检测类型：full/incremental/manual
- 进度跟踪和统计信息
- 检测结果记录

##### 4. 新增 `system_logs` 表
系统级别的操作日志记录：
- 多级别日志：debug/info/warn/error
- 模块化日志：notification/validation/file_service
- 操作追踪和问题诊断

##### 5. 配置管理模块化
**新增配置模块（存储在 `configs` 表）：**

- **ALIST** - AList 连接配置
- **EMBY** - Emby 服务器和通知配置
- **TELEGRAM** - Telegram Bot 和消息模板配置
- **VALIDATION** - 失效检测策略配置
- **NOTIFICATION** - 通知系统全局配置
- **GENERATOR** - 文件生成器配置
- **SYSTEM** - 系统全局配置

## 管理页面模块设计

### 页面结构调整（2025-06-08）

为了更好地支持新功能，对管理页面进行了模块化重构：

#### 1. **仪表盘** (`/admin/`)
- 系统总体状态概览
- 任务执行统计图表
- 文件处理实时概览
- 失效检测摘要信息
- 通知发送统计面板

#### 2. **任务管理** (`/admin/task`)
- 保持原有任务CRUD功能
- 新增任务执行历史详情
- 任务性能统计分析

#### 3. **文件管理** (`/admin/files/`)
**子模块：**
- **生成记录** (`/admin/files/history`) - 增强的文件历史记录
- **失效文件** (`/admin/files/invalid`) - 检测到的失效文件列表
- **文件统计** (`/admin/files/statistics`) - 文件处理统计分析

**功能特性：**
- 文件状态可视化
- 批量操作支持
- 高级筛选和搜索
- 文件关联关系展示

#### 4. **监控中心** (`/admin/monitoring/`)
**子模块：**
- **失效检测** (`/admin/monitoring/validation`) - 检测任务管理和结果
- **通知队列** (`/admin/monitoring/notifications`) - 通知发送状态和队列
- **系统日志** (`/admin/monitoring/logs`) - 系统操作日志

**功能特性：**
- 实时监控面板
- 任务进度跟踪
- 错误诊断工具
- 性能指标分析

#### 5. **系统配置** (`/admin/config/`)
**模块化配置页面：**
- **AList 配置** (`/admin/config/alist`) - 连接设置和认证
- **Emby 配置** (`/admin/config/emby`) - 服务器配置和API设置
- **Telegram 配置** (`/admin/config/telegram`) - Bot配置和消息模板
- **失效检测** (`/admin/config/validation`) - 检测策略和时间配置
- **通知设置** (`/admin/config/notification`) - 全局通知策略
- **生成器设置** (`/admin/config/generator`) - 文件生成参数
- **系统设置** (`/admin/config/system`) - 系统级别配置

**配置功能特性：**
- 表单化配置编辑
- 配置测试和验证
- 配置历史版本管理
- 一键恢复默认设置

### 设计优势

1. **模块化设计** - 每个功能模块职责清晰，便于维护和扩展
2. **用户体验** - 层级化导航，操作路径短且直观
3. **功能完整** - 覆盖所有计划功能：监控、通知、检测
4. **可扩展性** - 预留扩展空间，支持未来功能增加
5. **数据驱动** - 基于新表结构设计，充分利用数据能力

### 实现计划

1. **Phase 1** - 基础表结构迁移和核心服务实现
2. **Phase 2** - 失效检测系统开发
3. **Phase 3** - 通知系统（Emby + Telegram）开发
4. **Phase 4** - 前端页面模块化重构
5. **Phase 5** - 监控和统计功能完善

### 功能计划


