package provider

import (
	"context"
	"io"
)

// Provider defines the interface that all LLM providers must implement
type Provider interface {
	// Generate performs a non-streaming text generation request
	Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error)

	// StreamGenerate performs a streaming text generation request
	StreamGenerate(ctx context.Context, req *GenerateRequest) (io.ReadCloser, error)

	// GetCapabilities returns the capabilities of this provider
	GetCapabilities() ProviderCapabilities

	// GetInfo returns information about this provider
	GetInfo() ProviderInfo

	// Close closes any underlying resources
	Close() error
}
