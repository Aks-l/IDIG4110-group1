CONTAINER_NAME := idig4110-group1-ingest-service-1

CONTAINER_RUNNING := $(shell docker ps --format '{{.Names}}' 2>/dev/null | grep -E '^$(CONTAINER_NAME)$$')

ifdef CONTAINER_RUNNING
	EXEC_CMD = docker exec $(CONTAINER_NAME)
	EXEC_CMD_INTERACTIVE = docker exec -i $(CONTAINER_NAME)
	ENV_MSG = Running in Docker container
else
	EXEC_CMD = 
	EXEC_CMD_INTERACTIVE = 
	ENV_MSG = Running on host (Docker not available)
endif

up:
	@echo "Starting docker containers"
	@docker compose up -d --build

down:
	@echo "Stopping docker containers"
	@docker compose down

logs:
	@docker compose logs -f

migrate-create:
ifndef NAME
	@echo "Error: NAME is required"
	@echo "Usage: make migrate-create NAME=add_user_avatar"
	@exit 1
endif
ifdef CONTAINER_RUNNING
	@echo "$(ENV_MSG)"
	@$(EXEC_CMD) go run cmd/migrate/main.go create $(NAME)
else
	@if command -v go >/dev/null 2>&1; then \
		echo "$(ENV_MSG)"; \
		go run cmd/migrate/main.go create $(NAME); \
	else \
		echo "Error: Docker container not running and Go not installed"; \
		echo "Please run: make up"; \
		exit 1; \
	fi
endif

migrate-up:
ifdef CONTAINER_RUNNING
	@echo "$(ENV_MSG)"
	@$(EXEC_CMD) go run cmd/migrate/main.go up
else
	@if command -v go >/dev/null 2>&1; then \
		echo "$(ENV_MSG)"; \
		go run cmd/migrate/main.go up; \
	else \
		echo "Error: Docker container not running and Go not installed"; \
		echo "Please run: make up"; \
		exit 1; \
	fi
endif

migrate-down:
ifdef CONTAINER_RUNNING
	@echo "$(ENV_MSG)"
ifdef STEPS
	@$(EXEC_CMD_INTERACTIVE) go run cmd/migrate/main.go down $(STEPS)
else
	@$(EXEC_CMD_INTERACTIVE) go run cmd/migrate/main.go down
endif
else
	@if command -v go >/dev/null 2>&1; then \
		echo "$(ENV_MSG)"; \
ifdef STEPS
		go run cmd/migrate/main.go down $(STEPS); \
else
		go run cmd/migrate/main.go down; \
endif
	else \
		echo "Error: Docker container not running and Go not installed"; \
		echo "Please run: make up"; \
		exit 1; \
	fi
endif

migrate-status:
ifdef CONTAINER_RUNNING
	@echo "$(ENV_MSG)"
	@$(EXEC_CMD) go run cmd/migrate/main.go version
else
	@if command -v go >/dev/null 2>&1; then \
		echo "$(ENV_MSG)"; \
		go run cmd/migrate/main.go version; \
	else \
		echo "Error: Docker container not running and Go not installed"; \
		echo "Please run: make up"; \
		exit 1; \
	fi
endif
