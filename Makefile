.PHONY: help build run test test-integ clean docker-up docker-down lint fmt deps

help:
	@echo "Comandos disponíveis:"
	@echo "  make build        - Compila a aplicação"
	@echo "  make run          - Executa a aplicação"
	@echo "  make test         - Roda os testes unitários"
	@echo "  make test-integ   - Roda os testes de integração"
	@echo "  make clean        - Remove artefatos de build"
	@echo "  make docker-up    - Sobe os containers Docker"
	@echo "  make docker-down  - Derruba os containers Docker"
	@echo "  make lint         - Roda o linter"
	@echo "  make fmt          - Formata o código"
	@echo "  make deps         - Baixa as dependências"

build:
	@echo "Compilando aplicação..."
	go build -o bin/dtc ./cmd/server 

run: build
	@echo "Executando aplicação..."
	./bin/dtc

test:
	@echo "Rodando testes unitários..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-integ:
	@echo "Rodando testes de integração..."
	docker-compose -f docker-compose.yml up -d
	sleep 3
	go test -v -race -tags=integration ./tests/integration/...
	docker-compose -f docker-compose.yml down

clean:
	@echo "Limpando artefatos..."
	rm -rf bin/ coverage.out coverage.html
	go clean

docker-up:
	docker-compose -f docker-compose.yml up -d

docker-down:
	docker-compose -f docker-compose.yml down

lint:
	@echo "Rodando linter..."
	golangci-lint run ./...

fmt:
	@echo "Formatando código..."
	go fmt ./...
	goimports -w .

deps:
	@echo "Baixando dependências..."
	go mod download
	go mod tidy