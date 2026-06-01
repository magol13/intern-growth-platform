package service

import (
	"errors"
	"fmt"
	"time"

	"platform-intern-growth/models"
	"platform-intern-growth/repository"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type CreateTaskInput struct {
	Title             string   `json:"title"               binding:"required,max=255"`
	Description       string   `json:"description"`
	Deadline          string   `json:"deadline"            binding:"required"`
	AssignedStudentID string   `json:"assigned_student_id" binding:"required,uuid"`
	Competences       []string `json:"competences"         binding:"required,min=1"`
}

type UpdateTaskInput struct {
	Title             *string  `json:"title"`
	Description       *string  `json:"description"`
	Deadline          *string  `json:"deadline"`
	AssignedStudentID *string  `json:"assigned_student_id"`
	Competences       []string `json:"competences"`
	Status            *string  `json:"status"`
}

func CreateTask(mentorID uuid.UUID, input CreateTaskInput) (*models.Task, error) {
	deadline, err := time.Parse("2006-01-02", input.Deadline)
	if err != nil {
		return nil, fmt.Errorf("неверный формат даты дедлайна, используйте YYYY-MM-DD")
	}

	if deadline.Before(time.Now()) {
		return nil, fmt.Errorf("дедлайн должен быть позже даты создания")
	}

	studentID, err := uuid.Parse(input.AssignedStudentID)
	if err != nil {
		return nil, fmt.Errorf("неверный формат UUID стажёра")
	}

	student, err := repository.FindUserByID(studentID)
	if err != nil {
		return nil, fmt.Errorf("стажёр с указанным ID не найден")
	}
	if student.Role != models.RoleIntern {
		return nil, fmt.Errorf("указанный пользователь не является стажёром")
	}

	for _, competenceName := range input.Competences {
		_, err := repository.FindCompetenceByName(competenceName)
		if err != nil {
			return nil, fmt.Errorf("компетенция '%s' не найдена в справочнике", competenceName)
		}
	}

	task := &models.Task{
		ID:                uuid.New(),
		Title:             input.Title,
		Description:       input.Description,
		Deadline:          deadline,
		AssignedStudentID: studentID,
		MentorID:          mentorID,
		Status:            models.StatusNew,
		Competences:       pq.StringArray(input.Competences),
		Artefacts:         pq.StringArray{},
	}

	err = repository.CreateTask(task)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать задачу: %w", err)
	}

	return task, nil
}

func GetAllTasks(statusFilter string) ([]models.Task, error) {
	if statusFilter != "" {
		if !models.AllowedStatuses[models.TaskStatus(statusFilter)] {
			return nil, fmt.Errorf("недопустимый статус фильтра: %s", statusFilter)
		}
	}
	return repository.FindAllTasks(statusFilter)
}

func UpdateTask(taskID uuid.UUID, mentorID uuid.UUID, input UpdateTaskInput) (*models.Task, error) {
	task, err := repository.FindTaskByID(taskID)
	if err != nil {
		return nil, fmt.Errorf("задача не найдена")
	}

	if task.MentorID != mentorID {
		return nil, fmt.Errorf("вы можете редактировать только свои задачи")
	}

	if input.Title != nil {
		task.Title = *input.Title
	}
	if input.Description != nil {
		task.Description = *input.Description
	}
	if input.Deadline != nil {
		deadline, err := time.Parse("2006-01-02", *input.Deadline)
		if err != nil {
			return nil, fmt.Errorf("неверный формат даты дедлайна, используйте YYYY-MM-DD")
		}
		task.Deadline = deadline
	}
	if input.AssignedStudentID != nil {
		studentID, err := uuid.Parse(*input.AssignedStudentID)
		if err != nil {
			return nil, fmt.Errorf("неверный формат UUID стажёра")
		}
		task.AssignedStudentID = studentID
	}
	if len(input.Competences) > 0 {
		task.Competences = pq.StringArray(input.Competences)
	}
	if input.Status != nil {
		newStatus := models.TaskStatus(*input.Status)
		if !models.AllowedStatuses[newStatus] {
			return nil, fmt.Errorf("недопустимый статус: %s", *input.Status)
		}
		task.Status = newStatus
	}

	err = repository.UpdateTask(task)
	if err != nil {
		return nil, fmt.Errorf("не удалось обновить задачу: %w", err)
	}

	return task, nil
}

func DeleteTask(taskID uuid.UUID, mentorID uuid.UUID) error {
	task, err := repository.FindTaskByID(taskID)
	if err != nil {
		return fmt.Errorf("задача не найдена")
	}

	if task.MentorID != mentorID {
		return fmt.Errorf("вы можете удалять только свои задачи")
	}

	return repository.DeleteTask(taskID)
}

func UpdateInternTaskStatus(taskID uuid.UUID, internID uuid.UUID, newStatus models.TaskStatus) (*models.Task, error) {
	if !models.AllowedStatuses[newStatus] {
		return nil, fmt.Errorf("недопустимый статус: %s", newStatus)
	}

	var updatedTask *models.Task

	err := repository.UpdateTaskInTransaction(taskID, func(task *models.Task, tx *gorm.DB) error {
		if task.AssignedStudentID != internID {
			return fmt.Errorf("эта задача назначена другому стажёру")
		}

		if newStatus == models.StatusNeedsHelp && task.Status != models.StatusInProgress {
			return fmt.Errorf("статус NeedsHelp можно установить только из статуса InProgress")
		}

		task.Status = newStatus
		updatedTask = task
		return nil
	})

	return updatedTask, err
}

func AddArtefactToTask(taskID uuid.UUID, internID uuid.UUID, artefactURL string) (*models.Task, error) {
	if artefactURL == "" {
		return nil, errors.New("URL артефакта не может быть пустым")
	}

	var updatedTask *models.Task

	err := repository.UpdateTaskInTransaction(taskID, func(task *models.Task, tx *gorm.DB) error {
		if task.AssignedStudentID != internID {
			return fmt.Errorf("эта задача назначена другому стажёру")
		}

		task.Artefacts = append(task.Artefacts, artefactURL)
		updatedTask = task
		return nil
	})

	return updatedTask, err
}

func UpdateTaskComment(taskID uuid.UUID, internID uuid.UUID, comment string) (*models.Task, error) {
	var updatedTask *models.Task

	err := repository.UpdateTaskInTransaction(taskID, func(task *models.Task, tx *gorm.DB) error {
		if task.AssignedStudentID != internID {
			return fmt.Errorf("эта задача назначена другому стажёру")
		}

		task.Comments = comment
		updatedTask = task
		return nil
	})

	return updatedTask, err
}
