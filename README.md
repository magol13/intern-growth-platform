# Platform Intern Growth — Backend на Go

Учебный проект платформы управления ростом стажёров. Backend на Go + Gin + PostgreSQL.

## Структура проекта

```
platform-intern-growth/
├── main.go                    # Точка входа, настройка роутинга
├── go.mod                     # Зависимости
├── db/
│   └── db.go                  # Подключение к PostgreSQL, AutoMigrate, сид-данные
├── models/
│   ├── user.go                # Модель пользователя (ментор/стажёр)
│   ├── task.go                # Модель задачи и статусы
│   └── competence.go          # Модель компетенции и матрица
├── middleware/
│   ├── auth.go                # Авторизация по Bearer-токену, проверка ролей
│   └── logger.go              # Логирование HTTP-запросов
├── repository/
│   ├── user_repo.go           # Запросы к таблице users
│   ├── task_repo.go           # Запросы к таблице tasks
│   └── competence_repo.go     # Запросы к таблице competences
├── service/
│   ├── auth_service.go        # Валидация токена
│   ├── task_service.go        # Бизнес-логика задач (CRUD, статусы)
│   ├── competence_service.go  # Расчёт матрицы компетенций
│   └── report_service.go      # Генерация отчётов (JSON + PDF)
└── handlers/
    ├── auth_handler.go        # GET /auth/validate
    ├── task_handler.go        # CRUD задач для ментора
    ├── intern_handler.go      # Эндпоинты стажёра
    ├── competence_handler.go  # Матрица компетенций
    └── report_handler.go      # Отчёты
```

## Требования

- Go 1.22+
- PostgreSQL (установленный через Homebrew)

## Запуск на macOS (Homebrew PostgreSQL)

### Шаг 1 — Проверка, что PostgreSQL запущен

```bash
brew services start postgresql@14
# или для другой версии:
brew services start postgresql@16
```

Проверка, что запущен:
```bash
psql postgres -c "\l"
```

### Шаг 2 — Создать базу данных


```bash
createdb intern_platform
```

Проверить что база создалась:
```bash
psql intern_platform -c "\dt"
```

### Шаг 3 — Установить зависимости Go

```bash
go mod tidy
```

### Шаг 4 — Запустить сервер

```bash
go run main.go
```

Сервер стартует на `http://localhost:8080`. При первом запуске автоматически создадутся таблицы и тестовые данные.

---

## Переменные окружения

Все параметры имеют умные умолчания для macOS. 

| Переменная    | Умолчание              | Описание                        |
|---------------|------------------------|---------------------------------|
| `DB_HOST`     | `localhost`            | Хост PostgreSQL                 |
| `DB_PORT`     | `5432`                 | Порт PostgreSQL                 |
| `DB_USER`     | текущий пользователь macOS | Имя роли в PostgreSQL       |
| `DB_PASSWORD` | _(пусто)_              | Пароль (обычно не нужен на mac) |
| `DB_NAME`     | `intern_platform`      | Имя базы данных                 |

Пример если нужно переопределить:
```bash
DB_USER=myuser DB_NAME=mydb go run main.go
```

---

## Если всё ещё ошибка подключения

```bash
# Узнать точное имя своей роли в PostgreSQL
psql postgres -c "\du"

# Запустить с явным указанием имени пользователя
DB_USER=maria go run main.go
```

---

## Тестовые токены (создаются при первом запуске)

| Роль    | Токен                        |
|---------|------------------------------|
| Ментор  | `mentor-secret-token-abc`    |
| Стажёр 1 | `intern1-secret-token-xyz` |
| Стажёр 2 | `intern2-secret-token-qwe` |

---

## API — примеры запросов

### Проверка работы сервера
```bash
curl http://localhost:8080/health
```

### Валидация токена
```bash
curl -H "Authorization: Bearer mentor-secret-token-abc" \
  http://localhost:8080/auth/validate
```

