package server

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"sync"

	"github.com/100nandoo/vocalize/internal/config"
	"github.com/100nandoo/vocalize/internal/gemini"
	"github.com/100nandoo/vocalize/internal/summarizer"
)

type activeSumConfig struct {
	mu       sync.RWMutex
	Provider string
	Model    string
	APIKey   string
}

func (a *activeSumConfig) get() (provider, model, apiKey string) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Provider, a.Model, a.APIKey
}

func (a *activeSumConfig) set(provider, model, apiKey string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Provider = provider
	a.Model = model
	a.APIKey = apiKey
}

func loadActiveConfig(cfg *config.Config) *activeSumConfig {
	asc := &activeSumConfig{
		Provider: cfg.SummarizerProvider,
		APIKey:   apiKeyForProvider(cfg.SummarizerProvider, cfg),
	}
	fileMu.Lock()
	vc := readVocalizeConfigUnlocked()
	fileMu.Unlock()
	if vc.Summarizer.Provider != "" {
		asc.Provider = vc.Summarizer.Provider
		asc.APIKey = vc.Summarizer.APIKey
	}
	asc.Model = vc.Summarizer.Model
	return asc
}

func saveActiveConfig(provider, model, apiKey string) error {
	fileMu.Lock()
	defer fileMu.Unlock()
	vc := readVocalizeConfigUnlocked()
	vc.Summarizer = summarizerSection{
		Provider: provider,
		Model:    model,
		APIKey:   apiKey,
	}
	return writeVocalizeConfigUnlocked(vc)
}

func apiKeyForProvider(provider string, cfg *config.Config) string {
	switch provider {
	case "groq":
		return cfg.GroqAPIKey
	case "openrouter":
		return cfg.OpenRouterAPIKey
	default:
		return cfg.GeminiAPIKey
	}
}

func Start(cfg *config.Config, webFS embed.FS) error {
	var g *gemini.Client
	if cfg.GeminiAPIKey != "" {
		var err error
		g, err = gemini.New(cfg.GeminiAPIKey)
		if err != nil {
			return fmt.Errorf("init gemini: %w", err)
		}
	}

	sum, err := summarizer.New(cfg)
	if err != nil {
		// Non-fatal: summarize endpoint will return an error per-request.
		sum = nil
	}

	asc := loadActiveConfig(cfg)
	ks := loadAPIKeyStore()

	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/speak", handleSpeak(g, cfg))
	mux.HandleFunc("/api/voices", handleVoices(cfg))
	mux.HandleFunc("/api/models", handleModels(cfg))
	mux.HandleFunc("/api/ocr", handleOCR())
	mux.HandleFunc("/api/summarize", handleSummarize(sum, asc, cfg))
	mux.HandleFunc("/api/summarizer-config", handleSummarizerConfig(asc, cfg))

	// API key management routes
	mux.HandleFunc("GET /api/admin/keys", handleAdminListKeys(ks))
	mux.HandleFunc("POST /api/admin/keys", handleAdminCreateKey(ks))
	mux.HandleFunc("DELETE /api/admin/keys/{id}", handleAdminDeleteKey(ks))

	// Static files — strip the "web/" prefix from the embedded FS
	webRoot, err := fs.Sub(webFS, "web")
	if err != nil {
		return fmt.Errorf("embed sub: %w", err)
	}
	mux.Handle("/", http.FileServer(http.FS(webRoot)))

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	return http.ListenAndServe(addr, requireAPIKey(cfg.MasterKey, ks, mux))
}
