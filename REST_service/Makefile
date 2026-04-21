.PHONY: help build run test clean docker-up docker-down migrate-up migrate-down swagger

help:
	@echo "Available commands:"
	@echo "  make build       - Build the application"
	@echo "  make run         - Run the application locally"
	@echo "  make test        - Run tests"
	@echo "  make docker-up   - Start all services with docker-compose"
	@echo "  make docker-down - Stop all services"
	@echo "  make migrate-up  - Run database migrations"
	@echo "  make migrate-down- Rollback database migrations"
	@echo "  make swagger     - Generate Swagger documentation"

build:
	go build -o bin/server ./cmd

run:
	go run ./cmd

test:
	go test -v ./...

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

migrate-up:
	migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/subscriptions?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/subscriptions?sslmode=disable" down

swagger:
	swag init -g cmd/main.go -o docs

clean:
	rm -rf bin/
	rm -rf docs/