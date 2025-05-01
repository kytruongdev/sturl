PROJECT_NAME=sturl
DB_CONTAINER=database
MIGRATE_CONTAINER=migrate
SERVER_CONTAINER=server

db:
	docker-compose -p $(PROJECT_NAME) up -d $(DB_CONTAINER)

migrate:
	docker-compose -p $(PROJECT_NAME) run --rm $(MIGRATE_CONTAINER)

setup: db migrate

run:
	docker-compose -p $(PROJECT_NAME) up --build $(SERVER_CONTAINER)

down:
	docker-compose down

generate-models:
	sqlboiler psql --no-tests

