package cmd

import (
	"fmt"

	"github.com/100nandoo/inti/internal/summarizer"
	"github.com/spf13/cobra"
)

var summarizeCmd = &cobra.Command{
	Use:   "summarize <text>",
	Short: "Summarize text using a configured AI provider",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		instruction, _ := cmd.Flags().GetString("instruction")
		provider, _ := cmd.Flags().GetString("provider")
		apiKey, _ := cmd.Flags().GetString("api-key")

		s, err := summarizer.NewFromRequest(provider, apiKey, "", cfg)
		if err != nil {
			return fmt.Errorf("init summarizer: %w", err)
		}

		fmt.Println("summarizing…")
		summary, err := s.Summarize(cmd.Context(), args[0], instruction)
		if err != nil {
			return fmt.Errorf("summarize: %w", err)
		}

		fmt.Println("\n--- Summary ---")
		fmt.Println(summary)

		return nil
	},
}

func init() {
	summarizeCmd.Flags().String("instruction", "", "custom summarization instruction (optional)")
	summarizeCmd.Flags().String("provider", "", "AI provider: gemini, groq, openrouter (overrides config)")
	summarizeCmd.Flags().String("api-key", "", "API key for the provider (overrides env var)")
}
