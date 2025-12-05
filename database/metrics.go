package database

import (
	"book-api/metrics"
	"time"
)

// CollectDBMetrics collects database connection pool metrics
func CollectDBMetrics() {
	if DB == nil {
		return
	}

	stats := DB.Stats()

	// Update connection pool metrics
	metrics.DbConnectionsOpen.Set(float64(stats.OpenConnections))
	metrics.DbConnectionsInUse.Set(float64(stats.InUse))
	metrics.DbConnectionsIdle.Set(float64(stats.Idle))
	metrics.DbConnectionsWaitCount.Add(float64(stats.WaitCount))
	metrics.DbConnectionsWaitDuration.Add(stats.WaitDuration.Seconds())
}

// StartMetricsCollection starts periodic collection of DB metrics
func StartMetricsCollection(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			CollectDBMetrics()
		}
	}()
}