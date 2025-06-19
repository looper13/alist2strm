-- 完整的表结构设计 - 支持 Emby 通知、Telegram 通知和失效检测功能
-- Created: 2025-06-08

-- 配置表
CREATE TABLE `configs` (
	`id` INTEGER PRIMARY KEY,
	`createdAt` DATETIME,
	`updatedAt` DATETIME,
	`name` VARCHAR(255) NOT NULL UNIQUE,
	`code` VARCHAR(255) NOT NULL UNIQUE,
	`value` TEXT NOT NULL
);

-- 用户表
CREATE TABLE `users` (
	`id` INTEGER PRIMARY KEY,
	`createdAt` DATETIME,
	`updatedAt` DATETIME,
	`username` VARCHAR(255) NOT NULL UNIQUE,
	`password` VARCHAR(255) NOT NULL,
	`nickname` VARCHAR(255),
	`status` TEXT NOT NULL DEFAULT 'active',
	`lastLoginAt` DATETIME
);

-- 任务表
CREATE TABLE `tasks` (
    `id` INTEGER PRIMARY KEY,
    `createdAt` DATETIME,
    `updatedAt` DATETIME,
    `name` VARCHAR(255) NOT NULL,
    `mediaType` VARCHAR(50) NOT NULL DEFAULT 'movie',  -- 媒体类型：movie/tv
    `sourcePath` VARCHAR(255) NOT NULL,
    `targetPath` VARCHAR(255) NOT NULL,
    `fileSuffix` VARCHAR(255) NOT NULL,
    `overwrite` TINYINT (1) NOT NULL DEFAULT 0,
    `enabled` TINYINT (1) NOT NULL DEFAULT 1,
    `cron` VARCHAR(255),
    `running` TINYINT (1) NOT NULL DEFAULT 0,
    `lastRunAt` DATETIME,
    `downloadMetadata` TINYINT(1) NOT NULL DEFAULT 0,  -- 是否下载刮削数据
    `downloadSubtitle` TINYINT(1) NOT NULL DEFAULT 0,  -- 是否下载字幕
    `metadataExtensions` VARCHAR(255) DEFAULT '.nfo,.jpg,.png',  -- 刮削数据文件扩展名
    `subtitleExtensions` VARCHAR(255) DEFAULT '.srt,.ass,.ssa'   -- 字幕文件扩展名
);

-- 任务日志表
CREATE TABLE `task_logs` (
	`id` INTEGER PRIMARY KEY,
	`createdAt` DATETIME,
	`updatedAt` DATETIME,
	`taskId` INTEGER NOT NULL,
	`status` VARCHAR(255) NOT NULL,
	`message` TEXT,
	`startTime` DATETIME NOT NULL,
	`endTime` DATETIME,
	`totalFile` INTEGER NOT NULL DEFAULT '0',
	`generatedFile` INTEGER NOT NULL DEFAULT '0',
	`skipFile` INTEGER NOT NULL DEFAULT 0,
	`metadataCount` INTEGER NOT NULL DEFAULT 0,  -- 下载的刮削数据文件数
	`subtitleCount` INTEGER NOT NULL DEFAULT 0    -- 下载的字幕文件数
);

-- 增强的文件历史表 - 支持失效检测、通知和扩展功能
C-- 增强的文件历史表 - 支持失效检测、通知和扩展功能
CREATE TABLE `file_histories` (
    `id` INTEGER PRIMARY KEY,
    `taskId` INTEGER NOT NULL,
    `taskLogId` INTEGER NOT NULL,
    `createdAt` DATETIME,
    `updatedAt` DATETIME,

    -- 文件基本信息
    `fileName` VARCHAR(255) NOT NULL,
    `sourcePath` VARCHAR(255) NOT NULL,                -- AList 源路径
    `targetFilePath` VARCHAR(255) NOT NULL,            -- 目标文件路径（strm文件路径）
    `fileSize` BIGINT NOT NULL,
    `fileType` VARCHAR(255) NOT NULL,
    `modifiedAt` DATETIME NOT NULL,                    -- 文件修改时间
    `fileSuffix` VARCHAR(255) NOT NULL,
    `hash` VARCHAR(64),                                -- 文件哈希值（用于变更检测）
    -- 索引定义
    INDEX `idx_task_id` (`taskId`),
    INDEX `idx_task_log_id` (`taskLogId`),
    INDEX `idx_created_at` (`createdAt`),
    INDEX `idx_source_path` (`sourcePath`),
    INDEX `idx_file_name` (`fileName`)
);


