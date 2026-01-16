.PHONY: build build-tinygo dev clean generate setup setup-tinygo docker docker-run docker-push test-setup test test-a11y test-report

# First-time setup: copy wasm_exec.js from Go installation
# Note: Go 1.24+ moved wasm_exec.js from misc/wasm/ to lib/wasm/
setup:
	cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" .

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
	rm -f main.wasm server

# Docker build (run from project root: cd .. && make -C example docker)
docker:
	cd .. && docker build -f example/Dockerfile -t gux-example .

# Run Docker container locally
docker-run: docker
	docker run --rm -p 8080:8080 gux-example

# Push to registry (set DOCKER_REGISTRY env var)
docker-push: docker
	docker tag gux-example $${DOCKER_REGISTRY}/gux-example:latest
	docker push $${DOCKER_REGISTRY}/gux-example:latest

# ============================================================================
# Testing
# ============================================================================

# One-time test setup: install npm dependencies and Playwright browser
test-setup:
	npm install
	npx playwright install chromium

# Run accessibility tests (builds WASM first)
test: build
	npm test

# Alias for running accessibility tests
test-a11y: test

# Run tests with HTML reporter and open report
test-report: build
	npm test -- --reporter=html
	open playwright-report/index.html
