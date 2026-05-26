BINARY := linkedin-post-executor

.PHONY: help build build-ui build-go dev-api dev-ui run test lint tidy clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

build: build-ui build-go ## Full build (UI + single binary)

build-ui: ## Build the UI (React/Vite) into web/dist
	cd web && npm install && npm run build

build-go: ## Build the Go binary (embeds the already-built UI)
	go build -o $(BINARY) .

run: ## Run the compiled binary
	./$(BINARY)

dev-api: ## Run the Go API in dev mode (port 8080)
	go run .

dev-ui: ## Run the UI in dev mode (port 5173, proxies to the API)
	cd web && npm install && npm run dev

test: ## Run Go tests
	go test ./...

lint: ## Go vet + UI type checking
	go vet ./...
	cd web && npm run lint

tidy: ## Tidy Go dependencies
	go mod tidy

clean: ## Remove the binary and the local database
	rm -f $(BINARY) *.db *.db-shm *.db-wal
