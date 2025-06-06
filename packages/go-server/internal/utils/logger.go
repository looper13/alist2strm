package utils

import (
	"alist2strm/config"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *zap.Logger

// InitLogger 初始化日志
func InitLogger() {
	// 确保日志目录存在
	logDir := filepath.Join(config.GlobalConfig.Log.BaseDir, config.GlobalConfig.Log.AppName)
	os.MkdirAll(filepath.Join(logDir, "info"), 0755)
	os.MkdirAll(filepath.Join(logDir, "error"), 0755)

	// 配置 info 日志
	infoLogger := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, "info/app.log"),
		MaxSize:    config.GlobalConfig.Log.MaxFileSize, // 每个日志文件的最大大小（MB）
		MaxBackups: 30,                                  // 保留的旧日志文件的最大数量
		MaxAge:     config.GlobalConfig.Log.MaxDays,     // 保留旧日志文件的最大天数
		Compress:   true,                                // 是否压缩/归档旧文件
	}

	// 配置 error 日志
	errorLogger := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, "error/error.log"),
		MaxSize:    10,
		MaxBackups: 30,
		MaxAge:     7,
		Compress:   true,
	}

	// 自定义日志编码配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    "func",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,                       // 使用大写带颜色的级别
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"), // 更友好的时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 创建自定义级别的核心
	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.ErrorLevel
	})

	// 配置控制台输出
	consoleDebugging := zapcore.Lock(os.Stdout)
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	// 配置核心，同时输出到文件和控制台
	core := zapcore.NewTee(
		// 普通日志文件输出
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(infoLogger),
			lowPriority,
		),
		// 错误日志文件输出
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(errorLogger),
			highPriority,
		),
		// 控制台输出
		zapcore.NewCore(
			consoleEncoder,
			consoleDebugging,
			zap.InfoLevel,
		),
	)

	// 创建 logger
	Logger = zap.New(core,
		zap.AddCaller(),                   // 添加调用者信息
		zap.AddCallerSkip(1),              // 跳过一层调用堆栈
		zap.AddStacktrace(zap.ErrorLevel), // Error 级别及以上添加堆栈跟踪
	)

	// 替换全局 logger
	zap.ReplaceGlobals(Logger)
}

// Info wrapper for zap.Logger.Info
func Info(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

// Error wrapper for zap.Logger.Error
func Error(msg string, fields ...zap.Field) {
	Logger.Error(msg, fields...)
}

// Debug wrapper for zap.Logger.Debug
func Debug(msg string, fields ...zap.Field) {
	Logger.Debug(msg, fields...)
}

// Warn wrapper for zap.Logger.Warn
func Warn(msg string, fields ...zap.Field) {
	Logger.Warn(msg, fields...)
}

// Fatal wrapper for zap.Logger.Fatal
func Fatal(msg string, fields ...zap.Field) {
	Logger.Fatal(msg, fields...)
}
