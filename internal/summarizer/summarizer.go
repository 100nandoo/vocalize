package summarizer

import (
	"context"
	"fmt"

	"github.com/100nandoo/vocalize/internal/config"
	"github.com/100nandoo/vocalize/internal/gemini"
)

// Summarizer is implemented by any AI backend that can summarize text.
type Summarizer interface {
	Summarize(ctx context.Context, text, instruction string) (string, error)
}

// New returns the server-configured Summarizer based on cfg.SummarizerProvider.
func New(cfg *config.Config) (Summarizer, error) {
	return newForProvider(cfg.SummarizerProvider, "", "", cfg)
}

// NewFromRequest creates a Summarizer using a provider, key, and model supplied at
// request time (e.g. from the web UI). Falls back to New(cfg) when all are empty.
func NewFromRequest(provider, apiKey, model string, cfg *config.Config) (Summarizer, error) {
	if provider == "" && apiKey == "" && model == "" {
		return New(cfg)
	}
	if provider == "" {
		provider = cfg.SummarizerProvider
	}
	return newForProvider(provider, apiKey, model, cfg)
}

func newForProvider(provider, apiKeyOverride, modelOverride string, cfg *config.Config) (Summarizer, error) {
	switch provider {
	case "gemini":
		key := cfg.GeminiAPIKey
		if apiKeyOverride != "" {
			key = apiKeyOverride
		}
		if key == "" {
			return nil, fmt.Errorf("GEMINI_API_KEY required for provider 'gemini'")
		}
		return gemini.New(key)

	case "groq":
		key := cfg.GroqAPIKey
		if apiKeyOverride != "" {
			key = apiKeyOverride
		}
		if key == "" {
			return nil, fmt.Errorf("GROQ_API_KEY required for provider 'groq'")
		}
		m := cfg.GroqModel
		if modelOverride != "" {
			m = modelOverride
		}
		return &GroqClient{apiKey: key, model: m}, nil

	case "openrouter":
		key := cfg.OpenRouterAPIKey
		if apiKeyOverride != "" {
			key = apiKeyOverride
		}
		if key == "" {
			return nil, fmt.Errorf("OPENROUTER_API_KEY required for provider 'openrouter'")
		}
		m := cfg.OpenRouterModel
		if modelOverride != "" {
			m = modelOverride
		}
		return &OpenRouterClient{apiKey: key, model: m}, nil

	case "":
		return nil, fmt.Errorf("no summarizer provider configured — set SUMMARIZER_PROVIDER or provide an API key in the request")

	default:
		return nil, fmt.Errorf("unknown summarizer provider %q", provider)
	}
}
