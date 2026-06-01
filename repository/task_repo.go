package repository

import (
	"platform-intern-growth/db"
	"platform-intern-growth/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func CreateTask(task *models.Task) error {
	return db.DB.Create(task).Error
}

func FindTaskByID(taskID uuid.UUID) (*models.Task, error) {
	var task models.Task
	result := db.DB.Where("id = ?", taskID).First(&task)
	if result.Error != nil {
		return nil, result.Error
	}
	return &task, nil
}

func FindAllTasks(statusFilter string) ([]models.Task, error) {
	var tasks []models.Task
	query := db.DB

	if statusFilter != "" {
		query = query.Where("status = ?", statusFilter)
	}

	result := query.Find(&tasks)
	return tasks, result.Error
}

func FindTasksByStudentID(studentID uuid.UUID) ([]models.Task, error) {
	var tasks []models.Task
	result := db.DB.Where("assigned_student_id = ?", studentID).Find(&tasks)
	return tasks, result.Error
}

func UpdateTask(task *models.Task) error {
	return db.DB.Save(task).Error
}

func DeleteTask(taskID uuid.UUID) error {
	return db.DB.Where("id = ?", taskID).Delete(&models.Task{}).Error
}

func UpdateTaskInTransaction(taskID uuid.UUID, updateFunc func(task *models.Task, tx *gorm.DB) error) error {
	return db.DB.Transaction(func(transaction *gorm.DB) error {
		var task models.Task
		if err := transaction.Where("id = ?", taskID).First(&task).Error; err != nil {
			return err
		}
		if err := updateFunc(&task, transaction); err != nil {
			return err
		}
		return transaction.Save(&task).Error
	})
}
