package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func RequestLogger() gin.HandlerFunc {
	return func(context *gin.Context) {
		startTime := time.Now()

		context.Next()

		duration := time.Since(startTime)
		statusCode := context.Writer.Status()
		method := context.Request.Method
		path := context.Request.URL.Path
		clientIP := context.ClientIP()

		log.Printf("[%s] %s %s | Статус: %d | Время: %v | IP: %s",
			time.Now().Format("2006-01-02 15:04:05"),
			method,
			path,
			statusCode,
			duration,
			clientIP,
		)
	}
}
