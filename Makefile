.PHONY: help generate migrate migrate-down run build tidy

help:
	@echo "make generate     - generate API code from OpenAPI"
	@echo "make migrate      - apply migrations"
	@echo "make migrate-down - rollback migrations"
	@echo "make run          - run the service"
	@echo "make build        - build the binary"
	@echo "make tidy         - go mod tidy"

generate:
	@echo "TODO: oapi-codegen"

migrate:
	@echo "TODO: goose up"

migrate-down:
	@echo "TODO: goose down"

run:
	@echo "TODO: go run ./cmd/trip-service"

build:
	go build ./...

tidy:
	go mod tidy