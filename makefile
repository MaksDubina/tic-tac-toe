# Запуск приложения локально с подтягиванием окружения
run:
	docker-compose up -d postgres
	go run main.go

# Альтернативный вариант запуска через shell-скрипт, если не хотите использовать .env файл:
run-env:
	DB_USER=postgres \
	DB_PASSWORD=my_secret_password \
	DB_HOST=localhost \
	DB_PORT=5432 \
	DB_NAME=tictactoe \
	DB_SSLMODE=disable \
	go run main.go
