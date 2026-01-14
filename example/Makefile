.PHONY: build dev clean generate setup

# First-time setup: copy wasm_exec.js from Go installation
setup:
	cp "$$(go env GOROOT)/misc/wasm/wasm_exec.js" .

# Generate API clients from interface definitions
generate:
	cd .. && go run ./cmd/apigen -source=example/api/posts.go -output=posts_client_gen.go

# Build WASM module
build:
	cd .. && GOOS=js GOARCH=wasm go build -o example/main.wasm ./example/app

# Build and run dev server
dev: build
	cd .. && go run ./example/server -port 8093 -dir ./example

# Clean build artifacts
clean:
	rm -f main.wasm
