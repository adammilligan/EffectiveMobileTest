# EffectiveMobileTest — сервис агрегации подписок

REST-сервис для CRUDL-операций над онлайн-подписками пользователей и подсчёта суммарной стоимости подписок за период.

## Запуск

1) Создай `.env` на основе примера:

```bash
cp .env.example .env
```

2) Подними сервисы:

```bash
docker compose -f deployments/docker-compose.yml up -d --build
```

После запуска:
- **Front**: `http://localhost:3000`
- **API**: `http://localhost:8080`
- **Swagger UI**: `http://localhost:8081`

Остановить:

```bash
docker compose -f deployments/docker-compose.yml down
```

## Документация (Swagger UI)

- **Swagger UI**: `http://localhost:8081`
- **Спецификация OpenAPI**: `api/openapi.yaml`

## Быстрая проверка

Healthcheck:

```bash
curl http://localhost:8080/healthz
```

Создание подписки (пример из ТЗ):

```bash
curl -X POST http://localhost:8080/subscriptions \
  -H 'Content-Type: application/json' \
  -d '{"service_name":"Yandex Plus","price":400,"user_id":"60601fee-2bf1-4721-ae6f-7636e79a0cba","start_date":"07-2025"}'
```

Список подписок:

```bash
curl 'http://localhost:8080/subscriptions?user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba&limit=10&offset=0'
```

Сумма за период:

```bash
curl 'http://localhost:8080/subscriptions/total?user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba&service_name=Yandex%20Plus&from=07-2025&to=09-2025'
```

## Конфигурация

Сервис читает YAML-конфиг (по умолчанию `configs/config.yaml`) и позволяет переопределить значения через env.

- **CONFIG_PATH**: путь до YAML (в `docker-compose` выставлен `CONFIG_PATH=/app/configs/config.yaml`)
- **SERVER_HOST**, **SERVER_PORT**
- **LOG_LEVEL**: `debug|info|warn|error`
- **DB_HOST**, **DB_PORT**, **DB_NAME**, **DB_USER**, **DB_PASSWORD**, **DB_SSL_MODE**

## Миграции

Миграции лежат в `migrations/` и прогоняются при старте сервиса автоматически.

Если добавил новую миграцию и хочешь применить её локально:

```bash
docker compose -f deployments/docker-compose.yml up -d --build api
docker compose -f deployments/docker-compose.yml logs -f --tail=50 api
```

Полный сброс локальной БД (все данные будут удалены):

```bash
docker compose -f deployments/docker-compose.yml down -v
docker compose -f deployments/docker-compose.yml up -d --build
```

## Разработка без полного compose

### Backend (Go) + Postgres (в контейнере)

Поднять только Postgres:

```bash
docker compose -f deployments/docker-compose.yml up -d postgres
```

Запустить API локально:

```bash
cp .env.example .env
go run ./cmd/subscriptions-api
```

### Frontend (Vite)

Запуск в dev-режиме (прокси `/api` → `http://localhost:8080` настроен в `front/vite.config.ts`):

```bash
cd front
npm install
npm run dev
```