-- 通知任务队列表 - 异步处理通知
CREATE TABLE `notification_queue` (
    `id` INTEGER PRIMARY KEY,
    `createdAt` DATETIME,
    `updatedAt` DATETIME,
    `type` VARCHAR(50) NOT NULL,                       -- 通知类型：wxwork/telegram
    `event` VARCHAR(100) NOT NULL,                     -- 事件类型：task
    `payload` TEXT NOT NULL,                           -- 通知内容（JSON格式）
    `status` VARCHAR(50) NOT NULL DEFAULT 'pending',   -- 状态：pending/processing/completed/failed
    `retryCount` INTEGER NOT NULL DEFAULT 0,           -- 重试次数
    `maxRetries` INTEGER NOT NULL DEFAULT 3,           -- 最大重试次数
    `nextRetryAt` DATETIME,                            -- 下次重试时间
    `processedAt` DATETIME,                            -- 处理完成时间
    `errorMessage` TEXT,                               -- 错误信息
    `priority` INTEGER NOT NULL DEFAULT 5,             -- 优先级（1-10，数字越小优先级越高）
    
    INDEX `idx_status` (`status`),
    INDEX `idx_type` (`type`),
    INDEX `idx_event` (`event`),
    INDEX `idx_next_retry_at` (`nextRetryAt`),
    INDEX `idx_priority` (`priority`),
    INDEX `idx_created_at` (`createdAt`)
);

-- 失效检测任务表 - 管理失效检测任务
CREATE TABLE `validation_tasks` (
    `id` INTEGER PRIMARY KEY,
    `createdAt` DATETIME,
    `updatedAt` DATETIME,
    `type` VARCHAR(50) NOT NULL,                       -- 检测类型：full/incremental/manual
    `status` VARCHAR(50) NOT NULL DEFAULT 'pending',   -- 状态：pending/running/completed/failed
    `startTime` DATETIME,                              -- 开始时间
    `endTime` DATETIME,                                -- 结束时间
    `totalFiles` INTEGER NOT NULL DEFAULT 0,           -- 总文件数
    `checkedFiles` INTEGER NOT NULL DEFAULT 0,         -- 已检查文件数
    `invalidFiles` INTEGER NOT NULL DEFAULT 0,         -- 失效文件数
    `errorFiles` INTEGER NOT NULL DEFAULT 0,           -- 检查出错文件数
    `progress` INTEGER NOT NULL DEFAULT 0,             -- 进度百分比
    `message` TEXT,                                    -- 任务消息
    `config` TEXT,                                     -- 任务配置（JSON格式）
    
    INDEX `idx_status` (`status`),
    INDEX `idx_type` (`type`),
    INDEX `idx_created_at` (`createdAt`)
);

-- 系统日志表 - 记录系统级别的操作日志
CREATE TABLE `system_logs` (
    `id` INTEGER PRIMARY KEY,
    `createdAt` DATETIME,
    `level` VARCHAR(20) NOT NULL,                      -- 日志级别：debug/info/warn/error
    `module` VARCHAR(100) NOT NULL,                    -- 模块名称：notification/validation/file_service
    `operation` VARCHAR(100) NOT NULL,                 -- 操作名称
    `message` TEXT NOT NULL,                           -- 日志消息
    `data` TEXT,                                       -- 相关数据（JSON格式）
    `userId` INTEGER,                                  -- 用户ID（如果是用户操作）
    `ip` VARCHAR(45),                                  -- IP地址
    `userAgent` VARCHAR(500),                          -- User Agent
    
    INDEX `idx_level` (`level`),
    INDEX `idx_module` (`module`),
    INDEX `idx_operation` (`operation`),
    INDEX `idx_created_at` (`createdAt`),
    INDEX `idx_user_id` (`userId`)
);

