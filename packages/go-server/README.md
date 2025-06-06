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
