DB_URL=postgres://postgres:postgres@localhost:5432/wallet_indexer?sslmode=disable

run:
	go run cmd/worker/main.go

serve:
	go run cmd/api/main.go

migrate:
	migrate -database "$(DB_URL)" -path db/migrations up

migrate-new:
	migrate create -ext sql -dir db/migrations -seq $(name)

.PHONY: run-indexer run-server migrate new-migration
