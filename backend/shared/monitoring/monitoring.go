package monitoring

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	HttpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_request_duration_seconds",
		Help: "Duration of HTTP requests in seconds",
	}, []string{"method", "path", "status"})

	HttpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"method", "path", "status"})

	GrpcRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "grpc_request_duration_seconds",
		Help: "Duration of gRPC requests in seconds",
	}, []string{"service", "method", "status"})

	GrpcRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "grpc_requests_total",
		Help: "Total number of gRPC requests",
	}, []string{"service", "method", "status"})

	DbQueryDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "db_query_duration_seconds",
		Help: "Duration of database queries in seconds",
	}, []string{"operation", "table"})

	KafkaMessagesProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "kafka_messages_processed_total",
		Help: "Total number of Kafka messages processed",
	}, []string{"topic", "status"})

	NotificationsCreated = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "notifications_created_total",
		Help: "Total number of notifications created",
	}, []string{"type"})

	ReviewsCreated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "reviews_created_total",
		Help: "Total number of reviews created",
	})
)

func GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		status := c.Writer.Status()

		HttpRequestDuration.WithLabelValues(
			c.Request.Method,
			c.Request.URL.Path,
			http.StatusText(status),
		).Observe(duration)

		HttpRequestsTotal.WithLabelValues(
			c.Request.Method,
			c.Request.URL.Path,
			http.StatusText(status),
		).Inc()
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func StartMetricsServer(port string) {
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		http.ListenAndServe(":"+port, nil)
	}()
}
