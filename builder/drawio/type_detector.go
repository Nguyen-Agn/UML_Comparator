package drawio

import "strings"

// ─────────────────────────────────────────────────────────────────────────────
// typeDetector — SRP: node type and relation type classification only.
// OCP: new UML types can be added without touching other concerns.
// ─────────────────────────────────────────────────────────────────────────────

type typeDetector struct {
	san ITextSanitizer // depends on abstraction, not *htmlSanitizer (DIP)
}

// Compile-time interface satisfaction check.
var _ ITypeDetector = (*typeDetector)(nil)

// nodeType returns a normalized UML node type string based on the Draw.io
// style attribute and optionally the sanitized cell value text.
//
// Detection priority:
//  1. Stereotype keyword in value text (<<interface>>, <<abstract>>, <<enum>>)
//  2. Style attribute keywords (shape=actor, ellipse, enumeration, swimlane, …)
func (d *typeDetector) nodeType(style, sanitizedValue string) string {
	v := strings.ToLower(sanitizedValue)
	s := strings.ToLower(style)

	// Priority 1: stereotype in value text
	switch {
	case strings.Contains(v, "<<interface>>") || strings.Contains(v, "«interface»") || strings.Contains(v, "<interface>"):
		return "Interface"
	case strings.Contains(v, "<<abstract>>") || strings.Contains(v, "«abstract»"):
		return "Abstract"
	case strings.Contains(v, "<<enum>>") || strings.Contains(v, "«enum»") ||
		strings.Contains(v, "<<enumeration>>") || strings.Contains(v, "«enumeration»"):
		return "Enum"
	case strings.Contains(v, "<i>") || strings.Contains(s, "fontstyle=2") || strings.Contains(s, "fontstyle=3"):
		// Italic often represents Abstract or Interface in Draw.io UML.
		if strings.Contains(s, "ellipse") || strings.Contains(s, "shape=mxgraph.uml2") {
			return "Interface"
		}
		return "Abstract"
	}

	// Priority 2: style attribute keywords
	switch {
	case strings.Contains(s, "shape=actor"):
		return "Actor"
	case strings.Contains(s, "ellipse"):
		return "Interface"
	case strings.Contains(s, "enumeration"):
		return "Enum"
	case strings.Contains(s, "umlclass") || strings.Contains(s, "swimlane"):
		return "Class"
	case strings.Contains(s, "shape=mxgraph.uml2"):
		return "Interface"
	default:
		if style != "" {
			return "Class"
		}
		return ""
	}
}

// relationType maps a Draw.io edge style string to a normalized UML
// relation type (Inheritance, Realization, Association, Composition, etc.).
func (d *typeDetector) relationType(style string) string {
	s := strings.ToLower(style)
	isComposition := (strings.Contains(s, "endarrow=diamond") && strings.Contains(s, "endfill=1")) ||
		(strings.Contains(s, "startarrow=diamond") && strings.Contains(s, "startfill=1"))
	isAggregation := strings.Contains(s, "endarrow=diamond") || strings.Contains(s, "startarrow=diamond")
	isRealization := (strings.Contains(s, "endarrow=block") || strings.Contains(s, "startarrow=block")) && strings.Contains(s, "dashed=1")
	isInheritance := strings.Contains(s, "endarrow=block") || strings.Contains(s, "startarrow=block")

	switch {
	case isRealization:
		return "Realization"
	case isInheritance:
		return "Inheritance"
	case isComposition:
		return "Composition"
	case isAggregation:
		return "Aggregation"
	case strings.Contains(s, "endarrow=open") || strings.Contains(s, "endarrow=none") ||
		strings.Contains(s, "startarrow=open") || strings.Contains(s, "startarrow=none"):
		return "Association"
	case strings.Contains(s, "dashed=1"):
		return "Dependency"
	default:
		return "Association"
	}
}

// isEnumByChildPattern heuristically detects Draw.io Enum nodes that carry
// NO <<enum>> stereotype. Returns true when ALL non-separator, non-empty
// children look like enum constants:
//   - No visibility marker at start (+/-/#/~)
//   - No parentheses (would indicate a method)
//   - No colon (would indicate a typed attribute)
//
// NOTE: portConstraint=eastwest is present on ALL Draw.io UML child cells,
// so it is NOT a valid discriminator and must not be used here.
func (d *typeDetector) isEnumByChildPattern(children []mxCell) bool {
	if len(children) == 0 {
		return false
	}
	relevant, enumLike := 0, 0
	for _, c := range children {
		if c.Vertex != "1" || strings.Contains(c.Style, "line;") {
			continue // skip edges and divider lines
		}
		v := strings.TrimSpace(d.san.clean(c.Value))
		if v == "" {
			continue
		}
		relevant++
		if d.looksLikeEnumConstant(v) {
			enumLike++
		}
	}
	return relevant > 0 && enumLike == relevant
}

// looksLikeEnumConstant returns true when every non-empty line in the text
// block has no visibility marker, no parentheses, and no colon-type separator.
func (d *typeDetector) looksLikeEnumConstant(text string) bool {
	for _, line := range strings.Split(text, "\n") {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}
		if len(t) > 0 && (t[0] == '+' || t[0] == '-' || t[0] == '#' || t[0] == '~') {
			return false
		}
		if strings.ContainsAny(t, "()") {
			return false
		}
		if strings.Contains(t, ":") {
			return false
		}
	}
	return true
}
