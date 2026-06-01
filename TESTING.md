# Тестирование

## Структура тестов

```
platform-intern-growth/
├── service/
│   ├── competence_service_test.go   # T-3: юнит-тесты расчёта матрицы компетенций
│   └── task_service_test.go         # T-2: юнит-тесты валидации задач
├── middleware/
│   └── auth_test.go                 # T-1: юнит-тесты парсинга токена
└── handlers/
    └── handlers_integration_test.go # T-1, T-2, C-1, C-3, C-5: интеграционные тесты
```

Тесты делятся на два вида:

| Вид | Нужна БД? | Как запустить |
|-----|-----------|---------------|
| **Юнит-тесты** | Нет | `go test ./service/... ./middleware/...` |
| **Интеграционные** | Да (тестовая БД) | `go test ./handlers/ -tags=integration` |

---

## Запуск юнит-тестов (без базы данных)

```bash
go test ./service/... ./middleware/... -v
```


---

## Запуск интеграционных тестов (нужна PostgreSQL)

### Шаг 1: Создать тестовую базу данных

```bash
createdb intern_platform_test
```

> Использовать **отдельную** базу для тестов, чтобы не затереть рабочие данные.

### Шаг 2: Запустить тесты с тегом `integration`

```bash
DB_NAME=intern_platform_test go test ./handlers/ -tags=integration -v
```


| Тест | Что проверяет |
|------|---------------|
| `TestAuthValidate_ValidToken_Returns200` | T-1: верный токен → 200 |
| `TestAuthValidate_InvalidToken_Returns401` | T-1: несуществующий токен → 401 |
| `TestAuthValidate_EmptyHeader_Returns401` | T-1: нет заголовка → 401 |
| `TestCreateTask_AllRequiredFields_Returns201` | T-2: все поля → задача создана |
| `TestCreateTask_MissingTitle_Returns400` | T-2: нет title → 400 |
| `TestCreateTask_MissingCompetences_Returns400` | T-2: нет competences → 400 |
| `TestCreateTask_AsIntern_Returns403` | Стажёр не может создавать задачи |
| `TestScenarioC1_MentorCreatesTask_InternSeesIt` | C-1: ментор создал → стажёр видит |
| `TestScenarioC3_WrongToken_Returns401` | C-3: чужой токен → 401 |
| `TestScenarioC5_FilterByNeedsHelp_ReturnsOnlyNeedsHelp` | C-5: фильтр по статусу работает |

---

## Запуск всех тестов сразу

```bash
# Только юнит-тесты (быстро, без БД)
go test ./service/... ./middleware/... -v

# Юнит + интеграционные (нужна тестовая БД)
DB_NAME=intern_platform_test go test ./... -tags=integration -v
```

---

## Флаги для удобства

```bash
# -v — подробный вывод (показывать каждый тест)
go test ./service/... -v

# -run — запустить только конкретный тест
go test ./service/... -run TestCalculateCompetenceStatus -v

# -count=1 — отключить кеширование результатов
go test ./service/... -count=1 -v

# Вывести покрытие кода тестами
go test ./service/... -cover
```
