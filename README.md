# AList2Strm

AList2Strm 是一个用于将 AList 媒体文件转换为 Strm 格式的工具，支持定时任务和批量处理。



## 功能特性

- 🎯 支持从 AList 获取媒体文件列表
- 🔄 自动将媒体文件转换为 Strm 格式
- ⏰ 支持定时任务调度
- 📊 任务执行日志记录
- 🔍 文件处理历史记录
- ⚙️ 可配置的文件后缀和路径
- 🚀 支持批量处理和

## 技术栈

### 后端
- Node.js
- Express.js
- TypeScript
- Sequelize
- node-cron

### 前端
- Vue 3
- TypeScript
- Naive UI
- Vite

## 界面一览

#### 移动端适配
[![移动端适配](./screenshot/screenshot20250524011706@2x.png)](https://github.com/MccRay-s/alist2strm/raw/main/screenshot/screenshot20250524011706@2x.png)

#### 首页
[![任务管理](./screenshot/screenshot20250524011249@2x.png)](https://github.com/MccRay-s/alist2strm/raw/main/screenshot/screenshot20250524011249@2x.png)

#### 任务管理
[![任务管理](./screenshot/screenshot20250524011222@2x.png)](https://github.com/MccRay-s/alist2strm/raw/main/screenshot/screenshot20250524011222@2x.png)

#### 配置管理
[![任务管理](./screenshot/screenshot20250524011243@2x.png)](https://github.com/MccRay-s/alist2strm/raw/main/screenshot/screenshot20250524011243@2x.png)

#### 文件记录
[![文件记录](./screenshot/screenshot20250524011029@2x.png)](https://github.com/MccRay-s/alist2strm/raw/main/screenshot/screenshot20250524011029@2x.png)

## 功能计划
- [x] 用户授权中心，公网暴露的情况确实挺不安全的 `2025-05-24 22:41`
- [ ] Emby 入库通知，strm 实时入库支持
- [ ] strm 失效检测，预估方案应该是每个 strm 都需要一次网络请求来判断是否有效
- [ ] Telgram 消息通知，具体通知规则还没想好


## 项目结构

```
alist2strm/
├── packages/
│   ├── server/          # 后端服务
│   └── frontend/        # 前端应用
├── package.json
└── README.md
```

## 安装说明

#### docker-compose

```yml
networks:
  media_network:
    external: true

services:
  alist2strm:
    image: mccray/alist2strm:latest
    container_name: alist2strm
    restart: unless-stopped
    networks:
      - media_network
    ports:
      - "3456:80"   # 前端访问端口
      - "4567:3210" # 后端API端口
    volumes:
      # 数据挂载目录
      - /share/Docker/data/alist2strm/data:/app/data
      # 媒体目录
      - /share/MediaCenter:/media
    environment:
      - 'PUID=1000'
      - 'PGID=0'
      - 'UMASK=000'
      - 'TZ=Asia/Shanghai'
      # 用户相关
      - 'JWT_SECRET={你的JWT密钥}'
      - 'USER_NAME={管理员账号}'
      - 'USER_PASSWORD={管理员密码}'
```

#### docker run 
```**bash**
docker run -d \
  --name alist2strm \
  --restart unless-stopped \
  -p 3456:80 \
  -p 4567:3210 \
  -v /share/Docker/data/alist2strm/data:/app/data \
  -v /share/MediaCenter:/media \
  -e PUID=1000 \
  -e PGID=0 \
  -e UMASK=000 \
  -e TZ=Asia/Shanghai \
  -e JWT_SECRET={你的JWT密钥} \
  -e USER_NAME={管理员账号} \
  -e USER_PASSWORD={管理员密码} \
  mccray/alist2strm:latest
```

#### 环境变量说明

| 变量名称    | 说明 | 默认值 |
| -------- | ------- |------- |
| PORT  | 后台服务端口    |`3210` |
| LOG_BASE_DIR | 日志目录     |`/app/data/logs`|
| LOG_LEVEL    | 日志级别    |`info`|
| LOG_LEVEL    | 日志级别    |`info`|
| LOG_APP_NAME    | App名称    |`alist2strm`|
| LOG_MAX_DAYS    | 日志保留天数    |`30`|
| LOG_MAX_FILE_SIZE    | 日志单文件大小/M    |`10`|
| DB_BASE_DIR    | 数据库目录    |`/app/data/db`|
| DB_NAME    | 日志级别    |`database.sqlite`|
| JWT_SECRET    | JWT密钥，自行处理   ||
| USER_NAME    | 管理员账号    |`admin`|
| USER_PASSWORD    | 管理员密码，不填随机生成   |见日志内容|



## 任务配置

1. 创建新任务
   - 设置任务名称
   - 配置源路径（AList 路径）
   - 设置目标路径
   - 选择需要处理的文件后缀
   - 配置定时执行计划（Cron 表达式）

2. 任务管理
   - 启用/禁用任务
   - 手动执行任务
   - 查看执行日志
   - 监控任务状态

## 开发说明

### 后端开发
```bash
cd packages/server
npm run dev
```

### 前端开发
```bash
cd packages/frontend
npm run dev
```

### 构建部署
```bash
# 构建前端
cd packages/frontend
npm run build

# 构建后端
cd packages/server
npm run build
```

### 更新日志
V1.0.1: `2025-05-22 23:45`
- 任务界面调整，Table View 改成 Cards View
- 任务编辑取消cron 必填项目
  
V1.0.2: `2025-05-24 01:28`
- 添加参数 `Alist外网地址`，`strm` 内容优先使用外网地址生成
- 添加参数 `strm` 生成选项 `替换扩展名`，开启后文件名称显示为 `xxx_4k.strm`，未开启则显示为 `xxx_4k.mp4.strm`
- 添加参数 `strm` 生成选项 `URL编码`，启用后会对 `strm` 内容进行URL 编码
- 适配移动端界面

V1.0.3: `2025-05-25 23:05`
- fix `strm 替换扩展名` 配置项无效的问题
- fix 配置修改未实时生效
- 添加用户授权相关表结构，以及路由拦截
- 添加用户登录、注册、退出功能
- 其他代码优化

V1.0.4: `2025-05-26 22:20`
- 增强安全性，移除用户注册功能
- 增加容器启动变量，初始管理员账号、密码 `USER_NAME`、`USER_PASSWORD` 
- 个人信息修改和密码修改



### Bugs 
1. AList 原路径，未编码导致API 调用异常，例如：`我的接收/【BD-ISO】`

## 许可证

MIT License 