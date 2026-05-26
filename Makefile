BINARY := linkedin-post-executor

.PHONY: help build build-ui build-go dev-api dev-ui run test lint tidy clean

help: ## Mostra esta ajuda
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

build: build-ui build-go ## Build completo (UI + binário único)

build-ui: ## Compila a UI (React/Vite) para web/dist
	cd web && npm install && npm run build

build-go: ## Compila o binário Go (embute a UI já construída)
	go build -o $(BINARY) .

run: ## Sobe o binário compilado
	./$(BINARY)

dev-api: ## Sobe a API Go em modo dev (porta 8080)
	go run .

dev-ui: ## Sobe a UI em modo dev (porta 5173, proxy para a API)
	cd web && npm install && npm run dev

test: ## Roda os testes Go
	go test ./...

lint: ## Vet do Go + checagem de tipos da UI
	go vet ./...
	cd web && npm run lint

tidy: ## Organiza as dependências Go
	go mod tidy

clean: ## Remove binário e banco local
	rm -f $(BINARY) *.db *.db-shm *.db-wal
