package service

import (
	"testing"

	"platform-intern-growth/models"
)

// Тест T-3: Проверка расчёта статуса компетенции при смене статуса задачи.

func TestCalculateCompetenceStatus_AllCompleted_ReturnsCovered(t *testing.T) {
	statuses := []models.TaskStatus{
		models.StatusCompleted,
		models.StatusCompleted,
	}

	result := calculateCompetenceStatus(statuses)

	if result != models.CompetenceStatusCovered {
		t.Errorf("Ожидался статус Covered, получен: %s", result)
	}
}

func TestCalculateCompetenceStatus_OneInProgress_ReturnsInProgress(t *testing.T) {
	statuses := []models.TaskStatus{
		models.StatusInProgress,
		models.StatusCompleted,
	}

	result := calculateCompetenceStatus(statuses)

	if result != models.CompetenceStatusInProgress {
		t.Errorf("Ожидался статус In Progress, получен: %s", result)
	}
}

func TestCalculateCompetenceStatus_OneBlocked_ReturnsInProgress(t *testing.T) {
	statuses := []models.TaskStatus{
		models.StatusBlocked,
	}

	result := calculateCompetenceStatus(statuses)

	if result != models.CompetenceStatusInProgress {
		t.Errorf("Заблокированная задача должна давать статус In Progress, получен: %s", result)
	}
}

func TestCalculateCompetenceStatus_OneNeedsHelp_ReturnsInProgress(t *testing.T) {
	statuses := []models.TaskStatus{
		models.StatusNeedsHelp,
	}

	result := calculateCompetenceStatus(statuses)

	if result != models.CompetenceStatusInProgress {
		t.Errorf("Задача NeedsHelp должна давать статус In Progress, получен: %s", result)
	}
}

func TestCalculateCompetenceStatus_AllNew_ReturnsNotCovered(t *testing.T) {
	statuses := []models.TaskStatus{
		models.StatusNew,
		models.StatusNew,
	}

	result := calculateCompetenceStatus(statuses)

	if result != models.CompetenceStatusNotCovered {
		t.Errorf("Все задачи в статусе New — должен быть Not Covered, получен: %s", result)
	}
}

func TestCalculateCompetenceStatus_EmptyStatuses_ReturnsNotCovered(t *testing.T) {
	statuses := []models.TaskStatus{}

	result := calculateCompetenceStatus(statuses)

	if result != models.CompetenceStatusNotCovered {
		t.Errorf("Пустой список статусов — должен быть Not Covered, получен: %s", result)
	}
}

func TestCalculateMatrix_NoTasks_AllNotCovered(t *testing.T) {
	allCompetences := []models.Competence{
		{Name: "Go programming"},
		{Name: "PostgreSQL"},
	}
	studentTasks := []models.Task{} // задач нет

	result := calculateMatrix(allCompetences, studentTasks)

	if len(result) != 2 {
		t.Fatalf("Ожидалось 2 элемента в матрице, получено: %d", len(result))
	}

	for _, item := range result {
		if item.Status != models.CompetenceStatusNotCovered {
			t.Errorf("Без задач все компетенции должны быть Not Covered, получен: %s для %s", item.Status, item.Name)
		}
	}
}

func TestCalculateMatrix_CompletedTask_RelatedCompetenceCovered(t *testing.T) {
	allCompetences := []models.Competence{
		{Name: "Go programming"},
		{Name: "Docker basics"},
	}

	studentTasks := []models.Task{
		{
			Status:      models.StatusCompleted,
			Competences: []string{"Go programming"},
		},
	}

	result := calculateMatrix(allCompetences, studentTasks)

	statusMap := make(map[string]models.CompetenceStatus)
	for _, item := range result {
		statusMap[item.Name] = item.Status
	}

	if statusMap["Go programming"] != models.CompetenceStatusCovered {
		t.Errorf("Go programming должна быть Covered после выполнения задачи, получен: %s", statusMap["Go programming"])
	}

	if statusMap["Docker basics"] != models.CompetenceStatusNotCovered {
		t.Errorf("Docker basics без задач должна быть Not Covered, получен: %s", statusMap["Docker basics"])
	}
}

func TestCalculateMatrix_MixedStatuses_CorrectResult(t *testing.T) {
	allCompetences := []models.Competence{
		{Name: "Go programming"},
	}

	// Одна задача выполнена, другая в процессе — итог: In Progress (не все Completed)
	studentTasks := []models.Task{
		{Status: models.StatusCompleted, Competences: []string{"Go programming"}},
		{Status: models.StatusInProgress, Competences: []string{"Go programming"}},
	}

	result := calculateMatrix(allCompetences, studentTasks)

	if len(result) != 1 {
		t.Fatalf("Ожидался 1 элемент, получено: %d", len(result))
	}

	if result[0].Status != models.CompetenceStatusInProgress {
		t.Errorf("При смешанных статусах ожидается In Progress, получен: %s", result[0].Status)
	}
}
