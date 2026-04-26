package cmd

import (
	"fmt"
	"os"

	"github.com/100nandoo/inti/internal/config"
	"github.com/100nandoo/inti/internal/gemini"
	"github.com/100nandoo/inti/internal/tui"
	"github.com/spf13/cobra"
)

var cfg *config.Config

var rootCmd = &cobra.Command{
	Use:   "inti",
	Short: "Text-to-speech powered by Gemini",
	Long:  "Inti converts text to speech using Google Gemini. Run without subcommands for interactive TUI.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfg.GeminiAPIKey == "" {
			return fmt.Errorf("GEMINI_API_KEY is required for TTS — set it in your environment or .env file")
		}
		g, err := gemini.New(cfg.GeminiAPIKey)
		if err != nil {
			return fmt.Errorf("init gemini: %w", err)
		}
		return tui.Run(cfg, g)
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.AddCommand(speakCmd)
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(pdfCmd)
	rootCmd.AddCommand(ocrCmd)
	rootCmd.AddCommand(summarizeCmd)
}

func initConfig() {
	var err error
	cfg, err = config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
