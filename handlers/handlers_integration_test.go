//go:build integration

package handlers

// Интеграционные тесты — запускаются с тегом:
//   go test ./handlers/ -tags=integration
//
// Требуют запущенного PostgreSQL и переменной окружения:
//   TEST_DB_URL=postgres://user@localhost/intern_platform_test
// или отдельных переменных DB_HOST, DB_USER, DB_NAME=intern_platform_test
//
// Создать тестовую базу:
//   createdb intern_platform_test

import (
        "bytes"
        "encoding/json"
        "fmt"
        "net/http"
        "net/http/httptest"
        "os"
        "testing"

        "platform-intern-growth/db"
        "platform-intern-growth/middleware"

        "github.com/gin-gonic/gin"
        "github.com/google/uuid"
)

var testRouter *gin.Engine

func TestMain(m *testing.M) {
        gin.SetMode(gin.TestMode)

        // Подключаемся к тестовой базе данных
        testDBName := os.Getenv("DB_NAME")
        if testDBName == "" {
                os.Setenv("DB_NAME", "intern_platform_test")
        }

        db.Init()
        testRouter = setupTestRouter()

        exitCode := m.Run()
        os.Exit(exitCode)
}

func setupTestRouter() *gin.Engine {
        router := gin.New()
        router.Use(gin.Recovery())

        router.GET("/health", func(ctx *gin.Context) {
                ctx.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"status": "ok"}})
        })

        router.GET("/auth/validate", ValidateToken)

        protected := router.Group("/")
        protected.Use(middleware.AuthRequired())
        {
                mentor := protected.Group("/")
                mentor.Use(middleware.MentorOnly())
                {
                        mentor.GET("/tasks", GetAllTasks)
                        mentor.POST("/tasks", CreateTask)
                        mentor.PUT("/tasks/:id", UpdateTask)
                        mentor.DELETE("/tasks/:id", DeleteTask)
                        mentor.POST("/competences", CreateCompetence)
                        mentor.GET("/students/:id/matrix", GetStudentMatrix)
                        mentor.GET("/students/:id/report", GetMentorStudentReport)
                }

                intern := protected.Group("/")
                intern.Use(middleware.InternOnly())
                {
                        intern.GET("/my-tasks", GetMyTasks)
                        intern.PATCH("/my-tasks/:id/status", UpdateMyTaskStatus)
                        intern.PATCH("/my-tasks/:id/artefacts", AddMyTaskArtefact)
                        intern.PATCH("/my-tasks/:id/comment", UpdateMyTaskComment)
                        intern.GET("/my-matrix", GetMyMatrix)
                }
        }

        return router
}

const (
        mentorToken  = "mentor-secret-token-abc"
        intern1Token = "intern1-secret-token-xyz"
        intern2Token = "intern2-secret-token-qwe"
        intern1ID    = "22222222-2222-2222-2222-222222222222"
)

func makeRequest(method, path, token string, body interface{}) *httptest.ResponseRecorder {
        var requestBody *bytes.Buffer
        if body != nil {
                bodyBytes, _ := json.Marshal(body)
                requestBody = bytes.NewBuffer(bodyBytes)
        } else {
                requestBody = bytes.NewBuffer(nil)
        }

        request := httptest.NewRequest(method, path, requestBody)
        if token != "" {
                request.Header.Set("Authorization", "Bearer "+token)
        }
        if body != nil {
                request.Header.Set("Content-Type", "application/json")
        }

        recorder := httptest.NewRecorder()
        testRouter.ServeHTTP(recorder, request)
        return recorder
}

// T-1: Проверка валидации токена

func TestAuthValidate_ValidToken_Returns200(t *testing.T) {
        recorder := makeRequest(http.MethodGet, "/auth/validate", mentorToken, nil)

        if recorder.Code != http.StatusOK {
                t.Errorf("Корректный токен → ожидался 200, получен: %d | тело: %s", recorder.Code, recorder.Body.String())
        }

        var response map[string]interface{}
        json.Unmarshal(recorder.Body.Bytes(), &response)

        if response["success"] != true {
                t.Error("Поле success должно быть true")
        }
        data := response["data"].(map[string]interface{})
        if data["role"] != "mentor" {
                t.Errorf("Ожидалась роль mentor, получена: %v", data["role"])
        }
}

func TestAuthValidate_InvalidToken_Returns401(t *testing.T) {
        recorder := makeRequest(http.MethodGet, "/auth/validate", "несуществующий-токен", nil)

        if recorder.Code != http.StatusUnauthorized {
                t.Errorf("Несуществующий токен → ожидался 401, получен: %d", recorder.Code)
        }
}

func TestAuthValidate_EmptyHeader_Returns401(t *testing.T) {
        recorder := makeRequest(http.MethodGet, "/auth/validate", "", nil) // без токена

        // Хендлер сам проверяет заголовок
        if recorder.Code != http.StatusUnauthorized {
                t.Errorf("Пустой заголовок → ожидался 401, получен: %d", recorder.Code)
        }
}

// T-2: Проверка создания задачи

