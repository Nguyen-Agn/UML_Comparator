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

func logToFile(message string) {
	f, err := os.OpenFile("debug_log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	f.WriteString(fmt.Sprintf("[%s] %s\n", runtime.GOOS, message))
}

func GetHybridMatcher() (domain.IHybridMatcher, error) {
	filename := findFileAI()
	if filename == "" {
		logToFile("AI file not found, falling back to Levenshtein")
		// Fallback an toàn: Trả về Levenshtein thuần túy thay vì nil
		fallback := &HybridMatcher{
			levenshtein:    NewLevenshteinMatcher(),
			semantic:       nil,
			levenThreshold: 0.8,
		}
		return fallback, nil
	}

	logToFile(fmt.Sprintf("Found AI file: %s. Attempting to load...", filename))
	similar_component, err := NewHybridMatcher("./" + filename)
	if err != nil {
		logToFile(fmt.Sprintf("FAILED to load AI model: %v", err))
		return nil, fmt.Errorf("fail to load model: %v", err)
	}

	logToFile("SUCCESS: AI model loaded successfully")
	return similar_component, nil
}
