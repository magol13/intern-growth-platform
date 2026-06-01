package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type TaskStatus string

const (
	StatusNew        TaskStatus = "New"
	StatusInProgress TaskStatus = "InProgress"
	StatusBlocked    TaskStatus = "Blocked"
	StatusNeedsHelp  TaskStatus = "NeedsHelp"
	StatusPaused     TaskStatus = "Paused"
	StatusCompleted  TaskStatus = "Completed"
)

var AllowedStatuses = map[TaskStatus]bool{
	StatusNew:        true,
	StatusInProgress: true,
	StatusBlocked:    true,
	StatusNeedsHelp:  true,
	StatusPaused:     true,
	StatusCompleted:  true,
}

type Task struct {
	ID                uuid.UUID      `gorm:"type:uuid;primaryKey"             json:"id"`
	Title             string         `gorm:"size:255;not null"                json:"title"`
	Description       string         `gorm:"type:text"                        json:"description"`
	Deadline          time.Time      `gorm:"not null"                         json:"deadline"`
	AssignedStudentID uuid.UUID      `gorm:"type:uuid;not null"               json:"assigned_student_id"`
	MentorID          uuid.UUID      `gorm:"type:uuid;not null"               json:"mentor_id"`
	Status            TaskStatus     `gorm:"type:varchar(20);not null;default:'New'" json:"status"`
	Competences       pq.StringArray `gorm:"type:text[]"                      json:"competences"`
	Artefacts         pq.StringArray `gorm:"type:text[]"                      json:"artefacts"`
	Comments          string         `gorm:"type:text"                        json:"comments"`
	CreatedAt         time.Time      `                                        json:"created_at"`
	UpdatedAt         time.Time      `                                        json:"updated_at"`
}

type InternTaskView struct {
	ID          uuid.UUID      `json:"id"`
	Title       string         `json:"title"`
	Status      TaskStatus     `json:"status"`
	Deadline    time.Time      `json:"deadline"`
	Competences pq.StringArray `json:"competences"`
}
