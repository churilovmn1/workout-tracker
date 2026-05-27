.PHONY: build run test migrate-up migrate-down lint

DATABASE_URL ?= postgres://postgres:postgres@localhost:5432/workout_tracker?sslmode=disable

build:
	go build -o bin/workout-tracker ./cmd/

run:
	go run ./cmd/

test:
	go test ./... -v

lint:
	golangci-lint run ./...

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down 1
