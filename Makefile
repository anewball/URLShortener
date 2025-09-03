MIGRATIONS_DIR := internal/db/migrations

.PHONY: migrate-up migrate-down migrate-force migrate-new

migrate-up:
	migrate -database "$(DB_URL)" -path $(MIGRATIONS_DIR) up

migrate-down:
	migrate -database "$(DB_URL)" -path $(MIGRATIONS_DIR) down 1

migrate-force:
	@read -p "Version to force: " v; \
	migrate -database "$(DB_URL)" -path $(MIGRATIONS_DIR) force $$v

migrate-new:
	@read -p "Name (snake_case): " name; \
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $$name