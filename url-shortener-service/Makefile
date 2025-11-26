.PHONY: create-network db migrate redis setup run down generate-orm generate-mock test reset-db boilerplate

PROJECT_NAME = url-shortener
DB_CONTAINER = database
DB_CONTAINER_NAME = pg
DB_NAME = url-shortener
MIGRATE_CONTAINER = migrate
SERVER_CONTAINER = server
PRODUCER_CONTAINER = producer
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
	docker exec kafka sh -c "\
	  kafka-topics --bootstrap-server kafka:9092 \
	    --create \
	    --topic urlshortener.metadata.requested.v1 \
	    --partitions 1 \
	    --replication-factor 1 \
	    || true"

kafka-console:
	docker exec -it $(KAFKA_CONTAINER) kafka-console-consumer \
	  --topic $(TOPIC) \
	  --bootstrap-server $(BROKER) \
	  --from-beginning