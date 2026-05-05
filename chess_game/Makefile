-include .env
export

DB_URL ?= postgres://postgres:secret@localhost:5432/simple_bank?sslmode=disable

postgres:
	docker run --name postgres18 -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:18-alpine
	
postgrescontainer:
	docker run --name postgres18 -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:18

createdb:
	docker exec -it postgres18 createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it postgres18 dropdb simple_bank

migrateup:
	goose -dir sql/schema postgres "$(DB_URL)" up

migratedown:
	goose -dir sql/schema postgres "$(DB_URL)" down

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

run-server:
	go run ./cmd/server/main.go

build-server:
	go build ./cmd/server   # builds the server binary

mock:
	mockgen -destination internal/db/mock/store.go github.com/Glenn444/golang-chess/internal/db Store

createschema:
	goose -dir sql/schema create create_users_table sql

.PHONY: postgres createdb dropdb migrateup migratedown generatesql test server mock