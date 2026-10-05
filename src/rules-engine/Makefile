# Runs from within this service directory.
# Config via env: DB_URL (required), MIGRATION_DIR (default: migrations)

migrate-create:
ifndef NAME
	@echo "Error: NAME is required. Usage: make migrate-create NAME=add_widgets"
	@exit 1
endif
	go run cmd/migrate/main.go create $(NAME)

migrate-up:
	go run cmd/migrate/main.go up

migrate-down:
	go run cmd/migrate/main.go down $(STEPS)

migrate-status:
	go run cmd/migrate/main.go version
