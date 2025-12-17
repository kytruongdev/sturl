# =========================================================
# Global config
# =========================================================
ENV ?= local
COMPOSE_FILE := build/docker-compose.$(ENV).yml
PROJECT_NAME := api-gateway

GATEWAY_CONTAINER := gateway

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
# Core lifecycle
# =========================================================
.PHONY: run down logs

run:
ifeq ($(ENV),prod)
	docker compose -f $(COMPOSE_FILE) -p $(PROJECT_NAME) up --build $(GATEWAY_CONTAINER)
else
	docker compose -f $(COMPOSE_FILE) -p $(PROJECT_NAME) up $(GATEWAY_CONTAINER)
endif

logs:
	docker compose -f $(COMPOSE_FILE) -p $(PROJECT_NAME) logs -f

down:
	docker compose -f $(COMPOSE_FILE) -p $(PROJECT_NAME) down

# =========================================================
# Observability (local only)
# =========================================================
.PHONY: jaeger

jaeger:
	docker compose -f $(COMPOSE_FILE) -p $(PROJECT_NAME) up -d jaeger

# =========================================================
# Testing
# =========================================================
.PHONY: test

# run single test case:
# env $(grep -v '^#' ./local.env | egrep -v '^#' | xargs) go test -cover -mod=vendor -v ./internal/infra/proxy -run TestRegister
test:
	env $(shell cat local.env | egrep -v '^#' | xargs) \
	sh -c 'go test -p 1 -failfast -timeout 5m ./...' \
	&& echo "\n[Test OK]" || (echo "\n[Test FAILED]" && exit 1)
