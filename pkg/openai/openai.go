// pkg/openai/openai.go
package openai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"document-embeddings/internal/config"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	model      string
	maxRetries int
}

type EmbeddingRequest struct {
	Input []string `json:"input"`
	Model string   `json:"model"`
}

type EmbeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

type ChatRequest struct {
	Model    string `json:"model"`
	Messages []struct {
		Role    string `json:"role"`
		Content []struct {
			Type     string `json:"type"`
			Text     string `json:"text,omitempty"`
			ImageURL *struct {
				URL string `json:"url"`
			} `json:"image_url,omitempty"`
		} `json:"content"`
	} `json:"messages"`
	MaxTokens int `json:"max_tokens"`
}

type ChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func New(cfg config.OpenAIConfig) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		baseURL:    cfg.BaseURL,
		apiKey:     cfg.APIKey,
		model:      cfg.Model,
		maxRetries: cfg.MaxRetries,
	}
}

func (c *Client) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	req := EmbeddingRequest{
		Input: texts,
		Model: c.model,
	}

	var resp EmbeddingResponse
	if err := c.makeRequest(ctx, "POST", "/embeddings", req, &resp); err != nil {
		return nil, err
	}

	embeddings := make([][]float32, len(resp.Data))
	for _, data := range resp.Data {
		embeddings[data.Index] = data.Embedding
	}

	return embeddings, nil
}

func (c *Client) ExtractTextFromImage(ctx context.Context, imageData []byte, mimeType string) (string, error) {
	imageURL := fmt.Sprintf("data:%s;base64,%s", mimeType, encodeBase64(imageData))

	req := ChatRequest{
		Model: "gpt-4o-mini",
		Messages: []struct {
			Role    string `json:"role"`
			Content []struct {
				Type     string `json:"type"`
				Text     string `json:"text,omitempty"`
				ImageURL *struct {
					URL string `json:"url"`
				} `json:"image_url,omitempty"`
			} `json:"content"`
		}{
			{
				Role: "user",
				Content: []struct {
					Type     string `json:"type"`
					Text     string `json:"text,omitempty"`
					ImageURL *struct {
						URL string `json:"url"`
					} `json:"image_url,omitempty"`
				}{
					{
						Type: "text",
						Text: "Extract all text from this image. Return only the extracted text without any additional commentary or formatting.",
					},
					{
						Type: "image_url",
						ImageURL: &struct {
							URL string `json:"url"`
						}{URL: imageURL},
					},
				},
			},
		},
		MaxTokens: 4000,
	}

	var resp ChatResponse
	if err := c.makeRequest(ctx, "POST", "/chat/completions", req, &resp); err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return resp.Choices[0].Message.Content, nil
}

func (c *Client) makeRequest(ctx context.Context, method, endpoint string, body interface{}, response interface{}) error {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+endpoint, reqBody)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	var resp *http.Response
	for i := 0; i <= c.maxRetries; i++ {
		resp, err = c.httpClient.Do(req)
		if err == nil && resp.StatusCode < 500 {
			break
		}
		if i < c.maxRetries {
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("OpenAI API error: %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(response)
}

func encodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}
