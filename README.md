# P2P цифровой кошелек

## Зависимости

- Go `1.25+`
- Docker + Docker Compose (для локального Postgres/Jaeger и деплоя контейнера)
- `make` не нужен, используется `Taskfile` (`go-task` опционально)

Опциональные локальные dev-tools (ставятся в `./bin` через задачи):
- `golangci-lint`
- `oapi-codegen`
- `mockery`
- `goose`

## Конфигурация

1. Скопируйте переменные окружения:
```bash
cp .env.example .env
```
2. При запуске приложения на хосте (не в Docker) установите в `.env`:
```env
HTTP_HOST=localhost
DB_HOST=localhost
```
3. При запуске через `docker compose` оставьте:
```env
HTTP_HOST=0.0.0.0
DB_HOST=postgres
```

## Локальный запуск

### Вариант 1: полностью в Docker (рекомендуется)

```bash
docker compose up -d postgres jaeger
task migrate-up
task docker-run
```

Проверка:
- `GET http://localhost:8080/health` -> `200`
- `GET http://localhost:8080/ready` -> `200`, если доступна БД
- `GET http://localhost:8080/metric` -> Prometheus-метрики

Jaeger UI: `http://localhost:16686`

### Вариант 2: приложение на хосте, зависимости в Docker

```bash
docker compose up -d postgres jaeger
task migrate-up
task build
task run
```

Важно: для этого варианта в `.env` должен быть `DB_HOST=localhost`.

## Тесты и проверки

```bash
task unit-test
task api-test
task integration-test
task lint
task check
```

## Деплой

Ниже базовый Docker-поток деплоя.

1. Собрать образ:
```bash
VERSION=vX.Y.Z COMMIT_SHA=$(git rev-parse --short HEAD) BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
docker build \
  --build-arg VERSION=$VERSION \
  --build-arg COMMIT_SHA=$COMMIT_SHA \
  --build-arg BUILD_TIME=$BUILD_TIME \
  -t p2p-wallet-auth:$VERSION .
```

2. Запушить в registry:
```bash
docker tag p2p-wallet-auth:$VERSION <registry>/p2p-wallet-auth:$VERSION
docker push <registry>/p2p-wallet-auth:$VERSION
```

3. Запустить в окружении:
- передать все переменные из `.env` через секреты/конфиг окружения;
- обеспечить доступ к Postgres;
- открыть HTTP-порт сервиса (`HTTP_PORT`, по умолчанию `8080`);
- использовать `health` и `ready` как liveness/readiness пробы.

## Полезные задачи

- `task install-deps` - установить dev-инструменты в `./bin`
- `task gen-api` - сгенерировать OpenAPI-код
- `task gen-mocks` - сгенерировать моки
- `task migrate-up` / `task migrate-down` - миграции БД
- `task docker-run` - пересобрать и поднять сервис `p2p-wallet` в Docker
