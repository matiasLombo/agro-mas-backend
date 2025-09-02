package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"agro-mas-backend/internal/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthMiddleware validates JWT tokens and sets user context
func AuthMiddleware(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractTokenFromHeader(c.GetHeader("Authorization"))
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization token required",
				"code":  "AUTH_TOKEN_REQUIRED",
			})
			c.Abort()
			return
		}

		claims, err := jwtManager.VerifyToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
				"code":  "AUTH_TOKEN_INVALID",
			})
			c.Abort()
			return
		}

		// Set user context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Set("user_cuit", claims.CUIT)
		c.Set("user_province", claims.Province)
		c.Set("user_verification_level", claims.VerificationLevel)
		c.Set("user_is_verified", claims.IsVerified)
		c.Set("user_claims", claims)

		c.Next()
	}
}

// OptionalAuthMiddleware validates JWT tokens if present but doesn't require them
func OptionalAuthMiddleware(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractTokenFromHeader(c.GetHeader("Authorization"))
		if token != "" {
			claims, err := jwtManager.VerifyToken(token)
			if err == nil {
				// Set user context if token is valid
				c.Set("user_id", claims.UserID)
				c.Set("user_email", claims.Email)
				c.Set("user_role", claims.Role)
				c.Set("user_cuit", claims.CUIT)
				c.Set("user_province", claims.Province)
				c.Set("user_verification_level", claims.VerificationLevel)
				c.Set("user_is_verified", claims.IsVerified)
				c.Set("user_claims", claims)
			}
		}

		c.Next()
	}
}

// RequireRole middleware ensures user has required role
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User authentication required",
				"code":  "AUTH_REQUIRED",
			})
			c.Abort()
			return
		}

		role := userRole.(string)
		for _, requiredRole := range roles {
			if role == requiredRole {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error": "Insufficient permissions",
			"code":  "INSUFFICIENT_PERMISSIONS",
		})
		c.Abort()
	}
}

// RequireVerificationLevel middleware ensures user meets minimum verification level
func RequireVerificationLevel(minLevel int) gin.HandlerFunc {
	return func(c *gin.Context) {
		verificationLevel, exists := c.Get("user_verification_level")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User authentication required",
				"code":  "AUTH_REQUIRED",
			})
			c.Abort()
			return
		}

		level := verificationLevel.(int)
		if level < minLevel {
			c.JSON(http.StatusForbidden, gin.H{
				"error":           "Higher verification level required",
				"code":            "VERIFICATION_REQUIRED",
				"required_level":  minLevel,
				"current_level":   level,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireVerified middleware ensures user is verified
func RequireVerified() gin.HandlerFunc {
	return func(c *gin.Context) {
		isVerified, exists := c.Get("user_is_verified")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User authentication required",
				"code":  "AUTH_REQUIRED",
			})
			c.Abort()
			return
		}

		if !isVerified.(bool) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Account verification required",
				"code":  "VERIFICATION_REQUIRED",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// SellerOnly middleware restricts access to sellers only
func SellerOnly() gin.HandlerFunc {
	return RequireRole("seller", "admin")
}

// AdminOnly middleware restricts access to admins only
func AdminOnly() gin.HandlerFunc {
	return RequireRole("admin")
}

// ModeratorOrAdmin middleware restricts access to moderators and admins
func ModeratorOrAdmin() gin.HandlerFunc {
	return RequireRole("moderator", "admin")
}

// Helper function to extract token from Authorization header
func extractTokenFromHeader(header string) string {
	if len(header) < 7 || !strings.HasPrefix(header, "Bearer ") {
		return ""
	}
	return header[7:]
}

// GetUserID extracts user ID from context
func GetUserID(c *gin.Context) (string, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", fmt.Errorf("user ID not found in context")
	}
	return userID.(uuid.UUID).String(), nil
}

// GetUserClaims extracts user claims from context
func GetUserClaims(c *gin.Context) (*auth.UserClaims, error) {
	claims, exists := c.Get("user_claims")
	if !exists {
		return nil, fmt.Errorf("user claims not found in context")
	}
	return claims.(*auth.UserClaims), nil
}