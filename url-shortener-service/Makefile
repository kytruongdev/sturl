PROJECT_NAME=url-shortener
DB_CONTAINER=database
DB_CONTAINER_NAME=pg
DB_NAME=url-shortener
MIGRATE_CONTAINER=migrate
SERVER_CONTAINER=server
REDIS_CONTAINER=redis
MOCKERY_CONTAINER=mockery

db:
	docker-compose -p $(PROJECT_NAME) up -d $(DB_CONTAINER)

migrate:
	docker-compose -p $(PROJECT_NAME) run --rm $(MIGRATE_CONTAINER)

redis:
	docker-compose -p $(PROJECT_NAME) up -d $(REDIS_CONTAINER)

setup: db migrate redis

run:
	docker-compose -p $(PROJECT_NAME) up --build $(SERVER_CONTAINER)

down:
	docker-compose down

generate-orm:
	sqlboiler psql --no-tests

generate-mock:
	docker-compose run --rm mockery

reset-db:
	docker exec -i $(DB_CONTAINER_NAME) psql -U postgres -d $(DB_NAME) -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
