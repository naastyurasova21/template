# Trip Service

Сервис поездок TripGo. HTTP API на `chi`, PostgreSQL через `pgx`, менеджер транзакций на
контексте, код по OpenAPI генерируется через `oapi-codegen`.

## Требования

- **Go 1.24+** (для `go get -tool`; проект собирается на 1.26)
- **Docker** (для `tripgoctl` — поднимает PostgreSQL в контейнере)
- **`tripgoctl`** — утилита курса.
  Установка: [`course-infra`](https://github.com/course-go-autumn-2026/course-infra)
- **`psql`** — для проверки БД (опционально)
- **`jq`** — для удобного чтения JSON (опционально)

## Быстрый старт

```bash
# 1. Поднять окружение (один раз на машине)
tripgoctl cluster start
tripgoctl environment start      # создаст .env с адресом PostgreSQL

# 2. Накатить миграции
make migrate

# 3. Запустить сервис
make run
```

Сервис слушает `:8080`. Проверка:

```bash
curl -i http://localhost:8080/health
curl -i http://localhost:8080/ready
```

Остановка — `Ctrl+C` или `kill -TERM <pid>`. Graceful shutdown ждёт активные
запросы до `SHUTDOWN_TIMEOUT`.

## Переменные окружения

Все переменные приходят из окружения (`.env` создаёт `tripgoctl`). Пример —
в [`.env.example`](.env.example).

| Переменная | Обязательна | Описание | Пример |
|---|---|---|---|
| `HTTP_ADDR` | да | адрес HTTP-сервера | `:8080` |
| `LOG_LEVEL` | нет | уровень логирования: `debug`, `info`, `warn`, `error` | `info` |
| `SHUTDOWN_TIMEOUT` | да | таймаут graceful shutdown | `10s` |
| `DATABASE_URL` | да | строка подключения к PostgreSQL | `postgres://tripgo:tripgo@localhost:21032/tripgo?sslmode=disable` |
| `DATABASE_MAX_CONNS` | да | максимум соединений в пуле | `10` |
| `DATABASE_MIN_CONNS` | да | минимум соединений в пуле | `2` |
| `DATABASE_MAX_CONN_LIFETIME` | да | время жизни соединения | `30m` |
| `DATABASE_CONNECT_TIMEOUT` | да | таймаут подключения к БД | `5s` |
| `DATABASE_QUERY_TIMEOUT` | да | таймаут одного запроса | `3s` |

При отсутствии или невалидности обязательной переменной сервис падает на
старте с понятной ошибкой.

## HTTP API

Контракт — [`contracts/openapi/trip-service.openapi.yaml`](contracts/openapi/trip-service.openapi.yaml).
В работе 1 реализованы ручки:

| Метод | Путь | Коды |
|---|---|---|
| `POST` | `/api/v1/trips` | `201`, `400`, `409`, `500` |
| `GET` | `/api/v1/trips/{tripId}` | `200`, `400`, `404`, `500` |
| `POST` | `/api/v1/trips/{tripId}/finish` | `200`, `400`, `404`, `409`, `500` |
| `GET` | `/health` | `200` |
| `GET` | `/ready` | `200`, `503` |

Ошибки — по RFC 9457 в `application/problem+json`:

```json
{
  "type": "https://tripgo.example/problems/driver-busy",
  "title": "Driver busy",
  "status": 409,
  "detail": "Driver already has an active trip",
  "instance": "/api/v1/trips",
  "code": "driver_busy"
}
```

Коды ошибок: `invalid_request`, `trip_not_found`, `trip_completed`,
`driver_busy`, `internal_error`.

### Пример

```bash
TRIP_ID=$(curl -sS -X POST http://localhost:8080/api/v1/trips \
  -H 'content-type: application/json' \
  -d '{"user_id":"5cb72c04-7650-45c9-a79b-bcdba0631e0c",
       "driver_id":"8860b315-ec86-42eb-a17c-7c163d721ff5",
       "start_point":{"latitude":59.9398,"longitude":30.3146},
       "end_point":{"latitude":59.9290,"longitude":30.3626},
       "price":1450}' | jq -r .id)

curl -sS http://localhost:8080/api/v1/trips/$TRIP_ID | jq
curl -sS -X POST http://localhost:8080/api/v1/trips/$TRIP_ID/finish -i
```

## Команды

| Команда | Что делает |
|---|---|
| `make run` | собрать бинарь и запустить сервис |
| `make build` | собрать бинарь в `bin/trip-service` |
| `make migrate` | накатить миграции (`goose up`) |
| `make migrate-down` | откатить последнюю миграцию (`goose down`) |
| `make generate` | сгенерировать код по OpenAPI |
| `make tidy` | `go mod tidy` |

## Миграции

Только миграции — никаких ручных `CREATE TABLE` в живой БД. Схема описана в
[`contracts/schema.md`](contracts/schema.md).

```bash
make migrate         # накатить все
make migrate-down    # откатить последнюю
```

Применённые миграции не редактируются — при ошибке выпускается новая.

## Тесты

В работе 1 тестов нет (это работа 2). В следующих работах:
`make test` — все тесты с `-race`.

## Решения

### Уровень изоляции транзакций

Выбран **`READ COMMITTED`** (дефолт PostgreSQL). Указан явно при старте
транзакции в `TxManager.Do`.

**Почему:**
- Конкурентность на `create` защищена **partial unique index** на уровне БД:
  `trips_driver_active_uidx ON trips(driver_id) WHERE status = 'active'`.
  При попытке вставить второго активного водителя PostgreSQL вернёт `23505`
  независимо от уровня изоляции.
- Конкурентность на `finish` защищена **условным `UPDATE`**:
  `UPDATE trips SET status = 'completed' WHERE id = $1 AND status = 'active'`.
  Второй параллельный запрос увидит `0 rows affected`. Это тоже не зависит от
  уровня изоляции.
- `SERIALIZABLE` и `REPEATABLE READ` избыточны для наших запросов и потребовали
  бы retry-логики на `serialization_failure`.

### Менеджер транзакций

`TxManager.Do(ctx, fn)` — центральный компонент, позволяет бизнес-коду не знать
про `pgx`, `pgxpool` и `pgx.Tx`.

**Как работает:**

1. `Do` проверяет, есть ли в контексте уже открытая транзакция
   (`db.TxFromContext`). Если есть — **переиспользует её** (вложенный вызов
   не открывает вторую).
2. Если транзакции нет — открывает новую через `pool.BeginTx` с уровнем
   изоляции `READ COMMITTED`.
3. Кладёт транзакцию в контекст через `db.WithTx(ctx, tx)` и вызывает `fn` с
   обогащённым контекстом.
4. `fn` вернула `nil` → `COMMIT`.
5. `fn` вернула ошибку → `ROLLBACK`, ошибка пробрасывается наружу.
6. `fn` паникует → `ROLLBACK`, паника пробрасывается (через `defer recover`).

**Как репозиторий видит транзакцию:**

Репозиторий не принимает транзакцию аргументом. Он вызывает
`db.ExecutorFromContext(ctx, r.pool)`:

- если в контексте есть `pgx.Tx` — работает через неё;
- если нет — работает через `pgxpool.Pool`.

Интерфейс `db.DBTX` (`Exec`/`Query`/`QueryRow`) реализуют **и** `pgx.Tx`,
**и** `pgxpool.Pool`. Репозиторий не знает, что именно ему передали.

**Что произойдёт при вызове репозитория вне `Do`:**

Запрос выполнится **вне транзакции**, каждый запрос — отдельная транзакция
PostgreSQL. Для чтения это норма. Для двух INSERT (`Create`) — **не атомарно**:
если второй упадёт, первый останется. Поэтому `Create` **обязательно**
вызывается внутри `txManager.Do` — это делает handler.

**Проверка атомарности:** в `Create` временно заменили `to_status` на
невалидное значение → второй INSERT упал на CHECK → первый откатился,
`SELECT count(*) FROM trips` вернул `0`.

### Запрет двух активных поездок

Обеспечен **partial unique index** на уровне БД:

```sql
CREATE UNIQUE INDEX trips_driver_active_uidx
    ON trips (driver_id)
    WHERE status = 'active';
```

**Почему так:**
- Индекс уникален **только для активных поездок** (`WHERE status = 'active'`).
  Один водитель может иметь много завершённых, но **не больше одной активной**.
- Проверка `SELECT` перед `INSERT` **не годится** — два параллельных запроса
  могут пройти проверку до того, как любой из них вставит строку.
- Ограничение **нельзя обойти** из кода — оно в схеме.

**Обработка в коде:** при `INSERT` PostgreSQL возвращает `pgconn.PgError` с
кодом `23505`. Репозиторий ловит его и возвращает `repository.ErrDriverBusy`.
Handler мапит в HTTP `409` с `code: driver_busy`.
**Проверено:** `seq 20 | xargs -P20` на одного водителя → ровно `1×201`, `19×409`.

### Защита от конкурентного `finish`

Завершение защищено **условным `UPDATE`**:

```sql
UPDATE trips
SET status = 'completed', finished_at = now()
WHERE id = $1 AND status = 'active';
```

Ключевое — `WHERE status = 'active'`: два параллельных `finish` не могут оба
завершить поездку, второй получит `409 trip_completed`. 

### `problem+json` для ошибок парсинга параметров

`oapi-codegen` по умолчанию возвращает `400 text/plain` с сырым текстом ошибки
при невалидном `tripId` (не UUID). Чтобы соблюсти требование задания
(`application/problem+json` с `code: invalid_request`), в `api.ChiServerOptions`
передан кастомный `ErrorHandlerFunc: handler.ErrorHandler`. Сгенерированный код
при этом не правится.
`ErrorHandler` распознаёт `*api.InvalidParamFormatError` и возвращает **краткий читаемый** `detail` без внутренних деталей парсера:

- для `tripId` → `"invalid UUID format"`;
- для остальных параметров → `"invalid parameter format"`.

Если ошибка другого типа — отдаётся общий `"invalid request parameters"`.
## Структура

```
cmd/trip-service/        composition root (main)
internal/
  config/                загрузка конфигурации из env
  db/                    пул pgxpool, tx в контексте, интерфейс DBTX
  txmanager/             менеджер транзакций
  repository/            репозитории на pgx + squirrel
  httpserver/            HTTP-сервер, таймауты, graceful shutdown
  handler/               реализация api.ServerInterface
api/                     код, сгенерированный по OpenAPI
migrations/              SQL-миграции goose
contracts/               контракты курса (не редактируются)
deploy/                  Dockerfile (в следующих работах)
```