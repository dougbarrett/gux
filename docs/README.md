# Gux

> A full-stack Go framework for building modern web applications with WebAssembly.

## What is Gux?

Gux enables you to write entire web applications in Go:

- **Frontend**: Compiles to WebAssembly, runs natively in the browser
- **Backend**: Standard Go HTTP server with generated handlers
- **API**: Type-safe clients and servers generated from Go interfaces

## Quick Start

```bash
# Install
go install github.com/dougbarrett/gux/cmd/gux@latest

# Create new project
gux init --module github.com/youruser/myapp myapp
cd myapp

# Run development server
gux dev
```

## Features

- **Type-Safe APIs** - Define Go interfaces, get generated clients and servers
- **SSR + WASM Hydration** - Server-side rendering with client-side interactivity
- **Reactive State** - Type-safe state management with `r.StateInt`, `r.StateString`, etc.
- **CRUD Generation** - Automatic REST API generation with DTOs and hooks
- **CSRF Protection** - Built-in security with automatic token handling
- **Hot Reload Development** - Fast development cycle with `gux dev`

## Links

- [Live Demo](https://gux-demo.production.app.dbb1.dev/) — Try Gux in your browser
- [GitHub Repository](https://github.com/dougbarrett/gux)
- [Getting Started Guide](getting-started.md)
