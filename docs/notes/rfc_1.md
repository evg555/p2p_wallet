Критичные/важные замечания (Clean Architecture + DDD)

1. Слой application (service) зависит от transport DTO (internal/api)
    - В use-case слое используются api.RegisterRequest и api.LoginRequest, что связывает бизнес-логику с OpenAPI-генерацией: auth.go:9, auth.go:48, auth.go:67.
    - В чистой архитектуре service должен принимать собственные команды/DTO (application layer), а mapping из HTTP делать в handler.
2. Инверсия зависимостей нарушена в infra/repository
    - Репозитории импортируют internal/service и реализуют его интерфейсы: user_postgres.go:10, session.go:9, user_metrics.go:11, user_tracing.go:8.
    - По Clean Architecture интерфейсы портов должны жить во внутреннем слое (application/domain), а infra зависеть от них, но не от конкретного service-пакета.
3. Сущность User смешивает домен и технические детали + риск по безопасности
    - В доменной сущности есть JSON-теги (деталь API): user.go:15.
    - Хэш пароля через sha256 без соли/адаптивного cost: user.go:47. Для auth это небезопасно; нужен bcrypt/argon2id.
    - Домен сам назначает ID через rand.Int63(), хотя БД использует BIGSERIAL: user.go:37, 20260223143000_create_users.sql:4. Это дублирование responsibility.
4. Доменный слой знает про HTTP/cookie/context-ключи
    - SessionKey и CtxKey живут в domain: auth_result.go:7, auth_result.go:9.
    - Middleware и service завязаны на этот transport-механизм: middleware.go:88, middleware.go:94, auth.go:102.
    - Лучше вынести cookie/context-контракт в delivery/auth package или в отдельный application port.
5. Непоследовательность бизнес-семантики в Logout
    - Сначала проверяется сессия, потом пользователь: auth.go:95.
    - Из-за этого ErrUserNotFound в handler для logout может быть трудно достижим в реальном потоке: handler.go:122. Нужно явно зафиксировать желаемую семантику (401 vs
        404) и упорядочить проверки.

Что сделать в первую очередь

1. Вынести входные модели use-case из internal/api в internal/service (или internal/application) и сделать mapping в handler.
2. Перенести интерфейсы UserRepo/SessionRepo из service в отдельный port пакет (или рядом с use-case), чтобы repository зависел только от порта.
3. Переделать credential policy: bcrypt/argon2id, убрать password-хэширование из сущности User в domain service/value object.
4. Нормализовать контракт репозитория: not found всегда через typed error (errs.ErrUserNotFound), убрать nil,nil и небезопасные type assertion.