package provider

import (
	"testing"

	"github.com/luguanyu1234/letllm-go/internal/config"
)

func TestNewRegistry(t *testing.T) {
	cfg := &config.Config{
		OpenAI: struct {
			APIKey       string `yaml:"api_key"`
			DefaultModel string `yaml:"default_model"`
		}{
			APIKey:       "test-openai-key",
			DefaultModel: "gpt-4",
		},
		Gemini: struct {
			APIKey       string `yaml:"api_key"`
			DefaultModel string `yaml:"default_model"`
		}{
			APIKey:       "test-gemini-key",
			DefaultModel: "gemini-pro",
		},
	}

	registry, err := NewRegistry(cfg)
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	if registry == nil {
		t.Fatal("Registry should not be nil")
	}

	// Check that providers were registered
	providers := registry.ListProviders()
	if len(providers) != 2 {
		t.Errorf("Expected 2 providers, got %d", len(providers))
	}

	// Check OpenAI provider
	openaiProvider, exists := registry.GetProvider("openai")
	if !exists {
		t.Error("OpenAI provider should be registered")
	}
	if openaiProvider == nil {
		t.Error("OpenAI provider should not be nil")
	}

	// Check Gemini provider
	geminiProvider, exists := registry.GetProvider("gemini")
	if !exists {
		t.Error("Gemini provider should be registered")
	}
	if geminiProvider == nil {
		t.Error("Gemini provider should not be nil")
	}
}

func TestRegistryRouting(t *testing.T) {
	cfg := &config.Config{
		Routes: []config.Route{
			{Prefix: "gpt-", Provider: "openai"},
			{Prefix: "gemini-", Provider: "gemini"},
		},
		OpenAI: struct {
			APIKey       string `yaml:"api_key"`
			DefaultModel string `yaml:"default_model"`
		}{
			APIKey:       "test-openai-key",
			DefaultModel: "gpt-4",
		},
		Gemini: struct {
			APIKey       string `yaml:"api_key"`
			DefaultModel string `yaml:"default_model"`
		}{
			APIKey:       "test-gemini-key",
			DefaultModel: "gemini-pro",
		},
	}

	registry, err := NewRegistry(cfg)
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	// Test routing by prefix
	req := &RouteRequest{Model: "gpt-4"}
	provider, err := registry.Route(req)
	if err != nil {
		t.Errorf("Failed to route gpt-4: %v", err)
	}
	if provider == nil {
		t.Error("Provider should not be nil for gpt-4")
	}

	req = &RouteRequest{Model: "gemini-pro"}
	provider, err = registry.Route(req)
	if err != nil {
		t.Errorf("Failed to route gemini-pro: %v", err)
	}
	if provider == nil {
		t.Error("Provider should not be nil for gemini-pro")
	}

	// Test fallback routing
	req = &RouteRequest{Model: "gpt-3.5-turbo"}
	provider, err = registry.Route(req)
	if err != nil {
		t.Errorf("Failed to route gpt-3.5-turbo with fallback: %v", err)
	}
	if provider == nil {
		t.Error("Provider should not be nil for gpt-3.5-turbo")
	}

	// Test unknown model
	req = &RouteRequest{Model: "unknown-model"}
	provider, err = registry.Route(req)
	if err == nil {
		t.Error("Expected error for unknown model")
	}
	if provider != nil {
		t.Error("Provider should be nil for unknown model")
	}
}

func TestRegistryProviderManagement(t *testing.T) {
	cfg := &config.Config{}

	registry, err := NewRegistry(cfg)
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	// Test registering a provider
	mockProvider, err := NewOpenAIProvider("test-key", "gpt-4")
	if err != nil {
		t.Fatalf("Failed to create mock provider: %v", err)
	}

	err = registry.RegisterProvider("test-provider", mockProvider)
	if err != nil {
		t.Errorf("Failed to register provider: %v", err)
	}

	// Test getting the registered provider
	provider, exists := registry.GetProvider("test-provider")
	if !exists {
		t.Error("Registered provider should exist")
	}
	if provider != mockProvider {
		t.Error("Retrieved provider should match registered provider")
	}

	// Test listing providers
	providers := registry.ListProviders()
	if len(providers) != 1 {
		t.Errorf("Expected 1 provider, got %d", len(providers))
	}

	// Test registering nil provider
	err = registry.RegisterProvider("nil-provider", nil)
	if err == nil {
		t.Error("Expected error when registering nil provider")
	}
}

func TestRegistryClose(t *testing.T) {
	cfg := &config.Config{
		OpenAI: struct {
			APIKey       string `yaml:"api_key"`
			DefaultModel string `yaml:"default_model"`
		}{
			APIKey:       "test-openai-key",
			DefaultModel: "gpt-4",
		},
	}

	registry, err := NewRegistry(cfg)
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	// Close should not return an error for properly configured providers
	err = registry.Close()
	if err != nil {
		t.Errorf("Unexpected error closing registry: %v", err)
	}
}

func TestBackwardCompatibility(t *testing.T) {
	cfg := &config.Config{
		OpenAI: struct {
			APIKey       string `yaml:"api_key"`
			DefaultModel string `yaml:"default_model"`
		}{
			APIKey:       "test-openai-key",
			DefaultModel: "gpt-4",
		},
	}

	// Test NewRouter (backward compatibility)
	router, err := NewRouter(cfg)
	if err != nil {
		t.Fatalf("Failed to create router: %v", err)
	}

	// Test ForModel (backward compatibility)
	provider, err := router.ForModel("gpt-4")
	if err != nil {
		t.Errorf("Failed to get provider for model: %v", err)
	}
	if provider == nil {
		t.Error("Provider should not be nil")
	}
}
