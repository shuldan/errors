# `errors` — Управление ошибками с кодами и деталями для Go-приложений

[![Go CI](https://github.com/shuldan/errors/workflows/Go%20CI/badge.svg)](https://github.com/shuldan/errors/actions)
[![codecov](https://codecov.io/gh/shuldan/errors/branch/main/graph/badge.svg)](https://codecov.io/gh/shuldan/errors)
[![Go Report Card](https://goreportcard.com/badge/github.com/shuldan/errors)](https://goreportcard.com/report/github.com/shuldan/errors)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

Этот пакет предоставляет расширенную систему управления ошибками для Go-приложений, позволяя прикреплять коды, детали и стек вызовов к ошибкам. Он упрощает отладку, логирование и обработку ошибок в распределённых системах.

---

## 🚀 Основные возможности

- **Коды ошибок**: уникальные идентификаторы для классификации ошибок.
- **Шаблонизация сообщений**: подстановка значений в сообщения об ошибках.
- **Детали ошибки**: структурированные данные, связанные с конкретным экземпляром ошибки.
- **Цепочки ошибок**: поддержка `Unwrap` для отслеживания причин.
- **Снимок стека**: автоматический захват трассировки стека в момент создания.
- **Вспомогательные функции**: обёртки для стандартного пакета `errors`.
- **Потокобезопасная генерация кодов**: удобный генератор кодов с префиксами и счётчиками.

---

## 📦 Установка зависимостей и инструментов

Для работы с проектом убедитесь, что у вас установлен Go 1.24+.

Установите необходимые инструменты:

```sh
make install-tools
```

Это установит:
- `golangci-lint` (v2.4.0)
- `goimports`
- `gosec`

---

## 🛠️ Работа с проектом

### Запуск локальной проверки

```sh
make all
```

Выполняет:
- проверку форматирования кода,
- статический анализ (`golangci-lint`),
- security-сканирование (`gosec`),
- запуск тестов.

### Проверка в CI

```sh
make ci
```

Запускает:
- форматирование,
- `go vet`,
- линтер,
- тесты с отчётом о покрытии.

### Форматирование кода

```sh
make fmt
```

Автоматически форматирует `.go` файлы и сортирует импорты.

### Запуск тестов

```sh
make test
make test-coverage
```

---

## 🧱 Архитектура

### `Code`

Псевдоним для `string`, представляющий уникальный код ошибки. Позволяет создавать новые экземпляры `*Error` с помощью метода `New`.

### `Error`

Основная структура ошибки, включающая:
- `Code`: уникальный код ошибки.
- `Message`: сообщение, может содержать шаблоны Go `text/template`.
- `Details`: мапа для хранения произвольных данных.
- `Cause`: внутренняя ошибка, если таковая имеется.
- `Stack`: строка с трассировкой стека.
- `Timestamp`: время создания ошибки.

### `WithPrefix`

Функция-генератор, создающая фабрику для `Code` с общим префиксом и уникальными суффиксами или счётчиками.

### `utils.go`

Содержит удобные функции-обёртки (`Is`, `As`, `Unwrap`, `Join`, `GetErrorCode`) для работы с ошибками, совместимые со стандартным пакетом `errors`.

---

## 🧪 Пример использования

```go
package main

import (
	"fmt"

	"github.com/shuldan/errors"
)

func main() {
	// 1. Создание генератора кодов
	generateCode := errors.WithPrefix("AUTH")

	// 2. Создание ошибки с кодом и шаблонизированным сообщением
	authErr := generateCode("0001").New("failed to authenticate user {{.username}}")

	// 3. Добавление деталей и причины
	authErr = authErr.WithDetail("username", "admin").
		WithDetail("ip", "192.168.1.1").
		WithCause(fmt.Errorf("invalid password"))

	// 4. Использование в коде
	if authErr != nil {
		fmt.Printf("Error: %v\n", authErr)
		// Вывод: AUTH_0001: failed to authenticate user admin (caused by: invalid password)

		// 5. Извлечение кода ошибки
		code := errors.GetErrorCode(authErr)
		fmt.Printf("Error code: %s\n", code) // AUTH_0001

		// 6. Проверка типа ошибки
		var customErr *errors.Error
		if errors.As(authErr, &customErr) {
			fmt.Printf("Stack: %s\n", customErr.Stack)
		}
	}
}
```

---

## 📄 Лицензия

Проект распространяется под лицензией [MIT](LICENSE).

---

## 🤝 Вклад в проект

PR и issue приветствуются! Следуйте стандартам форматирования и покрывайте новый код тестами.

---

> **Автор**: MSeytumerov  
> **Репозиторий**: `github.com/shuldan/errors`  
> **Go**: `1.24.2`