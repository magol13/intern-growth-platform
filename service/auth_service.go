package service

import (
	"platform-intern-growth/models"
	"platform-intern-growth/repository"
)

type AuthResult struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	Name   string `json:"name"`
}

func ValidateToken(token string) (*AuthResult, error) {
	user, err := repository.FindUserByToken(token)
	if err != nil {
		return nil, err
	}

	result := &AuthResult{
		UserID: user.ID.String(),
		Role:   string(user.Role),
		Name:   user.Name,
	}

	return result, nil
}

func GetCurrentUser(token string) (*models.User, error) {
	return repository.FindUserByToken(token)
}
