package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const GrokAPIURL = "https://api.x.ai/v1/chat/completions"

// GrokClient mengelola integrasi dengan Grok API
type GrokClient struct {
	apiKey string
	client *http.Client
}

// GrokMessage merepresentasikan message dalam API call
type GrokMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// GrokRequest merepresentasikan request ke Grok API
type GrokRequest struct {
	Model    string         `json:"model"`
	Messages []GrokMessage  `json:"messages"`
	MaxToken int            `json:"max_tokens"`
	Temp     float64        `json:"temperature"`
}

// GrokChoice merepresentasikan pilihan dalam response
type GrokChoice struct {
	Index   int          `json:"index"`
	Message GrokMessage  `json:"message"`
	Delta   *GrokMessage `json:"delta,omitempty"`
}

// GrokResponse merepresentasikan response dari Grok API
type GrokResponse struct {
	ID      string       `json:"id"`
	Object  string       `json:"object"`
	Created int64        `json:"created"`
	Model   string       `json:"model"`
	Choices []GrokChoice `json:"choices"`
}

// NewGrokClient membuat instance GrokClient baru
func NewGrokClient() *GrokClient {
	apiKey := os.Getenv("GROK_API_KEY")
	if apiKey == "" {
		// Fallback API key untuk testing (ganti dengan key asli)
		apiKey = "test-key-placeholder"
	}

	return &GrokClient{
		apiKey: apiKey,
		client: &http.Client{},
	}
}

// GenerateResponse menghasilkan response dari Grok AI
func (gc *GrokClient) GenerateResponse(userMessage, conversationContext string) (string, error) {
	messages := []GrokMessage{
		{
			Role:    "system",
			Content: "You are a helpful AI assistant in a chat room. Keep responses concise and friendly.",
		},
		{
			Role:    "user",
			Content: userMessage,
		},
	}

	if conversationContext != "" {
		messages = append(messages, GrokMessage{
			Role:    "assistant",
			Content: conversationContext,
		})
	}

	req := GrokRequest{
		Model:    "grok-2",
		Messages: messages,
		MaxToken: 256,
		Temp:     0.7,
	}

	jsonBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest(
		"POST",
		GrokAPIURL,
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", gc.apiKey))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := gc.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to call Grok API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Grok API error: %s (status: %d)", string(body), resp.StatusCode)
	}

	var grokResp GrokResponse
	if err := json.Unmarshal(body, &grokResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(grokResp.Choices) == 0 {
		return "", fmt.Errorf("no response choices from Grok")
	}

	return grokResp.Choices[0].Message.Content, nil
}

// HealthCheck mengecek konektivitas dengan Grok API
func (gc *GrokClient) HealthCheck() error {
	messages := []GrokMessage{
		{
			Role:    "user",
			Content: "Hello",
		},
	}

	req := GrokRequest{
		Model:    "grok-2",
		Messages: messages,
		MaxToken: 10,
	}

	jsonBody, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("POST", GrokAPIURL, bytes.NewBuffer(jsonBody))
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", gc.apiKey))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := gc.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: %d", resp.StatusCode)
	}

	return nil
}
