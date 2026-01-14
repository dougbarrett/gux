.PHONY: build build-tinygo dev clean generate setup setup-tinygo

# First-time setup: copy wasm_exec.js from Go installation
setup:
	cp "$$(go env GOROOT)/misc/wasm/wasm_exec.js" .

# TinyGo setup: copy wasm_exec.js from TinyGo installation
setup-tinygo:
	cp "$$(tinygo env TINYGOROOT)/targets/wasm_exec.js" .

# Generate API clients from interface definitions
generate:
	cd .. && go run ./cmd/apigen -source=example/api/posts.go -output=posts_client_gen.go

# Build WASM module with standard Go
build:
	cd .. && GOOS=js GOARCH=wasm go build -o example/main.wasm ./example/app

# Build WASM module with TinyGo (much smaller output)
build-tinygo:
	cd .. && tinygo build -o example/main.wasm -target wasm -no-debug ./example/app

# Build and run dev server
dev: build
	cd .. && go run ./example/server -port 8093 -dir ./example

# Build with TinyGo and run dev server
dev-tinygo: build-tinygo
	cd .. && go run ./example/server -port 8093 -dir ./example

# Clean build artifacts
clean:
	rm -f main.wasm
