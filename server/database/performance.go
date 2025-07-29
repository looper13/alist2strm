package database

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// PerformanceMetrics 性能指标结构
type PerformanceMetrics struct {
	// 连接池指标
	ConnectionPool ConnectionPoolMetrics `json:"connection_pool"`
	// 查询性能指标
	QueryPerformance QueryPerformanceMetrics `json:"query_performance"`
	// 慢查询统计
	SlowQueries []SlowQueryRecord `json:"slow_queries"`
	// 统计时间范围
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// ConnectionPoolMetrics 连接池指标
type ConnectionPoolMetrics struct {
	MaxOpenConnections    int           `json:"max_open_connections"`
	OpenConnections       int           `json:"open_connections"`
	InUse                 int           `json:"in_use"`
	Idle                  int           `json:"idle"`
	WaitCount             int64         `json:"wait_count"`
	WaitDuration          time.Duration `json:"wait_duration"`
	MaxIdleClosed         int64         `json:"max_idle_closed"`
	MaxIdleTimeClosed     int64         `json:"max_idle_time_closed"`
	MaxLifetimeClosed     int64         `json:"max_lifetime_closed"`
	ConnectionUtilization float64       `json:"connection_utilization"` // 连接使用率
	AverageWaitTime       time.Duration `json:"average_wait_time"`      // 平均等待时间
}

// QueryPerformanceMetrics 查询性能指标
type QueryPerformanceMetrics struct {
	TotalQueries       int64         `json:"total_queries"`        // 总查询数
	SlowQueries        int64         `json:"slow_queries"`         // 慢查询数
	AverageQueryTime   time.Duration `json:"average_query_time"`   // 平均查询时间
	MaxQueryTime       time.Duration `json:"max_query_time"`       // 最大查询时间
	MinQueryTime       time.Duration `json:"min_query_time"`       // 最小查询时间
	QueriesPerSecond   float64       `json:"queries_per_second"`   // 每秒查询数
	SlowQueryThreshold time.Duration `json:"slow_query_threshold"` // 慢查询阈值
}

// SlowQueryRecord 慢查询记录
type SlowQueryRecord struct {
	SQL          string        `json:"sql"`           // SQL语句
	Duration     time.Duration `json:"duration"`      // 执行时间
	Timestamp    time.Time     `json:"timestamp"`     // 执行时间戳
	RowsAffected int64         `json:"rows_affected"` // 影响行数
}

// PerformanceMonitor 性能监控器
type PerformanceMonitor struct {
	db                 *gorm.DB
	slowQueryThreshold time.Duration
	metrics            *PerformanceMetrics
	slowQueries        []SlowQueryRecord
	queryStats         struct {
		totalQueries  int64
		totalDuration time.Duration
		maxDuration   time.Duration
		minDuration   time.Duration
		startTime     time.Time
	}
	mutex sync.RWMutex
}

// NewPerformanceMonitor 创建性能监控器
func NewPerformanceMonitor(db *gorm.DB, slowQueryThreshold time.Duration) *PerformanceMonitor {
	monitor := &PerformanceMonitor{
		db:                 db,
		slowQueryThreshold: slowQueryThreshold,
		slowQueries:        make([]SlowQueryRecord, 0),
	}

	// 初始化查询统计
	monitor.queryStats.startTime = time.Now()
	monitor.queryStats.minDuration = time.Hour // 设置一个较大的初始值

	// 设置自定义日志记录器来捕获慢查询
	monitor.setupSlowQueryLogger()

	return monitor
}

// setupSlowQueryLogger 设置慢查询日志记录器
func (pm *PerformanceMonitor) setupSlowQueryLogger() {
	customLogger := &SlowQueryLogger{
		monitor: pm,
		logger:  logger.Default,
	}

	pm.db.Logger = customLogger
}

// SlowQueryLogger 慢查询日志记录器
type SlowQueryLogger struct {
	monitor *PerformanceMonitor
	logger  logger.Interface
}

// LogMode 实现logger.Interface的LogMode方法
func (sql *SlowQueryLogger) LogMode(level logger.LogLevel) logger.Interface {
	return &SlowQueryLogger{
		monitor: sql.monitor,
		logger:  sql.logger.LogMode(level),
	}
}

// Info 实现logger.Interface的Info方法
func (sql *SlowQueryLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	sql.logger.Info(ctx, msg, data...)
}

// Warn 实现logger.Interface的Warn方法
func (sql *SlowQueryLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	sql.logger.Warn(ctx, msg, data...)
}

// Error 实现logger.Interface的Error方法
func (sql *SlowQueryLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	sql.logger.Error(ctx, msg, data...)
}

// Trace 实现logger.Interface的Trace方法
func (sql *SlowQueryLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sqlStr, rows := fc()

	// 更新查询统计
	sql.monitor.updateQueryStats(elapsed)

	// 如果是慢查询，记录到慢查询列表
	if elapsed >= sql.monitor.slowQueryThreshold {
		sql.monitor.recordSlowQuery(sqlStr, elapsed, rows)
		log.Printf("慢查询检测: SQL=%s, 耗时=%v, 影响行数=%d", sqlStr, elapsed, rows)
	}

	// 调用原始logger的Trace方法
	sql.logger.Trace(ctx, begin, fc, err)
}

// updateQueryStats 更新查询统计信息
func (pm *PerformanceMonitor) updateQueryStats(duration time.Duration) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	pm.queryStats.totalQueries++
	pm.queryStats.totalDuration += duration

	if duration > pm.queryStats.maxDuration {
		pm.queryStats.maxDuration = duration
	}

	if duration < pm.queryStats.minDuration {
		pm.queryStats.minDuration = duration
	}
}

