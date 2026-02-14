# `errors` — Структурированные ошибки для Go-приложений

[![Go CI](https://github.com/shuldan/errors/workflows/Go%20CI/badge.svg)](https://github.com/shuldan/errors/actions)
[![codecov](https://codecov.io/gh/shuldan/errors/branch/main/graph/badge.svg)](https://codecov.io/gh/shuldan/errors)
[![Go Report Card](https://goreportcard.com/badge/github.com/shuldan/errors)](https://goreportcard.com/report/github.com/shuldan/errors)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

Пакет предоставляет систему типизированных ошибок с кодами, семантическими категориями, уровнями серьёзности и структурированными деталями. Спроектирован для приложений, построенных по принципам DDD, где ошибки являются частью доменного контракта.

---

## Основные возможности

- **Коды ошибок** — стабильные строковые идентификаторы с префиксами (`CATALOG_NOT_FOUND`).
- **Семантические категории (Kind)** — `Validation`, `NotFound`, `Conflict`, `DomainRule`, `Infrastructure` и др.
- **Уровни серьёзности (Severity)** — `Error`, `Warning`, `Critical`.
- **Шаблонизация сообщений** — Go `text/template` с подстановкой деталей, парсится один раз.
- **Иммутабельность** — методы `With*` возвращают копию, оригинал неизменен.
- **Ленивый стек вызовов** — захватывается как `[]uintptr`, форматируется только при обращении.
- **Маппинг на HTTP-статусы** — автоматическое преобразование `Kind` → HTTP status code.
- **Безопасная JSON-сериализация** — полная (для логов) и публичная (для клиента).
- **Совместимость** — поддержка `errors.Is`, `errors.As`, `errors.Unwrap`, `fmt.Formatter`.

---

## Установка

```sh
go get github.com/shuldan/errors
```

Требуется Go 1.24+.

---

## Быстрый старт

```go
package main

import (
    "fmt"

    "github.com/shuldan/errors"
)

// 1. Фабрика кодов с префиксом
var authCode = errors.WithPrefix("AUTH")

// 2. Определение ошибок-шаблонов
var (
    ErrInvalidCredentials = authCode("INVALID_CREDENTIALS").
        Kind(errors.Authentication).
        New("неверные учётные данные для пользователя {{.username}}")

    ErrAccountLocked = authCode("ACCOUNT_LOCKED").
        Kind(errors.Authorization).
        Severity(errors.SeverityWarning).
        New("аккаунт {{.username}} заблокирован")
)

// 3. Использование в бизнес-логике
func Authenticate(username, password string) error {
    user, err := findUser(username)
    if err != nil {
        return ErrInvalidCredentials.
            WithDetail("username", username).
            WithCause(err)
    }

    if user.IsLocked {
        return ErrAccountLocked.
            WithDetail("username", username)
    }

    return nil
}

func main() {
    err := Authenticate("admin", "wrong")
    if err != nil {
        fmt.Println(err)
        // AUTH_INVALID_CREDENTIALS: неверные учётные данные для пользователя admin: user not found

        fmt.Println(errors.GetKind(err))
        // authentication

        fmt.Println(errors.ToHTTPStatus(err))
        // 401
    }
}
```

---

## Архитектура

### Поток создания ошибки

```
WithPrefix("CATALOG")                      — фабрика кодов
    ↓
catalogCode("NOT_FOUND")                   — CodeBuilder с кодом CATALOG_NOT_FOUND
    ↓
.Kind(errors.NotFound)                     — семантическая категория
    ↓
.Severity(errors.SeverityError)            — уровень серьёзности (опционально)
    ↓
.New("категория {{.id}} не найдена")       — шаблон ошибки (парсинг шаблона)
    ↓
*Error (шаблон)                            — без стека, без timestamp
    ↓
.WithDetail("id", 42) / .WithCause(err)   — инстанс ошибки (клонирование + стек + timestamp)
```

### Шаблон vs Инстанс

**Шаблон** — объект, созданный через `.New()` на уровне `var`. Не содержит стек и timestamp. Используется для определения типа ошибки и сравнения через `errors.Is`.

**Инстанс** — объект, созданный через `.WithDetail()`, `.WithDetails()`, `.WithCause()` или `.WithSeverity()`. Содержит стек вызовов, timestamp и контекстные данные.

```go
// Шаблон — определяется один раз на уровне пакета
var ErrNotFound = catalogCode("NOT_FOUND").
    Kind(errors.NotFound).
    New("категория {{.id}} не найдена")

// Инстанс — создаётся в месте возникновения ошибки
return ErrNotFound.WithDetail("id", categoryID)
```

---

## API

### Создание кодов

#### `WithPrefix`

Возвращает фабрику `CodeBuilder` с заданным префиксом. Код формируется как `PREFIX_NAME`.

```go
var catalogCode = errors.WithPrefix("CATALOG")
// catalogCode("NOT_FOUND") → код "CATALOG_NOT_FOUND"
```

#### `NewCode`

Создаёт `CodeBuilder` без префикса.

```go
var ErrInternal = errors.NewCode("INTERNAL_ERROR").
    Kind(errors.Internal).
    Severity(errors.SeverityCritical).
    New("внутренняя ошибка сервера")
```

### CodeBuilder

Цепочечный билдер для настройки ошибки перед созданием.

```go
catalogCode("NOT_FOUND").     // CodeBuilder с кодом
    Kind(errors.NotFound).    // семантическая категория
    Severity(errors.SeverityError). // уровень серьёзности
    New("сообщение")          // → *Error
```

| Метод | Описание |
|-------|----------|
| `.Kind(k Kind)` | Устанавливает семантическую категорию |
| `.Severity(s Severity)` | Устанавливает уровень серьёзности |
| `.New(msg string)` | Создаёт шаблон `*Error` |

### Kind — семантические категории

| Kind | Описание | HTTP-статус |
|------|----------|-------------|
| `Unknown` | Не классифицировано (по умолчанию) | 500 |
| `Validation` | Ошибка валидации входных данных | 400 |
| `NotFound` | Ресурс не найден | 404 |
| `Conflict` | Конфликт состояния (дубликат и т.п.) | 409 |
| `DomainRule` | Нарушение бизнес-правила | 422 |
| `Authorization` | Недостаточно прав | 403 |
| `Authentication` | Ошибка аутентификации | 401 |
| `Infrastructure` | Инфраструктурная ошибка (БД, сеть) | 503 |
| `Internal` | Внутренняя ошибка | 500 |

### Severity — уровни серьёзности

| Severity | Описание |
|----------|----------|
| `SeverityError` | Стандартная ошибка (по умолчанию) |
| `SeverityWarning` | Предупреждение, операция может продолжиться |
| `SeverityCritical` | Критическая ошибка, требует немедленного внимания |

### Error — методы модификации

Все методы возвращают **новый** `*Error`, не изменяя оригинал.

#### `WithDetail`

Добавляет одну деталь.

```go
return ErrNotFound.WithDetail("id", categoryID)
```

#### `WithDetails`

Добавляет несколько деталей за одно клонирование.

```go
return ErrStatusTransition.WithDetails(errors.D{
    "from": currentStatus,
    "to":   targetStatus,
})
```

`errors.D` — алиас для `map[string]any`.

#### `WithCause`

Устанавливает причину ошибки.

```go
return ErrNotFound.
    WithDetail("id", id).
    WithCause(sql.ErrNoRows)
```

#### `WithSeverity`

Переопределяет severity для конкретного инстанса.

```go
return ErrTimeout.
    WithSeverity(errors.SeverityCritical).
    WithCause(ctx.Err())
```

### Error — методы доступа

| Метод | Возвращаемый тип | Описание |
|-------|-------------------|----------|
| `GetCode()` | `Code` | Код ошибки |
| `GetKind()` | `Kind` | Семантическая категория |
| `GetSeverity()` | `Severity` | Уровень серьёзности |
| `GetMessage()` | `string` | Исходный шаблон сообщения |
| `GetTimestamp()` | `time.Time` | Время создания инстанса |
| `GetCause()` | `error` | Причина (аналог `Unwrap`) |
| `Detail(key)` | `(any, bool)` | Значение отдельной детали |
| `Details()` | `map[string]any` | Копия всех деталей |
| `StackTrace()` | `string` | Отформатированный стек вызовов |
| `RootCause()` | `error` | Самая глубокая причина в цепочке |

### Error — форматирование

Реализует `fmt.Formatter`:

```go
err := ErrNotFound.WithDetail("id", 42).WithCause(sql.ErrNoRows)

fmt.Sprintf("%s", err)
// CATALOG_NOT_FOUND: категория 42 не найдена: sql: no rows in result set

fmt.Sprintf("%v", err)
// CATALOG_NOT_FOUND: категория 42 не найдена: sql: no rows in result set

fmt.Sprintf("%+v", err)
// CATALOG_NOT_FOUND: категория 42 не найдена: sql: no rows in result set
//
// Stack trace:
// main.findCategory
//     /app/catalog/repository.go:42
// main.HandleRequest
//     /app/catalog/handler.go:18
// ...
// Caused by: sql: no rows in result set

fmt.Sprintf("%q", err)
// "CATALOG_NOT_FOUND: категория 42 не найдена: sql: no rows in result set"
```

### Error — JSON-сериализация

Реализует `json.Marshaler`. Включает все поля кроме стека:

```go
data, _ := json.Marshal(err)
```

```json
{
  "code": "CATALOG_NOT_FOUND",
  "kind": "not_found",
  "severity": "error",
  "message": "категория 42 не найдена",
  "details": {"id": 42},
  "cause": "sql: no rows in result set",
  "timestamp": "2025-01-15T10:30:00Z"
}
```

---

## HTTP-интеграция

### `ToHTTPStatus`

Маппит `Kind` ошибки на HTTP-статус.

```go
status := errors.ToHTTPStatus(err)
// NotFound → 404, Validation → 400, DomainRule → 422, ...
```

Если ошибка не является `*Error`, возвращает `500`.

### `ToPublicJSON`

Возвращает JSON, безопасный для отправки клиенту. Исключает стек, причину и severity.

```go
body := errors.ToPublicJSON(err)
```

```json
{
  "code": "CATALOG_NOT_FOUND",
  "message": "категория 42 не найдена",
  "details": {"id": 42}
}
```

### `ToPublicError`

Возвращает структуру `PublicError` для дальнейшей обработки.

```go
pe := errors.ToPublicError(err)
// pe.Code, pe.Message, pe.Details
```

### Пример HTTP-обработчика

```go
func (h *Handler) GetCategory(w http.ResponseWriter, r *http.Request) {
    category, err := h.service.FindByID(r.Context(), id)
    if err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(errors.ToHTTPStatus(err))
        w.Write(errors.ToPublicJSON(err))
        return
    }
    // ...
}
```

---

## Утилиты

Обёртки над стандартным пакетом `errors` для единого импорта:

| Функция | Описание |
|---------|----------|
| `Is(err, target)` | Проверяет совпадение в цепочке ошибок |
| `As(err, &target)` | Типизированное приведение (с дженериками) |
| `Unwrap(err)` | Возвращает следующую ошибку в цепочке |
| `Join(errs...)` | Объединяет несколько ошибок |
| `Wrap(cause, template)` | Оборачивает cause в шаблон (алиас для `template.WithCause(cause)`) |
| `GetCode(err)` | Извлекает `Code` из ошибки |
| `GetKind(err)` | Извлекает `Kind` из ошибки |
| `GetSeverity(err)` | Извлекает `Severity` из ошибки |

### Проверка ошибок

```go
// Проверка по типу (Code)
if errors.Is(err, ErrCategoryNotFound) {
    // обработка "не найдено"
}

// Извлечение структурированной ошибки
var domainErr *errors.Error
if errors.As(err, &domainErr) {
    log.Printf("code=%s kind=%s details=%v",
        domainErr.GetCode(),
        domainErr.GetKind(),
        domainErr.Details(),
    )
}

// Проверка категории
if errors.GetKind(err) == errors.Validation {
    // все ошибки валидации обрабатываем одинаково
}
```

---

## Примеры использования в DDD

### Определение ошибок домена

```go
// catalog/errors.go
package catalog

import "github.com/shuldan/errors"

var catalogCode = errors.WithPrefix("CATALOG")

var (
    ErrCategoryNotFound = catalogCode("CATEGORY_NOT_FOUND").
        Kind(errors.NotFound).
        New("категория {{.id}} не найдена")

    ErrCategoryStatusTransition = catalogCode("INVALID_STATUS_TRANSITION").
        Kind(errors.DomainRule).
        New("переход статуса категории из {{.from}} в {{.to}} невозможен")

    ErrCategoryDuplicateName = catalogCode("DUPLICATE_CATEGORY_NAME").
        Kind(errors.Conflict).
        New("категория с именем {{.name}} уже существует")

    ErrCategoryTreeDepth = catalogCode("TREE_DEPTH_EXCEEDED").
        Kind(errors.DomainRule).
        New("превышена максимальная глубина вложенности: {{.max}}")
)
```

### Доменная логика

```go
// catalog/category_status.go
func (s CategoryStatus) TransitionTo(target CategoryStatus) (CategoryStatus, error) {
    allowed, exists := categoryStatusTransitions[s]
    if !exists || !slices.Contains(allowed, target) {
        return "", ErrCategoryStatusTransition.WithDetails(errors.D{
            "from": s,
            "to":   target,
        })
    }
    return target, nil
}
```

### Репозиторий (инфраструктура)

```go
// catalog/infrastructure/repository.go
func (r *CategoryRepo) FindByID(ctx context.Context, id uuid.UUID) (*catalog.Category, error) {
    row := r.db.QueryRowContext(ctx, "SELECT ... WHERE id = $1", id)

    var cat catalog.Category
    if err := row.Scan(&cat.ID, &cat.Name, &cat.Status); err != nil {
        if stderrors.Is(err, sql.ErrNoRows) {
            return nil, catalog.ErrCategoryNotFound.
                WithDetail("id", id).
                WithCause(err)
        }
        return nil, ErrDatabaseQuery.
            WithDetail("query", "FindCategoryByID").
            WithCause(err)
    }

    return &cat, nil
}
```

### Application Service

```go
// catalog/application/service.go
func (s *CategoryService) Activate(ctx context.Context, id uuid.UUID) error {
    cat, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return err // ошибка уже обёрнута с кодом и деталями
    }

    newStatus, err := cat.Status.TransitionTo(catalog.StatusActive)
    if err != nil {
        return err // CATALOG_INVALID_STATUS_TRANSITION с деталями
    }

    cat.Status = newStatus
    return s.repo.Save(ctx, cat)
}
```

### Structured logging

```go
func logError(logger *slog.Logger, err error) {
    var domainErr *errors.Error
    if errors.As(err, &domainErr) {
        attrs := []any{
            "error_code", domainErr.GetCode(),
            "error_kind", domainErr.GetKind(),
            "severity", domainErr.GetSeverity(),
            "details", domainErr.Details(),
            "timestamp", domainErr.GetTimestamp(),
        }

        if cause := domainErr.RootCause(); cause != nil {
            attrs = append(attrs, "root_cause", cause.Error())
        }

        switch domainErr.GetSeverity() {
        case errors.SeverityCritical:
            logger.Error("critical error", attrs...)
        case errors.SeverityWarning:
            logger.Warn("warning", attrs...)
        default:
            logger.Error("error", attrs...)
        }
    } else {
        logger.Error("unstructured error", "error", err.Error())
    }
}
```

---

## Сборка и тестирование

### Установка инструментов

```sh
make install-tools
```

Устанавливает `golangci-lint`, `goimports`, `gosec`.

### Команды

| Команда | Описание |
|---------|----------|
| `make all` | Форматирование + линтер + security + тесты |
| `make ci` | Проверки для CI-пайплайна |
| `make fmt` | Форматирование кода и сортировка импортов |
| `make test` | Запуск тестов |
| `make test-coverage` | Тесты с отчётом о покрытии |

---

## Лицензия

Проект распространяется под лицензией [MIT](LICENSE).

---

## Вклад в проект

PR и issue приветствуются. Следуйте стандартам форматирования и покрывайте новый код тестами.

---

> **Автор**: MSeytumerov
> **Репозиторий**: `github.com/shuldan/errors`
> **Go**: `1.24.2`
