package similarity

import (
	"fmt"
	"os"
	"path/filepath"
	"uml_compare/domain"
)

func findFileAI() string {
	filename := "minilm.ai"
	//check if file exists and is executable
	if _, err := os.Stat(filename); err != nil {
		// find file extends with .ai
		files, _ := os.ReadDir(".")

		for _, file := range files {
			if file.IsDir() {
				continue
			}
			if filepath.Ext(file.Name()) == ".ai" {
				return file.Name()
			}
		}

		return ""
	}
	return filename
}

func GetHybridMatcher() (domain.IHybridMatcher, error) {
	filename := findFileAI()
	if filename == "" {
		// Fallback an toàn: Trả về Levenshtein thuần túy thay vì nil
		fallback := &HybridMatcher{
			levenshtein:    NewLevenshteinMatcher(),
			semantic:       nil,
			levenThreshold: 0.8,
		}
		return fallback, nil
	}

	similar_component, err := NewHybridMatcher("./" + filename)
	if err != nil {
		return nil, fmt.Errorf("fail to load model: %v", err)
	}

	return similar_component, nil
}
