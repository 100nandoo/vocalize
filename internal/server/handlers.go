package server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/100nandoo/inti/internal/audio"
	"github.com/100nandoo/inti/internal/config"
	"github.com/100nandoo/inti/internal/gemini"
	"github.com/100nandoo/inti/internal/ocr"
	"github.com/100nandoo/inti/internal/summarizer"
)

type speakRequest struct {
	Text  string `json:"text"`
	Voice string `json:"voice"`
	Model string `json:"model"`
}

type speakResponse struct {
	Opus string `json:"opus"`
}

type voicesResponse struct {
	Voices  []string `json:"voices"`
	Default string   `json:"default"`
}

type modelsResponse struct {
	Models  []string `json:"models"`
	Default string   `json:"default"`
}

type errResponse struct {
	Error string `json:"error"`
}

type summarizeRequest struct {
	Text        string `json:"text"`
	Instruction string `json:"instruction"`
	Provider    string `json:"provider"` // optional: override server-configured provider
	APIKey      string `json:"apiKey"`   // optional: key supplied from web UI
	Model       string `json:"model"`    // optional: override provider default model
	Mock        bool   `json:"mock"`     // optional: return a mock summary
}

type summarizerConfigResponse struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
}

type summarizeResponse struct {
	Summary    string                  `json:"summary"`
	Provider   string                  `json:"provider"`
	Model      string                  `json:"model"`
	RateLimits *summarizer.RateLimits  `json:"rateLimits,omitempty"`
}

func handleSpeak(g *gemini.Client, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if g == nil {
			writeJSON(w, http.StatusServiceUnavailable, errResponse{"TTS unavailable — GEMINI_API_KEY not configured"})
			return
		}
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, errResponse{"method not allowed"})
			return
		}

		var req speakRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errResponse{"invalid request body"})
			return
		}

		if req.Text == "" {
			writeJSON(w, http.StatusBadRequest, errResponse{"text is required"})
			return
		}

		voice := req.Voice
		if voice == "" {
			voice = cfg.DefaultVoice
		}
		if !config.IsValidVoice(voice) {
			writeJSON(w, http.StatusBadRequest, errResponse{"invalid voice: " + voice})
			return
		}

		model := req.Model
		if model == "" {
			model = cfg.DefaultModel
		}
		if !config.IsValidModel(model) {
			writeJSON(w, http.StatusBadRequest, errResponse{"invalid model: " + model})
			return
		}

		pcm, err := g.GenerateSpeech(r.Context(), req.Text, voice, model)
		if err != nil {
			if gemini.IsRateLimit(err) {
				writeJSON(w, http.StatusTooManyRequests, errResponse{"rate limited — wait a moment and try again"})
			} else {
				writeJSON(w, http.StatusInternalServerError, errResponse{err.Error()})
			}
			return
		}

		opusBytes, err := audio.EncodePCMToOpus(pcm, 24000)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errResponse{err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, speakResponse{
			Opus: base64.StdEncoding.EncodeToString(opusBytes),
		})
	}
}

func handleVoices(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, voicesResponse{
			Voices:  config.ValidVoices(),
			Default: cfg.DefaultVoice,
		})
	}
}

func handleModels(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, modelsResponse{
			Models:  config.ValidModels(),
			Default: cfg.DefaultModel,
		})
	}
}

type ocrResponse struct {
	Text string `json:"text"`
}

func handleOCR() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, errResponse{"method not allowed"})
			return
		}

		if err := r.ParseMultipartForm(50 << 20); err != nil {
			writeJSON(w, http.StatusBadRequest, errResponse{"invalid multipart form"})
			return
		}

		// Accept "files" (multi-upload) with "file" as a fallback for single-file requests.
		fileHeaders := r.MultipartForm.File["files"]
		if len(fileHeaders) == 0 {
			fileHeaders = r.MultipartForm.File["file"]
		}
		if len(fileHeaders) == 0 {
			writeJSON(w, http.StatusBadRequest, errResponse{"at least one file is required"})
			return
		}

		var parts []string
		for _, fh := range fileHeaders {
			f, err := fh.Open()
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, errResponse{"open file: " + err.Error()})
				return
			}
			imageBytes, err := io.ReadAll(f)
			f.Close()
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, errResponse{"read file: " + err.Error()})
				return
			}
			text, err := ocr.ExtractText(imageBytes)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, errResponse{err.Error()})
				return
			}
			if text != "" {
				parts = append(parts, text)
			}
		}

		writeJSON(w, http.StatusOK, ocrResponse{Text: strings.Join(parts, "\n\n")})
	}
}

