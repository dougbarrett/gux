# Multi-stage Dockerfile for Gux Example
# Builds both the WASM frontend and Go server

# Stage 1: Build WASM with TinyGo
FROM tinygo/tinygo:latest AS wasm-builder

WORKDIR /app

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum* ./
RUN go mod download || true

# Copy source code
COPY . .

# Build WASM module (output to /app which is writable)
RUN GOFLAGS=-buildvcs=false tinygo build -o main.wasm -target wasm -no-debug ./example/app

# Stage 2: Build Go server
FROM golang:1.24-alpine AS server-builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go.mod and go.sum
COPY go.mod go.sum* ./
RUN go mod download || true

# Copy source code
COPY . .

# Build server binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server ./example/server

# Stage 3: Production image
FROM alpine:3.21

WORKDIR /app

# Install CA certificates for HTTPS
RUN apk add --no-cache ca-certificates

# Copy server binary
COPY --from=server-builder /app/server .

# Copy static files
COPY example/index.html .
COPY example/wasm_exec.js .
COPY example/manifest.json .
COPY example/offline.html .
COPY example/service-worker.js .

# Copy built WASM
COPY --from=wasm-builder /app/main.wasm .

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1

# Run server
CMD ["./server", "-port", "8080", "-dir", "."]
