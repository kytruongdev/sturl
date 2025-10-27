.PHONY: run down test

PROJECT_NAME=api-gateway
SERVER_CONTAINER=gateway

run:
	docker-compose -p $(PROJECT_NAME) up --build $(SERVER_CONTAINER)

down:
	docker-compose down

# run single test case:
# env $(grep -v '^#' ./local.env | egrep -v '^#' | xargs) go test -cover -mod=vendor -v ./internal/infra/proxy -run TestRegister
test:
	env $(shell cat local.env | egrep -v '^#' | xargs) \
	sh -c 'go test -p 1 -mod=vendor -failfast -timeout 5m ./...' \
	&& echo "\n[Test okay -- `date`]" || (echo "\n[Test FAILED -- `date`]" && exit 1)
