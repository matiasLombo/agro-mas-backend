package middleware

import (
	"fmt"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware configures CORS for the agricultural marketplace
func CORSMiddleware() gin.HandlerFunc {
	config := cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000",  // React dev port
			"http://localhost:4200",  // Angular default dev port
			"http://localhost:4201",  // Angular alternative dev port
			"http://localhost:3001",  // Alternative dev port
			// Firebase projects for environments
			"https://agro-mas-fe-dev-2025.web.app",     // Firebase Development
			"https://agro-mas-fe-dev-2025.firebaseapp.com", // Firebase Development Alt
			"https://agro-mas-fe-prod-2025.web.app",    // Firebase Production
			"https://agro-mas-fe-prod-2025.firebaseapp.com", // Firebase Production Alt
		},
		AllowMethods: []string{
			"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS",
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Content-Length",
			"Accept-Encoding",
			"X-CSRF-Token",
			"Authorization",
			"accept",
			"origin",
			"Cache-Control",
			"X-Requested-With",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"X-Total-Count",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	return cors.New(config)
}

// SecurityHeadersMiddleware adds security headers
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'; img-src 'self' data: https:; script-src 'self' 'unsafe-inline'")
		
		// API-specific headers
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")

		c.Next()
	}
}

// LoggerMiddleware configures structured logging with debug for PUT requests
func LoggerMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Add extra debug for PUT requests
		if param.Method == "PUT" && param.StatusCode == 400 {
			fmt.Printf("[MIDDLEWARE DEBUG] PUT request failed with 400\n")
			fmt.Printf("[MIDDLEWARE DEBUG] Path: %s\n", param.Path)
			fmt.Printf("[MIDDLEWARE DEBUG] Content-Type: %s\n", param.Request.Header.Get("Content-Type"))
			fmt.Printf("[MIDDLEWARE DEBUG] Content-Length: %s\n", param.Request.Header.Get("Content-Length"))
			fmt.Printf("[MIDDLEWARE DEBUG] Error: %s\n", param.ErrorMessage)
		}
		
		return fmt.Sprintf(`{"time":"%s","method":"%s","path":"%s","protocol":"%s","status":%d,"latency":"%s","client_ip":"%s","user_agent":"%s","errors":"%s"}%s`,
			param.TimeStamp.Format(time.RFC3339),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.ClientIP,
			param.Request.UserAgent(),
			param.ErrorMessage,
			"\n",
		)
	})
}

// APIVersionMiddleware sets API version header
func APIVersionMiddleware(version string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("API-Version", version)
		c.Next()
	}
}

// ContentTypeMiddleware ensures JSON content type for API responses (not for requests)
func ContentTypeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process the request first
		c.Next()
		
		// Only set JSON content type if not already set (don't override multipart responses)
		if c.GetHeader("Content-Type") == "" {
			c.Header("Content-Type", "application/json; charset=utf-8")
		}
	}
}
