# MyHealth NextGen

SPA-каркас для хранения показателей здоровья: React + Vite на клиенте и
Go + Gin JSON API на сервере.

## Локальный запуск

Требуются Go 1.26.5 и Node.js 20+ (Node.js нужен только для frontend).

```bash
npm --prefix client install
```

Запустите API и frontend в двух терминалах:

```bash
go run ./cmd/server
npm --prefix client run dev
```

Интерфейс: http://localhost:5173. API: http://localhost:3000/api.

- `GET /api/health` — состояние сервиса;
- `GET /api/me` — пользователь из клиентского сертификата;
- `GET /api/measurements` — список измерений;
- `POST /api/measurements` — создание измерения.

```bash
go test ./...                          # тесты API
npm --prefix client test              # тесты frontend

go build -o dist/myhealth-server ./cmd/server
npm --prefix client run build
```

В development-режиме приложение использует локального тестового пользователя.
При серверной установке включается обязательный клиентский сертификат, а GUID
хранится только в памяти процесса и заново извлекается из каждого запроса.

## Деплоймент

Скрипты сборки, первоначальной установки, обновления и выпуска пользовательских
сертификатов описаны в [`deployment/README.md`](deployment/README.md).

Краткий сценарий:

```bash
./deployment/build-bundles.sh 1.0.0
sudo ./deployment/install-server.sh ./artifacts/myhealth-1.0.0.tar.gz health.example.com
sudo ./deployment/generate-client-cert.sh "Иван Иванов"
```

Сейчас медицинские измерения хранятся в памяти. API зависит от интерфейса
`MeasurementRepository`: для перехода на PostgreSQL нужно добавить реализацию
этого интерфейса в Go и передать её в `httpapi.NewRouter`, не меняя маршруты и
клиент. Настройки окружения перечислены в `.env.example`.
