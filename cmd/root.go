package cmd

import (
	"fmt"
	"os"

	"github.com/100nandoo/vocalize/internal/config"
	"github.com/100nandoo/vocalize/internal/gemini"
	"github.com/100nandoo/vocalize/internal/tui"
	"github.com/spf13/cobra"
)

var cfg *config.Config

var rootCmd = &cobra.Command{
	Use:   "vocalize",
	Short: "Text-to-speech powered by Gemini",
	Long:  "Vocalize converts text to speech using Google Gemini. Run without subcommands for interactive TUI.",
	RunE: func(cmd *cobra.Command, args []string) error {
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
}

func initConfig() {
	var err error
	cfg, err = config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
