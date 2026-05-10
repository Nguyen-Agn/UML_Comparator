package uml_generator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// iLLMProvider is a package-local interface for sending chat messages to an LLM.
// Using OpenAI-compatible /v1/chat/completions endpoint.
// Keeping it package-local (lowercase) — domain layer doesn't need to know about it.
type iLLMProvider interface {
	// ChatComplete sends a list of messages and returns the assistant's text reply.
	// Supports any OpenAI-compatible provider (OpenAI, Groq, Ollama, Azure, Gemini proxy, ...).
	ChatComplete(ctx context.Context, messages []llmMessage) (string, error)
}

// llmMessage represents a single chat message in the conversation.
type llmMessage struct {
	Role    string `json:"role"`    // "system", "user", or "assistant"
	Content string `json:"content"` // The text content of the message
}

// openAICompatibleProvider implements iLLMProvider for any OpenAI-compatible endpoint.
// A single implementation covers: OpenAI, Groq, Ollama, Azure OpenAI, Gemini (via proxy).
type openAICompatibleProvider struct {
	endpoint string       // Base URL, e.g. https://api.openai.com/v1
	apiKey   string       // Bearer token; empty string skips Authorization header (Ollama)
	model    string       // Model ID, e.g. gpt-4o-mini, gemma3
	client   *http.Client // Shared client with timeout
}

// newOpenAICompatibleProvider creates a provider with a 90-second timeout.
func newOpenAICompatibleProvider(endpoint, apiKey, model string) iLLMProvider {
	return &openAICompatibleProvider{
		endpoint: endpoint,
		apiKey:   apiKey,
		model:    model,
		client:   &http.Client{Timeout: 90 * time.Second},
	}
}

// chatRequest is the JSON body sent to /v1/chat/completions.
type chatRequest struct {
	Model    string       `json:"model"`
	Messages []llmMessage `json:"messages"`
}

// chatResponse is the subset of the /v1/chat/completions JSON response we care about.
type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// ChatComplete sends messages to {endpoint}/chat/completions and returns the reply text.
// One HTTP POST is made per call — no multi-turn agent loop.
func (p *openAICompatibleProvider) ChatComplete(ctx context.Context, messages []llmMessage) (string, error) {
	reqBody := chatRequest{Model: p.model, Messages: messages}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("llm_provider: marshal request: %w", err)
	}

	url := p.endpoint + "/chat/completions"
	log.Printf("[LLMProvider] POST %s  model=%q  body=%s\n", url, p.model, string(bodyBytes))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("llm_provider: build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.apiKey)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("llm_provider: http do: %w", err)
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("llm_provider: read body: %w", err)
	}
	log.Printf("[LLMProvider] status=%d body_len=%d\n", resp.StatusCode, len(rawBody))

	// Check HTTP error status BEFORE attempting JSON unmarshal.
	// Ollama and some proxies return plain text on 4xx (e.g. "404 page not found"),
	// which would cause a misleading unmarshal error if not caught here.
	if resp.StatusCode >= 400 {
		// Try structured JSON error first (OpenAI format)
		var errBody struct {
			Error *struct {
				Message string `json:"message"`
			} `json:"error,omitempty"`
		}
		apiErrorMessage := ""
		if jsonErr := json.Unmarshal(rawBody, &apiErrorMessage); jsonErr == nil && apiErrorMessage != "" {
			// Some providers return a plain string for error
		} else if jsonErr := json.Unmarshal(rawBody, &errBody); jsonErr == nil && errBody.Error != nil {
			apiErrorMessage = errBody.Error.Message
		}

		switch resp.StatusCode {
		case 401:
			return "", fmt.Errorf("llm_provider: 401 Unauthorized — Sai API Key hoặc không có quyền truy cập. Kiểm tra cấu hình.")
		case 403:
			return "", fmt.Errorf("llm_provider: 403 Forbidden — Yêu cầu bị server từ chối. Có thể do model không khả dụng hoặc bị chặn.")
		case 404:
			return "", fmt.Errorf(
				"llm_provider: 404 Not Found — Sai Endpoint URL hoặc Model không tồn tại.\n"+
					"URL: %s\nModel: %s\nRaw: %s", url, p.model, string(rawBody))
		case 429:
			return "", fmt.Errorf("llm_provider: 429 Too Many Requests — Hết quota hoặc bị giới hạn tần suất gọi API (Rate Limit).")
		case 500, 502, 503, 504:
			return "", fmt.Errorf("llm_provider: Lỗi Server (%d) — AI Provider đang gặp sự cố. Thử lại sau.", resp.StatusCode)
		default:
			if apiErrorMessage != "" {
				return "", fmt.Errorf("llm_provider: HTTP %d: %s", resp.StatusCode, apiErrorMessage)
			}
			return "", fmt.Errorf("llm_provider: HTTP %d: %s", resp.StatusCode, string(rawBody))
		}
	}

	var cr chatResponse
	if err := json.Unmarshal(rawBody, &cr); err != nil {
		return "", fmt.Errorf("llm_provider: unmarshal response: %w", err)
	}
	if len(cr.Choices) == 0 {
		return "", fmt.Errorf("llm_provider: no choices in response")
	}

	return cr.Choices[0].Message.Content, nil
}
