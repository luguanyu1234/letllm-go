# Product Overview

LetLLM-Go is an LLM API proxy server that provides a unified OpenAI-compatible interface for multiple language model providers. The service acts as a routing layer that can forward requests to different LLM providers (OpenAI, Google Gemini) based on model name prefixes.

## Key Features

- OpenAI-compatible API endpoint (`/v1/chat/completions`)
- Multi-provider support (OpenAI, Google Gemini)
- Model-based routing using configurable prefixes
- Streaming and non-streaming response support
- YAML-based configuration with environment variable overrides
- Graceful server lifecycle management

## Use Cases

- Unified interface for multiple LLM providers
- Model routing and load balancing
- API key management and abstraction
- Protocol translation between different provider APIs