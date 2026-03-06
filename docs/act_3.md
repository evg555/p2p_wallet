# Добавить сервис кошельков

## Требования

## 1. Создание нового кошелька

`POST /wallet`

### Request
- `header Cookie: session_id=string`
```json
{
  "user_id": 1,
  "currency": "string"
}
```

### Response
- `201 Created`

```json
{
  "id": 1,
  "user_id": 1,
  "currency": "string",
  "status": "string",
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
      "currency": "string",      
      "total_amount": 10010, // в минорах
      "held_amount": 10010, // в минорах
      "status": "string",      
      "created_at": "2026-02-19T13:00:00Z",
      "updated_at": "2026-02-19T13:00:00Z"
    },
    {}
  ]
}
```

## Хранение в БД

### Таблица `wallets`
| поле         | тип                     | ограничение      | название          |
|--------------|-------------------------|------------------|-------------------|
| `id`         | `bigint`                | `primary key`    | ID кошелька       |
| `currency`   | `enum(USD, EUR)`        | `not null`       | валюта кошелька   |
| `status`     | `enum(active, blocked)` | `default active` | статус кошелька   |
| `user_id`    | `bigint`                | `foreign key`    | кому принадлежит  |
| `created_at` | `timestamp`             | `default now()`  | дата создания     |
| `updated_at` | `timestamp`             | `null`           | дата обновления   |

- constraint uq_wallet_user_currency unique (user_id, currency)

### Таблица `wallet_balance_snapshots`
| поле               | тип              | ограничение              | название                      |
|--------------------|------------------|--------------------------|-------------------------------|
| `wallet_id`        | `bigint`         | `primary key`            | ID кошелька                   |
| `currency`         | `enum(USD, EUR)` | `not null`               | валюта                        |
| `held_amount`      | `bigint`         | `not null default 0`     | удержанный баланс             |
| `total_amount`     | `bigint`         | `not null default 0`     | общий баланс                  |
| `updated_at`       | `timestamp`      | `not null default now()` | дата и время обновления       |

- constraint fk_wallet_balance_snapshots_wallet_id foreign key (wallet_id) references wallets(id) on delete cascade
- constraint chk_balance_nonnegative check (held_amount >= 0 and total_amount >= 0 and held_amount <= total_amount)

## Что важно не упустить в сценариях

### Создание кошелька
- Операции создания кошелька должны быть идемпотентны (уникальный индекс по user_id и currency)
- Snapshot баланса создается автоматически при создании нового кошелька с дефолтными значениями
- `400 Bad Request` при невалидном `currency`.
- `409 Conflict`, если `wallet` уже занят.
- `401 Unauthorized`, если сессии нет или просрочена.

### Получение кошелька
- `401 Unauthorized`, если сессии нет или просрочена.
- Кэшировать на час данные по кошелька мпо пользователю

### Общее
- Деньги храним в минорах (в центах с точностью 2 знака после запятой)
- Единый формат ошибок (`code`, `message`).
- Ограничения на длину и допустимые символы для `currency`.
- Таймзона для `created_at`/`updated_at` (рекомендуется UTC).
