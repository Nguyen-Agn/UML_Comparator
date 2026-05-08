package uml_generator

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"uml_compare/domain"
)

const configFileName = ".uml_generator_config.json"

// configStore handles persistence of GeneratorConfig to a local JSON file.
// API keys are base64-encoded to avoid plaintext storage.
type configStore struct {
	path string // absolute path to the config file
}

// storedConfig is the on-disk representation (API key is base64-encoded).
type storedConfig struct {
	APIKeyB64   string `json:"api_key_b64"`
	APIEndpoint string `json:"api_endpoint"`
	Model       string `json:"model"`
}

// newConfigStore creates a configStore pointing to ~/.uml_generator_config.json.
func newConfigStore() (*configStore, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("config_store: get home dir: %w", err)
	}
	return &configStore{path: filepath.Join(home, configFileName)}, nil
}

// Load reads the config from disk.
// If the file does not exist, default config is returned without error.
func (s *configStore) Load() (domain.GeneratorConfig, error) {
	def := defaultConfig()
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return def, nil
	}
	if err != nil {
		return def, fmt.Errorf("config_store: read file: %w", err)
	}

	var sc storedConfig
	if err := json.Unmarshal(data, &sc); err != nil {
		return def, fmt.Errorf("config_store: unmarshal: %w", err)
	}

	apiKey, err := base64.StdEncoding.DecodeString(sc.APIKeyB64)
	if err != nil {
		// Tolerate corrupt key — reset to empty
		apiKey = []byte{}
	}

	return domain.GeneratorConfig{
		APIKey:      strings.TrimSpace(string(apiKey)),
		APIEndpoint: strings.TrimSpace(orDefault(sc.APIEndpoint, def.APIEndpoint)),
		Model:       strings.TrimSpace(orDefault(sc.Model, def.Model)),
	}, nil
}

// Save persists the config to disk (base64-encodes the API key).
func (s *configStore) Save(cfg domain.GeneratorConfig) error {
	sc := storedConfig{
		APIKeyB64:   base64.StdEncoding.EncodeToString([]byte(strings.TrimSpace(cfg.APIKey))),
		APIEndpoint: strings.TrimSpace(cfg.APIEndpoint),
		Model:       strings.TrimSpace(cfg.Model),
	}
	data, err := json.MarshalIndent(sc, "", "  ")
	if err != nil {
		return fmt.Errorf("config_store: marshal: %w", err)
	}
	if err := os.WriteFile(s.path, data, 0600); err != nil {
		return fmt.Errorf("config_store: write file: %w", err)
	}
	return nil
}

func defaultConfig() domain.GeneratorConfig {
	return domain.GeneratorConfig{
		APIKey:      "",
		APIEndpoint: "https://api.openai.com/v1",
		Model:       "gpt-4o-mini",
	}
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
