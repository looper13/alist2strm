# 数据库配置和使用指南

## 概述

alist2strm 支持两种数据库类型：
- **SQLite**: 轻量级文件数据库，适合单用户或小规模部署
- **PostgreSQL**: 企业级关系数据库，适合多用户或高并发场景

## 数据库配置

### 1. 配置方式

数据库配置可以通过以下方式设置：
1. 环境变量（推荐）
2. 配置文件
3. 默认配置

### 2. SQLite 配置

SQLite 是默认的数据库类型，配置简单：

```bash
# 环境变量配置
DB_TYPE=sqlite
DB_BASE_DIR=../data/db
DB_NAME=go_database.sqlite
```

**配置参数说明：**
- `DB_TYPE`: 设置为 `sqlite`
- `DB_BASE_DIR`: 数据库文件存储目录
- `DB_NAME`: 数据库文件名

### 3. PostgreSQL 配置

PostgreSQL 提供更强大的功能和性能：

```bash
# 基本连接配置
DB_TYPE=postgresql
DB_HOST=localhost
DB_PORT=5432
DB_DATABASE=alist2strm
DB_USERNAME=postgres
DB_PASSWORD=your_password_here
DB_SSL_MODE=disable

# 连接池配置（可选）
DB_MAX_OPEN_CONNS=25      # 最大打开连接数
DB_MAX_IDLE_CONNS=5       # 最大空闲连接数
DB_CONN_MAX_LIFETIME=300  # 连接最大生存时间（秒）
```

**配置参数说明：**
- `DB_TYPE`: 设置为 `postgresql`
- `DB_HOST`: PostgreSQL 服务器地址
- `DB_PORT`: PostgreSQL 服务器端口（默认 5432）
- `DB_DATABASE`: 数据库名称
- `DB_USERNAME`: 数据库用户名
- `DB_PASSWORD`: 数据库密码
- `DB_SSL_MODE`: SSL 连接模式（disable/require/verify-ca/verify-full）

**连接池参数：**
- `DB_MAX_OPEN_CONNS`: 最大同时打开的连接数
- `DB_MAX_IDLE_CONNS`: 连接池中保持的空闲连接数
- `DB_CONN_MAX_LIFETIME`: 连接的最大生存时间

## 数据库类型切换

### 从 SQLite 切换到 PostgreSQL

1. **准备 PostgreSQL 环境**
   ```bash
   # 使用 Docker 快速启动 PostgreSQL
   docker run --name postgres-alist2strm \
     -e POSTGRES_DB=alist2strm \
     -e POSTGRES_USER=postgres \
     -e POSTGRES_PASSWORD=your_password \
     -p 5432:5432 \
     -d postgres:13
   ```

2. **更新配置**
   ```bash
   # 修改环境变量或 .env 文件
   DB_TYPE=postgresql
   DB_HOST=localhost
   DB_PORT=5432
   DB_DATABASE=alist2strm
   DB_USERNAME=postgres
   DB_PASSWORD=your_password
   DB_SSL_MODE=disable
   ```

3. **重启应用**
   ```bash
   # 停止应用
   # 启动应用，系统会自动创建表结构
   ```

### 从 PostgreSQL 切换到 SQLite

1. **更新配置**
   ```bash
   DB_TYPE=sqlite
   DB_BASE_DIR=../data/db
   DB_NAME=go_database.sqlite
   ```

2. **重启应用**

**注意：** 数据库类型切换不会自动迁移数据，需要手动处理数据迁移。

## PostgreSQL 部署最佳实践

### 1. 生产环境部署

#### Docker Compose 部署
```yaml
version: '3.8'
services:
  postgres:
    image: postgres:13
    environment:
      POSTGRES_DB: alist2strm
      POSTGRES_USER: alist2strm_user
      POSTGRES_PASSWORD: secure_password_here
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./init.sql:/docker-entrypoint-initdb.d/init.sql
    ports:
      - "5432:5432"
    restart: unless-stopped
    
  app:
    build: .
    environment:
      DB_TYPE: postgresql
      DB_HOST: postgres
      DB_PORT: 5432
      DB_DATABASE: alist2strm
      DB_USERNAME: alist2strm_user
      DB_PASSWORD: secure_password_here
      DB_SSL_MODE: disable
    depends_on:
      - postgres
    restart: unless-stopped

volumes:
  postgres_data:
```

#### 独立 PostgreSQL 服务器
```bash
# 1. 安装 PostgreSQL
sudo apt-get install postgresql postgresql-contrib

# 2. 创建数据库和用户
sudo -u postgres psql
CREATE DATABASE alist2strm;
CREATE USER alist2strm_user WITH PASSWORD 'secure_password';
GRANT ALL PRIVILEGES ON DATABASE alist2strm TO alist2strm_user;
\q

# 3. 配置应用连接
DB_TYPE=postgresql
DB_HOST=your_postgres_server
DB_PORT=5432
DB_DATABASE=alist2strm
DB_USERNAME=alist2strm_user
DB_PASSWORD=secure_password
DB_SSL_MODE=require
```

