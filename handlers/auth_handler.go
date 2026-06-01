package handlers

import (
	"net/http"
	"strings"

	"platform-intern-growth/service"

	"github.com/gin-gonic/gin"
)

func ValidateToken(context *gin.Context) {
	authHeader := context.GetHeader("Authorization")
	if authHeader == "" {
		context.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Заголовок Authorization отсутствует",
		})
		return
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		context.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Неверный формат токена. Используйте: Bearer <token>",
		})
		return
	}

	token := parts[1]

	authResult, err := service.ValidateToken(token)
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Токен не найден или недействителен",
		})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    authResult,
	})
}
