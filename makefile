run:
	go mod tidy
	docker-compose up -d postgres
	go run main.go

