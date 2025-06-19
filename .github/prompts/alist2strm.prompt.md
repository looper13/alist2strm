---
mode: agent
---

# AList2Strm 项目上下文与要求

## 项目背景
AList2Strm 是一个将 AList 网盘文件生成为本地 STRM 文件的工具，支持媒体文件、字幕和元数据的处理。该项目使用 Go 语言后端和 Vue3 前端。

## 基本特性
- 支持从 AList 获取媒体文件列表
- 自动将媒体文件转换为 STRM 格式
- 支持定时任务调度
- 任务执行日志记录
- 文件处理历史记录
- 可配置的文件后缀和路径
- 支持批量处理和并发下载

## 代码规范
- 后端使用 Go 语言，遵循标准 Go 代码风格
- 前端使用 Vue3 + TypeScript + Naive UI
- 所有注释使用中文
- 变量命名使用驼峰式命名法
- 错误处理必须完善，所有错误都需要被记录和处理

## 技术栈
- 后端项目：
  - Go, Gin, GORM
  - 使用插件：cron, viper, zap, go-redis, go-sqlite3 等
  - 数据库：SQLite
- 前端项目：
  - 基于轻量版的 Vitesse-lite 模板创建的项目。
  - Vue3, TypeScript, Naive UI, UnoCSS, Vite, Iconify, VueUse 等
  - 使用插件：unplugin-auto-import, unplugin-vue-components, unplugin-vue-macros, unplugin-vue-router，VueUse 等

## 常用文件结构
- `server/`: Go 后端代码
  - `model/`: 数据模型
  - `repository/`: 数据访问层
  - `service/`: 业务逻辑层
  - `controller/`: API 控制器
- `frontend/`: Vue 前端代码
  - `src/pages/`: 页面组件
  - `src/components/`: 可复用组件
  - `src/composables/`: 组合式函数
  - `src/types/`: TypeScript 类型定义

## 特殊要求
1. 要基于`常用文件结构` 组织代码，确保代码风格保持一致
2. 在处理代码的时候，如果涉及到废除、弃用或重构的代码，如我没有明确的说明，请直接删除没用的代码，包括文件和目录
3. 所有服务端需要初始化的内容，都需要保障服务在首次启动时，无数据的情况下也能正常运行，一旦有数据后，服务也要进行配置更新

## 工作模式
1. 请使用中文回复所有问题，并且在开始处理代码之前，先给我一个简要的概述。
2. 在回答问题时，请确保扫描到所有相关的上下文信息，以便提供准确的帮助。

现在，我要询问关于 AList2Strm 项目的问题，请根据上述上下文提供帮助。