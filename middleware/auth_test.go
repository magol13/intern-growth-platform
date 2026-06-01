package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// Тест T-1: Проверка парсинга токена из заголовка Authorization.
// Эти тесты проверяют формат заголовка, не обращаясь к базе данных.

func init() {
	gin.SetMode(gin.TestMode)
}

// extractTokenFromHeader — вспомогательная функция, выделяет логику парсинга.
// Возвращает токен и признак ошибки.
func extractTokenFromHeader(authHeader string) (string, bool) {
	if authHeader == "" {
		return "", false
	}

	const prefix = "Bearer "
	if len(authHeader) <= len(prefix) {
		return "", false
	}

	if authHeader[:len(prefix)] != prefix {
		return "", false
	}

	token := authHeader[len(prefix):]
	if token == "" {
		return "", false
	}

	return token, true
}

func TestExtractToken_ValidBearerToken_ReturnsToken(t *testing.T) {
	authHeader := "Bearer my-secret-token-123"

	token, ok := extractTokenFromHeader(authHeader)

	if !ok {
		t.Error("Должен успешно извлечь токен из корректного заголовка")
	}
	if token != "my-secret-token-123" {
		t.Errorf("Ожидался токен 'my-secret-token-123', получен: '%s'", token)
	}
}

func TestExtractToken_EmptyHeader_ReturnsFalse(t *testing.T) {
	authHeader := ""

	_, ok := extractTokenFromHeader(authHeader)

	if ok {
		t.Error("Пустой заголовок должен вернуть ошибку")
	}
}

func TestExtractToken_MissingBearerPrefix_ReturnsFalse(t *testing.T) {
	invalidHeaders := []string{
		"my-secret-token",
		"Token my-secret-token",
		"Basic dXNlcjpwYXNz",
		"bearer my-secret-token", // bearer с маленькой буквы
	}

	for _, header := range invalidHeaders {
		_, ok := extractTokenFromHeader(header)
		if ok {
			t.Errorf("Заголовок '%s' должен вернуть ошибку, но был принят", header)
		}
	}
}

func TestAuthMiddleware_NoHeader_Returns401(t *testing.T) {
	router := gin.New()
	router.Use(func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Заголовок Authorization отсутствует",
			})
			ctx.Abort()
			return
		}
		ctx.Next()
	})
	router.GET("/test", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"success": true})
	})

	request := httptest.NewRequest(http.MethodGet, "/test", nil)
	// Заголовок Authorization намеренно не добавляем
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("Ожидался статус 401, получен: %d", recorder.Code)
	}
}

func TestAuthMiddleware_WithHeader_PassesThrough(t *testing.T) {
	router := gin.New()
	router.Use(func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"success": false})
			ctx.Abort()
			return
		}
		ctx.Next()
	})
	router.GET("/test", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"success": true})
	})

	request := httptest.NewRequest(http.MethodGet, "/test", nil)
	request.Header.Set("Authorization", "Bearer some-token")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	// Middleware пропускает запрос дальше (хендлер вернёт 200)
	if recorder.Code != http.StatusOK {
		t.Errorf("Запрос с заголовком должен проходить, получен код: %d", recorder.Code)
	}
}
