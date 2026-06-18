package main

import (
	"log"
	"net/http"

	"platform-intern-growth/db"
	"platform-intern-growth/handlers"
	"platform-intern-growth/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	db.Init()

	router := gin.New()
	router.Use(middleware.RequestLogger())
	router.Use(gin.Recovery())

	router.GET("/health", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"status":  "ok",
				"service": "platform-intern-growth",
			},
		})
	})

	router.GET("/auth/validate", handlers.ValidateToken)

	protected := router.Group("/")
	protected.Use(middleware.AuthRequired())
	{
		mentorRoutes := protected.Group("/")
		mentorRoutes.Use(middleware.MentorOnly())
		{
			mentorRoutes.GET("/tasks", handlers.GetAllTasks)
			mentorRoutes.POST("/tasks", handlers.CreateTask)
			mentorRoutes.PUT("/tasks/:id", handlers.UpdateTask)
			mentorRoutes.DELETE("/tasks/:id", handlers.DeleteTask)

			mentorRoutes.POST("/students", handlers.CreateStudent)
			mentorRoutes.POST("/competences", handlers.CreateCompetence)
			mentorRoutes.GET("/students/:id/matrix", handlers.GetStudentMatrix)
			mentorRoutes.GET("/students/:id/report", handlers.GetMentorStudentReport)
		}

		internRoutes := protected.Group("/")
		internRoutes.Use(middleware.InternOnly())
		{
			internRoutes.GET("/my-tasks", handlers.GetMyTasks)
			internRoutes.PATCH("/my-tasks/:id/status", handlers.UpdateMyTaskStatus)
			internRoutes.PATCH("/my-tasks/:id/artefacts", handlers.AddMyTaskArtefact)
			internRoutes.PATCH("/my-tasks/:id/comment", handlers.UpdateMyTaskComment)

			internRoutes.GET("/my-matrix", handlers.GetMyMatrix)
			internRoutes.GET("/my-report", handlers.GetMyReport)
		}
	}

	log.Println("Сервер запущен на http://localhost:8080")
	log.Println("Проверьте работу: GET http://localhost:8080/health")
	log.Println("")
	log.Println("Тестовые токены для запросов:")
	log.Println("  Ментор:   Authorization: Bearer mentor-secret-token-abc")
	log.Println("  Стажёр 1: Authorization: Bearer intern1-secret-token-xyz")
	log.Println("  Стажёр 2: Authorization: Bearer intern2-secret-token-qwe")

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Не удалось запустить сервер: %v", err)
	}
}
