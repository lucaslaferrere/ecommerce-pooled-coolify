.PHONY: up down seed mongo mongo-stop logs build test tidy

mongo:
	@echo ">>> Levantando MongoDB..."
	docker compose up -d
	@echo ">>> Esperando que MongoDB este listo..."
	@timeout /t 5 /nobreak > nul
	@echo ">>> MongoDB listo."

seed:
	@echo ">>> Ejecutando seed de admin..."
	go run ./cmd/seed/main.go
	@echo ">>> Seed completado."

up: mongo seed
	@echo ">>> Iniciando backend Go..."
	go run ./cmd/api/...

down:
	docker compose down

logs:
	docker compose logs -f mongo

build:
	go build -o bin/api.exe ./cmd/api/...

test:
	go test ./...

tidy:
	go mod tidy