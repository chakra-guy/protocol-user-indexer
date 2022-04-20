DB_URL=postgres://postgres:postgres@localhost:5432/wallet_explorer_db?sslmode=disable

run:
	go run main.go

migrate:
	migrate -database "$(DB_URL)" -path db/migrations up

new-migration:
	migrate create -ext sql -dir db/migrations -seq $(name)

.PHONY: run migrate new-migration
