# Deployment

This guide covers deploying Gux applications to production environments.

## Single Binary Deployment

Gux applications compile to a **single binary** with all static assets embedded. This makes deployment simple—just copy and run.

### Building for Production

```bash
# Build the production binary
gux setup --tinygo
gux build --tinygo

# Run locally
./server

# Or deploy the binary anywhere
scp ./server user@server:/app/
ssh user@server '/app/server -port 8080'
```

### What's Embedded

The `./server` binary contains:
- WASM frontend (`main.wasm`)
- Go WASM runtime (`wasm_exec.js`)
- HTML, manifest, and service worker
- Any CSS, JS, images, or other files in `public/`

Cache-busting is handled automatically—the server computes a hash of `main.wasm` at startup and injects it into `index.html`.

## Docker

Gux includes a 2-stage Dockerfile optimized for production:

### Building the Image

```bash
docker build -t myapp .
```

### Running Locally

```bash
docker run --rm -p 8080:8080 myapp
```

Open http://localhost:8080

### Dockerfile Explained

```dockerfile
# Stage 1: Build with TinyGo using gux
FROM tinygo/tinygo:latest AS builder
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download || true
RUN go install github.com/dougbarrett/gux/cmd/gux@latest
COPY . .
RUN gux setup --tinygo && gux build --tinygo

# Stage 2: Minimal production image
FROM alpine:3.21
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/server .
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1
CMD ["./server", "-port", "8080"]
```

### Key Optimizations

1. **Single binary** — All assets embedded, nothing to copy separately
2. **2-stage build** — Only the binary goes into production image
3. **TinyGo** — Produces ~500KB WASM vs ~5MB with standard Go
4. **Alpine base** — Minimal ~5MB base image
5. **Runtime cache-busting** — No build-time index.html modification needed
6. **Health checks** — Built-in health monitoring

### Large Assets

For large files (videos, images >1MB), consider using a CDN instead of embedding:

```html
<!-- In your index.html or WASM code -->
<video src="https://cdn.example.com/video.mp4"></video>
```

## Docker Compose

For applications with multiple services:

```yaml
# docker-compose.yml
version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: example/Dockerfile
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgres://postgres:password@db:5432/myapp
    depends_on:
      - db
    healthcheck:
      test: ["CMD", "wget", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 3s
      retries: 3

  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: myapp
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

```bash
docker-compose up -d
```

## Cloud Platforms

### Fly.io

1. Install flyctl: https://fly.io/docs/hands-on/install-flyctl/

2. Create `fly.toml`:

```toml
app = "my-gux-app"
primary_region = "ord"

[build]
  dockerfile = "example/Dockerfile"

[http_service]
  internal_port = 8080
  force_https = true
  auto_stop_machines = true
  auto_start_machines = true
  min_machines_running = 1

[[vm]]
  cpu_kind = "shared"
  cpus = 1
  memory_mb = 256
```

3. Deploy:

```bash
fly launch
fly deploy
```

### Railway

1. Connect your GitHub repository at https://railway.app

2. Railway auto-detects the Dockerfile

3. Set environment variables in the dashboard

4. Deploy automatically on push

### Google Cloud Run

```bash
# Build and push
gcloud builds submit --tag gcr.io/PROJECT_ID/gux-app

# Deploy
gcloud run deploy gux-app \
  --image gcr.io/PROJECT_ID/gux-app \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated
```

### AWS ECS/Fargate

1. Push to ECR:

```bash
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin ACCOUNT.dkr.ecr.us-east-1.amazonaws.com

docker build -t gux-app -f example/Dockerfile .
docker tag gux-app:latest ACCOUNT.dkr.ecr.us-east-1.amazonaws.com/gux-app:latest
docker push ACCOUNT.dkr.ecr.us-east-1.amazonaws.com/gux-app:latest
```

2. Create task definition and service via AWS Console or Terraform

### DigitalOcean App Platform

1. Create `app.yaml`:

```yaml
name: gux-app
services:
  - name: web
    dockerfile_path: example/Dockerfile
    github:
      repo: yourusername/gux
      branch: main
    http_port: 8080
    instance_size_slug: basic-xxs
    instance_count: 1
```

2. Deploy via DigitalOcean dashboard or CLI

## Kubernetes

### Deployment

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gux-app
  labels:
    app: gux-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: gux-app
  template:
    metadata:
      labels:
        app: gux-app
    spec:
      containers:
        - name: gux-app
          image: your-registry/gux-app:latest
          ports:
            - containerPort: 8080
          resources:
            requests:
              memory: "64Mi"
              cpu: "100m"
            limits:
              memory: "128Mi"
              cpu: "200m"
          livenessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 5
```

### Service

```yaml
# k8s/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: gux-app
spec:
  selector:
    app: gux-app
  ports:
    - protocol: TCP
      port: 80
      targetPort: 8080
  type: ClusterIP
```

### Ingress

```yaml
# k8s/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: gux-app
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  tls:
    - hosts:
        - myapp.example.com
      secretName: gux-app-tls
  rules:
    - host: myapp.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: gux-app
                port:
                  number: 80
```

### Deploy

