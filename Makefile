DB_URL='postgres://eutychus:eutychus@127.0.0.1:5432/goquizbox?sslmode=disable'
MIGRATION_DIR="migrations"


.PHONY: help
help: ## Display available commands.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

compile: ## Compile the app into /tmp/goquizbox
	go build -o /tmp/goquizbox cmd/server/main.go

compile_cli: ## Compile the cli app
	go build -o /tmp/goquizboxcli cmd/client/main.go

docker-ui: ## Docker build the ui into ektowett/goquizbox-ui:latest
	@cd ui && docker build -t ektowett/goquizbox-ui:latest . && cd ..

docker: ## Docker build the app into ektowett/goquizbox:latest
	docker build -t ektowett/goquizbox:latest .

up: ## Docker Compose bring up all containers in detatched mode
	docker-compose up -d

ps: ## Docker Compose check docker processes
	docker-compose ps

logs: ## Docker Compose tail follow logs
	docker-compose logs -f

stop: ## Docker Compose stop all containers
	docker-compose stop

rm: stop ## Docker Compose stop and force remove all containers
	docker-compose rm -f

GOFMT_FILES = $(shell go list -f '{{.Dir}}' ./... | grep -v '/pb')
HTML_FILES = $(shell find . -name \*.html)
GO_FILES = $(shell find . -name \*.go)
MD_FILES = $(shell find . -name \*.md)

# diff-check runs git-diff and fails if there are any changes.
diff-check:
	@FINDINGS="$$(git status -s -uall)" ; \
		if [ -n "$${FINDINGS}" ]; then \
			echo "Changed files:\n\n" ; \
			echo "$${FINDINGS}\n\n" ; \
			echo "Diffs:\n\n" ; \
			git diff ; \
			git diff --cached ; \
			exit 1 ; \
		fi
.PHONY: diff-check

generate:
	@go generate ./...
.PHONY: generate

# lint uses the same linter as CI and tries to report the same results running
# locally. There is a chance that CI detects linter errors that are not found
# locally, but it should be rare.
lint:
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint
	@golangci-lint run --config .golangci.yaml
.PHONY: lint

generate-check: generate diff-check
.PHONY: generate-check

test:
	@go test \
		-shuffle=on \
		-count=1 \
		-short \
		-timeout=5m \
		./...
.PHONY: test

test-acc:
	@go test \
		-shuffle=on \
		-count=1 \
		-race \
		-timeout=10m \
		./... \
		-coverprofile=coverage.out
.PHONY: test-acc

test-coverage:
	@go tool cover -func=./coverage.out
.PHONY: test-coverage

zapcheck:
	@go install github.com/sethvargo/zapw/cmd/zapw
	@zapw ./...
.PHONY: zapcheck

# make migration name=create_users
migration: ## Create golang goose migrations
	@echo "Creating migration $(name)!"
	@goose -dir $(MIGRATION_DIR) create $(name) sql
	@echo "Done!"

migrate_up: ## Golang goose up migrations
	@echo "Migrating up!"
	@goose -dir $(MIGRATION_DIR) postgres $(DB_URL) up
	@echo "Done!"

migrate_down: ## Golang goose down migrations
	@echo "Migrating down!"
	@goose -dir $(MIGRATION_DIR) postgres $(DB_URL) down
	@echo "Done!"

migrate_status: ## Golang goose status migrations
	@echo "Getting migration status!"
	@goose -dir $(MIGRATION_DIR) postgres $(DB_URL) status
	@echo "Done!"

migrate_reset: ## Golang goose reset migrations
	@echo "Resetting $(MIGRATION_DIR)!"
	@goose -dir $(MIGRATION_DIR) postgres $(DB_URL) reset
	@echo "Done!"

migrate_version: ## Golang goose version  migrations
	@echo "Getting migration version!"
	@goose -dir $(MIGRATION_DIR) postgres $(DB_URL) version
	@echo "Done!"

migrate_redo: migrate_reset migrate_up ## Golang goose redo migrations - reset then up
