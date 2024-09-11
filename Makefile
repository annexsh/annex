SQLC_VERSION = 1.27.0

# Run

.PHONY:run-server
run-server:
	go run main.go -config-file local.yaml

# Tests

.PHONY: unit
unit:
	go test ./... -race -count=1

# Docker

.PHONY: compose-up
compose-up:
	docker-compose up -d --build --remove-orphans

.PHONY: compose-down
compose-down:
	docker-compose down -v


# Sqlc

.PHONY: sqlc-gen
sqlc-gen:
	make sqlc-gen-postgres
	make sqlc-gen-sqlite

.PHONY: sqlc-gen-postgres
sqlc-gen-postgres:
	rm -rf postgres/sqlc/*.sql.go
	docker run --rm -v $(shell pwd)/postgres:/src -w /src sqlc/sqlc:$(SQLC_VERSION) generate

.PHONY: sqlc-gen-sqlite
sqlc-gen-sqlite:
	rm -rf sqlite/sqlc/*.sql.go
	docker run --rm -v $(shell pwd)/sqlite:/src -w /src sqlc/sqlc:$(SQLC_VERSION) generate