```bash
kubectl apply -f k8s/
```

## Environment Configuration

### Environment Variables

```go
// server/main.go
import "os"

func main() {
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    dbURL := os.Getenv("DATABASE_URL")
    if dbURL == "" {
        log.Fatal("DATABASE_URL required")
    }

    // ...
}
```

### Docker Environment

```bash
docker run -p 8080:8080 \
  -e DATABASE_URL=postgres://... \
  -e API_KEY=secret \
  gux-app
```

### Kubernetes ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: gux-config
data:
  LOG_LEVEL: "info"
  CORS_ORIGIN: "https://myapp.com"
```

### Kubernetes Secrets

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: gux-secrets
type: Opaque
stringData:
  DATABASE_URL: postgres://user:pass@host:5432/db
  API_KEY: your-secret-key
```

## Health Checks

### Basic Health Endpoint

```go
mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
})
```

### Comprehensive Health Check

```go
type HealthStatus struct {
    Status   string            `json:"status"`
    Version  string            `json:"version"`
    Checks   map[string]string `json:"checks"`
}

mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    status := HealthStatus{
        Status:  "healthy",
        Version: version,
        Checks:  make(map[string]string),
    }

    // Check database
    if err := db.Ping(); err != nil {
        status.Status = "unhealthy"
        status.Checks["database"] = "failed: " + err.Error()
    } else {
        status.Checks["database"] = "ok"
    }

    // Check cache
    if err := cache.Ping(); err != nil {
        status.Status = "degraded"
        status.Checks["cache"] = "failed: " + err.Error()
    } else {
        status.Checks["cache"] = "ok"
    }

    w.Header().Set("Content-Type", "application/json")
    if status.Status != "healthy" {
        w.WriteHeader(http.StatusServiceUnavailable)
    }
    json.NewEncoder(w).Encode(status)
})
```

## Logging

### Structured Logging

```go
import (
    "log/slog"
    "os"
)

func main() {
    // JSON logging for production
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
    slog.SetDefault(logger)

    slog.Info("server starting",
        "port", port,
        "version", version,
    )
}
```

### Request Logging

The built-in `server.Logger()` middleware logs all requests:

```
2024/01/15 10:30:45 GET /api/posts 200 15.234ms
```

## Performance Optimization

### WASM Size

```bash
# Use TinyGo (recommended)
tinygo build -o main.wasm -target wasm -no-debug ./app
# ~500KB

# Standard Go (larger)
GOOS=js GOARCH=wasm go build -o main.wasm ./app
# ~5MB
```

### Compression

```go
// Add gzip middleware
import "github.com/NYTimes/gziphandler"

handler := gziphandler.GzipHandler(yourHandler)
```

### Caching Headers

```go
func (h *SPAHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Cache WASM and static assets
    if strings.HasSuffix(r.URL.Path, ".wasm") ||
       strings.HasSuffix(r.URL.Path, ".js") ||
       strings.HasSuffix(r.URL.Path, ".css") {
        w.Header().Set("Cache-Control", "public, max-age=31536000")
    }
    // ...
}
```

## Security

### HTTPS

Always use HTTPS in production. Most platforms handle this automatically.

For self-hosted:

```go
log.Fatal(http.ListenAndServeTLS(":443", "cert.pem", "key.pem", mux))
```

### Security Headers

```go
func SecurityHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("X-XSS-Protection", "1; mode=block")
        w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
        next.ServeHTTP(w, r)
    })
}
```

### Rate Limiting

```go
import "golang.org/x/time/rate"

var limiter = rate.NewLimiter(100, 200) // 100 req/s, burst 200

func RateLimit(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !limiter.Allow() {
            http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

## Monitoring

### Prometheus Metrics

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    requestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total HTTP requests",
        },
        []string{"method", "path", "status"},
    )
)

func init() {
    prometheus.MustRegister(requestsTotal)
}

func main() {
    mux.Handle("/metrics", promhttp.Handler())
}
```

### Graceful Shutdown

```go
func main() {
    srv := &http.Server{Addr: ":8080", Handler: mux}

    go func() {
        if err := srv.ListenAndServe(); err != http.ErrServerClosed {
            log.Fatal(err)
        }
    }()

    // Wait for interrupt
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    // Graceful shutdown with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("Shutdown error:", err)
    }
    log.Println("Server stopped")
}
```

## CI/CD

### GitHub Actions

```yaml
# .github/workflows/deploy.yml
name: Deploy

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Build Docker image
        run: docker build -f example/Dockerfile -t gux-app .

      - name: Push to registry
        run: |
          echo ${{ secrets.DOCKER_PASSWORD }} | docker login -u ${{ secrets.DOCKER_USERNAME }} --password-stdin
          docker tag gux-app ${{ secrets.DOCKER_REGISTRY }}/gux-app:${{ github.sha }}
          docker push ${{ secrets.DOCKER_REGISTRY }}/gux-app:${{ github.sha }}

      - name: Deploy to production
        run: |
          # Your deployment command
          kubectl set image deployment/gux-app gux-app=${{ secrets.DOCKER_REGISTRY }}/gux-app:${{ github.sha }}
```
