package middleware

import (
	"JustChat/internal/auth/usecase"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
)

func AuthMiddleware(authUC usecase.JWTUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		userID, err := authUC.ParseToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		// Устанавливаем в контекст
		c.Set("userID", userID)

		// Устанавливаем в заголовок X-User-ID
		c.Request.Header.Set("X-User-ID", strconv.FormatInt(userID.UserID, 10))
		c.Next()
	}
}
