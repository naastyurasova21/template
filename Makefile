.PHONY: help generate migrate migrate-down run build tidy

help:
	@echo "make generate     - generate API code from OpenAPI"
	@echo "make migrate      - apply migrations"
	@echo "make migrate-down - rollback migrations"
	@echo "make run          - build and run the service"
	@echo "make build        - build the binary"
	@echo "make tidy         - go mod tidy"

generate:
	go tool oapi-codegen \
		-generate types,chi-server \
		-package api \
		-o api/api.gen.go \
		contracts/openapi/trip-service.openapi.yaml

migrate:
	@echo "TODO: goose up"

migrate-down:
	@echo "TODO: goose down"

build:
	go build -o bin/trip-service ./cmd/trip-service

run: build
	./bin/trip-service

tidy:
	go mod tidy