package server

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/100nandoo/vocalize/internal/config"
	"github.com/100nandoo/vocalize/internal/gemini"
)

func Start(cfg *config.Config, webFS embed.FS) error {
	g, err := gemini.New(cfg.GeminiAPIKey)
	if err != nil {
		return fmt.Errorf("init gemini: %w", err)
	}

	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/speak", handleSpeak(g, cfg))
	mux.HandleFunc("/api/voices", handleVoices(cfg))
	mux.HandleFunc("/api/models", handleModels(cfg))
	mux.HandleFunc("/api/ocr", handleOCR())

	// Static files — strip the "web/" prefix from the embedded FS
	webRoot, err := fs.Sub(webFS, "web")
	if err != nil {
		return fmt.Errorf("embed sub: %w", err)
	}
	mux.Handle("/", http.FileServer(http.FS(webRoot)))

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	return http.ListenAndServe(addr, mux)
}
