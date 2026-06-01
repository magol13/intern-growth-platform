package repository

import (
	"platform-intern-growth/db"
	"platform-intern-growth/models"

	"github.com/google/uuid"
)

func FindUserByToken(token string) (*models.User, error) {
	var user models.User
	result := db.DB.Where("token = ?", token).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func FindUserByID(userID uuid.UUID) (*models.User, error) {
	var user models.User
	result := db.DB.Where("id = ?", userID).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func FindAllInterns() ([]models.User, error) {
	var interns []models.User
	result := db.DB.Where("role = ?", models.RoleIntern).Find(&interns)
	return interns, result.Error
}
