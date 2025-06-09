package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/MccRay-s/alist2strm/config"
	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	InfoLogger   *zap.SugaredLogger
	ErrorLogger  *zap.SugaredLogger
	DebugLogger  *zap.SugaredLogger
	WarnLogger   *zap.SugaredLogger
	AccessLogger *zap.SugaredLogger
)

// InitLogger 初始化日志系统
func InitLogger(cfg *config.AppConfig) error {
	if err := ensureLogDirs(cfg.Log.BaseDir, cfg.Log.AppName); err != nil {
		return fmt.Errorf("创建日志目录失败: %w", err)
	}

	InfoLogger = createSugaredLogger(cfg, "info")
	ErrorLogger = createSugaredLogger(cfg, "error")
	DebugLogger = createSugaredLogger(cfg, "debug")
	WarnLogger = createSugaredLogger(cfg, "warn")
	AccessLogger = createSugaredLogger(cfg, "access")

	return nil
}

// ensureLogDirs 确保日志目录存在
func ensureLogDirs(baseDir, appName string) error {
	dirs := []string{"info", "error", "debug", "warn", "access"}
	for _, dir := range dirs {
		logDir := filepath.Join(baseDir, appName, dir)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return err
		}
	}
	return nil
}

// createSugaredLogger 创建糖化日志记录器
func createSugaredLogger(cfg *config.AppConfig, level string) *zap.SugaredLogger {
	logPath := filepath.Join(cfg.Log.BaseDir, cfg.Log.AppName, level, level+".log")

	w := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    cfg.Log.MaxFileSize, // MB
		MaxBackups: cfg.Log.MaxBackups,
		MaxAge:     cfg.Log.MaxDays,
		Compress:   cfg.Log.Compress,
	}

	// 使用自定义编码器，混合字符串和JSON格式
	encoder := &customEncoder{appName: cfg.Log.AppName}

	core := zapcore.NewCore(encoder, zapcore.AddSync(w), getLogLevel(level))
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.ErrorLevel))

	return logger.Sugar()
}

// customEncoder 自定义编码器，实现混合格式
type customEncoder struct {
	appName string
}

// 这些方法是 zapcore.Encoder 接口的要求，但我们在 EncodeEntry 中直接处理所有逻辑
func (e *customEncoder) AddArray(key string, marshaler zapcore.ArrayMarshaler) error   { return nil }
func (e *customEncoder) AddObject(key string, marshaler zapcore.ObjectMarshaler) error { return nil }
func (e *customEncoder) AddBinary(key string, value []byte)                            {}
func (e *customEncoder) AddByteString(key string, value []byte)                        {}
func (e *customEncoder) AddBool(key string, value bool)                                {}
func (e *customEncoder) AddComplex128(key string, value complex128)                    {}
func (e *customEncoder) AddComplex64(key string, value complex64)                      {}
func (e *customEncoder) AddDuration(key string, value time.Duration)                   {}
func (e *customEncoder) AddFloat64(key string, value float64)                          {}
func (e *customEncoder) AddFloat32(key string, value float32)                          {}
func (e *customEncoder) AddInt(key string, value int)                                  {}
func (e *customEncoder) AddInt64(key string, value int64)                              {}
func (e *customEncoder) AddInt32(key string, value int32)                              {}
func (e *customEncoder) AddInt16(key string, value int16)                              {}
func (e *customEncoder) AddInt8(key string, value int8)                                {}
func (e *customEncoder) AddString(key, value string)                                   {}
func (e *customEncoder) AddTime(key string, value time.Time)                           {}
func (e *customEncoder) AddUint(key string, value uint)                                {}
func (e *customEncoder) AddUint64(key string, value uint64)                            {}
func (e *customEncoder) AddUint32(key string, value uint32)                            {}
func (e *customEncoder) AddUint16(key string, value uint16)                            {}
func (e *customEncoder) AddUint8(key string, value uint8)                              {}
func (e *customEncoder) AddUintptr(key string, value uintptr)                          {}
func (e *customEncoder) AddReflected(key string, value interface{}) error              { return nil }
func (e *customEncoder) OpenNamespace(key string)                                      {}
func (e *customEncoder) Clone() zapcore.Encoder                                        { return &customEncoder{appName: e.appName} }

func (e *customEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	// 构建基础日志格式：[时间] [APP] [级别] [文件:行号] 消息
	timeStr := entry.Time.Format("2006-01-02 15:04:05")
	caller := ""
	if entry.Caller.Defined {
		caller = fmt.Sprintf("%s:%d", filepath.Base(entry.Caller.File), entry.Caller.Line)
	}

	// 基础信息字符串格式
	basicInfo := fmt.Sprintf("[%s] [%s] [%s] [%s] %s",
		timeStr,
		e.appName,
		strings.ToUpper(entry.Level.String()),
		caller,
		entry.Message,
	)

	// 如果有额外字段，转换为JSON
	if len(fields) > 0 {
		fieldMap := make(map[string]interface{})
		for _, field := range fields {
			switch field.Type {
			case zapcore.StringType:
				fieldMap[field.Key] = field.String
			case zapcore.Int64Type:
				fieldMap[field.Key] = field.Integer
			case zapcore.BoolType:
				fieldMap[field.Key] = field.Integer == 1
			case zapcore.DurationType:
				fieldMap[field.Key] = time.Duration(field.Integer).String()
			case zapcore.TimeType:
				fieldMap[field.Key] = time.Unix(0, field.Integer).Format(time.RFC3339)
			default:
				fieldMap[field.Key] = field.Interface
			}
		}

		if jsonData, err := json.Marshal(fieldMap); err == nil {
			basicInfo += " " + string(jsonData)
		}
	}

	buf := buffer.NewPool().Get()
	buf.WriteString(basicInfo)
	buf.WriteByte('\n')
	return buf, nil
}

// getLogLevel 获取日志级别
func getLogLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// 便捷方法
func Info(msg string, fields ...interface{}) {
	if InfoLogger != nil {
		InfoLogger.Infow(msg, fields...)
	}
}

func Error(msg string, fields ...interface{}) {
	if ErrorLogger != nil {
		ErrorLogger.Errorw(msg, fields...)
	}
}

func Debug(msg string, fields ...interface{}) {
	if DebugLogger != nil {
		DebugLogger.Debugw(msg, fields...)
	}
}

func Warn(msg string, fields ...interface{}) {
	if WarnLogger != nil {
		WarnLogger.Warnw(msg, fields...)
	}
}

// Sync 同步所有日志
func Sync() {
	if InfoLogger != nil {
		InfoLogger.Sync()
	}
	if ErrorLogger != nil {
		ErrorLogger.Sync()
	}
	if DebugLogger != nil {
		DebugLogger.Sync()
	}
	if WarnLogger != nil {
		WarnLogger.Sync()
	}
	if AccessLogger != nil {
		AccessLogger.Sync()
	}
}
