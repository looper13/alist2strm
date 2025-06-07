
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

-- 文件历史表
CREATE TABLE `file_histories` (
    `id` INTEGER PRIMARY KEY,
    `taskId` INTEGER NOT NULL,
    `taskLogId` INTEGER NOT NULL,
    `createdAt` DATETIME,
    `updatedAt` DATETIME,
    `fileName` VARCHAR(255) NOT NULL,
    `sourcePath` VARCHAR(255) NOT NULL,
    `targetFilePath` VARCHAR(255) NOT NULL,
    `fileSize` BIGINT NOT NULL,
    `fileType` VARCHAR(255) NOT NULL,
    `fileSuffix` VARCHAR(255) NOT NULL,
    `isMainFile` TINYINT(1) NOT NULL DEFAULT 1,  -- 是否为主文件（视频文件）
    `mainFileId` INTEGER,  -- 关联的主文件ID（如果是刮削数据或字幕文件）
    `fileCategory` VARCHAR(50)  -- 文件类别：main/metadata/subtitle
);

