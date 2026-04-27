package similarity

import "fmt"

// HybridMatcher combines Levenshtein (typo detection) with Semantic (synonym detection).
// It implements matcher.IFuzzyMatcher so it can be a drop-in replacement.
//
// Strategy:
//   - If Levenshtein score >= threshold (default 0.8), return it immediately (typo case).
//   - Otherwise, run Semantic comparison and return max(levenshtein, semantic).
type HybridMatcher struct {
	levenshtein    IFuzzyMatcher
	semantic       ISemanticMatcher
	levenThreshold float64
}

// Compile-time check: HybridMatcher must satisfy matcher.IFuzzyMatcher
var _ IFuzzyMatcher = (*HybridMatcher)(nil)

// NewHybridMatcher creates a HybridMatcher.
// If semanticMatcher is nil, it falls back to pure Levenshtein.
func NewHybridMatcher(zipPath string) (*HybridMatcher, error) {
	//
	semanticMatcher, err := NewMiniLMSemanticMatcher(zipPath)
	if err != nil {
		fmt.Println("fail to load model: ", err)
		return nil, nil
	}

	return &HybridMatcher{
		levenshtein:    NewLevenshteinMatcher(),
		semantic:       semanticMatcher,
		levenThreshold: 0.8,
	}, nil
}

// Compare returns the best similarity score between two strings,
// using Levenshtein for typos and Semantic for synonyms.
func (h *HybridMatcher) Compare(s1, s2 string) float64 {
	// Step 1: Fast Levenshtein check
	// levenScore := h.levenshtein.Compare(s1, s2)

	// // If Levenshtein already says they're similar enough, it's a typo — no need for ML.
	// if levenScore >= h.levenThreshold {
	// 	return levenScore
	// }

	// // Step 2: If semantic matcher is available, check for synonym
	// if h.semantic != nil {
	// 	semanticScore := h.semantic.Compare(s1, s2)
	// 	if semanticScore > levenScore {
	// 		return semanticScore
	// 	}
	// }

	if h.semantic != nil {
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

// Close releases resources held by the semantic matcher.
func (h *HybridMatcher) Close() error {
	if h.semantic != nil {
		return h.semantic.Close()
	}
	return nil
}
