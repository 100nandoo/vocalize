package cmd

import (
	"fmt"

	"github.com/100nandoo/inti/internal/audio"
	"github.com/100nandoo/inti/internal/config"
	"github.com/100nandoo/inti/internal/gemini"
	"github.com/spf13/cobra"
)

var speakCmd = &cobra.Command{
	Use:   "speak <text>",
	Short: "Synthesize and play text",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		voice, _ := cmd.Flags().GetString("voice")
		if voice == "" {
			voice = cfg.DefaultVoice
		}
		if !config.IsValidVoice(voice) {
			return fmt.Errorf("invalid voice %q, valid voices: %v", voice, config.ValidVoices())
		}

		model, _ := cmd.Flags().GetString("model")
		if model == "" {
			model = cfg.DefaultModel
		}
		if !config.IsValidModel(model) {
			return fmt.Errorf("invalid model %q, valid models: %v", model, config.ValidModels())
		}

		exportPath, _ := cmd.Flags().GetString("export")

		if cfg.GeminiAPIKey == "" {
			return fmt.Errorf("GEMINI_API_KEY is required for TTS — set it in your environment or .env file")
		}
		g, err := gemini.New(cfg.GeminiAPIKey)
		if err != nil {
			return fmt.Errorf("init gemini: %w", err)
		}

		fmt.Printf("synthesizing with voice %s (model: %s)...\n", voice, model)
		pcm, err := g.GenerateSpeech(cmd.Context(), args[0], voice, model)
		if err != nil {
			if gemini.IsRateLimit(err) {
				return fmt.Errorf("rate limited — wait a moment and try again")
			}
			return fmt.Errorf("generate speech: %w", err)
		}

		if exportPath != "" {
			if err := audio.WriteOpusFile(exportPath, pcm, 24000); err != nil {
				return fmt.Errorf("write opus: %w", err)
			}
			fmt.Printf("saved to %s\n", exportPath)
		}

		if exportPath == "" || mustBool(cmd.Flags().GetBool("play")) {
			fmt.Println("playing...")
			if err := audio.Play(pcm); err != nil {
				return fmt.Errorf("play audio: %w", err)
			}
		}

		return nil
	},
}

func init() {
	speakCmd.Flags().String("voice", "", "voice name (default from config)")
	speakCmd.Flags().String("model", "", fmt.Sprintf("TTS model (default: %s)", config.DefaultModelName))
	speakCmd.Flags().String("export", "", "save audio to this .opus file path")
	speakCmd.Flags().Bool("play", false, "play audio even when --export is set")
}

func mustBool(v bool, _ error) bool { return v }
