# go-adv

Сервис коротких ссылок с аналитикой кликов. Писал его, чтобы прокачаться в Go посерьёзнее: без Gin и прочих батареек, на голом `net/http` + `ServeMux` из 1.22. БД — Postgres через GORM, аутентификация на JWT, статистика собирается асинхронно через свою шину событий.

## Что умеет

- Регистрация/логин, пароли хешируются bcrypt-ом.
- JWT в заголовке `Authorization: Bearer ...`, защищённые роуты через middleware.
- CRUD коротких ссылок + редирект по хешу (`GET /{hash}`).
- Сбор кликов: при каждом переходе публикуется событие в `EventBus`, отдельная горутина слушает канал и инкрементит счётчик в БД (чтобы не тормозить редирект).
- Агрегированная статистика за период с группировкой по дням или месяцам.

## Стек

- Go 1.25, `net/http` (без фреймворков)
- PostgreSQL 16 + GORM (драйвер на pgx)
- JWT (`github.com/golang-jwt/jwt/v5`)
- Валидация — `go-playground/validator`
- Docker Compose для локальной БД
- Тесты: стандартный `testing` + `httptest`, для сервисов — `go-sqlmock`

## Структура

```
cmd/            точка входа + e2e-тест на /auth/login
configs/        загрузка .env
internal/
  auth/         регистрация, логин, выдача токенов
  link/         CRUD ссылок и редирект
  stat/         подписчик событий и агрегации
  user/         модель юзера и репозиторий
migrations/     отдельный main с AutoMigrate
pkg/
  db/           обёртка над *gorm.DB
  di/           интерфейсы для мока репозиториев в тестах
  event/        простая in-memory шина событий
  jwt/          создание/парсинг токенов
  middleware/   Auth, CORS, Logging + Chain
  req/ res/     хелперы для парсинга тела и JSON-ответов
```

## Запуск

1. Поднять Postgres:

```bash
docker compose up -d
```

   Порт наружу — `5440`, внутри контейнера стандартный `5432`.

   Перед первым стартом нужно руками создать базу `link` (или поменять имя в DSN).

2. Накатить схему:

```bash
go run ./migrations
```

   Внутри обычный `AutoMigrate` по моделям `Link`, `User`, `Stat` - плодить отдельный инструмент вроде goose не хотелось.

3. Стартануть апи:

```bash
go run ./cmd
```

   Слушает `:8081`.

## Эндпоинты

### Auth

| Метод | URL              | Тело                          |
|-------|------------------|-------------------------------|
| POST  | `/auth/register` | `{ name, email, password }`   |
| POST  | `/auth/login`    | `{ email, password }`         |

Оба возвращают `{ "token": "..." }`. Токен дальше нужно класть в `Authorization: Bearer <token>`.

### Links (требуют токен)

| Метод  | URL            | Описание                                  |
|--------|----------------|-------------------------------------------|
| POST   | `/link`        | создать ссылку (`{ "url": "..." }`)       |
| PATCH  | `/link/{id}`   | обновить url/hash                         |
| DELETE | `/link/{id}`   | мягкое удаление (gorm.Model → deleted_at) |
| GET    | `/link`        | список, параметры `?limit=&offset=`       |

### Redirect

| Метод | URL        | Описание                                   |
|-------|------------|--------------------------------------------|
| GET   | `/{hash}`  | 307 редирект на оригинальный URL + событие в шину |

### Stats (требует токен)

```
GET /stat?from=YYYY-MM-DD&to=YYYY-MM-DD&by=day|month
```

Возвращает массив `{ period, sum }`, сгруппированный по выбранному интервалу.

## Пример использования

```bash
# регистрация
curl -X POST localhost:8081/auth/register \
  -H 'content-type: application/json' \
  -d '{"name":"arseniy","email":"a@a.com","password":"123"}'

# ответ: {"token":"eyJhbGciOi..."}

# создание ссылки
curl -X POST localhost:8081/link \
  -H 'Authorization: Bearer eyJhbGciOi...' \
  -H 'content-type: application/json' \
  -d '{"url":"https://github.com"}'

# редирект
curl -i localhost:8081/aBcDeF
```

## Тесты

```bash
go test ./...
```

- `pkg/jwt` — юниты на создание и парсинг токена.
- `internal/auth` — сервис покрыт с `go-sqlmock`, без живой БД.
- `cmd/auth_test.go` — e2e на логин через `httptest.NewServer(App())`. **Важно:** этот тест ходит в реальную Postgres из `.env`, так что перед запуском БД должна быть поднята.

