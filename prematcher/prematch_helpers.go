package prematcher

import (
	"regexp"
	"strings"
)

// cleanText decodes common HTML entities left by drawio and trims whitespace.
// It does NOT strip <...> to preserve UML generics like List<String>.
func cleanText(text string) string {
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	return strings.TrimSpace(text)
}

// cleanMemberString removes UML annotations like {static} and keyword modifiers
// so that the structural regex can run cleanly on the remaining string.
func cleanMemberString(s string) string {
	// 1. Remove all { ... } annotations
	reAnn := regexp.MustCompile(`\{[^}]*\}`)
	res := reAnn.ReplaceAllString(s, "")

	// 2. Remove modifier keywords
	keywords := []string{"static", "final", "const", "abstract"}
	for _, kw := range keywords {
		re := regexp.MustCompile("(?i)\\b" + kw + "\\b")
		res = re.ReplaceAllString(res, "")
	}
	return strings.TrimSpace(res)
}

// isScopeChar returns true if the byte is a UML visibility scope character.
func isScopeChar(c byte) bool {
	return c == '+' || c == '-' || c == '#' || c == '~'
}

// isPureShortcut returns true if the raw member string is a standalone
// getter/setter instruction (e.g. "getter", "getters and setters").
func isPureShortcut(s string) bool {
	s = strings.TrimLeft(strings.TrimSpace(s), "+-#~ ")
	clean := strings.ReplaceAll(strings.ToLower(cleanMemberString(s)), "/", " ")
	tokens := strings.Fields(clean)
	if len(tokens) == 0 {
		return false
	}
	for _, t := range tokens {
		if t != "getter" && t != "setter" && t != "getters" && t != "setters" && t != "and" {
			return false
		}
	}
	return true
}

// fuzzySimilarity returns a similarity score [0,1] between two strings
// using Levenshtein distance, case-insensitive.
func fuzzySimilarity(s1, s2 string) float64 {
	s1 = strings.ToLower(strings.TrimSpace(s1))
	s2 = strings.ToLower(strings.TrimSpace(s2))
	if s1 == s2 {
		return 1.0
	}
	if len(s1) == 0 || len(s2) == 0 {
		return 0.0
	}
	dist := levenshteinDistance(s1, s2)
	maxLen := len(s1)
	if len(s2) > maxLen {
		maxLen = len(s2)
	}
	return 1.0 - float64(dist)/float64(maxLen)
}

// levenshteinDistance computes the edit distance between two strings.
func levenshteinDistance(s1, s2 string) int {
	m := len(s1)
	n := len(s2)
	d := make([][]int, m+1)
	for i := range d {
		d[i] = make([]int, n+1)
		d[i][0] = i
	}
	for j := 0; j <= n; j++ {
		d[0][j] = j
	}
	for j := 1; j <= n; j++ {
		for i := 1; i <= m; i++ {
			if s1[i-1] == s2[j-1] {
				d[i][j] = d[i-1][j-1]
			} else {
				d[i][j] = minInt(d[i-1][j]+1, minInt(d[i][j-1]+1, d[i-1][j-1]+1))
			}
		}
	}
	return d[m][n]
}

// minInt returns the smaller of two ints.
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// minU32 returns the smaller of two uint32 values.
func minU32(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}

// splitOR splits a string on the '|' character and trims whitespace from each token.
// Returns a slice of at least 1 element (the trimmed input itself if no '|' found).
// Empty tokens after trimming are skipped.
//
// Example: "int | long" -> ["int", "long"]
// Example: "void"       -> ["void"]
func splitOR(s string) []string {
	parts := strings.Split(s, "|")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		return []string{strings.TrimSpace(s)}
	}
	return result
}

// splitParams splits a parameter string on commas while respecting angle brackets
// (generics like "Map<String, int>" are not split at the inner comma).
func splitParams(paramStr string) []string {
	var parts []string
	depth := 0
	start := 0
	for i, ch := range paramStr {
		switch ch {
		case '<':
			depth++
		case '>':
			depth--
		case ',':
			if depth == 0 {
				parts = append(parts, paramStr[start:i])
				start = i + 1
			}
		}
	}
	parts = append(parts, paramStr[start:])
	return parts
}
