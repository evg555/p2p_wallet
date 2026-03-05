# Добавить сервис кошельков

## Требования

## 1. Создание нового кошелька

`POST /wallet`

### Request
- `header Cookie: session_id=string`
```json
{
  "user_id": 1,
  "title": "string",
  "currency": "string"
}
```

### Response
- `201 Created`

```json
{
  "id": 1,
  "user_id": 1,
  "title": "string",
  "currency": "string",
  "created_at": "2026-02-19T13:00:00Z"
}
```

### Инварианты
- user_id + currency должен быть уникальным

## 2. Получение кошельков пользователя

`GET /wallets/{user_id}`
- `header Cookie: session_id=string`

### Response
- `200 OK`

```json
{
  "wallets": [
    {
      "id": 1,
      "user_id": 1,
      "title": "string",
      "currency": "string",
      "created_at": "2026-02-19T13:00:00Z",
      "updated_at": "2026-02-19T13:00:00Z"
    },
    {}
  ]
}
```

## Хранение в БД

### Таблица `wallets`
| поле         | тип              | ограничение     | название          |
|--------------|------------------|-----------------|-------------------|
| `id`         | `bigint`         | `primary key`   | ID кошелька       |
| `title`      | `varchar(50)`    | `not null`      | название кошелька |
| `currency`   | `enum(USD, EUR)` | `not null`      | валюта кошелька   |
| `user_id`    | `bigint`         | `foreign key`   | кому принадлежит  |
| `balance_id` | `bigint`         | `foreign key`   | связанный баланс  |
| `created_at` | `timestamp`      | `default now()` | дата создания     |
| `updated_at` | `timestamp`      | `null`          | дата обновления   |

## Что важно не упустить в сценариях

### Создание кошелька
- операции создания кошелька должны быть идемпотентны (хэш по user_id и currency)
- `400 Bad Request` при невалидном `currency`/`title`/`name`/`last_name`.
- `409 Conflict`, если `wallet` уже занят.
- `401 Unauthorized`, если сессии нет или просрочена.

### Получение кошелька
- `401 Unauthorized`, если сессии нет или просрочена.
- Кэшировать на час данные по кошелька мпо пользователю

### Общее
- Единый формат ошибок (`code`, `message`).
- Ограничения на длину и допустимые символы для `title`, `currency`.
- Таймзона для `created_at`/`updated_at` (рекомендуется UTC).
