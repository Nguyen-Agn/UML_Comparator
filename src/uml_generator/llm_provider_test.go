package uml_generator

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// makeTestServer creates a mock HTTP server returning the given OpenAI-format response.
func makeTestServer(t *testing.T, statusCode int, body map[string]any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(body)
	}))
}

func TestChatComplete_ParsesChoicesCorrectly(t *testing.T) {
	srv := makeTestServer(t, 200, map[string]any{
		"choices": []map[string]any{
			{"message": map[string]any{"content": "classDiagram\n    A-->B"}},
		},
	})
	defer srv.Close()

	p := newOpenAICompatibleProvider(srv.URL, "test-key", "test-model")
	got, err := p.ChatComplete(context.Background(), []llmMessage{{Role: "user", Content: "hi"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "classDiagram\n    A-->B" {
		t.Errorf("unexpected content: %q", got)
	}
}

func TestChatComplete_ErrorOn4xx(t *testing.T) {
	srv := makeTestServer(t, 401, map[string]any{
		"error": map[string]any{"message": "invalid api key"},
	})
	defer srv.Close()

	p := newOpenAICompatibleProvider(srv.URL, "bad-key", "model")
	_, err := p.ChatComplete(context.Background(), []llmMessage{{Role: "user", Content: "hi"}})
	if err == nil {
		t.Fatal("expected error on 401, got nil")
	}
}

func TestChatComplete_ErrorOnEmptyChoices(t *testing.T) {
	srv := makeTestServer(t, 200, map[string]any{
		"choices": []map[string]any{},
	})
	defer srv.Close()

	p := newOpenAICompatibleProvider(srv.URL, "", "model")
	_, err := p.ChatComplete(context.Background(), []llmMessage{{Role: "user", Content: "hi"}})
	if err == nil {
		t.Fatal("expected error for empty choices, got nil")
	}
}

func TestChatComplete_SkipsAuthHeaderWhenKeyEmpty(t *testing.T) {
	var authHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"content": "ok"}},
			},
		})
	}))
	defer srv.Close()

	p := newOpenAICompatibleProvider(srv.URL, "", "model") // empty key = Ollama mode
	_, _ = p.ChatComplete(context.Background(), []llmMessage{{Role: "user", Content: "hi"}})

	if authHeader != "" {
		t.Errorf("expected no Authorization header when key is empty, got %q", authHeader)
	}
}
