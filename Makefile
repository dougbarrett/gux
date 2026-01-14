.PHONY: build wasm serve dev prod dist clean generate

# Generate API clients from interfaces
generate:
	go generate ./internal/api/...

# Build WASM module
wasm:
	GOOS=js GOARCH=wasm go build -o main.wasm main.go

# Alias for wasm
build: wasm

# Build and run dev server with SPA support
dev: wasm
	go run ./cmd/server -port 8083 -dir .

# Alias for dev
serve: dev

# Create dist directory with all static files
dist: wasm
	@mkdir -p cmd/prod/dist
	@cp index.html cmd/prod/dist/
	@cp main.wasm cmd/prod/dist/
	@cp wasm_exec.js cmd/prod/dist/
	@echo "Distribution files copied to cmd/prod/dist/"

# Build single production binary with embedded files
prod: dist
	go build -o goquery-server ./cmd/prod
	@echo "Production binary built: ./goquery-server"
	@echo "Run with: ./goquery-server -port 8080"

# Clean build artifacts
clean:
	rm -f main.wasm
	rm -f goquery-server
	rm -rf cmd/prod/dist
