package cmd

import (
	"fmt"
	"os"

	"github.com/100nandoo/vocalize/internal/audio"
	"github.com/100nandoo/vocalize/internal/config"
	"github.com/100nandoo/vocalize/internal/gemini"
	"github.com/100nandoo/vocalize/internal/ocr"
	"github.com/spf13/cobra"
)

var ocrCmd = &cobra.Command{
	Use:   "ocr <image-path>",
	Short: "Extract text from an image and optionally synthesize it",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		imageBytes, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("read image: %w", err)
		}

		fmt.Println("extracting text from image...")
		text, err := ocr.ExtractText(imageBytes)
		if err != nil {
			return fmt.Errorf("ocr: %w", err)
		}

		if text == "" {
			fmt.Println("no text found in image")
			return nil
		}

		fmt.Println(text)

		speak, _ := cmd.Flags().GetBool("speak")
		if !speak {
			return nil
		}

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

		g, err := gemini.New(cfg.GeminiAPIKey)
		if err != nil {
			return fmt.Errorf("init gemini: %w", err)
		}

		fmt.Printf("synthesizing with voice %s (model: %s)...\n", voice, model)
		pcm, err := g.GenerateSpeech(cmd.Context(), text, voice, model)
		if err != nil {
			if gemini.IsRateLimit(err) {
				return fmt.Errorf("rate limited — wait a moment and try again")
			}
			return fmt.Errorf("generate speech: %w", err)
		}

		exportPath, _ := cmd.Flags().GetString("export")
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
	ocrCmd.Flags().Bool("speak", false, "synthesize extracted text with TTS")
	ocrCmd.Flags().String("voice", "", "voice name (default from config)")
	ocrCmd.Flags().String("model", "", fmt.Sprintf("TTS model (default: %s)", config.DefaultModelName))
	ocrCmd.Flags().String("export", "", "save audio to this .opus file path")
	ocrCmd.Flags().Bool("play", false, "play audio even when --export is set")
}
