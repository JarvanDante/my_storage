APP := my_storage
BIN := bin
MYDOCKER := /Users/wangdante/D/mydocker

.PHONY: build dev tidy clean migrate docker-up docker-down docker-logs docker-rebuild

build:
	go build -o $(BIN)/storageapi .

dev:
	gf run main.go

tidy:
	go mod tidy

migrate:
	goose -dir manifest/sql/migrations postgres "host=127.0.0.1 port=5432 user=postgres password=654321 dbname=my_storage sslmode=disable" up

clean:
	rm -rf $(BIN)

docker-up:
	cd $(MYDOCKER) && docker compose up -d --build my_storage
	@echo "探活: curl -sS http://127.0.0.1:8015/healthz"

docker-down:
	cd $(MYDOCKER) && docker compose stop my_storage

docker-logs:
	docker logs -f my_storage

docker-rebuild:
	cd $(MYDOCKER) && docker compose up -d --build --force-recreate my_storage
