# Tic-Tac-Toe API

REST API для игры в крестики-нолики, написанный на Go.

## Стек технологий

- **Go** — язык разработки
- **PostgreSQL** — база данных
- **uber-fx** — dependency injection
- **net/http** — HTTP-сервер
- **JWT** — аутентификация
- **Swagger (swaggo)** — документация API

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

## Требования

- Go 1.2x+
- PostgreSQL (запущенный локально или доступный по сети)

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
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=tictactoe

# JWT
JWT_SECRET=your_secret_key
JWT_ACCESS_SECRET=your_secret_key
```

### 3. Запустите проект

Всё поднимается одной командой через Makefile — она сама поднимет PostgreSQL в Docker и запустит приложение:

```bash
make run
```

## Swagger-документация

После запуска сервера документация API доступна автоматически по адресу:

```
http://localhost:8080/swagger/index.html#/
```

Через Swagger UI можно посмотреть все доступные эндпоинты, модели запросов/ответов и протестировать API прямо в браузере.

## Структура проекта

```
tic-tac-toe/
│
├── src/
│   ├── cmd/
│   │   └── main.go         # точка входа в приложение
│   ├── docs/               #  Swagger документация
│   │   ├── docs.go         
│   │   ├── swagger.json
│   │   └── swagger.yaml
│   └── internal/
│       ├── api/            # хендлеры, middleware, модели запросов
│       │   ├── auth_handler.go
│       │   ├── auth_model.go
│       │   ├── handler.go
│       │   ├── interface.go
│       │   ├── jwt_extention.go
│       │   ├── mapper.go
│       │   ├── middleware.go
│       │   └── model.go
│       ├── app/             # бизнес-логика / сервисный слой
│       │   ├── auth_service.go
│       │   ├── interface.go
│       │   └── service.go
│       ├── config/
│       │   └── config.go
│       ├── di/              # dependency injection (uber-fx) и миграции БД
│       │   ├── migrations/
│       │   └── di.go
│       ├── domain/          # доменные модели
│       │   ├── game.go
│       │   ├── jwt_model.go
│       │   ├── model.go
│       │   └── user.go
│       └── infra/           # инфраструктурный слой (Postgresql)
|           ├── connection.go
│           ├── jwt_provider.go
│           ├── mapper.go
│           ├── model.go
│           ├── repository.go
│           ├── user_models.go
│           └── user_repository.go
│
├── go.mod
├── README.md
├── makefile
├── .env
└── docker-compose.yaml

```
