package middleware

import (
	"dbms/internal/models"
	"dbms/pkg"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(config *pkg.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Yêu cầu token"})
			c.Abort()
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		claims := &models.UserClaims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			return config.JWTSecret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token không hợp lệ"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func RoleCheckMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if role != requiredRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "Bạn không có quyền thực hiện hành động này"})
			c.Abort()
			return
		}
		c.Next()
	}
}
