# Technology Stack

## Language & Runtime
- **Go 1.21+** - Primary programming language
- Standard library HTTP server with graceful shutdown

## Key Dependencies
- **Uber FX** (`go.uber.org/fx`) - Dependency injection framework for modular architecture
- **Google Generative AI Go** (`github.com/google/generative-ai-go`) - Gemini provider integration
- **OpenAI Go Client** (`github.com/sashabaranov/go-openai`) - OpenAI provider integration
- **YAML v3** (`gopkg.in/yaml.v3`) - Configuration file parsing
- **Google Cloud APIs** - Supporting Gemini integration

## Architecture Patterns
- **Dependency Injection**: Uses Uber FX for clean module separation and lifecycle management
- **Provider Pattern**: Pluggable LLM providers implementing common interface
- **Router Pattern**: Model-based request routing to appropriate providers
- **Module System**: Each package exports an FX module for clean dependency wiring

## Build & Development Commands

```bash
# Build the server
go build -o bin/server ./cmd/server

# Run the server (requires config.yaml)
go run ./cmd/server

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Format code
go fmt ./...

# Vet code for issues
go vet ./...

# Tidy dependencies
go mod tidy
```

## Configuration
- YAML-based configuration file
- Environment variable overrides for API keys (`OPENAI_API_KEY`, `GEMINI_API_KEY`)
- Default server address: `:8080`