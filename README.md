# P2P цифровой кошелек

## OpenAPI контракт

- Основная спецификация: `spec/openapi/users.yaml`
- Конфиги генерации:
  - `spec/openapi/oapi-codegen.types.yaml`
  - `spec/openapi/oapi-codegen.server.yaml`

## Генерация API кода

Сгенерировать модели и серверный слой:

```bash
task gen-api
```

Генерируемые файлы:

- `internal/api/types.gen.go`
- `internal/api/server.gen.go`

`*.gen.go` не редактируются вручную.

## Где писать handlers

- Ручная реализация хэндлеров: `internal/handler/handler.go`
- Хэндлер должен реализовывать интерфейс из generated-кода: `api.ServerInterface`

## Как обновлять контракт

1. Измени `spec/openapi/users.yaml`.
2. Запусти `task gen-api`.
3. Проверь, что generated-код актуален:

```bash
git diff --exit-code -- internal/api/*.gen.go
```

## Запуск и проверки

- Установить инструменты: `task install-deps`
- Линтер: `task lint`
- Тесты: `task test`
- Полная проверка: `task check`
- Запуск приложения: `task run` (или `go run ./cmd/app`)

CI также проверяет, что после `task gen-api` нет незакоммиченных изменений.
