package middleware

import (
	"slices"
	"strings"
	"time"

	"refina-transaction/config/env"
	"refina-transaction/internal/utils/data"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	config := cors.Config{
		// Gunakan AllowOriginFunc untuk wildcard pattern matching
		AllowOriginFunc: func(origin string) bool {
			// Development: allow localhost dengan port apapun
			if env.Cfg.Server.Mode == data.DEVELOPMENT_MODE {
				return strings.HasPrefix(origin, "http://localhost:") ||
					strings.HasPrefix(origin, "http://127.0.0.1:")
			}

			if strings.HasSuffix(origin, ".miftech.web.id") || strings.HasSuffix(origin, ".miv.best") {
				return true
			}

			allowedDomains := []string{
				"https://refina.miftech.web.id",
				"https://refina-staging.miftech.web.id",
			}

			return slices.Contains(allowedDomains, origin)
		},

		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "Accept", "Origin"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	return cors.New(config)
}

// Alternatif: Middleware manual untuk kontrol penuh
func CORSMiddlewareManual() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Cek apakah origin diperbolehkan
		if isAllowedOrigin(origin) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, Origin")
			c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type")
			c.Writer.Header().Set("Access-Control-Max-Age", "43200") // 12 hours
		}

		// Handle preflight request
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func isAllowedOrigin(origin string) bool {
	// Development mode: allow all localhost
	if env.Cfg.Server.Mode == data.DEVELOPMENT_MODE {
		if strings.HasPrefix(origin, "http://localhost:") ||
			strings.HasPrefix(origin, "http://127.0.0.1:") {
			return true
		}
	}

	// Production/Staging: specific domains
	allowedOrigins := map[string]bool{
		"https://refina.miftech.web.id":         true,
		"https://refina-staging.miftech.web.id": true,
	}

	// Atau gunakan pattern untuk subdomain wildcard
	// Pattern: *.miftech.web.id
	if strings.HasSuffix(origin, ".miftech.web.id") && strings.HasPrefix(origin, "https://") {
		return true
	}

	return allowedOrigins[origin]
}
