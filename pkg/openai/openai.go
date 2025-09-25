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
	"strings"
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

type ImageAnalysis struct {
	Summary  string            `json:"summary"`
	Metadata map[string]string `json:"metadata"`
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

// func (c *Client) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
// 	req := EmbeddingRequest{
// 		Input: texts,
// 		Model: c.model,
// 	}

// 	var resp EmbeddingResponse
// 	if err := c.makeRequest(ctx, "POST", "/embeddings", req, &resp); err != nil {
// 		return nil, err
// 	}

// 	embeddings := make([][]float32, len(resp.Data))
// 	for _, data := range resp.Data {
// 		embeddings[data.Index] = data.Embedding
// 	}

// 	return embeddings, nil
// }

func (c *Client) ExtractTextFromImage(ctx context.Context, imageData []byte, mimeType string) (string, error) {
	analysis, err := c.AnalyzeImage(ctx, imageData, mimeType)
	if err != nil {
		return "", err
	}

	// Extract raw text content from metadata if available
	if textContent, exists := analysis.Metadata["raw_text_content"]; exists {
		return textContent, nil
	}

	// Fallback to summary if no specific text content found
	return analysis.Summary, nil
}

func (c *Client) AnalyzeImage(ctx context.Context, imageData []byte, mimeType string) (*ImageAnalysis, error) {
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
						Text: `Analyze this image and provide:
1. A short summary of what you see
2. Extract ALL text content exactly as it appears in the image (do not translate, modify, or interpret)
3. Metadata including:
   - Image type/category
   - Colors (dominant colors)
   - Objects detected
   - Mood/atmosphere
   - Quality/technical aspects
json
{
  "summary": "",
  "raw_text_content": "",
  "metadata": {
    "image_type/category": "Presentation Slide",
    "colors": ["White", "Blue", "Red", "Black"],
    "objects_detected": ["Text", "Arrows", "Bullet Points"],
    "mood/atmosphere": "",
    "quality/technical_aspects": ""
  }
}


Return the response as a JSON object with "summary" and "metadata" fields. In the metadata, include "raw_text_content" with the exact text as it appears in the image.`,
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
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	var analysis ImageAnalysis
	content := resp.Choices[0].Message.Content

	// Remove markdown code blocks if present
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "```json") {
		content = strings.TrimPrefix(content, "```json")
	} else if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```")
	}
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	// Try to parse the cleaned JSON
	if err := json.Unmarshal([]byte(content), &analysis); err != nil {
		// If JSON parsing still fails, try to extract individual fields manually
		analysis = ImageAnalysis{
			Summary:  extractFieldFromJSON(content, "summary"),
			Metadata: make(map[string]string),
		}

		// Extract raw_text_content if available
		if rawText := extractFieldFromJSON(content, "raw_text_content"); rawText != "" {
			analysis.Metadata["raw_text_content"] = rawText
		}

		// If we still can't extract anything meaningful, use the original content
		if analysis.Summary == "" {
			analysis.Summary = resp.Choices[0].Message.Content
			analysis.Metadata["raw_response"] = resp.Choices[0].Message.Content
		}
	}

	return &analysis, nil
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

func extractFieldFromJSON(jsonStr, fieldName string) string {
	// Simple regex-like extraction for JSON fields
	// Look for "fieldName": "value"
	startPattern := fmt.Sprintf(`"%s":`, fieldName)
	startIdx := strings.Index(jsonStr, startPattern)
	if startIdx == -1 {
		return ""
	}

	// Find the start of the value (after the colon and optional whitespace)
	valueStart := startIdx + len(startPattern)
	// Skip whitespace
	for valueStart < len(jsonStr) && (jsonStr[valueStart] == ' ' || jsonStr[valueStart] == '\t') {
		valueStart++
	}

	// Find the end of the value (handle both string and object values)
	if valueStart >= len(jsonStr) {
		return ""
	}

	// If it's a string value (starts with quote)
	if jsonStr[valueStart] == '"' {
		valueStart++ // Skip opening quote
		// Find the closing quote, handling escaped quotes
		valueEnd := valueStart
		for valueEnd < len(jsonStr) {
			if jsonStr[valueEnd] == '"' && (valueEnd == valueStart || jsonStr[valueEnd-1] != '\\') {
				break
			}
			valueEnd++
		}
		if valueEnd < len(jsonStr) {
			return jsonStr[valueStart:valueEnd]
		}
	} else if jsonStr[valueStart] == '{' {
		// If it's an object, find the matching closing brace
		braceCount := 0
		valueEnd := valueStart
		for valueEnd < len(jsonStr) {
			if jsonStr[valueEnd] == '{' {
				braceCount++
			} else if jsonStr[valueEnd] == '}' {
				braceCount--
				if braceCount == 0 {
					valueEnd++
					break
				}
			}
			valueEnd++
		}
		if braceCount == 0 {
			return jsonStr[valueStart:valueEnd]
		}
	}

	return ""
}
