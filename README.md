# Tic-Tac-Toe API

REST API для игры в крестики-нолики, написанный на Go.

## Стек технологий

- **Go** — язык разработки
- **PostgreSQL** — база данных
- **uber-fx** — dependency injection
- **net/http** — HTTP-сервер
- **JWT** — аутентификация
- **Swagger (swaggo)** — документация API
- **Docker / Docker Compose** — контейнеризация приложения и всей инфраструктуры
- **Prometheus** — сбор метрик приложения
- **Grafana** — визуализация метрик и логов, дашборды
- **Loki + Promtail** — централизованный сбор и просмотр логов контейнеров

## Возможности

**Auth**
- Регистрация нового пользователя — `POST /auth/signup`
- Авторизация пользователя (вход) — `POST /auth/signin`
- Обновление access-токена — `POST /auth/refresh-access`
- Полное обновление пары токенов (ротация) — `POST /auth/refresh-token`

**Game**
- Создание новой игровой комнаты — `POST /game`
- Получение информации о конкретной игре — `GET /game/{current_game_UUID}`
- Совершение хода в игре — `POST /game/{current_game_UUID}`
- Присоединение к существующей игре — `POST /game/{current_game_UUID}/join`
- Получение списка доступных игр — `GET /games/available`
- Получение истории завершённых игр пользователя — `GET /games/history`
- Получение таблицы лидеров — `GET /games/leaderboard`

**User**
- Получение информации о текущем пользователе — `GET /user/me`
- Получение информации о пользователе по UUID — `GET /users/{user_uuid}`

**Observability**
- Метрики приложения в формате Prometheus — `GET /metrics`

## Мониторинг и наблюдаемость

Приложение инструментировано метриками Prometheus на двух уровнях:

- **Технические метрики** (через middleware): `http_requests_total`, `http_request_duration_seconds` — количество и латентность HTTP-запросов в разрезе метода, пути и статус-кода.
- **Бизнес-метрики** (в хендлерах): `games_created_total`, `games_active`, `moves_total`, `auth_attempts_total` — количество созданных/активных игр, совершённых ходов, попыток регистрации и входа (с разбивкой success/failure).

Логи контейнеров собираются Promtail и отправляются в Loki, откуда доступны для просмотра и поиска через Grafana.

Grafana поднимается с уже готовым datasource (Prometheus + Loki) и дашбордом через provisioning — заходить и настраивать вручную не требуется.

## Требования

- Go 1.26
- Docker и Docker Compose

## Установка и запуск

### 1. Клонируйте репозиторий

```bash
git clone https://github.com/MaksDubina/tic-tac-toe.git
cd tic-tac-toe
```

### 2. Создайте файл окружения

Переменные окружения не хранятся в репозитории из соображений безопасности. Создайте файл `.env` в корне проекта на основе примера ниже:

```bash
touch .env
```

и заполните его своими значениями:

```env
# Сервер
SERVER_PORT=8080

# База данных
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=tictactoe

# JWT
JWT_SECRET=your_secret_key
JWT_ACCESS_SECRET=your_secret_key
```

> **Важно:** `DB_HOST` должен быть `postgres` (имя сервиса в docker-compose), а не `localhost` — приложение теперь тоже запускается в контейнере и обращается к базе данных по внутренней docker-сети.

### 3. Запустите проект

Весь стек — приложение, база данных, Prometheus, Grafana, Loki и Promtail — поднимается одной командой:

```bash
docker compose up --build
```

При первом запуске Docker соберёт образ приложения и подтянет остальные образы — это может занять несколько минут.

## Доступные интерфейсы после запуска

| Сервис | Адрес | Назначение |
|---|---|---|
| API | http://localhost:8080 | Основные эндпоинты приложения |
| Swagger UI | http://localhost:8080/swagger/index.html#/ | Документация и тестирование API |
| Метрики (Prometheus формат) | http://localhost:8080/metrics | Сырые метрики приложения |
| Prometheus | http://localhost:9090 | Просмотр и запросы метрик (PromQL) |
| Grafana | http://localhost:3000 | Дашборды и визуализация (логин/пароль: `admin` / `admin`) |
| Loki | http://localhost:3100 | API логов (используется через Grafana, напрямую обычно не нужен) |

## Структура проекта

```
tic-tac-toe/
│
├── src/
│   ├── cmd/
│   │   └── main.go              # точка входа в приложение
│   ├── docs/                    # Swagger документация
│   │   ├── docs.go
│   │   ├── swagger.json
│   │   └── swagger.yaml
│   └── internal/
│       ├── api/                 # хендлеры, middleware, модели запросов
│       │   ├── auth_handler.go
│       │   ├── auth_model.go
│       │   ├── handler.go
│       │   ├── interface.go
│       │   ├── jwt_extention.go
│       │   ├── mapper.go
│       │   ├── middleware.go
│       │   └── model.go
│       ├── app/                 # бизнес-логика / сервисный слой
│       │   ├── auth_service.go
│       │   ├── interface.go
│       │   └── service.go
│       ├── config/
│       │   └── config.go
│       ├── di/                  # dependency injection (uber-fx) и миграции БД
│       │   ├── migrations/
│       │   └── di.go
│       ├── domain/              # доменные модели
│       │   ├── game.go
│       │   ├── jwt_model.go
│       │   ├── model.go
│       │   └── user.go
│       ├── infra/               # инфраструктурный слой (PostgreSQL)
│       │   ├── connection.go
│       │   ├── jwt_provider.go
│       │   ├── mapper.go
│       │   ├── model.go
│       │   ├── repository.go
│       │   ├── user_models.go
│       │   └── user_repository.go
│       └── metrics/             # регистрация и определения метрик Prometheus
│           └── metrics.go
│
├── prometheus/
│   └── prometheus.yml           # конфиг scrape-таргетов Prometheus
│
├── loki/
│   └── loki-config.yml          # конфиг хранилища и лимитов Loki
│
├── promtail/
│   └── promtail-config.yml      # конфиг сбора логов Docker-контейнеров
│
├── grafana/
│   └── provisioning/
│       ├── datasources/
│       │   └── datasources.yml  # автоподключение Prometheus и Loki
│       └── dashboards/
│           └── dashboards.yml   # автозагрузка дашбордов из файлов
│
├── Dockerfile                   # сборка образа приложения
├── docker-compose.yaml          # весь стек: app, postgres, prometheus, grafana, loki, promtail
├── go.mod
├── README.md
├── makefile
└── .env
```