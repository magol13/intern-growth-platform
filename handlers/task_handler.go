package handlers

import (
	"net/http"

	"platform-intern-growth/middleware"
	"platform-intern-growth/models"
	"platform-intern-growth/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GetAllTasks(context *gin.Context) {
	statusFilter := context.Query("status")

	tasks, err := service.GetAllTasks(statusFilter)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tasks,
	})
}

func CreateTask(context *gin.Context) {
	currentUser := context.MustGet(middleware.CurrentUserKey).(*models.User)

	var input service.CreateTaskInput
	if err := context.ShouldBindJSON(&input); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Ошибка валидации: " + err.Error(),
		})
		return
	}

	task, err := service.CreateTask(currentUser.ID, input)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	context.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    task,
	})
}

func UpdateTask(context *gin.Context) {
	currentUser := context.MustGet(middleware.CurrentUserKey).(*models.User)

	taskIDString := context.Param("id")
	taskID, err := uuid.Parse(taskIDString)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Неверный формат ID задачи",
		})
		return
	}

	var input service.UpdateTaskInput
	if err := context.ShouldBindJSON(&input); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Ошибка валидации: " + err.Error(),
		})
		return
	}

	task, err := service.UpdateTask(taskID, currentUser.ID, input)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    task,
	})
}

func DeleteTask(context *gin.Context) {
	currentUser := context.MustGet(middleware.CurrentUserKey).(*models.User)

	taskIDString := context.Param("id")
	taskID, err := uuid.Parse(taskIDString)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Неверный формат ID задачи",
		})
		return
	}

	err = service.DeleteTask(taskID, currentUser.ID)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    nil,
	})
}
