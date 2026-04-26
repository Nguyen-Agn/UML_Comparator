package prematcher

import (
	"regexp"
	"strconv"
	"strings"
)

// ScoreExtractor implements IScoreExtractor.
type ScoreExtractor struct {
	scoreRegex *regexp.Regexp
}

// NewScoreExtractor creates a new instance of ScoreExtractor.
func NewScoreExtractor() *ScoreExtractor {
	return &ScoreExtractor{
		scoreRegex: regexp.MustCompile(`__(\d+(?:\.\d+)?)__\s*$`),
	}
}

// ExtractScore pulls out the __d__ or __d.d__ point value from the end of a string.
func (e *ScoreExtractor) ExtractScore(raw string) (string, float64) {
	matches := e.scoreRegex.FindStringSubmatch(raw)
	if len(matches) > 0 {
		scoreStr := matches[1]
		score, err := strconv.ParseFloat(scoreStr, 64)
		if err == nil {
			cleaned := e.scoreRegex.ReplaceAllString(raw, "")
			return strings.TrimSpace(cleaned), score
		}
	}
	return raw, 1.0 // Default score is 1.0
}
