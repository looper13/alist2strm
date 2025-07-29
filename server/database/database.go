package database

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/MccRay-s/alist2strm/config"
	"github.com/MccRay-s/alist2strm/model/configs"
	"github.com/MccRay-s/alist2strm/model/filehistory"
	"github.com/MccRay-s/alist2strm/model/notification"
	"github.com/MccRay-s/alist2strm/model/task"
	"github.com/MccRay-s/alist2strm/model/tasklog"
	"github.com/MccRay-s/alist2strm/model/user"
	"gorm.io/gorm"
)

var (
	DB                 *gorm.DB
	performanceMonitor *PerformanceMonitor
)

// InitDatabase 初始化数据库连接
func InitDatabase(cfg *config.AppConfig) error {
	// 优化数据库配置参数
	optimizeConnectionConfig(&cfg.Database)

	// 创建数据库工厂
	factory := NewDatabaseFactory()

	// 使用工厂创建数据库连接
	db, err := factory.CreateConnection(&cfg.Database)
	if err != nil {
		return fmt.Errorf("创建数据库连接失败: %v", err)
	}

	// 设置全局数据库实例
	DB = db

	// 初始化性能监控器
	// 所有数据库类型都使用 PostgreSQL 配置中的性能监控参数
	slowQueryThreshold := time.Duration(cfg.Database.PostgreSQL.SlowQueryThreshold) * time.Millisecond
	if slowQueryThreshold <= 0 {
		slowQueryThreshold = 100 * time.Millisecond // 默认慢查询阈值100ms
	}
	performanceMonitor = NewPerformanceMonitor(db, slowQueryThreshold)

	// 记录连接成功信息
	switch cfg.Database.Type {
	case "postgresql":
		log.Printf("PostgreSQL数据库连接成功: %s:%d/%s",
			cfg.Database.PostgreSQL.Host,
			cfg.Database.PostgreSQL.Port,
			cfg.Database.PostgreSQL.Database)
	case "sqlite", "":
		dbPath := filepath.Join(cfg.Database.SQLite.BaseDir, cfg.Database.SQLite.Name)
		log.Printf("SQLite数据库连接成功: %s", dbPath)
	}

	// 执行数据库健康检查
	if err := healthCheck(db); err != nil {
		return fmt.Errorf("数据库健康检查失败: %v", err)
	}

	// 执行数据库迁移
	if err := migrateDatabase(db); err != nil {
		return fmt.Errorf("数据库迁移失败: %v", err)
	}

	// 启动定期性能报告
	if cfg.Database.PostgreSQL.EnablePerformanceLog {
		reportInterval := time.Duration(cfg.Database.PostgreSQL.PerformanceReportInterval) * time.Minute
		if reportInterval <= 0 {
			reportInterval = 5 * time.Minute // 默认5分钟
		}
		performanceMonitor.StartPeriodicReporting(reportInterval)
		log.Printf("性能监控已启用，报告间隔: %v", reportInterval)
	}

	// 记录初始连接池状态
	LogConnectionStats()

	return nil
}

// healthCheck 执行数据库健康检查
func healthCheck(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取数据库实例失败: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("数据库连接测试失败: %v", err)
	}

	log.Printf("数据库健康检查通过")
	return nil
}

// migrateDatabase 执行数据库表结构迁移
func migrateDatabase(db *gorm.DB) error {
	// 自动迁移数据库表结构
	if err := db.AutoMigrate(
		&user.User{},
		&configs.Config{},
		&task.Task{},
		&tasklog.TaskLog{},
		&filehistory.FileHistory{},
		&notification.Queue{},
	); err != nil {
		return fmt.Errorf("数据库表迁移失败: %v", err)
	}
	log.Printf("数据库表结构同步完成")
	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}

// GetConnectionStats 获取数据库连接池统计信息
func GetConnectionStats() (map[string]interface{}, error) {
	if DB == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return nil, fmt.Errorf("获取数据库实例失败: %v", err)
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}, nil
}

// LogConnectionStats 记录数据库连接池统计信息
func LogConnectionStats() {
	stats, err := GetConnectionStats()
	if err != nil {
		log.Printf("获取连接池统计信息失败: %v", err)
		return
	}

	log.Printf("数据库连接池统计: %+v", stats)
}

