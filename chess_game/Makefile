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

migratedown0:
	goose -dir sql/schema postgres "$(DB_URL)" down-to 0

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

run-server:
	go run ./cmd/server/main.go

build-server:
	go build ./cmd/server   # builds the server binary

build-linux:
	GOOS=linux GOARCH=amd64 go build -o chesske ./cmd/server

deploy:
	./deploy.sh

mock:
	mockgen -destination internal/db/mock/store.go github.com/Glenn444/golang-chess/internal/db Store

createschema:
	goose -dir sql/schema create create_users_table sql

# ── Docker / Production ─────────────────────────────────────────────────────
docker-build:
	docker build -t chess-server .

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f

docker-migrate:
	docker compose exec server sh -c 'goose -dir sql/schema postgres "$$DB_URL" up'

docker-restart:
	docker compose down && docker compose up -d --build

# ── Dev helpers ─────────────────────────────────────────────────────────────
swagger:
	swag init --generalInfo cmd/server/main.go --output ./docs

.PHONY: postgres createdb dropdb migrateup migratedown generatesql test server mock \
        docker-build docker-up docker-down docker-logs docker-migrate docker-restart swagger