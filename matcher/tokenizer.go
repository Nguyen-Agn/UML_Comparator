package matcher

import (
	"regexp"
	"strings"
)

var (
	// Matches any character that is not a letter or number
	nonAlphaNumReg = regexp.MustCompile(`[^a-zA-Z0-9]+`)
	
	// Matches lowercase followed by an uppercase letter (camel/Pascal case transition)
	camelCaseReg = regexp.MustCompile(`([a-z])([A-Z])`)
	
	// Matches sequence of uppercase letters followed by a CamelCase word (e.g. XMLParser -> XML Parser)
	acronymReg = regexp.MustCompile(`([A-Z]+)([A-Z][a-z])`)

	// Split between letters and numbers
	letterToNum = regexp.MustCompile(`([a-zA-Z])([0-9]+)`)
	numToLetter = regexp.MustCompile(`([0-9]+)([a-zA-Z])`)
)

// TokenizeIdentifier normalizes strings like PascalCase, camelCase, snake_case into 
// a slice of lowercase tokens to facilitate word-by-word semantic analysis.
func TokenizeIdentifier(identifier string) []string {
	if identifier == "" {
		return []string{}
	}

	// 1. Replace non-alphanumeric characters with spaces
	cleanStr := nonAlphaNumReg.ReplaceAllString(identifier, " ")

	// 2. Insert space between lowercase and next uppercase
	splitCamel := camelCaseReg.ReplaceAllString(cleanStr, "${1} ${2}")

	// 3. Handle acronyms before camelCase
	splitAcronym := acronymReg.ReplaceAllString(splitCamel, "${1} ${2}")

	// 4. Split letters from numbers
	splitNum1 := letterToNum.ReplaceAllString(splitAcronym, "${1} ${2}")
	splitNum2 := numToLetter.ReplaceAllString(splitNum1, "${1} ${2}")

	// 5. Lowercase everything and split by whitespace
	finalStr := strings.ToLower(splitNum2)
	tokens := strings.Fields(finalStr)

	if tokens == nil {
		return []string{}
	}

	return tokens
}