### 2. 性能优化配置

#### 连接池配置建议
```bash
# 根据服务器资源调整
DB_MAX_OPEN_CONNS=25      # CPU 核心数 * 2-4
DB_MAX_IDLE_CONNS=5       # MAX_OPEN_CONNS 的 20%
DB_CONN_MAX_LIFETIME=300  # 5分钟
```

#### PostgreSQL 服务器配置优化
```sql
-- postgresql.conf 关键配置
max_connections = 100
shared_buffers = 256MB
effective_cache_size = 1GB
work_mem = 4MB
maintenance_work_mem = 64MB
checkpoint_completion_target = 0.9
wal_buffers = 16MB
default_statistics_target = 100
```

### 3. 安全配置

#### SSL 连接配置
```bash
# 生产环境建议启用 SSL
DB_SSL_MODE=require  # 或 verify-ca, verify-full
```

#### 网络安全
```bash
# pg_hba.conf 配置示例
# 仅允许特定 IP 连接
host    alist2strm    alist2strm_user    10.0.0.0/8    md5
host    alist2strm    alist2strm_user    192.168.0.0/16    md5
```

### 4. 备份和恢复

#### 自动备份脚本
```bash
#!/bin/bash
# backup.sh
BACKUP_DIR="/backup/postgresql"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/alist2strm_$DATE.sql"

mkdir -p $BACKUP_DIR
pg_dump -h localhost -U alist2strm_user -d alist2strm > $BACKUP_FILE
gzip $BACKUP_FILE

# 保留最近 7 天的备份
find $BACKUP_DIR -name "*.gz" -mtime +7 -delete
```

#### 恢复数据
```bash
# 从备份恢复
gunzip -c backup_file.sql.gz | psql -h localhost -U alist2strm_user -d alist2strm
```

## 监控和维护

### 1. 连接监控

应用会自动记录数据库连接状态，可以通过日志查看：
```bash
# 查看数据库连接日志
tail -f logs/app.log | grep "database"
```

### 2. 性能监控

#### PostgreSQL 查询监控
```sql
-- 查看活跃连接
SELECT * FROM pg_stat_activity WHERE state = 'active';

-- 查看慢查询
SELECT query, mean_time, calls 
FROM pg_stat_statements 
ORDER BY mean_time DESC 
LIMIT 10;
```

#### 连接池监控
应用启动时会显示连接池配置信息，运行时可以通过健康检查接口查看连接状态。

### 3. 故障排除

#### 常见问题

1. **连接被拒绝**
   ```
   错误: connection refused
   解决: 检查 PostgreSQL 服务是否启动，端口是否正确
   ```

2. **认证失败**
   ```
   错误: authentication failed
   解决: 检查用户名、密码和 pg_hba.conf 配置
   ```

3. **数据库不存在**
   ```
   错误: database does not exist
   解决: 创建数据库或检查数据库名称配置
   ```

4. **连接池耗尽**
   ```
   错误: connection pool exhausted
   解决: 增加 MAX_OPEN_CONNS 或检查连接泄漏
   ```

#### 调试模式

启用详细日志记录：
```bash
LOG_LEVEL=debug
```

这将输出详细的数据库操作日志，帮助诊断问题。

## 开发环境配置

### 快速开始

1. **使用 Docker 启动 PostgreSQL**
   ```bash
   docker run --name dev-postgres \
     -e POSTGRES_DB=alist2strm_dev \
     -e POSTGRES_USER=dev \
     -e POSTGRES_PASSWORD=dev123 \
     -p 5432:5432 \
     -d postgres:13
   ```

2. **配置开发环境变量**
   ```bash
   DB_TYPE=postgresql
   DB_HOST=localhost
   DB_PORT=5432
   DB_DATABASE=alist2strm_dev
   DB_USERNAME=dev
   DB_PASSWORD=dev123
   DB_SSL_MODE=disable
   ```

3. **启动应用**
   ```bash
   go run main.go
   ```

### 测试环境

测试时可以使用内存数据库或临时 PostgreSQL 实例：
```bash
# 使用 SQLite 进行快速测试
DB_TYPE=sqlite
DB_BASE_DIR=/tmp
DB_NAME=test.sqlite
```

## 迁移和升级

### 数据库迁移

应用使用 GORM 的 AutoMigrate 功能自动处理表结构变更：
- 新表会自动创建
- 新列会自动添加
- 索引会自动创建
- 不会删除现有数据

### 版本升级

升级应用版本时：
1. 备份数据库
2. 停止应用
3. 更新应用代码
4. 启动应用（自动执行迁移）
5. 验证功能正常

如果迁移失败，应用会记录详细错误信息并停止启动，确保数据安全。

## 性能监控和优化

### 1. 性能监控配置

应用内置了完整的数据库性能监控系统，支持以下功能：

