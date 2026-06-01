package metrics

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

var (
	requestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"service", "method", "path", "status"},
	)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method", "path"},
	)

	kafkaConsumerLag = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kafka_consumer_lag",
			Help: "Approximate Kafka consumer group lag",
		},
		[]string{"service", "group", "topic"},
	)
)

func init() {
	prometheus.MustRegister(requestCount, requestDuration, kafkaConsumerLag)
}

// RegisterRuntimeCollectors is a no-op: prometheus/client_golang v1.19+ registers
// Go and process collectors on DefaultRegisterer in prometheus/registry.go init().
func RegisterRuntimeCollectors() {}

// SetKafkaConsumerLag updates lag for a consumer group/topic.
func SetKafkaConsumerLag(service, group, topic string, lag int64) {
	kafkaConsumerLag.WithLabelValues(service, group, topic).Set(float64(lag))
}

// RegisterRoutes mounts Prometheus metrics and observability middleware on Gin.
func RegisterRoutes(router *gin.Engine, serviceName string) {
	RegisterRuntimeCollectors()
	router.Use(otelgin.Middleware(serviceName))
	router.Use(metricsMiddleware(serviceName))
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
}

func metricsMiddleware(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		status := c.Writer.Status()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		method := c.Request.Method

		requestCount.WithLabelValues(serviceName, method, path, statusLabel(status)).Inc()
		requestDuration.WithLabelValues(serviceName, method, path).Observe(time.Since(start).Seconds())
	}
}

func statusLabel(status int) string {
	switch {
	case status >= 500:
		return "5xx"
	case status >= 400:
		return "4xx"
	case status >= 300:
		return "3xx"
	case status >= 200:
		return "2xx"
	default:
		return "other"
	}
}
