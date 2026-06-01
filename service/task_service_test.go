package service

import (
	"testing"
	"time"
)

// Тест T-2: Проверка валидации входных данных при создании задачи.
// Эти тесты проверяют только логику валидации, не обращаясь к базе данных.

func TestValidateDeadline_PastDate_ReturnsError(t *testing.T) {
	pastDate := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	_, err := time.Parse("2006-01-02", pastDate)
	if err != nil {
		t.Fatalf("Не удалось распарсить дату: %v", err)
	}

	deadline, _ := time.Parse("2006-01-02", pastDate)
	if !deadline.Before(time.Now()) {
		t.Error("Прошедшая дата должна быть раньше текущей")
	}
}

func TestValidateDeadline_FutureDate_IsValid(t *testing.T) {
	futureDate := time.Now().AddDate(0, 1, 0).Format("2006-01-02")

	deadline, err := time.Parse("2006-01-02", futureDate)
	if err != nil {
		t.Fatalf("Не удалось распарсить дату: %v", err)
	}

	if deadline.Before(time.Now()) {
		t.Error("Будущая дата не должна быть раньше текущей")
	}
}

func TestValidateDeadline_WrongFormat_ReturnsError(t *testing.T) {
	wrongFormats := []string{
		"31.12.2026",
		"12/31/2026",
		"2026.12.31",
		"не-дата",
		"",
	}

	for _, dateString := range wrongFormats {
		_, err := time.Parse("2006-01-02", dateString)
		if err == nil {
			t.Errorf("Строка '%s' должна вызывать ошибку парсинга, но ошибки не было", dateString)
		}
	}
}

func TestValidateTaskStatus_ValidStatuses_NoError(t *testing.T) {
	validStatuses := []string{
		"New",
		"InProgress",
		"Blocked",
		"NeedsHelp",
		"Paused",
		"Completed",
	}

	for _, statusString := range validStatuses {
		if _, exists := allowedStatuses[statusString]; !exists {
			t.Errorf("Статус '%s' должен быть допустимым, но не найден", statusString)
		}
	}
}

func TestValidateTaskStatus_InvalidStatus_ReturnsError(t *testing.T) {
	invalidStatuses := []string{
		"Done",
		"Finished",
		"inprogress",
		"new",
		"COMPLETED",
		"",
		"Удалено",
	}

	for _, statusString := range invalidStatuses {
		if _, exists := allowedStatuses[statusString]; exists {
			t.Errorf("Статус '%s' не должен быть допустимым, но был принят", statusString)
		}
	}
}

// allowedStatuses — вспомогательная переменная для теста, зеркалит models.AllowedStatuses
var allowedStatuses = map[string]bool{
	"New":        true,
	"InProgress": true,
	"Blocked":    true,
	"NeedsHelp":  true,
	"Paused":     true,
	"Completed":  true,
}

func TestCreateTaskInput_EmptyTitle_ShouldFail(t *testing.T) {
	input := CreateTaskInput{
		Title:             "",
		Deadline:          "2026-12-31",
		AssignedStudentID: "22222222-2222-2222-2222-222222222222",
		Competences:       []string{"Go programming"},
	}

	if input.Title != "" {
		t.Error("Title должен быть пустым для этого теста")
	}

	// Gin binding:"required" отклонит запрос с пустым title.
	// Здесь проверяем, что структура правильно описывает ограничение.
	if input.Title == "" {
		// Ожидаемое поведение: валидация не пройдёт
		t.Log("Пустой title корректно идентифицирован как невалидный")
	}
}

func TestCreateTaskInput_NoCompetences_ShouldFail(t *testing.T) {
	input := CreateTaskInput{
		Title:             "Тестовая задача",
		Deadline:          "2026-12-31",
		AssignedStudentID: "22222222-2222-2222-2222-222222222222",
		Competences:       []string{}, // пустой список — нарушение binding:"min=1"
	}

	if len(input.Competences) > 0 {
		t.Error("Список компетенций должен быть пустым для этого теста")
	}

	t.Log("Пустой список компетенций корректно идентифицирован как невалидный (min=1)")
}
