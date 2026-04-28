package similarity

import (
	"fmt"
	"reflect"
	"uml_compare/domain"
)

// HybridMatcher combines Levenshtein (typo detection) with Semantic (synonym detection).
// It implements matcher.IFuzzyMatcher so it can be a drop-in replacement.
//
// Strategy:
//   - If Levenshtein score >= threshold (default 0.8), return it immediately (typo case).
//   - Otherwise, run Semantic comparison and return max(levenshtein, semantic).
type HybridMatcher struct {
	levenshtein IFuzzyMatcher
	semantic    ISemanticMatcher
	config      domain.IAppConfig
}

// Compile-time check: HybridMatcher must satisfy matcher.IFuzzyMatcher
var _ IFuzzyMatcher = (*HybridMatcher)(nil)

// NewHybridMatcher creates a HybridMatcher.
// If semanticMatcher is nil, it falls back to pure Levenshtein.
func NewHybridMatcher(zipPath string, cfg domain.IAppConfig) (*HybridMatcher, error) {
	semanticMatcher, err := NewMiniLMSemanticMatcher(zipPath)
	if err != nil {
		fmt.Println("fail to load model: ", err)
		return &HybridMatcher{
			levenshtein: NewLevenshteinMatcher(),
			semantic:    nil,
			config:      cfg,
		}, nil
	}

	return &HybridMatcher{
		levenshtein: NewLevenshteinMatcher(),
		semantic:    semanticMatcher,
		config:      cfg,
	}, nil
}

// Compare returns the best similarity score between two strings,
// using Levenshtein for typos and Semantic for synonyms.
func (h *HybridMatcher) Compare(s1, s2 string) float64 {
	// If AI is disabled via config, fallback to pure Levenshtein
	if h.semantic != nil && h.config.UseAI() {
		return h.semantic.Compare(s1, s2)
	}

	return h.levenshtein.Compare(s1, s2)
}

func (h *HybridMatcher) CompareMultiple(candidate string, optionals []string) (float64, string) {
	// temp empty
	return 0, ""
}

func (h *HybridMatcher) CompareAttribute(s1, s2 string) float64 {
	return h.Compare(s1, s2)
}

func (h *HybridMatcher) CompareMethod(s1, s2 string) float64 {
	return h.Compare(s1, s2)
}

func (h *HybridMatcher) CompareField(s1, s2 string) float64 {
	return h.Compare(s1, s2)
}

func (h *HybridMatcher) GetThreshold() float64 {
	return h.config.GetThreshold()
}

func (h *HybridMatcher) IsAIAvailable() bool {
	return h.semantic != nil
}

// Close releases resources held by the semantic matcher.
func (h *HybridMatcher) Close() error {
	if h.semantic != nil {
		v := reflect.ValueOf(h.semantic)
		if v.Kind() == reflect.Ptr && v.IsNil() {
			return nil
		}
		return h.semantic.Close()
	}
	return nil
}
