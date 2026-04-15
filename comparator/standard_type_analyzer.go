package comparator

import "strings"

// StandardTypeAnalyzer implements ITypeAnalyzer with generic and type mapping support.
type StandardTypeAnalyzer struct{}

var _ ITypeAnalyzer = (*StandardTypeAnalyzer)(nil)

// NewStandardTypeAnalyzer creates a new instance of StandardTypeAnalyzer.
func NewStandardTypeAnalyzer() *StandardTypeAnalyzer {
	return &StandardTypeAnalyzer{}
}

// TranslateType converts a solution type name to its mapped student name if it exists.
func (a *StandardTypeAnalyzer) TranslateType(typeName string, typeMap map[string]string) string {
	if translated, ok := typeMap[typeName]; ok {
		return translated
	}
	return typeName
}

// CompareTypes checks if two type strings are compatible, considering generics and type mapping.
func (a *StandardTypeAnalyzer) CompareTypes(solType, stuType string, typeMap map[string]string) bool {
	solType = strings.TrimSpace(solType)
	stuType = strings.TrimSpace(stuType)

	// Base case: exactly equal after translation
	if a.TranslateType(solType, typeMap) == stuType {
		return true
	}

	// Generic case
	if strings.Contains(solType, "<") || strings.Contains(stuType, "<") {
		solOuter, solInners := a.splitGeneric(solType)
		stuOuter, stuInners := a.splitGeneric(stuType)

		// Outer check: "contains" rule (case-insensitive)
		if !a.isCompatibleOuter(solOuter, stuOuter) {
			return false
		}

		// If one of the types doesn't specify generic parameters, treat it as a match 
        // as long as the outer container matched (e.g. "List" matches "ArrayList<T>")
		if len(solInners) == 0 || len(stuInners) == 0 {
			return true
		}

		if len(solInners) != len(stuInners) {
			return false
		}

		// Recursive inner check
		for i := range solInners {
			if !a.CompareTypes(solInners[i], stuInners[i], typeMap) {
				return false
			}
		}
		return true
	}

	return false
}

// isCompatibleOuter checks if two outer container types (like List vs ArrayList) are compatible.
func (a *StandardTypeAnalyzer) isCompatibleOuter(sol, stu string) bool {
	s := strings.ToLower(sol)
	t := strings.ToLower(stu)
	if s == t {
		return true
	}
	// Rule: solution "List" matches student "ArrayList", or vice versa if they contain each other
	return strings.Contains(t, s) || strings.Contains(s, t)
}

// splitGeneric decomposes a type string into its outer container and a list of inner generic types.
func (a *StandardTypeAnalyzer) splitGeneric(t string) (string, []string) {
	idx := strings.Index(t, "<")
	if idx == -1 {
		return t, nil
	}
	outer := strings.TrimSpace(t[:idx])
	innerStr := t[idx+1:]
	if lastIdx := strings.LastIndex(innerStr, ">"); lastIdx != -1 {
		innerStr = innerStr[:lastIdx]
	}

	// Split by comma, respecting nested brackets
	var inners []string
	var current strings.Builder
	depth := 0
	for i := 0; i < len(innerStr); i++ {
		char := innerStr[i]
		if char == '<' {
			depth++
			current.WriteByte(char)
		} else if char == '>' {
			depth--
			current.WriteByte(char)
		} else if char == ',' && depth == 0 {
			inners = append(inners, strings.TrimSpace(current.String()))
			current.Reset()
		} else {
			current.WriteByte(char)
		}
	}
	if current.Len() > 0 {
		inners = append(inners, strings.TrimSpace(current.String()))
	}

	return outer, inners
}
