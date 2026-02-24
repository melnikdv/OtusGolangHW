BIN_DIR := ./bin
BIN_CALENDAR      := $(BIN_DIR)/calendar
BIN_SCHEDULER     := $(BIN_DIR)/calendar_scheduler
BIN_SENDER        := $(BIN_DIR)/calendar_sender
CALENDAR_CONFIG := "./configs/calendar.yaml"
SCHEDULER_CONFIG := "./configs/scheduler.yaml"
SENDER_CONFIG := "./configs/sender.yaml"
IMG_CALENDAR      := calendar:latest
IMG_SCHEDULER     := calendar-scheduler:latest
IMG_SENDER        := calendar-sender:latest
.DEFAULT_GOAL := help

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

# Автоопределение docker-compose vs docker compose
DOCKER_COMPOSE := $(shell command -v docker-compose 2> /dev/null)
ifeq ($(DOCKER_COMPOSE),)
	DOCKER_COMPOSE := LDFLAGS='$(LDFLAGS)' docker compose
else
	DOCKER_COMPOSE := LDFLAGS='$(LDFLAGS)' docker-compose
endif
COMPOSE_FILE := ./deployments/docker-compose.yaml

# PostgreSQL
POSTGRES_USER ?= postgres
POSTGRES_PASSWORD ?= password
POSTGRES_DB ?= backend
POSTGRES_PORT ?= 5432
POSTGRES_CONTAINER := postgres-calendar
POSTGRES_APP_DB_NAME ?= calendar
POSTGRES_APP_USER ?= calendar
POSTGRES_APP_PASS ?= calendar
POSTGRES_MIGRATIONS_DIR ?= ./internal/migration

# RabbitMQ
RABBIT_USER ?= calendar
RABBIT_PASSWORD ?= calendar
RABBIT_PORT ?= 5672
RABBIT_PORT_UI ?= 15672
RABBIT_CONTAINER := rabbitmq

## Запустить PostgreSQL в Docker
run-postgres:
	docker run -d --name $(POSTGRES_CONTAINER) \
	-e POSTGRES_USER=$(POSTGRES_USER) \
	-e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) \
	-e POSTGRES_DB=$(POSTGRES_DB) \
	-p $(POSTGRES_PORT):5432 \
	-v postgres-data:/var/lib/postgresql \
	postgres:18

## Остановить и удалить контейнер PostgreSQL
stop-postgres:
	docker rm -f $(POSTGRES_CONTAINER)

## Запустить RabbitMQ в Docker
run-rabbit:
	docker run -d --name $(RABBIT_CONTAINER) \
	-p $(RABBIT_PORT):5672 \
	-p $(RABBIT_PORT_UI):15672 \
	-e RABBITMQ_DEFAULT_USER=$(RABBIT_USER) \
	-e RABBITMQ_DEFAULT_PASS=$(RABBIT_PASSWORD) \
	-e RABBITMQ_DEFAULT_VHOST="/" \
	rabbitmq:3-management

## Остановить и удалить контейнер RabbitMQ
stop-rabbit:
	docker rm -f $(RABBIT_CONTAINER)

## Инициализировать БД и пользователя для приложения
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

## Применить SQL-миграции к локальной БД
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

## Собрать бинарники: calendar, scheduler, sender
build: $(BIN_CALENDAR) $(BIN_SCHEDULER) $(BIN_SENDER)

## Сборка calendar
$(BIN_CALENDAR):
	@mkdir -p $(BIN_DIR)
	go build -v -o $@ -ldflags "$(LDFLAGS)" ./cmd/calendar

## Сборка calendar_scheduler
$(BIN_SCHEDULER):
	@mkdir -p $(BIN_DIR)
	go build -v -o $@ -ldflags "$(LDFLAGS)" ./cmd/calendar_scheduler

## Сборка calendar_sender
$(BIN_SENDER):
	@mkdir -p $(BIN_DIR)
	go build -v -o $@ -ldflags "$(LDFLAGS)" ./cmd/calendar_sender

## Запустить calendar локально
run-calendar:
	$(BIN_CALENDAR) --config $(CALENDAR_CONFIG)