### Создать компетенцию (только ментор)
```bash
curl -X POST http://localhost:8080/competences \
  -H "Authorization: Bearer mentor-secret-token-abc" \
  -H "Content-Type: application/json" \
  -d '{"name": "React"}'
```

### Создать задачу (только ментор)
```bash
curl -X POST http://localhost:8080/tasks \
  -H "Authorization: Bearer mentor-secret-token-abc" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Реализовать REST API",
    "description": "Написать CRUD для сущности Task на Go + Gin",
    "deadline": "2026-12-31",
    "assigned_student_id": "22222222-2222-2222-2222-222222222222",
    "competences": ["Go programming", "REST API"]
  }'
```

### Список задач ментора (с фильтром по статусу)
```bash
curl -H "Authorization: Bearer mentor-secret-token-abc" \
  "http://localhost:8080/tasks?status=NeedsHelp"
```

### Добавление стажера нового
```bash
curl -X POST http://localhost:8080/students \
-H "Authorization: Bearer mentor-secret-token-abc" \
-H "Content-Type: application/json" \
-d '{"name": "Иван Петров"}'
```

### Задачи стажёра
```bash
curl -H "Authorization: Bearer intern1-secret-token-xyz" \
  http://localhost:8080/my-tasks
```

### Создаём задачу 
```bash
curl -s -X POST http://localhost:8080/tasks \
  -H "Authorization: Bearer mentor-secret-token-abc" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Реализовать REST API трекера задач",
    "description": "Написать CRUD для сущности Task на Go + Gin. Покрыть все маршруты.",
    "deadline": "2026-12-31",
    "assigned_student_id": "22222222-2222-2222-2222-222222222222",
    "competences": ["Go programming", "REST API"]
  }' | python3 -m json.tool
```

### Стажёр меняет статус задачи
```bash
curl -X PATCH "http://localhost:8080/my-tasks/<task_id>/status" \
  -H "Authorization: Bearer intern1-secret-token-xyz" \
  -H "Content-Type: application/json" \
  -d '{"status": "InProgress"}'
```

### Стажёр прикрепляет ссылку на выполненную работу
```bash
curl -s -X PATCH "http://localhost:8080/my-tasks/<task_id>/artefacts" \
-H "Authorization: Bearer intern1-secret-token-xyz" \
-H "Content-Type: application/json" \
-d '{"url": "https://github.com/artem/intern-project/pull/1"}' | python3 -m json.tool
```


### Матрица компетенций стажёра (для ментора)
```bash
curl -H "Authorization: Bearer mentor-secret-token-abc" \
  http://localhost:8080/students/22222222-2222-2222-2222-222222222222/matrix
```

### Отчёт в JSON (скачивается как файл)
```bash
curl -H "Authorization: Bearer mentor-secret-token-abc" \
  "http://localhost:8080/students/22222222-2222-2222-2222-222222222222/report?format=json" \
  --output report.json
```

### Отчёт в PDF
```bash
curl -H "Authorization: Bearer mentor-secret-token-abc" \
  "http://localhost:8080/students/22222222-2222-2222-2222-222222222222/report?format=pdf" \
  --output report.pdf
```

### Новая компетенция
```bash
curl -X POST http://localhost:8080/competences \
-H "Authorization: Bearer mentor-secret-token-abc" \
-H "Content-Type: application/json" \
-d '{"name": "Machine Learning"}'
```


---

## Статусы задач

| Статус      | Кто устанавливает             |
|-------------|-------------------------------|
| New         | Ментор (автоматически при создании) |
| InProgress  | Стажёр, Ментор                |
| Blocked     | Стажёр, Ментор                |
| NeedsHelp   | Стажёр (только из InProgress) |
| Paused      | Стажёр, Ментор                |
| Completed   | Стажёр, Ментор                |

## Формат всех ответов

```json
{ "success": true, "data": { ... } }
{ "success": false, "error": "описание ошибки" }
```
