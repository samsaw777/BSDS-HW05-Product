package main

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Rate: counts every request
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "route", "status"},
	)

	// Errors: counts only non-2xx responses
	httpErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_errors_total",
			Help: "Total number of non-2xx HTTP responses.",
		},
		[]string{"method", "route", "status"},
	)

	// Duration: how long each request took
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: []float64{0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
		[]string{"method", "route"},
	)
)

// metricsMiddleware records rate, errors, and duration for every request.
func metricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next() // run the handler

		route := c.FullPath() // e.g. "/products/:productId", not "/products/1100"
		if route == "/metrics" {
			return // don't count Prometheus scraping itself
		}
		if route == "" {
			route = "unmatched"
		}

		status := c.Writer.Status()
		statusLabel := strconv.Itoa(status)
		method := c.Request.Method

		httpRequestsTotal.WithLabelValues(method, route, statusLabel).Inc()
		if status < 200 || status >= 300 {
			httpErrorsTotal.WithLabelValues(method, route, statusLabel).Inc()
		}
		httpRequestDuration.WithLabelValues(method, route).Observe(time.Since(start).Seconds())
	}
}