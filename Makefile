include .envrc

# ============================================================================ #
# HELPERS
# ============================================================================ #

## help: Print this help message
.PHONY: help
help:
	@echo "Usage:"
	@sed -n "s/^##//p" ${MAKEFILE_LIST} | column -t -s ":" | sed -e "s/^/ /"

.PHONY: confirm
confirm:
	@echo -n "Are you sure? [y/N] " && read ans && [ $${ans:-N} = y ]

# ============================================================================ #
# DEVELOPMENT
# ============================================================================ #

## run/api: Run the cmd/api application
.PHONY: run/api
run/api:
	go run ./cmd/api/ --db-dsn=${GO_USER_DB_DSN} --addr=":8080"

# ============================================================================ #
# DATABASE
# ============================================================================ #

## db/psql: connect to the database using psql
.PHONY: db/psql
db/psql:
	psql ${GO_USER_DB_DSN}

## db/migrations/new name=$1: Create a new database migration
.PHONY: db/migrations/new
db/migrations/new:
	@echo "Creating migration files for ${name}..."
	migrate create --seq --ext .sql --dir ./migrations/ ${name}

## db/migrations/up: Apply all up database migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo "Running up migrations..."
	@# For some reason, migrate requires sslmode=disable in the DSN string.
	migrate \
	    --path ./migrations/ \
		--database ${GO_USER_DB_DSN}?sslmode=disable \
		up

## db/migrations/down: Apply all down database migrations
.PHONY: db/migrations/down
db/migrations/down: confirm
	@echo "Running down migrations..."
	@# For some reason, migrate requires sslmode=disable in the DSN string.
	migrate \
	    --path ./migrations/ \
		--database ${GO_USER_DB_DSN}?sslmode=disable \
		down

# ============================================================================ #
# QUALITY CONTROL
# ============================================================================ #

## audit: Tidy dependencies and format, vet and test all code
.PHONY: audit
audit: vendor
	@echo "Formatting code..."
	go fmt ./...
	@echo "Vetting code..."
	go vet ./...
	staticcheck ./...
	@echo "Running tests..."
	go test --race --vet off ./...

## vendor: Tidy and vendor dependencies
.PHONY: vendor
vendor:
	@echo "Tidying and verifying module dependencies..."
	go mod tidy
	go mod verify
	@echo "Vendoring dependencies..."
	go mod vendor

# ============================================================================ #
# BUILD
# ============================================================================ #

## build/api: Build the cmd/api application
.PHONY: build/api
build/api:
	@echo "Building cmd/api"
	go build --ldflags "-s" -o ./bin/api ./cmd/api
	GOOS=linux GOARCH=amd64 go build --ldflags "-s" -o ./bin/linux_amd64/api ./cmd/api
