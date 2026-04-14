.PHONY: dev test test-coverage build tidy format up down logs

# Inicia a aplicacao localmente rodando no terminal conectado no docker compose
dev: up
	@echo "Iniciando aplicação local..."
	@go run ./cmd/api/main.go

# Roda todos os testes da aplicação (TDD)
test:
	@echo "Rodando toda a bateria de testes unitários..."
	@go test -v ./...

# Roda os testes com cobertura apontando gargalos
test-coverage:
	@echo "Rodando testes de cobertura..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out

# Build pro binário que usaremos em PRD
build:
	@echo "Fazendo build do sistema..."
	@go build -o fitbrother.bin ./cmd/api

# Limpa go.mod resolvendo dependencias
tidy:
	@echo "Arrumando modulos (go mod tidy)..."
	@go mod tidy

# Formata o codigo no padrao da Google
format:
	@echo "Formatando fontes..."
	@go fmt ./...

# Docker comandos (Subir infra local PostgreSQL e Redis)
up:
	@echo "Subindo containers (Banco e Cache)..."
	docker-compose -f docker/docker-compose.yml up -d

down:
	@echo "Parando containers e desligando..."
	docker-compose -f docker/docker-compose.yml down

logs:
	docker-compose -f docker/docker-compose.yml logs -f
