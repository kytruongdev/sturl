# =========================================================
# Global config
# =========================================================
ENV ?= local
COMPOSE_FILE := build/docker-compose.$(ENV).yml
PROJECT_NAME := url-shortener

SERVER_CONTAINER   := server
PRODUCER_CONTAINER := producer
CONSUMER_CONTAINER := consumer
DB_CONTAINER       := database
REDIS_CONTAINER    := redis
KAFKA_CONTAINER    := kafka
MIGRATE_CONTAINER  := migrate

DB_CONTAINER_NAME := pg
DB_NAME := url-shortener
BROKER := kafka:9092

# =========================================================
# Build (prod only)
# =========================================================
.PHONY: build

build:
ifeq ($(ENV),prod)
	docker build -t $(PROJECT_NAME):1.0.0 .
else
	@echo "'make build' is for prod only. Use ENV=prod"
	@exit 1
endif

# =========================================================
# Core lifecycle (most used)
# =========================================================
.PHONY: run run-producer run-consumer logs down

run:
ifeq ($(ENV),prod)
	docker compose -f $(COMPOSE_FILE) -p $(PROJECT_NAME) up --build $(SERVER_CONTAINER)
else
	docker compose -f $(COMPOSE_FILE) -p $(PROJECT_NAME) up $(SERVER_CONTAINER)
endif

run-producer:
ifeq ($(ENV),prod)
	docker compose -f $(COMPOSE_FILE) -p $(PROJECT_NAME) up --build $(PRODUCER_CONTAINER)
else
	docker compose -f $(COMPOSE_FILE) -p $(PROJECT_NAME) up $(PRODUCER_CONTAINER)
endif

run-consumer:
ifeq ($(ENV),prod)
	docker compose -f $(COMPOSE_FILE) -p $(PROJECT_NAME) up --build $(CONSUMER_CONTAINER)
else
	docker compose -f $(COMPOSE_FILE) -p $(PROJECT_NAME) up $(CONSUMER_CONTAINER)
endif

logs:
	docker compose -f $(COMPOSE_FILE) -p $(PROJECT_NAME) logs -f

down:
	docker compose -f $(COMPOSE_FILE) -p $(PROJECT_NAME) down

# =========================================================
# Infra setup (DB / Redis / Network)
# =========================================================
.PHONY: create-network db redis migrate setup

create-network:
	docker network create sturl-net || true

db:
	docker compose -f $(COMPOSE_FILE) -p $(PROJECT_NAME) up -d $(DB_CONTAINER)

redis:
	docker compose -f $(COMPOSE_FILE) -p $(PROJECT_NAME) up -d $(REDIS_CONTAINER)

migrate:
	docker compose -f $(COMPOSE_FILE) -p $(PROJECT_NAME) run --rm $(MIGRATE_CONTAINER)

setup: create-network db migrate redis

# =========================================================
# Dev tooling (ORM / mocks)
# =========================================================
.PHONY: generate-orm generate-mock boilerplate

generate-orm:
	sqlboiler psql --no-tests

generate-mock:
	docker compose -f $(COMPOSE_FILE) run --rm --entrypoint '' mockery \
		/bin/sh -c "mockery --dir internal/controller --all --recursive --inpackage && \
		            mockery --dir internal/repository --all --recursive --inpackage"

boilerplate: migrate generate-orm generate-mock

# =========================================================
# Testing
# =========================================================
.PHONY: test reset-db

# run single test case:
# env $(grep -v '^#' ./local.env | egrep -v '^#' | xargs) go test -cover -mod=vendor -v ./internal/repository/shorturl -run TestGetByOriginalURL
test:
	env $(shell cat local.env | egrep -v '^#' | xargs) \
	sh -c 'go test -p 1 -failfast -timeout 5m ./...' \
	&& echo "\n[Test OK]" || (echo "\n[Test FAILED]" && exit 1)

reset-db:
	docker exec -i $(DB_CONTAINER_NAME) psql -U postgres -d $(DB_NAME) \
		-c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"

# =========================================================
# Kafka lifecycle
# =========================================================
.PHONY: kafka kafka-down

kafka:
	docker compose -f $(COMPOSE_FILE) -p $(PROJECT_NAME) up -d $(KAFKA_CONTAINER)

kafka-down:
	docker compose -f $(COMPOSE_FILE) -p $(PROJECT_NAME) stop $(KAFKA_CONTAINER)

# =========================================================
# Kafka operations
# =========================================================
.PHONY: kafka-topics kafka-create-topic

kafka-topics:
	docker exec $(KAFKA_CONTAINER) /opt/kafka/bin/kafka-topics.sh --list --bootstrap-server $(BROKER)

kafka-create-topic:
	@[ -n "$(topic)" ] || (echo "Missing TOPIC. Use: make kafka-create-topic topic=your.topic.name" && exit 1)

	@echo "Checking if topic '$(topic)' exists..."
	@if docker exec $(KAFKA_CONTAINER) /opt/kafka/bin/kafka-topics.sh --list --bootstrap-server $(BROKER) | grep -q "^$(topic)$$"; then \
		echo "Topic '$(topic)' already exists — skipping creation."; \
	else \
		echo "Creating topic '$(topic)'..."; \
		docker exec $(KAFKA_CONTAINER) /opt/kafka/bin/kafka-topics.sh --create \
			--topic $(topic) \
			--bootstrap-server kafka:9092 \
			--partitions 1 \
			--replication-factor 1; \
	fi

kafka-delete-topic:
	docker exec -it $(KAFKA_CONTAINER) /opt/kafka/bin/kafka-topics.sh \
	  --bootstrap-server $(BROKER) \
	  --delete \
	  --topic $(topic)

kafka-ui-up:
	docker compose -f $(COMPOSE_FILE) -p $(PROJECT_NAME) up -d redpanda-console

# =========================================================
# Convenience / info
# =========================================================
.PHONY: ps

ps:
	docker compose -f $(COMPOSE_FILE) -p $(PROJECT_NAME) ps
