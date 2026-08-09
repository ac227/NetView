package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Config struct {
	BaseURL string
	APIKey  string
	Model   string
}

type Client struct {
	cfg Config
	hc  *http.Client
}

func NewClient(cfg Config) *Client {
	return &Client{cfg: cfg, hc: &http.Client{Timeout: 60 * time.Second}}
}

func (c *Client) Enabled() bool {
	return c.cfg.APIKey != "" && c.cfg.BaseURL != ""
}

type chatMessage struct {
	Role    string        `json:"role"`
	Content []contentPart `json:"content"`
}

type contentPart struct {
	Type     string    `json:"type,omitempty"`
	Text     string    `json:"text,omitempty"`
	ImageURL *imageURL `json:"image_url,omitempty"`
}

type imageURL struct {
	URL string `json:"url"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
	ResponseFormat *responseFormat `json:"response_format,omitempty"`
	MaxTokens int `json:"max_tokens,omitempty"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

type TagResult struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

// TagImage analyzes an image file and returns suggested tags/title/description.
func (c *Client) TagImage(ctx context.Context, imagePath string) (*TagResult, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("AI not configured")
	}
	data, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, err
	}
	b64 := base64.StdEncoding.EncodeToString(data)
	parts := []contentPart{
		{Type: "text", Text: `Analyze this image and return JSON with fields:
"title" (short title in the image language), "description" (1-2 sentence description),
"tags" (array of 3-8 concise tags, in Chinese, lowercase, hyphen-free).
Only output the JSON object, no markdown.`},
		{Type: "image_url", ImageURL: &imageURL{URL: "data:image/jpeg;base64," + b64}},
	}
	return c.chat(ctx, parts)
}

// TagURL analyzes a remote image URL.
func (c *Client) TagURL(ctx context.Context, imgURL string) (*TagResult, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("AI not configured")
	}
	parts := []contentPart{
		{Type: "text", Text: `Analyze the image at the given URL and return JSON with fields:
"title", "description", "tags" (array of 3-8 concise tags, in Chinese).
Only output the JSON object, no markdown.`},
		{Type: "image_url", ImageURL: &imageURL{URL: imgURL}},
	}
	return c.chat(ctx, parts)
}

func (c *Client) chat(ctx context.Context, parts []contentPart) (*TagResult, error) {
	reqBody := chatRequest{
		Model: c.cfg.Model,
		Messages: []chatMessage{{Role: "user", Content: parts}},
		ResponseFormat: &responseFormat{Type: "json_object"},
		MaxTokens: 500,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := strings.TrimSuffix(c.cfg.BaseURL, "/") + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)

	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("AI API error %d: %s", resp.StatusCode, firstBytes(data, 300))
	}

	var cr chatResponse
	if err := json.Unmarshal(data, &cr); err != nil {
		return nil, err
	}
	if cr.Error != nil {
		return nil, fmt.Errorf("AI API: %s", cr.Error.Message)
	}
	if len(cr.Choices) == 0 {
		return nil, fmt.Errorf("AI API: empty response")
	}

	var result TagResult
	content := strings.TrimSpace(cr.Choices[0].Message.Content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("parse AI response: %w (content: %s)", err, firstBytes([]byte(content), 200))
	}
	return &result, nil
}

func firstBytes(b []byte, n int) string {
	if len(b) > n {
		return string(b[:n]) + "..."
	}
	return string(b)
}