func TestCreateTask_AllRequiredFields_Returns201(t *testing.T) {
        body := map[string]interface{}{
                "title":               "Тестовая задача для юнит-теста",
                "description":         "Описание тестовой задачи",
                "deadline":            "2026-12-31",
                "assigned_student_id": intern1ID,
                "competences":         []string{"Go programming"},
        }

        recorder := makeRequest(http.MethodPost, "/tasks", mentorToken, body)

        if recorder.Code != http.StatusCreated {
                t.Errorf("Все обязательные поля → ожидался 201, получен: %d | тело: %s", recorder.Code, recorder.Body.String())
        }

        var response map[string]interface{}
        json.Unmarshal(recorder.Body.Bytes(), &response)

        if response["success"] != true {
                t.Errorf("Ожидался success: true, получено: %v", response["success"])
        }

        data := response["data"].(map[string]interface{})
        if data["status"] != "New" {
                t.Errorf("Новая задача должна иметь статус New, получен: %v", data["status"])
        }
}

func TestCreateTask_MissingTitle_Returns400(t *testing.T) {
        body := map[string]interface{}{
                // title намеренно отсутствует
                "deadline":            "2026-12-31",
                "assigned_student_id": intern1ID,
                "competences":         []string{"Go programming"},
        }

        recorder := makeRequest(http.MethodPost, "/tasks", mentorToken, body)

        if recorder.Code != http.StatusBadRequest {
                t.Errorf("Отсутствующий title → ожидался 400, получен: %d", recorder.Code)
        }
}

func TestCreateTask_MissingCompetences_Returns400(t *testing.T) {
        body := map[string]interface{}{
                "title":               "Задача без компетенций",
                "deadline":            "2026-12-31",
                "assigned_student_id": intern1ID,
                // competences намеренно отсутствует
        }

        recorder := makeRequest(http.MethodPost, "/tasks", mentorToken, body)

        if recorder.Code != http.StatusBadRequest {
                t.Errorf("Отсутствующие competences → ожидался 400, получен: %d", recorder.Code)
        }
}

func TestCreateTask_AsIntern_Returns403(t *testing.T) {
        body := map[string]interface{}{
                "title":               "Стажёр пытается создать задачу",
                "deadline":            "2026-12-31",
                "assigned_student_id": intern1ID,
                "competences":         []string{"Go programming"},
        }

        recorder := makeRequest(http.MethodPost, "/tasks", intern1Token, body)

        if recorder.Code != http.StatusForbidden {
                t.Errorf("Стажёр не должен создавать задачи → ожидался 403, получен: %d", recorder.Code)
        }
}

// Сценарий C-1: Ментор создал задачу → Стажёр видит её в GET /my-tasks

func TestScenarioC1_MentorCreatesTask_InternSeesIt(t *testing.T) {
        // Шаг 1: Ментор создаёт задачу
        createBody := map[string]interface{}{
                "title":               fmt.Sprintf("Сценарий C1 — задача %s", uuid.New().String()[:8]),
                "deadline":            "2026-12-31",
                "assigned_student_id": intern1ID,
                "competences":         []string{"Go programming"},
        }
        createRecorder := makeRequest(http.MethodPost, "/tasks", mentorToken, createBody)
        if createRecorder.Code != http.StatusCreated {
                t.Fatalf("Не удалось создать задачу: %d — %s", createRecorder.Code, createRecorder.Body.String())
        }

        var createResponse map[string]interface{}
        json.Unmarshal(createRecorder.Body.Bytes(), &createResponse)
        createdTitle := createResponse["data"].(map[string]interface{})["title"].(string)

        // Шаг 2: Стажёр запрашивает свои задачи
        myTasksRecorder := makeRequest(http.MethodGet, "/my-tasks", intern1Token, nil)
        if myTasksRecorder.Code != http.StatusOK {
                t.Fatalf("Стажёр не смог получить список задач: %d", myTasksRecorder.Code)
        }

        var myTasksResponse map[string]interface{}
        json.Unmarshal(myTasksRecorder.Body.Bytes(), &myTasksResponse)
        tasks := myTasksResponse["data"].([]interface{})

        found := false
        for _, rawTask := range tasks {
                task := rawTask.(map[string]interface{})
                if task["title"] == createdTitle {
                        found = true
                        break
                }
        }

        if !found {
                t.Errorf("Созданная задача '%s' не найдена в списке задач стажёра", createdTitle)
        }
}

// Сценарий C-3: Стажёр с чужим токеном получает 401

func TestScenarioC3_WrongToken_Returns401(t *testing.T) {
        recorder := makeRequest(http.MethodGet, "/my-tasks", "чужой-несуществующий-токен", nil)

        if recorder.Code != http.StatusUnauthorized {
                t.Errorf("Чужой токен → ожидался 401, получен: %d", recorder.Code)
        }
}

// Сценарий C-5: Ментор фильтрует задачи по статусу NeedsHelp

func TestScenarioC5_FilterByNeedsHelp_ReturnsOnlyNeedsHelp(t *testing.T) {
        recorder := makeRequest(http.MethodGet, "/tasks?status=NeedsHelp", mentorToken, nil)

        if recorder.Code != http.StatusOK {
                t.Fatalf("Фильтрация задач → ожидался 200, получен: %d", recorder.Code)
        }

        var response map[string]interface{}
        json.Unmarshal(recorder.Body.Bytes(), &response)

        tasks, ok := response["data"].([]interface{})
        if !ok {
                // data может быть nil если задач нет — это нормально
                t.Log("Нет задач со статусом NeedsHelp — ответ корректен")
                return
        }

        for _, rawTask := range tasks {
                task := rawTask.(map[string]interface{})
                if task["status"] != "NeedsHelp" {
                        t.Errorf("В отфильтрованном списке найдена задача со статусом %v, ожидался только NeedsHelp", task["status"])
                }
        }
}
