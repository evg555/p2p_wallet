# Добавить перевод средств внутри сервиса

## Требования

## 1. Операция перевода средств с одного кошелька на другой

`POST /balance/transfer`

### Request
- `header Cookie: session_id=string`
- `header X-Idempotency-Key: string`
```json
{
  "from_wallet_id": 1,
  "to_wallet_id": 2,
  "amount": 10010, // в минорах
}
```

### Response
- `200 Ok`

```json
{
  "transaction": {
    "id": 1,
    "status": "string",
    "entries": [
      {
        "id": 1,
        "wallet_id": 1,
        "amount": 10010, // в минорах
        "currency": "string"
      },
      {
        "id": 2,
        "wallet_id": 1,
        "amount": -10010, // в минорах
        "currency": "string"
      }
    ],
    "created_at": "2026-02-19T13:00:00Z"    
  }
}
```

### Инварианты
- кошельки должны существовать
- from_wallet_id должен принадлежать user_id
- на кошельке с которого переводится должно быть достаточно средств для перевода (total - held)
- сумма перевода должна быть положительная
- нельзя переводить между кошельками с разными валютами

## Хранение в БД

### Таблица `ledger_transactions`
| поле           | тип                                                       | ограничение             | название                                 |
|----------------|-----------------------------------------------------------|-------------------------|------------------------------------------|
| `id`           | `bigint`                                                  | `primary key`           | ID транзакции                            |
| `type`         | `enum(topup, trasfer, hold, relese, capture, withdrawal)` | `not null`              | тип транзакции                           |
| `status`       | `enum(new, succeed, failed)`                              | `not null`              | статус транзакции                        |
| `reference_id` | `varchar(50)`                                             | `null`                  | ссылка на источник (номер заказа и т.д.) |
| `idemp_key`    | `char(16)\uuid`                                           | `unique,not null`       | uuid ключ идемпотентности                |
| `created_at`   | `timestamp`                                               | `default current_stamp` | дата создания                            |

- constraint uq_ledger_transactions_idemp_key unique (idemp_key)


### Таблица `ledger_entries`
| поле             | тип              | ограничение             | название                      |
|------------------|------------------|-------------------------|-------------------------------|
| `id`             | `bigint`         | `primary key`           | ID записи                     |
| `transaction_id` | `bigint`         | `not null`              | ID транзакции                 |
| `wallet_id`      | `bigint`         | `not null`              | ID кошелька                   |
| `amount`         | `bigint`         | `not null default 0`    | сумма перевода/списания с +/- |
| `currency`       | `enum(USD, EUR)` | `not null`              | валюта                        |
| `created_at`     | `timestamp`      | `default current_stamp` | дата создания                 |

- constraint fk_ledger_entries_transaction_id foreign key (transaction_id) references ledger_transactions(id) on delete cascade
- constraint fk_ledger_entries_wallet_id foreign key (wallet_id) references wallet_balance_snapshots(wallet_id) on delete nothing

## Что важно не упустить в сценариях

## Сценарий перевода средств
1. Получение кошельков с балансом (проверка принадлежности пользователю)
2. Проверка, что валюта перевода совпадает и что на кошельке источнике достаточно средств для перевода
3. Создание транзакции:
 - LedgerTransaction(id, type[transfer], status, reference_id[null], idemp_key[from header], created_at)
 - Создание двух entry
   - Entry(id, transaction_id, from_wallet_id, amount (-), currency, created_at)
   - Entry(id, transaction_id, to_wallet_id, amount (+), currency, created_at)   - 

### Перевод средств
- Операции перевода средств должны быть идемпотентны (генерируется ключ на фронте и вставляется в хэдер)
- Операции перевода выполняются в транзакции: создаются 2 записи entry (append-only), создается транзакция, обновляются балансы двух кошельков
- Балансы сортируются в порядке возрастания по id и блокируются (select for update), чтобы избежать потери обновлений, конкурентных обновлений и дедлоков
- `400 Bad Request` при невалидном `amount`.
- `401 Unauthorized`, если сессии нет или просрочена.
- `422 Unprocessable Entity`, если нет кошельков, кошелек не принадлежит текущему пользователю, недостаточно средств или не совпадают валюты перевода


### Общее
- Деньги храним в минорах (в центах с точностью 2 знака после запятой)
- Единый формат ошибок (`code`, `message`).
- Таймзона для `created_at`/`updated_at` (рекомендуется UTC).
