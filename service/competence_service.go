package service

import (
        "fmt"

        "platform-intern-growth/models"
        "platform-intern-growth/repository"

        "github.com/google/uuid"
)

func CreateCompetence(name string) (*models.Competence, error) {
        if name == "" {
                return nil, fmt.Errorf("название компетенции не может быть пустым")
        }

        existing, _ := repository.FindCompetenceByName(name)
        if existing != nil {
                return nil, fmt.Errorf("компетенция с названием '%s' уже существует", name)
        }

        competence := &models.Competence{
                ID:   uuid.New(),
                Name: name,
        }

        err := repository.CreateCompetence(competence)
        if err != nil {
                return nil, fmt.Errorf("не удалось создать компетенцию: %w", err)
        }

        return competence, nil
}

func GetStudentMatrix(studentID uuid.UUID) (*models.StudentMatrix, error) {
        student, err := repository.FindUserByID(studentID)
        if err != nil {
                return nil, fmt.Errorf("стажёр не найден")
        }

        if student.Role != models.RoleIntern {
                return nil, fmt.Errorf("указанный пользователь не является стажёром")
        }

        allCompetences, err := repository.FindAllCompetences()
        if err != nil {
                return nil, fmt.Errorf("не удалось получить справочник компетенций")
        }

        studentTasks, err := repository.FindTasksByStudentID(studentID)
        if err != nil {
                return nil, fmt.Errorf("не удалось получить задачи стажёра")
        }

        matrixItems := calculateMatrix(allCompetences, studentTasks)

        matrix := &models.StudentMatrix{
                StudentID:   studentID,
                StudentName: student.Name,
                Competences: matrixItems,
        }

        return matrix, nil
}

func calculateMatrix(allCompetences []models.Competence, studentTasks []models.Task) []models.CompetenceMatrixItem {
        competenceTaskMap := make(map[string][]models.TaskStatus)

        for _, task := range studentTasks {
                for _, competenceName := range task.Competences {
                        competenceTaskMap[competenceName] = append(competenceTaskMap[competenceName], task.Status)
                }
        }

        var matrixItems []models.CompetenceMatrixItem

        for _, competence := range allCompetences {
                statuses, hasTask := competenceTaskMap[competence.Name]

                var competenceStatus models.CompetenceStatus

                if !hasTask || len(statuses) == 0 {
                        competenceStatus = models.CompetenceStatusNotCovered
                } else {
                        competenceStatus = calculateCompetenceStatus(statuses)
                }

                matrixItems = append(matrixItems, models.CompetenceMatrixItem{
                        Name:   competence.Name,
                        Status: competenceStatus,
                })
        }

        return matrixItems
}

func calculateCompetenceStatus(statuses []models.TaskStatus) models.CompetenceStatus {
        if len(statuses) == 0 {
                return models.CompetenceStatusNotCovered
        }

        allCompleted := true
        hasActiveTask := false

        for _, status := range statuses {
                if status != models.StatusCompleted {
                        allCompleted = false
                }

                if status == models.StatusInProgress ||
                        status == models.StatusBlocked ||
                        status == models.StatusNeedsHelp ||
                        status == models.StatusPaused {
                        hasActiveTask = true
                }
        }

        if allCompleted {
                return models.CompetenceStatusCovered
        }

        if hasActiveTask {
                return models.CompetenceStatusInProgress
        }

        return models.CompetenceStatusNotCovered
}
