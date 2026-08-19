// internal/middleware/auth_middleware.go
package middleware

import (
    "net/http"
    "strings"

    "go-crud-jwt/internal/utils"

    "github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Get token from header
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
            c.Abort()
            return
        }

        // Check if token is Bearer format
        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format. Use Bearer token"})
            c.Abort()
            return
        }

        token := parts[1]

        // Validate token
        claims, err := utils.ValidateToken(token)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
            c.Abort()
            return
        }

        // Set user info in context
        c.Set("user_id", claims.UserID)
        c.Set("username", claims.Username)
        c.Set("email", claims.Email)
        c.Set("role", claims.Role)

        c.Next()
    }
}

// Role-based middleware
func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        role, exists := c.Get("role")
        if !exists {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
            c.Abort()
            return
        }

        roleStr := role.(string)
        allowed := false
        for _, r := range allowedRoles {
            if r == roleStr {
                allowed = true
                break
            }
        }

        if !allowed {
            c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
            c.Abort()
            return
        }

        c.Next()
    }
}