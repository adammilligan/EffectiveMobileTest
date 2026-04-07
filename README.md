# Запуск проекта

Проект состоит из:
- **API** (Go) — `http://localhost:8080`
- **Swagger UI** — `http://localhost:8081`
- **Front** (React/Vite/MUI, отдаётся через nginx) — `http://localhost:3000`

## Запуск в Dev Container

1) Установить расширение VS Code `Dev Containers` (Remote - Containers).

2) Открыть проект в контейнере одним из способов:
- через меню в левом нижнем углу VS Code → `Reopen in Container`
- через палитру команд (Cmd+Shift+P) → `Dev Containers: Rebuild and Reopen in Container`

После успешной сборки контейнера автоматически будут подняты сервисы:
- **Postgres** (для API)
- **Swagger UI** (документация)

Далее команды выполняются в терминале Dev Container.

Запуск API:

```bash
go run ./cmd/subscriptions-api
```

Запуск Front в dev‑режиме:

```bash
cd front
npm install
npm run dev -- --host 0.0.0.0 --port 5173
```

Адреса по умолчанию:
- **API**: `http://localhost:8080`
- **Swagger UI**: `http://localhost:8081`
- **Front (dev)**: `http://localhost:5173`

---

## Запуск через Docker Compose (рекомендуется)

1) Скопировать `.env.example` в `.env`

```bash
cp .env.example .env
```

2) Поднять все сервисы

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

# Разработка

1) Скопировать `.env.example` в `.env`.
2) Запустить сервисы выбранным способом (полный стек через Docker Compose или выборочно).

## Запуск сервисов

### Вариант A: поднять всё одной командой

```bash
docker compose -f deployments/docker-compose.yml up -d --build
```

### Вариант B: Postgres в контейнере, API локально

Поднять только Postgres:

```bash
docker compose -f deployments/docker-compose.yml up -d postgres
```

Запустить API локально:

```bash
go run ./cmd/subscriptions-api
```

### Вариант C: Front в dev‑режиме (Vite)

Dev‑режим (прокси `/api` → `http://localhost:8080` настроен в `front/vite.config.ts`):

```bash
cd front
npm install
npm run dev
```

## Миграции

Миграции лежат в `migrations/` и применяются **автоматически при старте API**.

Пересобрать и перезапустить только API (например, после добавления миграции):

```bash
docker compose -f deployments/docker-compose.yml up -d --build api
docker compose -f deployments/docker-compose.yml logs -f --tail=50 api
```

Полный сброс локальной БД (удалит данные):

```bash
docker compose -f deployments/docker-compose.yml down -v
docker compose -f deployments/docker-compose.yml up -d --build
```

## Документация (Swagger)

- **OpenAPI**: `api/openapi.yaml`
- **Swagger UI**: `http://localhost:8081`

## Быстрая проверка API

Healthcheck:

```bash
curl http://localhost:8080/healthz
```

Создание подписки:

```bash
curl -X POST http://localhost:8080/subscriptions \
  -H 'Content-Type: application/json' \
  -d '{"service_name":"Yandex Plus","price":400,"user_id":"60601fee-2bf1-4721-ae6f-7636e79a0cba","start_date":"07-2025"}'
```

Сумма за период:

```bash
curl 'http://localhost:8080/subscriptions/total?user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba&service_name=Yandex%20Plus&from=07-2025&to=09-2025'
```

## Конфигурация

Бэкенд читает YAML‑конфиг (по умолчанию `configs/config.yaml`) и позволяет переопределить значения через env.

- **CONFIG_PATH**: путь до YAML (в `docker-compose` используется `CONFIG_PATH=/app/configs/config.yaml`)
- **SERVER_HOST**, **SERVER_PORT**
- **LOG_LEVEL**: `debug|info|warn|error`
- **DB_HOST**, **DB_PORT**, **DB_NAME**, **DB_USER**, **DB_PASSWORD**, **DB_SSL_MODE**

## Тесты

```bash
go test ./...
```

## Линтинг

```bash
golangci-lint run ./...
```

## Частые проблемы

### API не стартует и ругается на БД

Проверить, что Postgres запущен и находится в состоянии `healthy`:

```bash
docker compose -f deployments/docker-compose.yml ps
```

### Front (dev) не ходит в API

- Убедиться, что API отвечает на `http://localhost:8080/healthz`.
- Убедиться, что Front запущен командой `npm run dev` в директории `front/` (используется прокси `/api`).

