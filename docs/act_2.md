# Добавить Definition of Done


## 1. Подключить БД
* Создать докер файлы БД и приложения (postgres)
* Версионирование: VERSION, COMMIT_SHA, BUILD_TIME зашиты в бинарь и логируются на старте
* Создать клиента к постгрес
* Написать миграции через goose
* Написать запросы в репо слое
* Поменять слой данных с кэша на БД
* Покрыть интеграционными тестами (через тестовую БД testcontainers)

## 2. Вынести основные параметры в конфиг
* Читается из env
* Усть пример env.example
* Валидация конфига при старте

## 3. Добавить структурированные логи
* С полями: service, env, version, request_id, trace_id, span_id
* У каждого запроса есть request_id
* Уровни логирования: debug/info/warn/error
* Есть лог ключевых событий старта/остановки, подключений, миграций

## 4. Добавить основные метрики и трейсы
* /metrics endpoint
*  Базовые метрики HTTP: 
  - requests_total (по route/method/status), 
  - request_duration_seconds histogram (p50/p95/p99)
  - in_flight_requests
* Метрики БД:

длительность запросов/ошибки (хотя бы на уровне repo)

pool stats (db/sql): open/idle/inuse, wait count/time
*  Метрики runtime:

go/process metrics включены (goroutines, GC, mem)

* Установить контейнеры otel-collector, kibana, prometheus, grafana
* Есть дашборд (Grafana, Kibana)

## 4. Добавить health и ready эндпойнты
* /health (liveness): процесс жив
* /ready (readiness): сервис готов принимать трафик (БД коннектится, миграции применены/не требуются, критичные зависимости доступны).
* Аккуратная деградация: если внешняя зависимость недоступна — readiness может быть false (по вашей политике).

## 5. e2e переписать на контрактные (API)

## 6. Поправить readme (назначение сервиса, зависимости, как запустить локально, как деплоить)