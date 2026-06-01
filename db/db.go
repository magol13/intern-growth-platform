package db

import (
	"fmt"
	"log"
	"os"

	"platform-intern-growth/models"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init() {
	host := getEnvOrDefault("DB_HOST", "localhost")
	port := getEnvOrDefault("DB_PORT", "5432")

	// На macOS (Homebrew) PostgreSQL создаёт роль с именем системного пользователя.
	// os.Getenv("USER") автоматически вернёт твоё имя (например, "maria").
	// Переопредели через переменную окружения DB_USER если нужно другое имя.
	systemUser := os.Getenv("USER")
	if systemUser == "" {
		systemUser = "postgres"
	}
	user := getEnvOrDefault("DB_USER", systemUser)

	password := os.Getenv("DB_PASSWORD") // пустой пароль — норма для локального Homebrew PG
	dbname := getEnvOrDefault("DB_NAME", "intern_platform")

	var dsn string
	if password == "" {
		// Без пароля — стандартная конфигурация macOS Homebrew
		dsn = fmt.Sprintf(
			"host=%s port=%s user=%s dbname=%s sslmode=disable TimeZone=Europe/Moscow",
			host, port, user, dbname,
		)
	} else {
		dsn = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=Europe/Moscow",
			host, port, user, password, dbname,
		)
	}

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v\n\nПодсказка: убедись, что PostgreSQL запущен и база данных создана:\n  createdb intern_platform\n", err)
	}

	err = database.AutoMigrate(
		&models.User{},
		&models.Competence{},
		&models.Task{},
	)
	if err != nil {
		log.Fatalf("Не удалось выполнить миграцию таблиц: %v", err)
	}

	DB = database
	log.Println("Подключение к базе данных успешно установлено")

	seedTestData(database)
}

func seedTestData(database *gorm.DB) {
	var userCount int64
	database.Model(&models.User{}).Count(&userCount)
	if userCount > 0 {
		return
	}

	mentorID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	intern1ID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	intern2ID := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	users := []models.User{
		{ID: mentorID, Name: "Мария Горбачёва", Role: models.RoleMentor, Token: "mentor-secret-token-abc"},
		{ID: intern1ID, Name: "Артём Жуков", Role: models.RoleIntern, Token: "intern1-secret-token-xyz"},
		{ID: intern2ID, Name: "Анна Смирнова", Role: models.RoleIntern, Token: "intern2-secret-token-qwe"},
	}

	for i := range users {
		database.Create(&users[i])
	}

	competences := []models.Competence{
		{ID: uuid.New(), Name: "Go programming"},
		{ID: uuid.New(), Name: "PostgreSQL"},
		{ID: uuid.New(), Name: "REST API"},
		{ID: uuid.New(), Name: "Docker basics"},
		{ID: uuid.New(), Name: "Git"},
	}

	for i := range competences {
		database.Create(&competences[i])
	}

	log.Println("Тестовые данные успешно добавлены в базу данных")
}

func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
