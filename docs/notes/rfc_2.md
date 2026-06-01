Расширение схемы ledger_entries для проводок между счетами

Так как есть внутренние проводки (холдирование, зачисление, списание средств), то одной сущности кошелька не достаточно. Нужно сделать дополнительную сущность Account, где хранить тип счета и свзяь с кошельком

### Таблица `accounts`
| поле           | тип                                                      | ограничение                  | название        |
|----------------|----------------------------------------------------------|------------------------------|-----------------|
| `id`           | `bigint`                                                 | `primary key`                | ID счета        |
| `wallet_id`    | `bigint null`                                            | `not null`                   | ID кошелька     |
| `account_type` | `enum(available, held, external_in, external_out, fees)` | `not null default available` | тип счета       |
| `currency`     | `wallet_currency`                                        | `not null`                   | валюта счета    |
| `status`       | `enum(active,blocked)`                                   | `not null`                   | статус          |
| `created_at`   | `timestamp`                                              | `default now()`              | дата создания   |
| `updated_at`   | `timestamp`                                              | `null`                       | дата обновления |

* В таблице `ledger_entries` сделать связь с account_id, а не с wallet_id
* Расширить сущность Entry в коде - добавить тип счета
* Если wallet_id == null, то это системный счет (external_in, external_out, fees) для каждой валюты

Тогда проводка может выглядеть так:
* Перевод между пользователями User1(available) -> User2(available)
* Холдирование средств User1(available) -> User1(held)
* Зачисление средств (external_in) -> User1(available)
* Списание средств User1(available) -> (external_out)
* Комиссия User1(available) -> (fees)
