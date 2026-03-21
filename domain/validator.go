package domain

import (
	"fmt"
	"strings"
)

// IntegrityError represents a single data integrity issue found in a UMLGraph.
type IntegrityError struct {
	Code     string // e.g., "EMPTY_GRAPH", "EMPTY_NODE_NAME", "INVALID_NODE_TYPE", "DANGLING_EDGE"
	Severity string // "ERROR" (pipeline must stop) or "WARN" (data is suspect but usable)
	Message  string
}

func (e IntegrityError) Error() string {
	return fmt.Sprintf("[%s][%s] %s", e.Severity, e.Code, e.Message)
}

// IsError returns true for pipeline-blocking issues.
func (e IntegrityError) IsError() bool { return e.Severity == "ERROR" }

// IsWarn returns true for non-blocking quality issues.
func (e IntegrityError) IsWarn() bool { return e.Severity == "WARN" }

// ValidNodeTypes is the set of accepted UMLNode.Type values.
var ValidNodeTypes = map[string]bool{
	"Class": true, "Interface": true, "Actor": true, "Abstract": true, "Enum": true,
}

// suspectNameChars are characters that indicate a corrupted / non-standard node name.
var suspectNameChars = []string{">", "<", "&", "&#", "<<", ">>"}

// ValidateGraph checks a UMLGraph for integrity (ERRORs) and UML quality (WARNs).
//
//   - ERRORs: pipeline MUST stop — data will produce meaningless comparison results.
//   - WARNs: data is usable BUT suspect — log for Grader to apply penalty or flag.
//
// Returns a slice of IntegrityErrors (empty = clean graph).
func ValidateGraph(g *UMLGraph, label string) []IntegrityError {
	var errs []IntegrityError

	// ── Level 1: Structural ERRORs ─────────────────────────────────────────────
	if g == nil || len(g.Nodes) == 0 {
		errs = append(errs, IntegrityError{
			Code:     "EMPTY_GRAPH",
			Severity: "ERROR",
			Message:  fmt.Sprintf("%s: graph has 0 nodes — empty file or builder failure", label),
		})
		return errs // no further checks possible
	}

	idSet := make(map[string]bool, len(g.Nodes))
	for _, n := range g.Nodes {
		idSet[n.ID] = true

		if n.Name == "" {
			errs = append(errs, IntegrityError{
				Code:     "EMPTY_NODE_NAME",
				Severity: "ERROR",
				Message:  fmt.Sprintf("%s: node ID=%s has empty name (swimlane/entity parse failure)", label, n.ID),
			})
		}
		if !ValidNodeTypes[n.Type] {
			errs = append(errs, IntegrityError{
				Code:     "INVALID_NODE_TYPE",
				Severity: "ERROR",
				Message:  fmt.Sprintf("%s: node '%s' has invalid Type='%s'", label, n.Name, n.Type),
			})
		}
	}

	for i, e := range g.Edges {
		if !idSet[e.SourceID] {
			errs = append(errs, IntegrityError{
				Code:     "DANGLING_EDGE_SOURCE",
				Severity: "ERROR",
				Message:  fmt.Sprintf("%s: edge[%d] SourceID='%s' not found in nodes", label, i, e.SourceID),
			})
		}
		if !idSet[e.TargetID] {
			errs = append(errs, IntegrityError{
				Code:     "DANGLING_EDGE_TARGET",
				Severity: "ERROR",
				Message:  fmt.Sprintf("%s: edge[%d] TargetID='%s' not found in nodes", label, i, e.TargetID),
			})
		}
		if e.SourceID == e.TargetID {
			errs = append(errs, IntegrityError{
				Code:     "SELF_LOOP_EDGE",
				Severity: "ERROR",
				Message:  fmt.Sprintf("%s: edge[%d] is a self-loop (src==tgt='%s')", label, i, e.SourceID),
			})
		}
	}

	// ── Level 2: UML Quality WARNings (non-standard UML patterns) ─────────────
	for _, n := range g.Nodes {
		// Suspect class name: contains HTML/entity remnants or only punctuation
		for _, ch := range suspectNameChars {
			if strings.Contains(n.Name, ch) {
				errs = append(errs, IntegrityError{
					Code:     "SUSPECT_NODE_NAME",
					Severity: "WARN",
					Message:  fmt.Sprintf("%s: node '%s' name contains suspect char '%s' (HTML decode error?)", label, n.Name, ch),
				})
				break
			}
		}
		// Very short name (1-2 chars) is likely a placeholder
		if len([]rune(n.Name)) <= 2 && n.Name != "" {
			errs = append(errs, IntegrityError{
				Code:     "TRIVIAL_NODE_NAME",
				Severity: "WARN",
				Message:  fmt.Sprintf("%s: node '%s' has a very short name (likely placeholder or error)", label, n.Name),
			})
		}

		// Incomplete attributes: attribute has no `:` (no type declared)
		// Skip for Enums as constants usually don't have types in UML
		for _, attr := range n.Attributes {
			if !strings.Contains(attr, ":") && attr != "" && !strings.EqualFold(n.Type, "Enum") {
				errs = append(errs, IntegrityError{
					Code:     "INCOMPLETE_ATTRIBUTE",
					Severity: "WARN",
					Message:  fmt.Sprintf("%s: node '%s' has attribute without type: '%s'", label, n.Name, attr),
				})
			}
		}
		// Incomplete methods: method ending with just `()` or `(` — no return type
		for _, m := range n.Methods {
			if strings.HasSuffix(strings.TrimSpace(m), "(") {
				errs = append(errs, IntegrityError{
					Code:     "INCOMPLETE_METHOD",
					Severity: "WARN",
					Message:  fmt.Sprintf("%s: node '%s' has unclosed method signature: '%s'", label, n.Name, m),
				})
			}
		}
	}

	return errs
}

// FilterErrors returns only ERROR-severity items (pipeline blockers).
func FilterErrors(errs []IntegrityError) []IntegrityError {
	var out []IntegrityError
	for _, e := range errs {
		if e.IsError() {
			out = append(out, e)
		}
	}
	return out
}

// FilterWarns returns only WARN-severity items (quality flags).
func FilterWarns(errs []IntegrityError) []IntegrityError {
	var out []IntegrityError
	for _, e := range errs {
		if e.IsWarn() {
			out = append(out, e)
		}
	}
	return out
}