## Запустить scheduler локально
run-scheduler:
	$(BIN_SCHEDULER) --config $(SCHEDULER_CONFIG)

## Запустить sender локально
run-sender:
	$(BIN_SENDER) --config $(SENDER_CONFIG)

## Запустить все сервисы локально (в фоне)
run-all: build
	@echo "Starting all services..."
	$(BIN_CALENDAR) --config $(CALENDAR_CONFIG) &
	$(BIN_SCHEDULER) --config $(SCHEDULER_CONFIG) &
	$(BIN_SENDER) --config $(SENDER_CONFIG) &
	@echo "All services started in background. PID: $$!"

## Убить все локальные процессы (calendar, scheduler, sender)
kill-all:
	-killall calendar calendar_scheduler calendar_sender 2>/dev/null || true

## Собрать Docker-образы всех сервисов
build-img:
	docker build \
		--build-arg LDFLAGS="$(LDFLAGS)" \
		-t $(IMG_CALENDAR) \
		-f build/Dockerfile.calendar .

	docker build \
		--build-arg LDFLAGS="$(LDFLAGS)" \
		-t $(IMG_SCHEDULER) \
		-f build/Dockerfile.scheduler .

	docker build \
		--build-arg LDFLAGS="$(LDFLAGS)" \
		-t $(IMG_SENDER) \
		-f build/Dockerfile.sender .

## Запустить всё через docker-compose (production-like)
up:
	$(DOCKER_COMPOSE) -f $(COMPOSE_FILE) up -d --build

## Остановить и удалить контейнеры из docker-compose
down:
	$(DOCKER_COMPOSE) -f $(COMPOSE_FILE) down

## Удалить Docker volumes (для полной очистки)
vrm:
	docker volume rm calendar_postgres_data calendar_rabbitmq_data

## Показать логи основных сервисов
logs:
	$(DOCKER_COMPOSE) -f $(COMPOSE_FILE) logs -f calendar calendar-scheduler calendar-sender

## Показать статус контейнеров
ps:
	$(DOCKER_COMPOSE) -f $(COMPOSE_FILE) ps -a

## Запустить интеграционные тесты в изолированном окружении
integration-tests:
	@set -e; \
	EXIT_CODE=0; \
	$(DOCKER_COMPOSE) -f tests/integration/docker-compose.test.yaml \
		up --build \
		--abort-on-container-exit \
		--exit-code-from integration-test \
		--attach integration-test \
	|| EXIT_CODE=$$?; \
	$(DOCKER_COMPOSE) -f tests/integration/docker-compose.test.yaml down -v --remove-orphans; \
	if [ $$EXIT_CODE -eq 0 ]; then \
	    echo "Integration tests passed"; \
	    exit 0; \
	else \
		echo "Integration tests failed"; \
		exit 1; \
	fi

## Показать версию calendar (требует сборки)
version: build
	$(BIN_CALENDAR) version

## Запустить unit-тесты
test:
	go test -race ./internal/...

## Установить golangci-lint (если отсутствует)
install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh   | sh -s -- -b $(shell go env GOPATH)/bin v1.62.2

## Запустить линтер
lint: install-lint-deps
	golangci-lint run ./...

## Сгенерировать gRPC код из .proto
generate:
	protoc --go_out=. --go-grpc_out=. api/EventService.proto

## Удалить собранные бинарники
clean:
	rm -f $(BIN_CALENDAR) $(BIN_SCHEDULER) $(BIN_SENDER)

## Показать эту справку
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk '/^[a-zA-Z0-9_-]+:/ { \
		help_line = match(prev, /^## (.*)/); \
		if (help_line) { \
			printf "  %-25s %s\n", $$1, substr(prev, RSTART + 3, RLENGTH - 3) \
		} else { \
			printf "  %-25s %s\n", $$1, "[no description]" \
		} \
	} \
	{ prev = $$0 }' $(MAKEFILE_LIST)

.PHONY: \
	help build run-calendar run-scheduler run-sender run-all kill-all \
	build-img up down vrm logs ps integration-tests version test \
	lint install-lint-deps generate clean \
	run-postgres stop-postgres run-rabbit stop-rabbit initdb migrate
