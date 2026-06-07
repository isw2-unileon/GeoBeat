TEST_DB_URL="postgresql://postgres:supersecret@localhost:5433/geobeat_test?sslmode=disable"
LOCAL_DB_URL="postgresql://postgres:supersecret@localhost:5432/geobeat_local?sslmode=disable"
MIGRATION_PATH="backend/internal/database/migrations"

.PHONY: install run-backend run-frontend build-backend build-frontend test lint e2e db-test-up db-test-down migrate-up migrate-down migrate-reset

## Install all dependencies
install:
	go install github.com/air-verse/air@latest
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $$(go env GOPATH)/bin
	go mod download
	cd frontend && npm ci
	cd e2e && npm ci

## Run backend with hot reload
run-backend:
	$(shell go env GOPATH)/bin/air -c backend/.air.toml

## Run frontend dev server
run-frontend:
	cd frontend && npm run dev

## Build backend binary
build-backend:
	go build -o backend/bin/server ./backend/cmd/server

## Build frontend for production
build-frontend:
	cd frontend && npm run build

## Run all tests
test:
	TEST_DATABASE_URL=$(TEST_DB_URL) go test -v -race ./...
	cd frontend && npm run test

## Run linters
lint:
	$(shell go env GOPATH)/bin/golangci-lint run
	cd frontend && npm run lint

## Run E2E tests (requires backend + frontend running)
e2e:
	cd e2e && npx playwright test

## Start the test database in the background
db-test-up:
	docker-compose up -d geobeat-test-db

## Stop and completely remove the test database
db-test-down:
	docker-compose down geobeat-test-db

## Apply all up migrations
migrate-up:
	migrate -path $(MIGRATION_PATH) -database $(TEST_DB_URL) up

## Rollback all migrations (Destroys all tables)
migrate-down:
	migrate -path $(MIGRATION_PATH) -database $(TEST_DB_URL) down -all

## Nuke the database and rebuild it fresh
migrate-reset: migrate-down migrate-up


## Start the local development database
db-local-up:
	docker compose up -d geobeat-local-db

## Stop the local development database
db-local-down:
	docker compose stop geobeat-local-db

## Apply all up migrations to the LOCAL database
migrate-up:
	migrate -path $(MIGRATION_PATH) -database $(LOCAL_DB_URL) up

## Rollback all migrations on the LOCAL database
migrate-down:
	migrate -path $(MIGRATION_PATH) -database $(LOCAL_DB_URL) down -all

## Nuke the LOCAL database and rebuild it fresh
migrate-reset: migrate-down migrate-up
