package provider

import (
	"fmt"
	"strings"
	"sync"

	"github.com/luguanyu1234/letllm-go/internal/config"
)

// RouteRequest represents a request for provider routing
type RouteRequest struct {
	Model    string            `json:"model"`
	ClientID string            `json:"client_id,omitempty"`
	Endpoint string            `json:"endpoint,omitempty"`
	Headers  map[string]string `json:"headers,omitempty"`
}

// RouterInterface defines the interface for provider routing
type RouterInterface interface {
	Route(req *RouteRequest) (Provider, error)
	GetProviderForModel(model string) (Provider, error)
	RegisterProvider(name string, provider Provider) error
	ListProviders() []ProviderInfo
	Close() error
}

// Registry manages provider instances and routing
type Registry struct {
	cfg       *config.Config
	providers map[string]Provider
	mu        sync.RWMutex
}

// NewRegistry creates a new provider registry
func NewRegistry(cfg *config.Config) (*Registry, error) {
	r := &Registry{
		cfg:       cfg,
		providers: make(map[string]Provider),
	}

	// Initialize providers if API keys are present
	if cfg.OpenAI.APIKey != "" {
		p, err := NewOpenAIProvider(cfg.OpenAI.APIKey, cfg.OpenAI.BaseURL, cfg.OpenAI.DefaultModel)
		if err != nil {
			return nil, fmt.Errorf("failed to create OpenAI provider: %w", err)
		}
		r.providers["openai"] = p
	}

	if cfg.Gemini.APIKey != "" {
		p, err := NewGeminiProvider(cfg.Gemini.APIKey, cfg.Gemini.BaseURL, cfg.Gemini.DefaultModel)
		if err != nil {
			return nil, fmt.Errorf("failed to create Gemini provider: %w", err)
		}
		r.providers["gemini"] = p
	}

	return r, nil
}

// Route routes a request to the appropriate provider based on routing rules
func (r *Registry) Route(req *RouteRequest) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// First, try explicit routing rules from config
	for _, rt := range r.cfg.Routes {
		if strings.HasPrefix(req.Model, rt.Prefix) {
			if provider, exists := r.providers[rt.Provider]; exists {
				return provider, nil
			}
			return nil, fmt.Errorf("provider %s not configured", rt.Provider)
		}
	}

	// Fallback: try provider default by model name hint
	return r.GetProviderForModel(req.Model)
}

// GetProviderForModel returns a provider for the given model using fallback logic
func (r *Registry) GetProviderForModel(model string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Try model name-based routing as fallback
	// More flexible OpenAI routing - check for common patterns and openai-compatible models
	if strings.HasPrefix(model, "gpt-") ||
		strings.HasPrefix(model, "gpt4") ||
		strings.Contains(model, "gpt") ||
		strings.HasSuffix(model, "-openai") ||
		strings.Contains(model, "openai") {
		if provider, exists := r.providers["openai"]; exists {
			return provider, nil
		}
	}

	// More flexible Gemini routing
	if strings.HasPrefix(model, "gemini-") ||
		strings.Contains(model, "gemini") ||
		strings.HasSuffix(model, "-gemini") {
		if provider, exists := r.providers["gemini"]; exists {
			return provider, nil
		}
	}

	return nil, fmt.Errorf("no provider matched model %q", model)
}

// RegisterProvider registers a new provider with the given name
func (r *Registry) RegisterProvider(name string, provider Provider) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	r.providers[name] = provider
	return nil
}

// ListProviders returns information about all registered providers
func (r *Registry) ListProviders() []ProviderInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	infos := make([]ProviderInfo, 0, len(r.providers))
	for _, provider := range r.providers {
		infos = append(infos, provider.GetInfo())
	}

	return infos
}

// GetProvider returns a provider by name
func (r *Registry) GetProvider(name string) (Provider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, exists := r.providers[name]
	return provider, exists
}

// Close closes all registered providers
func (r *Registry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var lastErr error
	for name, provider := range r.providers {
		if err := provider.Close(); err != nil {
			lastErr = fmt.Errorf("failed to close provider %s: %w", name, err)
		}
	}

	return lastErr
}

// Router is an alias for Registry to maintain backward compatibility
type Router = Registry

// NewRouter creates a new router (alias for NewRegistry)
func NewRouter(cfg *config.Config) (*Router, error) {
	return NewRegistry(cfg)
}

// ForModel is a backward compatibility method
func (r *Registry) ForModel(model string) (Provider, error) {
	return r.GetProviderForModel(model)
}
