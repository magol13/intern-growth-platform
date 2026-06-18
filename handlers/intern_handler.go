package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"platform-intern-growth/middleware"
	"platform-intern-growth/models"
	"platform-intern-growth/repository"
	"platform-intern-growth/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateStudentInput struct {
	Name string `json:"name" binding:"required"`
}

func CreateStudent(context *gin.Context) {
	var input CreateStudentInput
	if err := context.ShouldBindJSON(&input); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Поле name обязательно",
		})
		return
	}

	// Генерируем случайный токен (32 случайных байта → hex-строка)
	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Не удалось сгенерировать токен",
		})
		return
	}
	token := hex.EncodeToString(tokenBytes)

	newIntern := &models.User{
		ID:    uuid.New(),
		Name:  input.Name,
		Role:  models.RoleIntern,
		Token: token,
	}

	if err := repository.CreateUser(newIntern); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Не удалось создать стажёра",
		})
		return
	}

	context.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"id":    newIntern.ID,
			"name":  newIntern.Name,
			"role":  newIntern.Role,
			"token": newIntern.Token,
		},
	})
}

func GetMyTasks(context *gin.Context) {
	currentUser := context.MustGet(middleware.CurrentUserKey).(*models.User)

	tasks, err := repository.FindTasksByStudentID(currentUser.ID)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Не удалось получить список задач",
		})
		return
	}

	var internViews []models.InternTaskView
	for _, task := range tasks {
		internViews = append(internViews, models.InternTaskView{
			ID:          task.ID,
			Title:       task.Title,
			Status:      task.Status,
			Deadline:    task.Deadline,
			Competences: task.Competences,
		})
	}

	context.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    internViews,
	})
}

type UpdateStatusInput struct {
	Status string `json:"status" binding:"required"`
}

func UpdateMyTaskStatus(context *gin.Context) {
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

	var input UpdateStatusInput
	if err := context.ShouldBindJSON(&input); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Поле status обязательно",
		})
		return
	}

	task, err := service.UpdateInternTaskStatus(taskID, currentUser.ID, models.TaskStatus(input.Status))
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":     task.ID,
			"status": task.Status,
		},
	})
}

type AddArtefactInput struct {
	URL string `json:"url" binding:"required,url"`
}

func AddMyTaskArtefact(context *gin.Context) {
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

	var input AddArtefactInput
	if err := context.ShouldBindJSON(&input); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Поле url обязательно и должно быть валидным URL",
		})
		return
	}

	task, err := service.AddArtefactToTask(taskID, currentUser.ID, input.URL)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":        task.ID,
			"artefacts": task.Artefacts,
		},
	})
}

type UpdateCommentInput struct {
	Comment string `json:"comment" binding:"required"`
}

func UpdateMyTaskComment(context *gin.Context) {
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

	var input UpdateCommentInput
	if err := context.ShouldBindJSON(&input); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Поле comment обязательно",
		})
		return
	}

	task, err := service.UpdateTaskComment(taskID, currentUser.ID, input.Comment)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":       task.ID,
			"comments": task.Comments,
		},
	})
}
