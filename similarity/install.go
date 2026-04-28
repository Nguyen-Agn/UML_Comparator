package similarity

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"uml_compare/domain"
)

func findFileAI() string {
	var preferredFile string
	switch runtime.GOOS {
	case "windows":
		preferredFile = "minilm_win.ai"
	case "darwin":
		preferredFile = "minilm_mac.ai"
	default: // linux, freebsd, etc.
		preferredFile = "minilm_linux.ai"
	}

	// 1. Prioritize OS-specific AI file
	if _, err := os.Stat(preferredFile); err == nil {
		return preferredFile
	}

	// 2. Fallback to legacy generic name
	if _, err := os.Stat("minilm.ai"); err == nil {
		return "minilm.ai"
	}

	// 3. Fallback to any available .ai file
	files, _ := os.ReadDir(".")
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".ai" {
			return file.Name()
		}
	}

	return ""
}

func GetHybridMatcher(cfg domain.IAppConfig) (domain.IHybridMatcher, error) {
	filename := findFileAI()
	if filename == "" {
		fallback := &HybridMatcher{
			levenshtein: NewLevenshteinMatcher(),
			semantic:    nil,
			config:      cfg,
		}
		return fallback, nil
	}

	similar_component, err := NewHybridMatcher("./"+filename, cfg)
	if err != nil {

		return nil, fmt.Errorf("fail to load model: %v", err)
	}

	return similar_component, nil
}