// recordSlowQuery 记录慢查询
func (pm *PerformanceMonitor) recordSlowQuery(sql string, duration time.Duration, rowsAffected int64) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	record := SlowQueryRecord{
		SQL:          sql,
		Duration:     duration,
		Timestamp:    time.Now(),
		RowsAffected: rowsAffected,
	}

	pm.slowQueries = append(pm.slowQueries, record)

	// 限制慢查询记录数量，避免内存泄漏
	if len(pm.slowQueries) > 1000 {
		pm.slowQueries = pm.slowQueries[len(pm.slowQueries)-500:] // 保留最近500条
	}
}

// GetConnectionPoolMetrics 获取连接池指标
func (pm *PerformanceMonitor) GetConnectionPoolMetrics() (ConnectionPoolMetrics, error) {
	sqlDB, err := pm.db.DB()
	if err != nil {
		return ConnectionPoolMetrics{}, fmt.Errorf("获取数据库实例失败: %v", err)
	}

	stats := sqlDB.Stats()

	// 计算连接使用率
	var utilization float64
	if stats.MaxOpenConnections > 0 {
		utilization = float64(stats.InUse) / float64(stats.MaxOpenConnections) * 100
	}

	// 计算平均等待时间
	var avgWaitTime time.Duration
	if stats.WaitCount > 0 {
		avgWaitTime = stats.WaitDuration / time.Duration(stats.WaitCount)
	}

	return ConnectionPoolMetrics{
		MaxOpenConnections:    stats.MaxOpenConnections,
		OpenConnections:       stats.OpenConnections,
		InUse:                 stats.InUse,
		Idle:                  stats.Idle,
		WaitCount:             stats.WaitCount,
		WaitDuration:          stats.WaitDuration,
		MaxIdleClosed:         stats.MaxIdleClosed,
		MaxIdleTimeClosed:     stats.MaxIdleTimeClosed,
		MaxLifetimeClosed:     stats.MaxLifetimeClosed,
		ConnectionUtilization: utilization,
		AverageWaitTime:       avgWaitTime,
	}, nil
}

// GetQueryPerformanceMetrics 获取查询性能指标
func (pm *PerformanceMonitor) GetQueryPerformanceMetrics() QueryPerformanceMetrics {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	var avgQueryTime time.Duration
	var qps float64

	if pm.queryStats.totalQueries > 0 {
		avgQueryTime = pm.queryStats.totalDuration / time.Duration(pm.queryStats.totalQueries)

		// 计算QPS
		elapsed := time.Since(pm.queryStats.startTime)
		if elapsed > 0 {
			qps = float64(pm.queryStats.totalQueries) / elapsed.Seconds()
		}
	}

	return QueryPerformanceMetrics{
		TotalQueries:       pm.queryStats.totalQueries,
		SlowQueries:        int64(len(pm.slowQueries)),
		AverageQueryTime:   avgQueryTime,
		MaxQueryTime:       pm.queryStats.maxDuration,
		MinQueryTime:       pm.queryStats.minDuration,
		QueriesPerSecond:   qps,
		SlowQueryThreshold: pm.slowQueryThreshold,
	}
}

// GetSlowQueries 获取慢查询记录
func (pm *PerformanceMonitor) GetSlowQueries(limit int) []SlowQueryRecord {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	if limit <= 0 || limit > len(pm.slowQueries) {
		limit = len(pm.slowQueries)
	}

	// 返回最近的慢查询记录
	start := len(pm.slowQueries) - limit
	if start < 0 {
		start = 0
	}

	result := make([]SlowQueryRecord, limit)
	copy(result, pm.slowQueries[start:])
	return result
}

// GetPerformanceMetrics 获取完整的性能指标
func (pm *PerformanceMonitor) GetPerformanceMetrics() (*PerformanceMetrics, error) {
	connectionMetrics, err := pm.GetConnectionPoolMetrics()
	if err != nil {
		return nil, err
	}

	queryMetrics := pm.GetQueryPerformanceMetrics()
	slowQueries := pm.GetSlowQueries(100) // 获取最近100条慢查询

	return &PerformanceMetrics{
		ConnectionPool:   connectionMetrics,
		QueryPerformance: queryMetrics,
		SlowQueries:      slowQueries,
		StartTime:        pm.queryStats.startTime,
		EndTime:          time.Now(),
	}, nil
}