#### 基本配置
```bash
# 启用性能监控
DB_ENABLE_PERFORMANCE_LOG=true

# 慢查询阈值（毫秒）
DB_SLOW_QUERY_THRESHOLD=100

# 性能报告间隔（分钟）
DB_PERFORMANCE_REPORT_INTERVAL=5
```

#### 监控指标

**连接池指标：**
- 最大连接数和当前连接数
- 使用中连接数和空闲连接数
- 连接等待次数和等待时间
- 连接使用率和平均等待时间

**查询性能指标：**
- 总查询数和慢查询数
- 平均查询时间、最大/最小查询时间
- 每秒查询数（QPS）
- 慢查询阈值

**慢查询记录：**
- SQL 语句内容
- 执行时间和时间戳
- 影响行数

### 2. 性能监控功能

#### 自动性能报告
```bash
# 应用会定期输出性能报告
2025/01/24 10:00:00 数据库性能报告:
{
  "connection_pool": {
    "max_open_connections": 25,
    "open_connections": 8,
    "in_use": 3,
    "idle": 5,
    "connection_utilization": 32.0,
    "average_wait_time": "2ms"
  },
  "query_performance": {
    "total_queries": 1250,
    "slow_queries": 15,
    "average_query_time": "25ms",
    "queries_per_second": 4.2
  }
}
```

#### 优化建议
系统会自动分析性能指标并提供优化建议：
```bash
2025/01/24 10:00:00 数据库优化建议:
  1. 连接池使用率过高(85.2%)，建议增加最大连接数
  2. 慢查询比例过高(8.5%)，建议优化SQL语句或添加索引
  3. 平均查询时间过长(120ms)，建议检查查询效率
```

#### 健康状态监控
- **HEALTHY**: 所有指标正常
- **WARNING**: 存在性能问题但不严重
- **CRITICAL**: 存在严重性能问题

### 3. 性能优化建议

#### 连接池优化
```bash
# 根据应用负载调整连接池大小
# 高并发场景
DB_MAX_OPEN_CONNS=50
DB_MAX_IDLE_CONNS=10

# 低并发场景
DB_MAX_OPEN_CONNS=10
DB_MAX_IDLE_CONNS=2

# 连接生存时间（避免长时间空闲连接）
DB_CONN_MAX_LIFETIME=30  # 30分钟
```

#### 慢查询优化
1. **降低慢查询阈值进行更精确监控**
   ```bash
   DB_SLOW_QUERY_THRESHOLD=50  # 50ms
   ```

2. **分析慢查询日志**
   ```bash
   # 查看应用日志中的慢查询记录
   grep "慢查询检测" logs/app.log
   ```

3. **数据库索引优化**
   ```sql
   -- 为常用查询字段添加索引
   CREATE INDEX idx_user_email ON users(email);
   CREATE INDEX idx_task_status ON tasks(status);
   ```

#### PostgreSQL 服务器优化
```sql
-- postgresql.conf 性能优化配置
shared_buffers = 256MB          -- 共享缓冲区
effective_cache_size = 1GB      -- 有效缓存大小
work_mem = 4MB                  -- 工作内存
maintenance_work_mem = 64MB     -- 维护工作内存
checkpoint_completion_target = 0.9
random_page_cost = 1.1          -- SSD 存储优化
```

### 4. 监控最佳实践

#### 生产环境监控
1. **启用性能日志**
   ```bash
   DB_ENABLE_PERFORMANCE_LOG=true
   DB_PERFORMANCE_REPORT_INTERVAL=10  # 10分钟报告一次
   ```

2. **设置合理的慢查询阈值**
   ```bash
   # 根据业务需求设置
   DB_SLOW_QUERY_THRESHOLD=100  # 100ms
   ```

3. **监控关键指标**
   - 连接池使用率 < 80%
   - 慢查询比例 < 5%
   - 平均查询时间 < 50ms
   - 连接等待时间 < 100ms

#### 开发环境监控
```bash
# 开发环境可以设置更严格的阈值
DB_SLOW_QUERY_THRESHOLD=50   # 50ms
DB_PERFORMANCE_REPORT_INTERVAL=1  # 1分钟
```

#### 故障排除
1. **连接池耗尽**
   - 检查连接泄漏
   - 增加最大连接数
   - 减少连接生存时间

2. **查询性能差**
   - 分析慢查询日志
   - 检查数据库索引
   - 优化 SQL 语句

3. **内存使用过高**
   - 调整 work_mem 配置
   - 优化查询复杂度
   - 检查连接数配置

### 5. 性能测试

#### 基准测试
```bash
# 使用 pgbench 测试 PostgreSQL 性能
pgbench -i -s 10 alist2strm  # 初始化测试数据
pgbench -c 10 -j 2 -t 1000 alist2strm  # 运行性能测试
```

#### 应用性能测试
```bash
# 启用详细性能监控
DB_ENABLE_PERFORMANCE_LOG=true
DB_SLOW_QUERY_THRESHOLD=10  # 10ms 严格阈值

# 运行应用并观察性能指标
go run main.go
```

通过这些监控和优化措施，可以确保数据库在各种负载下都能保持良好的性能表现。