-- 默认配置数据插入
-- AList 配置
INSERT INTO `configs` (`name`, `code`, `value`, `createdAt`, `updatedAt`) VALUES 
('AList 配置', 'ALIST', '{"host":"http://localhost:5244","username":"admin","password":"","token":"","timeout":30,"retryTimes":3}', datetime('now'), datetime('now'));

-- Emby 配置
INSERT INTO `configs` (`name`, `code`, `value`, `createdAt`, `updatedAt`) VALUES 
('Emby 配置', 'EMBY', '{"enabled":false,"serverUrl":"","apiKey":"","userId":"","timeout":30,"autoRefreshLibrary":true,"notificationEvents":["task_completed","file_invalid"]}', datetime('now'), datetime('now'));

-- Telegram 配置
INSERT INTO `configs` (`name`, `code`, `value`, `createdAt`, `updatedAt`) VALUES 
('Telegram 配置', 'TELEGRAM', '{"enabled":false,"botToken":"","chatId":"","timeout":30,"templates":{"task_completed":"📊 任务执行完成\\n🎬 任务：{{.TaskName}}\\n📁 路径：{{.SourcePath}}\\n✅ 成功：{{.SuccessCount}} 个文件\\n❌ 失败：{{.FailedCount}} 个文件\\n⏩ 跳过：{{.SkippedCount}} 个文件\\n🕒 用时：{{.Duration}}","task_failed":"❌ 任务执行失败\\n🎬 任务：{{.TaskName}}\\n📁 路径：{{.SourcePath}}\\n💥 错误：{{.ErrorMessage}}","file_invalid":"⚠️ 文件失效检测\\n📁 共检测：{{.TotalFiles}} 个文件\\n❌ 失效文件：{{.InvalidFiles}} 个\\n🔗 主要原因：{{.MainReason}}"}}', datetime('now'), datetime('now'));

-- 失效检测配置
INSERT INTO `configs` (`name`, `code`, `value`, `createdAt`, `updatedAt`) VALUES 
('失效检测配置', 'VALIDATION', '{"enabled":true,"fullScanCron":"0 2 * * *","incrementalInterval":"1h","batchSize":100,"timeout":10,"retryTimes":3,"retryInterval":"5m","checkMethods":["http_head","file_exists"],"invalidThreshold":3}', datetime('now'), datetime('now'));

-- 通知配置
INSERT INTO `configs` (`name`, `code`, `value`, `createdAt`, `updatedAt`) VALUES 
('通知配置', 'NOTIFICATION', '{"enabled":true,"batchSize":50,"retryTimes":3,"retryInterval":"5m","queueProcessInterval":"30s","events":{"task_completed":{"emby":true,"telegram":true},"task_failed":{"emby":false,"telegram":true},"file_invalid":{"emby":true,"telegram":true}}}', datetime('now'), datetime('now'));

-- 生成器配置
INSERT INTO `configs` (`name`, `code`, `value`, `createdAt`, `updatedAt`) VALUES 
('生成器配置', 'GENERATOR', '{"replaceSuffix":"N","urlEncode":"N","replaceHost":"","concurrent":10}', datetime('now'), datetime('now'));

-- 系统配置
INSERT INTO `configs` (`name`, `code`, `value`, `createdAt`, `updatedAt`) VALUES 
('系统配置', 'SYSTEM', '{"logLevel":"info","logRetentionDays":30,"dbCleanupDays":90,"maxConcurrentTasks":5,"apiTimeout":30}', datetime('now'), datetime('now'));
