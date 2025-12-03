.PHONY: create-network db migrate redis setup run down generate-orm generate-mock test reset-db boilerplate run-producer run-consumer logs-consumer kafka-ui-up kafka-ui-down kafka-ui-restart

PROJECT_NAME = url-shortener
DB_CONTAINER = database
DB_CONTAINER_NAME = pg
DB_NAME = url-shortener
MIGRATE_CONTAINER = migrate
SERVER_CONTAINER = server
PRODUCER_CONTAINER = producer
CONSUMER_CONTAINER = consumer
REDIS_CONTAINER = redis
KAFKA_CONTAINER=kafka
BROKER=kafka:9092
TOPIC=urlshortener.metadata.requested.v1

create-network:
	docker network create sturl-net || true

db:
	docker compose -p $(PROJECT_NAME) up -d $(DB_CONTAINER)

migrate:
	docker compose -p $(PROJECT_NAME) run --rm $(MIGRATE_CONTAINER)

redis:
	docker compose -p $(PROJECT_NAME) up -d $(REDIS_CONTAINER)

setup: create-network db migrate redis

run:
	docker compose -p $(PROJECT_NAME) up --build $(SERVER_CONTAINER)

run-producer:
	docker compose -p $(PROJECT_NAME) up --build $(PRODUCER_CONTAINER)

run-consumer:
	docker compose -p $(PROJECT_NAME) up --build $(CONSUMER_CONTAINER)

logs-consumer:
	docker logs -f consumer

down:
	docker compose down

generate-orm:
	sqlboiler psql --no-tests

generate-mock:
	docker compose run --rm --entrypoint '' mockery /bin/sh -c "\
		mockery --dir internal/controller --all --recursive --inpackage && \
		mockery --dir internal/repository --all --recursive --inpackage"

boilerplate: migrate generate-orm generate-mock

# run single test case:
# env $(grep -v '^#' ./local.env | egrep -v '^#' | xargs) go test -cover -mod=vendor -v ./internal/repository/shorturl -run TestGetByOriginalURL
test:
	env $(shell cat local.env | egrep -v '^#' | xargs) \
	sh -c 'go test -p 1 -mod=vendor -failfast -timeout 5m ./...' \
	&& echo "\n[Test okay -- `date`]" || (echo "\n[Test FAILED -- `date`]" && exit 1)

reset-db:
	docker exec -i $(DB_CONTAINER_NAME) psql -U postgres -d $(DB_NAME) -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"

# kafka commands
kafka:
	docker compose -p $(PROJECT_NAME) up -d $(KAFKA_CONTAINER)

kafka-down:
	docker compose stop kafka

kafka-topics:
	docker exec -it $(KAFKA_CONTAINER) kafka-topics --list --bootstrap-server $(BROKER)

kafka-create-topic:
	@[ -n "$(topic)" ] || (echo "Missing TOPIC. Use: make kafka-create-topic topic=your.topic.name" && exit 1)

	@echo "Checking if topic '$(topic)' exists..."
	@if docker exec kafka /opt/kafka/bin/kafka-topics.sh --list --bootstrap-server kafka:9092 | grep -q "^$(topic)$$"; then \
		echo "Topic '$(topic)' already exists — skipping creation."; \
	else \
		echo "Creating topic '$(topic)'..."; \
		docker exec kafka /opt/kafka/bin/kafka-topics.sh --create \
			--topic $(topic) \
			--bootstrap-server kafka:9092 \
			--partitions 1 \
			--replication-factor 1; \
	fi

kafka-ui-up:
	docker compose -p $(PROJECT_NAME) up -d redpanda-console

kafka-ui-down:
	docker compose -p $(PROJECT_NAME) stop redpanda-console
	docker compose -p $(PROJECT_NAME) rm -f redpanda-console

kafka-ui-restart:
	docker compose -p $(PROJECT_NAME) restart redpanda-console