// ResetMetrics 重置性能指标
func (pm *PerformanceMonitor) ResetMetrics() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	pm.queryStats.totalQueries = 0
	pm.queryStats.totalDuration = 0
	pm.queryStats.maxDuration = 0
	pm.queryStats.minDuration = time.Hour
	pm.queryStats.startTime = time.Now()
	pm.slowQueries = pm.slowQueries[:0] // 清空慢查询记录
}

// LogPerformanceReport 记录性能报告
func (pm *PerformanceMonitor) LogPerformanceReport() {
	metrics, err := pm.GetPerformanceMetrics()
	if err != nil {
		log.Printf("获取性能指标失败: %v", err)
		return
	}

	// 将指标转换为JSON格式便于阅读
	jsonData, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		log.Printf("序列化性能指标失败: %v", err)
		return
	}

	log.Printf("数据库性能报告:\n%s", string(jsonData))
}

// StartPeriodicReporting 启动定期性能报告
func (pm *PerformanceMonitor) StartPeriodicReporting(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			pm.LogPerformanceReport()
			pm.LogOptimizationRecommendations()
		}
	}()
}

// LogOptimizationRecommendations 记录优化建议
func (pm *PerformanceMonitor) LogOptimizationRecommendations() {
	recommendations := pm.GetOptimizationRecommendations()
	if len(recommendations) > 0 {
		log.Printf("数据库优化建议:")
		for i, rec := range recommendations {
			log.Printf("  %d. %s", i+1, rec)
		}
	}
}

// GetOptimizationRecommendations 获取优化建议
func (pm *PerformanceMonitor) GetOptimizationRecommendations() []string {
	recommendations := make([]string, 0)

	// 获取连接池指标
	poolMetrics, err := pm.GetConnectionPoolMetrics()
	if err != nil {
		return recommendations
	}

	// 获取查询性能指标
	queryMetrics := pm.GetQueryPerformanceMetrics()

	// 连接池使用率过高
	if poolMetrics.ConnectionUtilization > 80 {
		recommendations = append(recommendations,
			fmt.Sprintf("连接池使用率过高(%.1f%%)，建议增加最大连接数", poolMetrics.ConnectionUtilization))
	}

	// 等待时间过长
	if poolMetrics.AverageWaitTime > 100*time.Millisecond {
		recommendations = append(recommendations,
			fmt.Sprintf("平均等待时间过长(%v)，建议优化连接池配置", poolMetrics.AverageWaitTime))
	}

	// 慢查询过多
	if queryMetrics.TotalQueries > 0 {
		slowQueryRate := float64(queryMetrics.SlowQueries) / float64(queryMetrics.TotalQueries) * 100
		if slowQueryRate > 5 {
			recommendations = append(recommendations,
				fmt.Sprintf("慢查询比例过高(%.1f%%)，建议优化SQL语句或添加索引", slowQueryRate))
		}
	}

	// 平均查询时间过长
	if queryMetrics.AverageQueryTime > 50*time.Millisecond {
		recommendations = append(recommendations,
			fmt.Sprintf("平均查询时间过长(%v)，建议检查查询效率", queryMetrics.AverageQueryTime))
	}

	// QPS过低可能表示性能问题
	if queryMetrics.QueriesPerSecond > 0 && queryMetrics.QueriesPerSecond < 1 {
		recommendations = append(recommendations,
			"查询频率较低，可能存在性能瓶颈或连接问题")
	}

	return recommendations
}

// GetDetailedPerformanceReport 获取详细性能报告
func (pm *PerformanceMonitor) GetDetailedPerformanceReport() map[string]interface{} {
	metrics, err := pm.GetPerformanceMetrics()
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	recommendations := pm.GetOptimizationRecommendations()

	return map[string]interface{}{
		"metrics":         metrics,
		"recommendations": recommendations,
		"health_status":   pm.GetHealthStatus(),
		"report_time":     time.Now(),
	}
}

// GetHealthStatus 获取数据库健康状态
func (pm *PerformanceMonitor) GetHealthStatus() string {
	poolMetrics, err := pm.GetConnectionPoolMetrics()
	if err != nil {
		return "UNKNOWN"
	}

	queryMetrics := pm.GetQueryPerformanceMetrics()

	// 检查各项指标
	issues := 0

	if poolMetrics.ConnectionUtilization > 90 {
		issues++
	}

	if poolMetrics.AverageWaitTime > 200*time.Millisecond {
		issues++
	}

	if queryMetrics.TotalQueries > 0 {
		slowQueryRate := float64(queryMetrics.SlowQueries) / float64(queryMetrics.TotalQueries) * 100
		if slowQueryRate > 10 {
			issues++
		}
	}

	if queryMetrics.AverageQueryTime > 100*time.Millisecond {
		issues++
	}

	switch issues {
	case 0:
		return "HEALTHY"
	case 1:
		return "WARNING"
	default:
		return "CRITICAL"
	}
}
