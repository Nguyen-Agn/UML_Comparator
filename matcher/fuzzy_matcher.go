package matcher

import (
	"strings"
)

// IFuzzyMatcher calculates the similarity score between two strings.
// The score is generalized between 0.0 (completely different) and 1.0 (exact match).
type IFuzzyMatcher interface {
	Compare(s1, s2 string) float64
}

// LevenshteinMatcher implements IFuzzyMatcher using Levenshtein distance.
type LevenshteinMatcher struct{}

var _ IFuzzyMatcher = (*LevenshteinMatcher)(nil)

func NewLevenshteinMatcher() *LevenshteinMatcher {
	return &LevenshteinMatcher{}
}

func (l *LevenshteinMatcher) Compare(s1, s2 string) float64 {
	s1 = strings.ToLower(strings.TrimSpace(s1))
	s2 = strings.ToLower(strings.TrimSpace(s2))

	if s1 == s2 {
		return 1.0
	}
	if len(s1) == 0 || len(s2) == 0 {
		return 0.0
	}

	dist := float64(levenshtein(s1, s2))
	maxLen := float64(len(s1))
	if float64(len(s2)) > maxLen {
		maxLen = float64(len(s2))
	}

	similarity := 1.0 - (dist / maxLen)
	if similarity < 0.0 {
		similarity = 0.0
	}
	return similarity
}

func levenshtein(s1, s2 string) int {
	len1 := len(s1)
	len2 := len(s2)

	// dp[i][j] gives the distance between s1[0:i] and s2[0:j]
	dp := make([][]int, len1+1)
	for i := range dp {
		dp[i] = make([]int, len2+1)
	}

	for i := 0; i <= len1; i++ {
		dp[i][0] = i
	}
	for j := 0; j <= len2; j++ {
		dp[0][j] = j
	}

	for i := 1; i <= len1; i++ {
		for j := 1; j <= len2; j++ {
			cost := 1
			if s1[i-1] == s2[j-1] {
				cost = 0
			}
			
			del := dp[i-1][j] + 1
			ins := dp[i][j-1] + 1
			sub := dp[i-1][j-1] + cost

			dp[i][j] = min(del, min(ins, sub))
		}
	}

	return dp[len1][len2]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