func handleSummarize(serverSum summarizer.Summarizer, asc *activeSumConfig, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, errResponse{"method not allowed"})
			return
		}

		var req summarizeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errResponse{"invalid request body"})
			return
		}

		if req.Text == "" {
			writeJSON(w, http.StatusBadRequest, errResponse{"text is required"})
			return
		}

		if req.Mock {
			writeJSON(w, http.StatusOK, summarizeResponse{
				Summary:  fmt.Sprintf("This is a mock summary of the provided text.\n\nOriginal text length: %d characters.\nInstruction: %q", len(req.Text), req.Instruction),
				Provider: "mock",
				Model:    "mock-model",
			})
			return
		}

		s := serverSum
		usedProvider := cfg.SummarizerProvider
		if req.Provider != "" || req.APIKey != "" || req.Model != "" {
			// explicit per-request override (web UI sends these)
			var err error
			s, err = summarizer.NewFromRequest(req.Provider, req.APIKey, req.Model, cfg)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, errResponse{err.Error()})
				return
			}
			if req.Provider != "" {
				usedProvider = req.Provider
			}
		} else {
			// no per-request override: use activeSumConfig saved by web UI
			activeProvider, activeModel, activeAPIKey := asc.get()
			if activeProvider != "" || activeAPIKey != "" || activeModel != "" {
				var err error
				s, err = summarizer.NewFromRequest(activeProvider, activeAPIKey, activeModel, cfg)
				if err != nil {
					writeJSON(w, http.StatusBadRequest, errResponse{err.Error()})
					return
				}
				if activeProvider != "" {
					usedProvider = activeProvider
				}
			}
		}
		if s == nil {
			writeJSON(w, http.StatusServiceUnavailable, errResponse{"no summarizer configured — set a provider and API key"})
			return
		}

		summary, err := s.Summarize(r.Context(), req.Text, req.Instruction)
		if err != nil {
			if gemini.IsRateLimit(err) || strings.Contains(err.Error(), "rate limited") {
				writeJSON(w, http.StatusTooManyRequests, errResponse{"rate limited — wait a moment and try again"})
			} else {
				writeJSON(w, http.StatusInternalServerError, errResponse{err.Error()})
			}
			return
		}

		usedModel := req.Model
		if usedModel == "" {
			usedModel = modelForProvider(usedProvider, cfg)
		}
		var rateLimits *summarizer.RateLimits
		if rl, ok := s.(summarizer.RateLimiter); ok {
			rateLimits = rl.GetLastRateLimits()
		}
		writeJSON(w, http.StatusOK, summarizeResponse{
			Summary:    summary,
			Provider:   usedProvider,
			Model:      usedModel,
			RateLimits: rateLimits,
		})
	}
}

type summarizerConfigRequest struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
	APIKey   string `json:"apiKey"`
}

func handleSummarizerConfig(asc *activeSumConfig, cfg *config.Config) http.HandlerFunc {
	validProviders := map[string]bool{"gemini": true, "groq": true, "openrouter": true, "mock": true, "": true}
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			provider, model, _ := asc.get()
			if model == "" {
				model = modelForProvider(provider, cfg)
			}
			writeJSON(w, http.StatusOK, summarizerConfigResponse{Provider: provider, Model: model})
		case http.MethodPost:
			var req summarizerConfigRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeJSON(w, http.StatusBadRequest, errResponse{"invalid request body"})
				return
			}
			if !validProviders[req.Provider] {
				writeJSON(w, http.StatusBadRequest, errResponse{"invalid provider"})
				return
			}
			asc.set(req.Provider, req.Model, req.APIKey)
			if err := saveActiveConfig(req.Provider, req.Model, req.APIKey); err != nil {
				// non-fatal: config will still work in-memory this session
				_ = err
			}
			model := req.Model
			if model == "" {
				model = modelForProvider(req.Provider, cfg)
			}
			writeJSON(w, http.StatusOK, summarizerConfigResponse{Provider: req.Provider, Model: model})
		default:
			writeJSON(w, http.StatusMethodNotAllowed, errResponse{"method not allowed"})
		}
	}
}

func modelForProvider(provider string, cfg *config.Config) string {
	switch provider {
	case "gemini":
		return "gemini-2.0-flash"
	case "groq":
		return cfg.GroqModel
	case "openrouter":
		return cfg.OpenRouterModel
	case "mock":
		return "mock-model"
	default:
		return ""
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
