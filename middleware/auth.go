package middleware

import (
	"net/http"
	"strings"

	"platform-intern-growth/models"
	"platform-intern-growth/repository"

	"github.com/gin-gonic/gin"
)

const CurrentUserKey = "currentUser"

func AuthRequired() gin.HandlerFunc {
	return func(context *gin.Context) {
		authHeader := context.GetHeader("Authorization")
		if authHeader == "" {
			context.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Заголовок Authorization отсутствует",
			})
			context.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			context.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Неверный формат токена. Используйте: Bearer <token>",
			})
			context.Abort()
			return
		}

		token := parts[1]
		if token == "" {
			context.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Токен не может быть пустым",
			})
			context.Abort()
			return
		}

		user, err := repository.FindUserByToken(token)
		if err != nil {
			context.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Токен не найден или недействителен",
			})
			context.Abort()
			return
		}

		context.Set(CurrentUserKey, user)
		context.Next()
	}
}

func MentorOnly() gin.HandlerFunc {
	return func(context *gin.Context) {
		user, exists := context.Get(CurrentUserKey)
		if !exists {
			context.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Пользователь не аутентифицирован",
			})
			context.Abort()
			return
		}

		currentUser, ok := user.(*models.User)
		if !ok || currentUser.Role != models.RoleMentor {
			context.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Доступ только для менторов",
			})
			context.Abort()
			return
		}

		context.Next()
	}
}

func InternOnly() gin.HandlerFunc {
	return func(context *gin.Context) {
		user, exists := context.Get(CurrentUserKey)
		if !exists {
			context.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Пользователь не аутентифицирован",
			})
			context.Abort()
			return
		}

		currentUser, ok := user.(*models.User)
		if !ok || currentUser.Role != models.RoleIntern {
			context.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Доступ только для стажёров",
			})
			context.Abort()
			return
		}

		context.Next()
	}
}
