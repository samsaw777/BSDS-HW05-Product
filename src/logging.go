package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

// loggingMiddleware writes one JSON line per request.
func loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next() // run the handler

		route := c.FullPath()
		if route == "/metrics" {
			return // skip Prometheus scrapes
		}
		if route == "" {
			route = "unmatched"
		}

		logger.Info("request",
			"method", c.Request.Method,
			"route", route,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration_ms", float64(time.Since(start).Microseconds())/1000,
		)
	}
}