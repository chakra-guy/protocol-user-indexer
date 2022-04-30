DB_URL=postgres://postgres:postgres@localhost:5432/wallet_indexer?sslmode=disable

run:
	go run main.go

migrate:
	migrate -database "$(DB_URL)" -path db/migrations up

migrate-new:
	migrate create -ext sql -dir db/migrations -seq $(name)

.PHONY: run migrate new-migration
