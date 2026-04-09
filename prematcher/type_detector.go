package prematcher

import "strings"

// TypeDetector implements ITypeDetector.
type TypeDetector struct{}

// NewTypeDetector creates a new instance of TypeDetector.
func NewTypeDetector() *TypeDetector {
	return &TypeDetector{}
}

// NormalizeNodeType maps various stereotype formats to a standard representation.
func (d *TypeDetector) NormalizeNodeType(t string) string {
	if d.IsEnumType(t) {
		return "Enum"
	}
	// Future-proofing: add other normalizations if needed (e.g., Interface, Abstract)
	return t
}

// IsEnumType checks if a given type string represents an enumeration (including stereotypes).
func (d *TypeDetector) IsEnumType(t string) bool {
	lower := strings.ToLower(t)
	// Match: enum, enumeration, <<enum>>, <<enumeration>>, «enum», «enumeration»
	return strings.Contains(lower, "enum") ||
		strings.Contains(lower, "enumeration") ||
		(strings.Contains(lower, "«") && strings.Contains(lower, "enu")) ||
		(strings.Contains(lower, "<<") && strings.Contains(lower, "enu"))
}
