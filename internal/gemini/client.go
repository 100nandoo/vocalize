package gemini

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"google.golang.org/genai"
)

type Client struct {
	inner *genai.Client
}

func New(apiKey string) (*Client, error) {
	c, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("create genai client: %w", err)
	}
	return &Client{inner: c}, nil
}

// sentenceRe matches a sentence: any chars up to and including a terminator (. ! ?),
// optionally followed by whitespace.
var sentenceRe = regexp.MustCompile(`[^.!?]*[.!?]+\s*`)

// splitChunks splits text into chunks of 120–150 words, breaking at sentence boundaries.
func splitChunks(text string) []string {
	sentences := sentenceRe.FindAllString(text, -1)
	// capture any trailing fragment without a terminator
	if tail := strings.TrimSpace(sentenceRe.ReplaceAllString(text, "")); tail != "" {
		sentences = append(sentences, tail)
	}

	const minWords = 120
	const maxWords = 150

	var chunks []string
	var cur strings.Builder
	wordCount := 0

	for _, s := range sentences {
		wc := len(strings.Fields(s))
		// close current chunk before adding if we'd overshoot maxWords and already hit minWords
		if wordCount >= minWords && wordCount+wc > maxWords {
			chunks = append(chunks, strings.TrimSpace(cur.String()))
			cur.Reset()
			wordCount = 0
		}
		cur.WriteString(s)
		wordCount += wc
	}
	if cur.Len() > 0 {
		if t := strings.TrimSpace(cur.String()); t != "" {
			chunks = append(chunks, t)
		}
	}
	return chunks
}

// GenerateSpeech returns raw PCM bytes (16-bit LE mono 24kHz) for the given text.
// Long texts are split by blank lines and each paragraph is requested separately,
// then the PCM streams are concatenated before returning.
func (c *Client) GenerateSpeech(ctx context.Context, text, voice, model string) ([]byte, error) {
	chunks := splitChunks(text)
	if len(chunks) == 0 {
		return nil, fmt.Errorf("no text provided")
	}
	if len(chunks) == 1 {
		return c.generateChunk(ctx, chunks[0], voice, model)
	}

	var combined []byte
	for i, chunk := range chunks {
		pcm, err := c.generateChunk(ctx, chunk, voice, model)
		if err != nil {
			return nil, fmt.Errorf("chunk %d: %w", i+1, err)
		}
		combined = append(combined, pcm...)
	}
	return combined, nil
}

func (c *Client) generateChunk(ctx context.Context, text, voice, model string) ([]byte, error) {
	contents := []*genai.Content{
		{Parts: []*genai.Part{{Text: text}}},
	}

	cfg := &genai.GenerateContentConfig{
		ResponseModalities: []string{"AUDIO"},
		SpeechConfig: &genai.SpeechConfig{
			VoiceConfig: &genai.VoiceConfig{
				PrebuiltVoiceConfig: &genai.PrebuiltVoiceConfig{
					VoiceName: voice,
				},
			},
		},
	}

	resp, err := c.inner.Models.GenerateContent(ctx, model, contents, cfg)
	if err != nil {
		if IsRateLimit(err) {
			return nil, fmt.Errorf("rate limited: %w", err)
		}
		return nil, fmt.Errorf("generate content: %w", err)
	}

	if len(resp.Candidates) == 0 ||
		resp.Candidates[0].Content == nil ||
		len(resp.Candidates[0].Content.Parts) == 0 ||
		resp.Candidates[0].Content.Parts[0].InlineData == nil {
		return nil, fmt.Errorf("no audio data in response")
	}

	return resp.Candidates[0].Content.Parts[0].InlineData.Data, nil
}

// IsRateLimit reports whether err is a quota / rate-limit error from the API.
func IsRateLimit(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "429") ||
		strings.Contains(s, "resource_exhausted") ||
		strings.Contains(s, "quota") ||
		strings.Contains(s, "rate limit") ||
		strings.Contains(s, "ratelimit")
}
