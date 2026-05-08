package uml_generator

import (
	"os"
	"path/filepath"
	"testing"

	"uml_compare/domain"
)

func makeTestStore(t *testing.T) *configStore {
	t.Helper()
	tmp := t.TempDir()
	return &configStore{path: filepath.Join(tmp, "test_config.json")}
}

func TestConfigStore_DefaultWhenMissing(t *testing.T) {
	s := makeTestStore(t)
	cfg, err := s.Load()
	if err != nil {
		t.Fatalf("Load on missing file returned error: %v", err)
	}
	if cfg.APIEndpoint == "" {
		t.Error("default config should have a non-empty APIEndpoint")
	}
	if cfg.Model == "" {
		t.Error("default config should have a non-empty Model")
	}
}

func TestConfigStore_RoundTrip(t *testing.T) {
	s := makeTestStore(t)
	want := domain.GeneratorConfig{
		APIKey:      "sk-test-key-12345",
		APIEndpoint: "http://localhost:11434/v1",
		Model:       "gemma3",
	}

	if err := s.Save(want); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	got, err := s.Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}

	if got.APIKey != want.APIKey {
		t.Errorf("APIKey: want %q, got %q", want.APIKey, got.APIKey)
	}
	if got.APIEndpoint != want.APIEndpoint {
		t.Errorf("APIEndpoint: want %q, got %q", want.APIEndpoint, got.APIEndpoint)
	}
	if got.Model != want.Model {
		t.Errorf("Model: want %q, got %q", want.Model, got.Model)
	}
}

func TestConfigStore_APIKeyNotPlaintext(t *testing.T) {
	s := makeTestStore(t)
	secret := "my-super-secret-key"
	_ = s.Save(domain.GeneratorConfig{APIKey: secret, APIEndpoint: "x", Model: "y"})

	data, _ := os.ReadFile(s.path)
	raw := string(data)

	// The plaintext key must NOT appear directly in the JSON file
	if contains(raw, secret) {
		t.Errorf("API key appears in plaintext in config file — should be base64 encoded")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub ||
		len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
