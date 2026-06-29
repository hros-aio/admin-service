.PHONY: run test lint docker-up docker-down migrate tidy build

APP_NAME=admin-service
MAIN_PATH=cmd/api/main.go

run:
	go run $(MAIN_PATH)

test:
	go test ./... -v -race -count=1

test-cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

lint:
	golangci-lint run ./...

format:
	go run mvdan.cc/gofumpt@latest -w -l .

docker-up:
	docker compose up -d

docker-down:
	docker compose down

migrate:
	@echo "Running migrations..."
	# In a real scenario, this would call a migration tool or a go command
	# For now, we'll assume the app runs AutoMigrate on startup or we add a command
	go run $(MAIN_PATH) -migrate

tidy:
	go mod tidy

build:
	go build -o bin/$(APP_NAME) $(MAIN_PATH)
