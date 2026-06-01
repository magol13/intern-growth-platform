package repository

import (
	"platform-intern-growth/db"
	"platform-intern-growth/models"
)

func CreateCompetence(competence *models.Competence) error {
	return db.DB.Create(competence).Error
}

func FindAllCompetences() ([]models.Competence, error) {
	var competences []models.Competence
	result := db.DB.Find(&competences)
	return competences, result.Error
}

func FindCompetenceByName(name string) (*models.Competence, error) {
	var competence models.Competence
	result := db.DB.Where("name = ?", name).First(&competence)
	if result.Error != nil {
		return nil, result.Error
	}
	return &competence, nil
}
