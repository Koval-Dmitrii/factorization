# 🔢 Factorization

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org)
[![Concurrency](https://img.shields.io/badge/concurrency-worker%20pool-success?style=flat)](https://go.dev/doc/effective_go#concurrency)
[![Algorithm](https://img.shields.io/badge/algorithm-prime%20factorization-blue?style=flat)](https://ru.wikipedia.org/wiki/Перебор_делителей)

Конкурентная реализация факторизации целых чисел с отдельными пулами воркеров для вычисления и записи результата.

## 📦 Установка

```bash
go get github.com/Koval-Dmitrii/factorization
```

## 🚀 Быстрый старт

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Koval-Dmitrii/factorization/internal/fact"
)

func main() {
	factorizer, err := fact.New(
		fact.WithFactorizationWorkers(4),
		fact.WithWriteWorkers(2),
	)
	if err != nil {
		panic(err)
	}

	numbers := []int{100, -17, 25, 38}
	if err = factorizer.Factorize(context.Background(), numbers, os.Stdout); err != nil {
		panic(err)
	}

	fmt.Println("done")
}
```

Пример результата (порядок строк может отличаться из-за конкурентной записи):

```text
100 = 2 * 2 * 5 * 5
-17 = -1 * 17
38 = 2 * 19
25 = 5 * 5
```

## 📚 API

```go
type Factorizer interface {
	Factorize(ctx context.Context, numbers []int, writer io.Writer) error
}
```

Конструктор:

```go
func New(opts ...FactorizeOption) (*factorizerImpl, error)
```

Опции:

```go
func WithFactorizationWorkers(workers int) FactorizeOption
func WithWriteWorkers(workers int) FactorizeOption
```

Ошибки:

```go
var (
	ErrFactorizationCancelled = errors.New("cancelled")
	ErrWriterInteraction      = errors.New("writer interaction")
)
```

## ⚙️ Поведение и гарантии

- Множители выводятся **по возрастанию**.
- Для отрицательных чисел добавляется множитель `-1`.
- Поддерживается `math.MinInt` (корректная обработка переполнения при смене знака).
- При отмене контекста возвращается `ErrFactorizationCancelled`.
- При ошибке записи — `ErrWriterInteraction`, после чего работа всех воркеров останавливается.
- `writer` должен быть потокобезопасным, так как запись выполняется конкурентно.

## 🧪 Тестирование

Запуск всех тестов:

```bash
go test ./...
```

Фокус на производительность:

```bash
go test ./internal/fact -run TestGeneralPerformance -count=1
```

## 🏗️ Структура проекта

```text
factorization/
├── internal/
│   └── fact/
│       ├── fact.go
│       ├── model_test.go
│       ├── performance_test.go
│       └── util_test.go
├── go.mod
└── README.md
```

## 👨‍💻 Автор

**Коваль Дмитрий**

- GitHub: [@Koval-Dmitrii](https://github.com/Koval-Dmitrii)
