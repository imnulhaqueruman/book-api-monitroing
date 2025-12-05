package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP Metrics
	HttpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	HttpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "api_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	HttpRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "api_http_requests_in_flight",
			Help: "Current number of HTTP requests being processed",
		},
	)

	// Database Connection Pool Metrics
	DbConnectionsOpen = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_open",
			Help: "Current number of open database connections",
		},
	)

	DbConnectionsInUse = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_in_use",
			Help: "Current number of in-use database connections",
		},
	)

	DbConnectionsIdle = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_idle",
			Help: "Current number of idle database connections",
		},
	)

	DbConnectionsWaitCount = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "db_connections_wait_count_total",
			Help: "Total number of times waited for a connection",
		},
	)

	DbConnectionsWaitDuration = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "db_connections_wait_duration_seconds_total",
			Help: "Total time waited for database connections in seconds",
		},
	)

	// Database Query Metrics
	DbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		},
		[]string{"operation"},
	)

	// Business Metrics
	BooksCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "api_books_created_total",
			Help: "Total number of books created",
		},
	)

	BooksUpdatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "api_books_updated_total",
			Help: "Total number of books updated",
		},
	)

	BooksDeletedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "api_books_deleted_total",
			Help: "Total number of books deleted",
		},
	)

	BooksTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "api_books_total",
			Help: "Current total number of books in the system",
		},
	)

	// Error Metrics
	ApiErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_errors_total",
			Help: "Total number of API errors",
		},
		[]string{"type", "endpoint"},
	)

	ValidationErrorsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "api_validation_errors_total",
			Help: "Total number of validation errors",
		},
	)
)