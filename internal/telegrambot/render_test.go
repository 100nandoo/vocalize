package telegrambot

import (
	"strings"
	"testing"

	"github.com/100nandoo/inti/internal/textprocessing"
)

func TestRenderTelegramMarkdown(t *testing.T) {
	input := "# Title\n\n- first item\n- **bold** and *italic*\n\nParagraph with `code` and [link](https://example.com).\n\n```go\nfmt.Println(\"hi\")\n```"
	got := renderTelegramMarkdown(input)

	wantParts := []string{
		`*Title*`,
		`• first item`,
		`• *bold* and _italic_`,
		"Paragraph with `code` and [link](https://example.com)\\.",
		"```\nfmt.Println(\"hi\")\n```",
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("renderTelegramMarkdown() missing %q in output:\n%s", want, got)
		}
	}
}

func TestFormatTelegramSummaryResultIncludesProviderAndModel(t *testing.T) {
	got := formatTelegramSummaryResult(textprocessing.SummaryResult{
		Summary:  "# Title\n\nCondensed text",
		Provider: "groq",
		Model:    "openai/gpt-oss-120b",
	})

	wantParts := []string{
		`*Title*`,
		`Condensed text`,
		`_Provider: groq • Model: openai/gpt\-oss\-120b_`,
	}
	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Fatalf("formatTelegramSummaryResult() missing %q in output:\n%s", want, got)
		}
	}
}
