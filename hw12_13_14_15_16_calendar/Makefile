BIN := "./bin/calendar"
CONFIG := "./configs/config.yaml"
DOCKER_IMG="calendar:develop"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

#Postrges
POSTGRES_USER ?= postgres
POSTGRES_PASSWORD ?= password
POSTGRES_DB ?= backend
POSTGRES_PORT ?= 5432
POSTGRES_CONTAINER := postgres-calendar
POSTGRES_APP_DB_NAME ?= calendar
POSTGRES_APP_USER ?= calendar
POSTGRES_APP_PASS ?= calendar
POSTGRES_MIGRATIONS_DIR ?= ./internal/migration

#Rabbit
RABBIT_USER ?= rabbit
RABBIT_PASSWORD ?= password
RABBIT_PORT ?= 5672
RABBIT_PORT_UI ?= 15672
RABBIT_CONTAINER := rabbitmq


run-postgres:
	docker run -d --name $(POSTGRES_CONTAINER) \
	-e POSTGRES_USER=$(POSTGRES_USER) \
	-e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) \
	-e POSTGRES_DB=$(POSTGRES_DB) \
	-p $(POSTGRES_PORT):5432 \
	-v postgres-data:/var/lib/postgresql \
	postgres:latest

stop-postgres:
	docker rm -f $(POSTGRES_CONTAINER)

run-rabbit:
	docker run -d --name $(RABBIT_CONTAINER) \
	-p $(RABBIT_PORT):5672 \
	-p $(RABBIT_PORT_UI):15672 \
	-e RABBITMQ_DEFAULT_USER=$(RABBIT_USER) \
	-e RABBITMQ_DEFAULT_PASS=$(RABBIT_PASSWORD) \
	rabbitmq:management

stop-rabbit:
	docker rm -f $(RABBIT_CONTAINER)

# Команда: инициализация БД и пользователя
initdb:
	@echo "🔍 Checking if user $(POSTGRES_APP_USER) exists..."
	@docker exec -i $(POSTGRES_CONTAINER) psql -U $(POSTGRES_USER) -tAc "SELECT 1 FROM pg_roles WHERE rolname='$(POSTGRES_APP_USER)';" | grep -q 1 || \
		(docker exec -i $(POSTGRES_CONTAINER) psql -U $(POSTGRES_USER) -c "CREATE USER $(POSTGRES_APP_USER) WITH PASSWORD '$(POSTGRES_APP_PASS)';" && echo "✅ User created.") || \
		(echo "❌ Failed to create user $(POSTGRES_APP_USER)" && exit 1)

	@echo "🔍 Checking if database $(POSTGRES_APP_DB_NAME) exists..."
	@docker exec -i $(POSTGRES_CONTAINER) psql -U $(POSTGRES_USER) -tAc "SELECT 1 FROM pg_database WHERE datname='$(POSTGRES_APP_DB_NAME)';" | grep -q 1 || \
		(docker exec -i $(POSTGRES_CONTAINER) psql -U $(POSTGRES_USER) -c "CREATE DATABASE $(POSTGRES_APP_DB_NAME) OWNER $(POSTGRES_APP_USER);" && echo "✅ Database created.") || \
		(echo "❌ Failed to create database $(POSTGRES_APP_DB_NAME)" && exit 1)

	@echo "✅ Database and user ready."

# Команда: применить все миграции по порядку (если нужно)
migrate:
	@echo "Applying migrations from $(POSTGRES_MIGRATIONS_DIR)..."
	@if [ ! -d "$(POSTGRES_MIGRATIONS_DIR)" ]; then \
		echo "❌ Migrations directory not found: $(POSTGRES_MIGRATIONS_DIR)"; \
		exit 1; \
	fi
	@set -o pipefail; \
	for f in $$(ls -1v $(POSTGRES_MIGRATIONS_DIR)/*.up.sql 2>/dev/null || true); do \
		if [ -n "$$f" ]; then \
			echo "→ Applying $$f"; \
			cat "$$f" | docker exec -i $(POSTGRES_CONTAINER) psql -v ON_ERROR_STOP=1 -U $(POSTGRES_APP_USER) -d $(POSTGRES_APP_DB_NAME); \
		fi; \
	done
	@echo "✅ All migrations applied."

build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/calendar

run: build
	$(BIN) --config $(CONFIG)

build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .

run-img: build-img
	docker run $(DOCKER_IMG)

version: build
	$(BIN) version

test:
	go test -race ./internal/...

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.62.2

lint: install-lint-deps
	golangci-lint run ./...

.PHONY: build run build-img run-img version test lint
