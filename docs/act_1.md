# Добавить сервис пользователей

## Требования

## 1. Регистрация пользователя

`POST /users/register`

### Request
```json
{
  "login": "string",
  "password": "string",
  "name": "string",
  "last_name": "string"
}
```

### Response
- `201 Created`

```json
{
  "id": 1,
  "login": "string",
  "name": "string",
  "last_name": "string",
  "created_at": "2026-02-19T13:00:00Z"
}
```

## 2. Вход в систему

`POST /users/login`

### Request
```json
{
  "login": "string",
  "password": "string"
}
```

### Response
- `200 OK`
- `header Cookie: session_id=string`

```json
{
  "id": 1,
  "login": "string",
  "name": "string",
  "last_name": "string",
}
```

### Сессии
- Сессия хранится в in-memory кэше.
- TTL: `1 час`.
- При повторном логине TTL обновляется.
- Формат хранения: `{"session:user_id": "session_id"}`.

## 3. Выход из системы

`POST /users/{id}/logout`
- `header Cookie: session_id=string`

## Хранение в БД

### Таблица `users`
| поле | тип | ограничение | название |
| --- | --- | --- | --- |
| `id` | `bigint` | `primary key` | ID пользователя |
| `name` | `varchar(50)` | `not null` | имя пользователя |
| `last_name` | `varchar(50)` | `not null` | фамилия пользователя |
| `login` | `varchar(255)` | `unique` | логин (латиница), уникальный |
| `password` | `varchar(255)` | `not null` | хэш пароля |
| `created_at` | `timestamp` | `default now()` | дата создания |
| `updated_at` | `timestamp` | `null` | дата обновления |

## Что важно не упустить в сценариях

### Регистрация
- `400 Bad Request` при невалидном `login`/`password`/`name`/`last_name`.
- `409 Conflict`, если `login` уже занят.
- Пароль сохраняется только в виде хэша (никогда не в открытом виде).

### Логин
- `401 Unauthorized` при неверной паре `login/password`.
- Явно определить поведение при повторном логине:
  - перезаписывать старую сессию, или
  - разрешать несколько сессий на пользователя.

### Логаут
- Проверять, что `session_id` принадлежит пользователю `{id}`.
- `401 Unauthorized` для невалидной/просроченной сессии.
- `404 Not Found`, если пользователь `{id}` не существует.

### Общее
- Единый формат ошибок (`code`, `message`).
- Ограничения на длину и допустимые символы для `login`.
- Таймзона для `created_at`/`updated_at` (рекомендуется UTC).
