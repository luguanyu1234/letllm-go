# Project Structure

## Directory Layout

```
letllm-go/
├── cmd/
│   └── server/           # Main application entry point
│       └── main.go       # FX app initialization
├── internal/             # Private application code
│   ├── config/           # Configuration management
│   │   ├── config.go     # Config struct and YAML loading
│   │   └── module.go     # FX module export
│   ├── provider/         # LLM provider implementations
│   │   ├── provider.go   # Provider interface definition
│   │   ├── registry.go   # Provider registry and routing
│   │   ├── openai.go     # OpenAI provider implementation
│   │   ├── gemini.go     # Google Gemini provider implementation
│   │   ├── models.go     # Standard request/response types
│   │   └── module.go     # FX module export
│   └── server/           # HTTP server implementation
│       ├── server.go     # HTTP handlers and server lifecycle
│       └── module.go     # FX module export (implied)
├── bin/                  # Build output directory
├── go.mod               # Go module definition
└── go.sum               # Dependency checksums
```

## Package Organization

### `/cmd/server`
- Application entry point
- Minimal main.go that wires FX modules
- No business logic

### `/internal/config`
- Configuration loading and validation
- YAML parsing with environment overrides
- Exports FX module for dependency injection

### `/internal/provider`
- Provider interface and implementations
- Request/response transformation between formats
- Model routing logic
- Each provider (OpenAI, Gemini) in separate files

### `/internal/server`
- HTTP server setup and lifecycle management
- OpenAI-compatible API endpoints
- Request handling and response formatting
- Streaming support implementation

## Naming Conventions

- **Packages**: lowercase, single word when possible
- **Files**: lowercase with underscores for separation
- **Interfaces**: descriptive names ending in -er when appropriate (Provider)
- **Structs**: PascalCase with clear, descriptive names
- **Methods**: PascalCase for exported, camelCase for private
- **Constants**: PascalCase or ALL_CAPS for package-level constants

## Module Pattern

Each package exports an FX module via `Module` variable:
```go
var Module = fx.Provide(NewSomething)
// or
var Module = fx.Module("name", fx.Provide(...), fx.Invoke(...))
```

This enables clean dependency injection in main.go.