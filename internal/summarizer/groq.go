package summarizer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// RateLimits holds Groq rate limit headers captured from the last API response.
type RateLimits struct {
	LimitRequests     string `json:"limitRequests"`
	LimitTokens       string `json:"limitTokens"`
	RemainingRequests string `json:"remainingRequests"`
	RemainingTokens   string `json:"remainingTokens"`
	ResetRequests     string `json:"resetRequests"`
	ResetTokens       string `json:"resetTokens"`
}

// RateLimiter is optionally implemented by providers that expose rate limit info.
type RateLimiter interface {
	GetLastRateLimits() *RateLimits
}

type GroqClient struct {
	apiKey     string
	model      string
	lastLimits *RateLimits
}

func (c *GroqClient) GetLastRateLimits() *RateLimits { return c.lastLimits }

func (c *GroqClient) Summarize(ctx context.Context, text, instruction string) (string, error) {
	if instruction == "" {
		instruction = "Please summarize the selection using precise and concise language. Use headers and bulleted lists in the summary, to make it scannable. Maintain the meaning and factual accuracy."
	}

	payload := map[string]any{
		"model": c.model,
		"messages": []map[string]string{
			{"role": "system", "content": instruction},
			{"role": "user", "content": text},
		},
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.groq.com/openai/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("groq request: %w", err)
	}
	defer resp.Body.Close()

	c.lastLimits = &RateLimits{
		LimitRequests:     resp.Header.Get("x-ratelimit-limit-requests"),
		LimitTokens:       resp.Header.Get("x-ratelimit-limit-tokens"),
		RemainingRequests: resp.Header.Get("x-ratelimit-remaining-requests"),
		RemainingTokens:   resp.Header.Get("x-ratelimit-remaining-tokens"),
		ResetRequests:     resp.Header.Get("x-ratelimit-reset-requests"),
		ResetTokens:       resp.Header.Get("x-ratelimit-reset-tokens"),
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return "", fmt.Errorf("rate limited")
	}
	if resp.StatusCode != http.StatusOK {
		var errBody map[string]any
		json.NewDecoder(resp.Body).Decode(&errBody)
		return "", fmt.Errorf("groq error %d: %v", resp.StatusCode, errBody)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no choices in groq response")
	}
	return result.Choices[0].Message.Content, nil
}
