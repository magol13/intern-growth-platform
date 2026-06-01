package models

import "github.com/google/uuid"

type Role string

const (
	RoleMentor Role = "mentor"
	RoleIntern Role = "intern"
)

type User struct {
	ID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name  string    `gorm:"size:255;not null"   json:"name"`
	Role  Role      `gorm:"type:varchar(10);not null" json:"role"`
	Token string    `gorm:"uniqueIndex;not null" json:"-"`
}