// optimizeConnectionConfig 优化数据库连接配置参数
func optimizeConnectionConfig(cfg *config.DatabaseConfig) {
	if cfg.Type == "postgresql" {
		// 根据系统资源动态优化PostgreSQL连接池参数
		if cfg.PostgreSQL.MaxOpenConns <= 0 {
			// 基于CPU核心数设置默认最大连接数
			cfg.PostgreSQL.MaxOpenConns = getOptimalMaxConnections()
		}
		if cfg.PostgreSQL.MaxIdleConns <= 0 {
			// 空闲连接数设置为最大连接数的20-40%
			cfg.PostgreSQL.MaxIdleConns = cfg.PostgreSQL.MaxOpenConns / 4
			if cfg.PostgreSQL.MaxIdleConns < 2 {
				cfg.PostgreSQL.MaxIdleConns = 2
			}
		}
		if cfg.PostgreSQL.ConnMaxLifetime <= 0 {
			cfg.PostgreSQL.ConnMaxLifetime = 60 // 默认连接生存时间60分钟
		}

		// 确保空闲连接数不超过最大连接数的一半
		if cfg.PostgreSQL.MaxIdleConns > cfg.PostgreSQL.MaxOpenConns/2 {
			cfg.PostgreSQL.MaxIdleConns = cfg.PostgreSQL.MaxOpenConns / 2
		}

		// 设置合理的慢查询阈值
		if cfg.PostgreSQL.SlowQueryThreshold <= 0 {
			cfg.PostgreSQL.SlowQueryThreshold = 100 // 默认100ms
		}

		log.Printf("优化后的PostgreSQL连接池配置: MaxOpen=%d, MaxIdle=%d, MaxLifetime=%d分钟, SlowQueryThreshold=%dms",
			cfg.PostgreSQL.MaxOpenConns,
			cfg.PostgreSQL.MaxIdleConns,
			cfg.PostgreSQL.ConnMaxLifetime,
			cfg.PostgreSQL.SlowQueryThreshold)
	}
}

// getOptimalMaxConnections 根据系统资源获取最优最大连接数
func getOptimalMaxConnections() int {
	// 基于经验值：每个CPU核心可以处理4-8个数据库连接
	// 这里使用保守的估算方式
	return 25 // 默认值，实际部署时可以根据具体情况调整
}

// GetPerformanceMonitor 获取性能监控器实例
func GetPerformanceMonitor() *PerformanceMonitor {
	return performanceMonitor
}

// GetPerformanceMetrics 获取数据库性能指标
func GetPerformanceMetrics() (*PerformanceMetrics, error) {
	if performanceMonitor == nil {
		return nil, fmt.Errorf("性能监控器未初始化")
	}
	return performanceMonitor.GetPerformanceMetrics()
}

// LogPerformanceReport 记录性能报告
func LogPerformanceReport() {
	if performanceMonitor != nil {
		performanceMonitor.LogPerformanceReport()
	}
}

// GetSlowQueries 获取慢查询记录
func GetSlowQueries(limit int) []SlowQueryRecord {
	if performanceMonitor == nil {
		return nil
	}
	return performanceMonitor.GetSlowQueries(limit)
}

// ResetPerformanceMetrics 重置性能指标
func ResetPerformanceMetrics() {
	if performanceMonitor != nil {
		performanceMonitor.ResetMetrics()
	}
}

// GetDetailedPerformanceReport 获取详细性能报告
func GetDetailedPerformanceReport() map[string]interface{} {
	if performanceMonitor == nil {
		return map[string]interface{}{
			"error": "性能监控器未初始化",
		}
	}
	return performanceMonitor.GetDetailedPerformanceReport()
}

// GetHealthStatus 获取数据库健康状态
func GetHealthStatus() string {
	if performanceMonitor == nil {
		return "UNKNOWN"
	}
	return performanceMonitor.GetHealthStatus()
}

// GetOptimizationRecommendations 获取优化建议
func GetOptimizationRecommendations() []string {
	if performanceMonitor == nil {
		return []string{"性能监控器未初始化"}
	}
	return performanceMonitor.GetOptimizationRecommendations()
}

// StartPerformanceMonitoring 启动性能监控
func StartPerformanceMonitoring(interval time.Duration) {
	if performanceMonitor != nil {
		performanceMonitor.StartPeriodicReporting(interval)
		log.Printf("性能监控已启动，报告间隔: %v", interval)
	}
}

// Close 关闭数据库连接
func Close() error {
	if DB == nil {
		return nil
	}

	// 记录最终性能报告
	if performanceMonitor != nil {
		log.Printf("记录最终性能报告...")
		performanceMonitor.LogPerformanceReport()
		performanceMonitor.LogOptimizationRecommendations()
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("获取数据库实例失败: %v", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("关闭数据库连接失败: %v", err)
	}

	log.Printf("数据库连接已关闭")
	return nil
}
