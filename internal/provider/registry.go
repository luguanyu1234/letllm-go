package provider

import (
	"fmt"
	"strings"

	"github.com/luguanyu1234/letllm-go/internal/config"
)

type Router struct {
	cfg    *config.Config
	openai *OpenAIProvider
	gemini *GeminiProvider
}

func NewRouter(cfg *config.Config) (*Router, error) {
	r := &Router{cfg: cfg}
	// Initialize providers if API keys are present
	if cfg.OpenAI.APIKey != "" {
		p, err := NewOpenAIProvider(cfg.OpenAI.APIKey, cfg.OpenAI.DefaultModel)
		if err != nil { return nil, err }
		r.openai = p
	}
	if cfg.Gemini.APIKey != "" {
		p, err := NewGeminiProvider(cfg.Gemini.APIKey, cfg.Gemini.DefaultModel)
		if err != nil { return nil, err }
		r.gemini = p
	}
	return r, nil
}

func (r *Router) ForModel(model string) (Provider, error) {
	// Choose by prefix routing
	for _, rt := range r.cfg.Routes {
		if strings.HasPrefix(model, rt.Prefix) {
			switch rt.Provider {
			case "openai":
				if r.openai != nil { return r.openai, nil }
				return nil, fmt.Errorf("openai provider not configured")
			case "gemini":
				if r.gemini != nil { return r.gemini, nil }
				return nil, fmt.Errorf("gemini provider not configured")
			default:
				return nil, fmt.Errorf("unknown provider: %s", rt.Provider)
			}
		}
	}
	// Fallback: try provider default by model name hint
	if strings.HasPrefix(model, "gpt-") || strings.HasPrefix(model, "gpt4") || strings.Contains(model, "gpt") {
		if r.openai != nil { return r.openai, nil }
	}
	if strings.HasPrefix(model, "gemini-") {
		if r.gemini != nil { return r.gemini, nil }
	}
	return nil, fmt.Errorf("no provider matched model %q", model)
}
