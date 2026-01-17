# Gux

> A full-stack Go framework for building modern web applications with WebAssembly.

## What is Gux?

Gux enables you to write entire web applications in Go:

- **Frontend**: Compiles to WebAssembly, runs natively in the browser
- **Backend**: Standard Go HTTP server with generated handlers
- **API**: Type-safe clients and servers generated from Go interfaces
- **Components**: 45+ production-ready UI components with Tailwind CSS

## Quick Start

```bash
# Install
go install github.com/dougbarrett/gux/cmd/gux@latest

# Create new project
gux init --module github.com/youruser/myapp myapp
cd myapp

# Setup WASM runtime
gux setup --tinygo

# Run development server
gux dev
```

## Features

- **Type-Safe APIs** - Define Go interfaces, get generated clients and servers
- **45+ Components** - Buttons, forms, tables, modals, charts, and more
- **State Management** - Reactive stores with localStorage/sessionStorage persistence
- **WebSocket Support** - Real-time updates with automatic reconnection
- **PWA Ready** - Service worker and manifest included
- **Docker Deployment** - Multi-stage Dockerfile included

## Links

- [Live Demo](https://gux-demo.production.app.dbb1.dev/) — Try Gux in your browser
- [GitHub Repository](https://github.com/dougbarrett/gux)
- [Getting Started Guide](getting-started.md)
- [Component Library](components.md)
