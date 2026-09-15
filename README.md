# MyHealth NextGen

Приложение для ведения данных о здоровье: справочник продуктов с КБЖУ и
пользовательская динамика веса. Backend — Go + Gin + PostgreSQL, frontend —
React + Vite.

## Локальный запуск

Требуются Go 1.26.5, Node.js 20+ и Docker с Compose. После однократной установки
frontend-зависимостей всё окружение можно запустить одним скриптом:

```bash
npm --prefix client install
./dev.sh
```

Скрипт поднимает в Compose только PostgreSQL, дожидается готовности базы, а
backend и Vite запускает локально через `go run` и `npm run dev`. При Ctrl-C
либо аварийном завершении одного из процессов скрипт останавливает приложение и
удаляет PostgreSQL container и volume. Загруженный образ PostgreSQL остаётся в
cache.

<!-- LEGACY IMPORT START: remove this paragraph after migration. -->
Перед запуском backend временный импортер наполняет базу данными из
`legacy_data` для локального пользователя. Он поддерживает независимые
загрузчики для нескольких разделов; инструкция по расширению и последующему
удалению находится в [`legacy_data/README.md`](legacy_data/README.md).
<!-- LEGACY IMPORT END -->

Для запуска только PostgreSQL используйте:

```bash
docker compose up -d --wait postgres
```

После этого backend и frontend можно запускать локально в отдельных терминалах
(потребуются Go 1.26.5, Node.js 20+ и `npm --prefix client install`):

```bash
go run ./cmd/server --database-url 'postgresql://myhealth:myhealth@127.0.0.1:5432/myhealth?sslmode=disable'
npm --prefix client run dev
```

Интерфейс доступен на http://localhost:5173. Vite проксирует относительные
запросы `/api` в Go-сервис на http://localhost:3000.

Локальная база использует официальный образ PostgreSQL 18.6 Alpine:

- база, пользователь и пароль — `myhealth`;
- контейнер использует host network и слушает локальный порт `5432`;
- данные сохраняются в именованном Docker volume;
- Goose автоматически создаёт схему при запуске API.

```bash
docker compose ps
docker compose down       # остановить, сохранив данные
docker compose down -v    # остановить и удалить локальные данные
```

## API

| Метод | Маршрут | Назначение |
| --- | --- | --- |
| `GET` | `/api/me` | Пользователь из клиентского сертификата |
| `GET` | `/api/food?q=...` | Список и поиск продуктов |
| `GET` | `/api/food/:key` | Получение продукта |
| `POST` | `/api/food` | Создание продукта |
| `PUT` | `/api/food/:key` | Изменение продукта |
| `DELETE` | `/api/food/:key` | Удаление продукта |
| `GET` | `/api/weight?from=YYYY-MM-DD&to=YYYY-MM-DD` | Вес текущего пользователя за период |
| `POST` | `/api/weight` | Создание или замена веса за дату |
| `DELETE` | `/api/weight/:dt` | Удаление веса за дату |

Ключ создаётся сервером как UUID. Поиск регистронезависимый и выполняется по
названию и бренду. Backend принимает и возвращает только значения КБЖУ на 100 г;
пересчёт данных для другого веса выполняет форма frontend.

Диапазон веса включителен с обеих сторон, его границы можно не указывать.
Повторный `POST` для той же даты заменяет значение. Пользователь определяется
только по проверенной backend identity и не передаётся клиентом в API.

## Архитектура

Backend следует гексагональной архитектуре:

- `internal/entity` — бизнес-сущности;
- `internal/port` — входные и выходные интерфейсы;
- `internal/cases` — бизнес-сценарии и валидация;
- `internal/adapter/httpapi` — JSON/HTTP;
- `internal/adapter/postgres` — обычные параметризованные SQL-запросы и
  встроенные миграции;
- `internal/service` — сборка зависимостей и жизненный цикл.

ORM и динамическая конкатенация SQL не используются.

## Проверки

```bash
go test ./...
go test -race ./...
go vet ./...
go build -o dist/myhealth-server ./cmd/server

TEST_DATABASE_URL='postgresql://myhealth:myhealth@127.0.0.1:5432/myhealth?sslmode=disable' go test ./internal/adapter/postgres

npm --prefix client test
npm --prefix client run build
bash -n deployment/*.sh
```

Интеграционные тесты очищают таблицы `food` и `weight` в указанной тестовой
базе. Не направляйте `TEST_DATABASE_URL` на базу с нужными данными.

## Серверный запуск

Бинарник принимает параметры:

```bash
myhealth-server \
  --host 127.0.0.1 \
  --port 3000 \
  --database-url 'postgresql://user:password@db.example/myhealth' \
  --require-client-cert=true
```

`--database-url` обязателен. PostgreSQL в production разворачивается отдельно.
Deployment и выпуск клиентских сертификатов описаны в
[`deployment/README.md`](deployment/README.md).
