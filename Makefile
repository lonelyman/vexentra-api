SHELL := /bin/sh

# If you run from vexentra-api/, compose file is one level up.
COMPOSE ?= docker compose -f ../docker-compose.yml
API_SERVICE ?= api
MIGRATIONS_DIR ?= database/migrations
GOOSE_RUN = goose -dir $(MIGRATIONS_DIR) postgres "$$DSN"
DSN_ENV = DSN="host=$$POSTGRES_PRIMARY_HOST port=$$POSTGRES_PRIMARY_PORT user=$$POSTGRES_PRIMARY_USER password=$$POSTGRES_PRIMARY_PASSWORD dbname=$$POSTGRES_PRIMARY_NAME sslmode=$$POSTGRES_PRIMARY_SSL_MODE"
GOOSE_BOOTSTRAP = export PATH="$$PATH:/usr/local/go/bin:/go/bin"; if ! command -v goose >/dev/null 2>&1; then /usr/local/go/bin/go install github.com/pressly/goose/v3/cmd/goose@latest; fi

.PHONY: migrate-status migrate-up migrate-down migrate-reset migrate-version migrate-create

migrate-status:
	@$(COMPOSE) exec $(API_SERVICE) sh -lc '$(GOOSE_BOOTSTRAP); $(DSN_ENV); $(GOOSE_RUN) status'

migrate-up:
	@$(COMPOSE) exec $(API_SERVICE) sh -lc '$(GOOSE_BOOTSTRAP); $(DSN_ENV); $(GOOSE_RUN) up'

migrate-down:
	@$(COMPOSE) exec $(API_SERVICE) sh -lc '$(GOOSE_BOOTSTRAP); $(DSN_ENV); $(GOOSE_RUN) down'

migrate-reset:
	@$(COMPOSE) exec $(API_SERVICE) sh -lc '$(GOOSE_BOOTSTRAP); $(DSN_ENV); $(GOOSE_RUN) reset'

migrate-version:
	@$(COMPOSE) exec $(API_SERVICE) sh -lc '$(GOOSE_BOOTSTRAP); $(DSN_ENV); $(GOOSE_RUN) version'

migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "usage: make migrate-create name=<migration_name>"; \
		exit 1; \
	fi
	@$(COMPOSE) exec $(API_SERVICE) sh -lc '$(GOOSE_BOOTSTRAP); $(DSN_ENV); $(GOOSE_RUN) create $(name) sql'
