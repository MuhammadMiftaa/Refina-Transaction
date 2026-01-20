package middleware

import (
	"fmt"
	"time"

	"refina-transaction/config/log"

	"github.com/gin-gonic/gin"
)

type middlewareConfig struct {
	// SkipPaths - daftar path yang tidak akan di-log (misal: /health, /metrics)
	SkipPaths []string
	// SkipUserAgents - daftar user agent yang tidak akan di-log (misal: monitoring tools)
	SkipUserAgents []string
	// LogRequestBody - apakah log request body (hati-hati dengan sensitive data)
	LogRequestBody bool
	// LogResponseBody - apakah log response body (hati-hati dengan ukuran response)
	LogResponseBody bool
}

func httpRequest(c *gin.Context, latency time.Duration, statusCode int) {
	// Get client IP dengan handling proxy
	clientIP := c.ClientIP()

	// Get user agent
	userAgent := c.Request.UserAgent()
	if userAgent == "" {
		userAgent = "-"
	}

	// Get referer
	referer := c.Request.Referer()
	if referer == "" {
		referer = "-"
	}

	// Format request size
	requestSize := c.Request.ContentLength
	if requestSize < 0 {
		requestSize = 0
	}

	// Response size (approximate, karena gin tidak menyediakan exact response size)
	responseSize := c.Writer.Size()
	if responseSize < 0 {
		responseSize = 0
	}

	// Fields untuk structured logging
	fields := map[string]interface{}{
		"method":        c.Request.Method,
		"uri":           c.Request.RequestURI,
		"status":        statusCode,
		"latency":       fmt.Sprintf("%.3fms", float64(latency.Nanoseconds())/1000000.0),
		"client_ip":     clientIP,
		"user_agent":    userAgent,
		"referer":       referer,
		"request_size":  requestSize,
		"response_size": responseSize,
		"protocol":      c.Request.Proto,
	}

	// Format message dalam style Apache Combined Log Format
	// Format: "METHOD URI PROTOCOL" status response_size "referer" "user_agent"
	message := fmt.Sprintf(`%s %s %s`,
		c.Request.Method,
		c.Request.RequestURI,
		c.Request.Proto,
	)

	// Tentukan log level berdasarkan status code
	switch {
	case statusCode >= 200 && statusCode < 300:
		log.Info(message, fields)
	case statusCode >= 300 && statusCode < 400:
		log.Info(message, fields)
	case statusCode >= 400 && statusCode < 500:
		log.Warn(message, fields)
	case statusCode >= 500:
		log.Error(message, fields)
	default:
		log.Info(message, fields)
	}
}

func GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get status code
		statusCode := c.Writer.Status()

		// Log the request
		httpRequest(c, latency, statusCode)
	}
}

func GinMiddlewareWithConfig(config middlewareConfig) gin.HandlerFunc {
	// Convert skip paths to map untuk O(1) lookup
	skipPaths := make(map[string]bool)
	for _, path := range config.SkipPaths {
		skipPaths[path] = true
	}

	// Convert skip user agents to map
	skipUserAgents := make(map[string]bool)
	for _, ua := range config.SkipUserAgents {
		skipUserAgents[ua] = true
	}

	return func(c *gin.Context) {
		// Skip jika path ada di skip list
		if skipPaths[c.Request.URL.Path] {
			c.Next()
			return
		}

		// Skip jika user agent ada di skip list
		if skipUserAgents[c.Request.UserAgent()] {
			c.Next()
			return
		}

		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get status code
		statusCode := c.Writer.Status()

		// Log the request
		httpRequest(c, latency, statusCode)
	}
}
