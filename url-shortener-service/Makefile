.PHONY: db migrate redis setup run down generate-orm generate-mock api-test reset-db

PROJECT_NAME = url-shortener
DB_CONTAINER = database
DB_CONTAINER_NAME = pg
DB_NAME = url-shortener
MIGRATE_CONTAINER = migrate
SERVER_CONTAINER = server
REDIS_CONTAINER = redis

db:
	docker compose -p $(PROJECT_NAME) up -d $(DB_CONTAINER)

migrate:
	docker compose -p $(PROJECT_NAME) run --rm $(MIGRATE_CONTAINER)

redis:
	docker compose -p $(PROJECT_NAME) up -d $(REDIS_CONTAINER)

setup: db migrate redis

run:
	docker compose -p $(PROJECT_NAME) up --build $(SERVER_CONTAINER)

down:
	docker compose down

generate-orm:
	sqlboiler psql --no-tests

generate-mock:
	docker compose run --rm --entrypoint '' mockery /bin/sh -c "\
		mockery --dir internal/controller --all --recursive --inpackage && \
		mockery --dir internal/repository --all --recursive --inpackage"

# run single test case:
# env $(grep -v '^#' ./local.env | egrep -v '^#' | xargs) go test -cover -mod=vendor -v ./internal/repository/shorturl -run TestGetByOriginalURL
api-test:
	env $(shell cat local.env | egrep -v '^#' | xargs) \
	sh -c 'go test -p 1 -mod=vendor -failfast -timeout 5m ./...' \
	&& echo "\n[Test okay -- `date`]" || (echo "\n[Test FAILED -- `date`]" && exit 1)

reset-db:
	docker exec -i $(DB_CONTAINER_NAME) psql -U postgres -d $(DB_NAME) -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
