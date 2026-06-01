package models

import "github.com/google/uuid"

type Competence struct {
	ID   uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name string    `gorm:"uniqueIndex;not null" json:"name"`
}

type CompetenceStatus string

const (
	CompetenceStatusCovered    CompetenceStatus = "Covered"
	CompetenceStatusInProgress CompetenceStatus = "In Progress"
	CompetenceStatusNotCovered CompetenceStatus = "Not Covered"
)

type CompetenceMatrixItem struct {
	Name   string           `json:"name"`
	Status CompetenceStatus `json:"status"`
}

type StudentMatrix struct {
	StudentID   uuid.UUID              `json:"student_id"`
	StudentName string                 `json:"student_name"`
	Competences []CompetenceMatrixItem `json:"competences"`
}
