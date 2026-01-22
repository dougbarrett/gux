.PHONY: build run clean

# Build everything into a single binary
build: cmd/app/dist/app.wasm cmd/app/dist/wasm_exec.js
	go build -o bin/app ./cmd/app

# Build WASM from internal entry point
cmd/app/dist/app.wasm: internal/wasm/main.go pages/home.go
	@mkdir -p cmd/app/dist
	GOOS=js GOARCH=wasm go build -o cmd/app/dist/app.wasm ./internal/wasm

# Copy wasm_exec.js from Go installation
cmd/app/dist/wasm_exec.js:
	@mkdir -p cmd/app/dist
	cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" cmd/app/dist/

# Build and run
run: build
	./bin/app

clean:
	rm -rf bin/ cmd/app/dist/
