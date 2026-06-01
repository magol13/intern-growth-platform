package handlers

import (
	"net/http"

	"platform-intern-growth/middleware"
	"platform-intern-growth/models"
	"platform-intern-growth/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateCompetenceInput struct {
	Name string `json:"name" binding:"required"`
}

func CreateCompetence(context *gin.Context) {
	var input CreateCompetenceInput
	if err := context.ShouldBindJSON(&input); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Поле name обязательно",
		})
		return
	}

	competence, err := service.CreateCompetence(input.Name)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	context.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    competence,
	})
}

func GetStudentMatrix(context *gin.Context) {
	studentIDString := context.Param("id")
	studentID, err := uuid.Parse(studentIDString)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Неверный формат ID стажёра",
		})
		return
	}

	matrix, err := service.GetStudentMatrix(studentID)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    matrix,
	})
}

func GetMyMatrix(context *gin.Context) {
	currentUser := context.MustGet(middleware.CurrentUserKey).(*models.User)

	matrix, err := service.GetStudentMatrix(currentUser.ID)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    matrix,
	})
}
