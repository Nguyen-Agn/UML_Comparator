package drawio

import (
	"strings"
)

// ─────────────────────────────────────────────────────────────────────────────
// memberParser — SRP: classify sanitized text lines into Attributes vs Methods.
// Handles multi-line method signatures (open-paren buffering).
// ─────────────────────────────────────────────────────────────────────────────

// lineKind categorizes a single line of text within a class member cell.
type lineKind int

const (
	lineSkip   lineKind = iota // dangling punctuation / lone visibility marker / return-type fragment
	lineAttr                   // attribute (typed field)
	lineMethod                 // complete method signature
)

type memberParser struct {
	san   ITextSanitizer // depends on abstraction, not *htmlSanitizer (DIP)
	style IStyleHelper
}

// Compile-time interface satisfaction check.
var _ IMemberParser = (*memberParser)(nil)

// parseChildren extracts and classifies all member lines from a list of child
// cells that belong to a single class container. Returns separate slices for
// attributes and methods.
func (m *memberParser) parseChildren(children []mxCell) (attrs, methods []string) {
	for _, child := range children {
		if child.Vertex != "1" || child.Edge == "1" {
			continue
		}

		// 1. Clean the text using HTML sanitizer
		text := strings.TrimSpace(m.san.clean(child.Value))
		if text == "" {
			continue
		}

		// 2. Detect semantic styling from Draw.io 'style' attribute bitmask
		// fontStyle bits: 1=Bold, 2=Italic (Abstract), 4=Underline (Static)
		style := child.Style
		if m.style.IsStyleBitSet(style, "fontStyle", 4) {
			if !strings.Contains(strings.ToLower(text), "{static}") {
				text += " {static}"
			}
		}
		if m.style.IsStyleBitSet(style, "fontStyle", 2) {
			if !strings.Contains(strings.ToLower(text), "{abstract}") {
				text += " {abstract}"
			}
		}

		// 3. Classify and collect
		a, mth := m.parseText(text)
		attrs = append(attrs, a...)
		methods = append(methods, mth...)
	}
	return
}

// parseText classifies all non-empty lines in a sanitized cell text block.
// Handles multi-line method signatures via open-parenthesis buffering.
func (m *memberParser) parseText(text string) (attrs, methods []string) {
	var pending string // buffer for multi-line method signature

	for _, line := range strings.Split(text, "\n") {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}

		// Continuation line for an open multi-line method signature
		if pending != "" {
			pending += " " + t
			if strings.Contains(t, ")") {
				methods = append(methods, m.san.normalizeSignature(pending))
				pending = ""
			}
			continue
		}

		switch m.classify(t) {
		case lineSkip:
			// nothing — discard
		case lineMethod:
			open := strings.Count(t, "(")
			close := strings.Count(t, ")")
			if open > close {
				pending = t // begin buffering multi-line signature
			} else {
				methods = append(methods, t)
			}
		case lineAttr:
			attrs = append(attrs, t)
		}
	}

	// Flush incomplete buffered signature (truncated Draw.io cell)
	if pending != "" {
		methods = append(methods, m.san.normalizeSignature(pending))
	}
	return
}

// classify determines how a single trimmed line should be treated.
func (m *memberParser) classify(t string) lineKind {
	// Lines with parentheses are methods (or start of multi-line method)
	if strings.Contains(t, "(") {
		return lineMethod
	}

	// Lone dangling punctuation — discard
	if t == ":" || t == ")" || t == "," {
		return lineSkip
	}

	// Return-type fragment from a split method (e.g. ": void")
	if strings.HasPrefix(t, ":") {
		return lineSkip
	}

	// Lone UML visibility marker used as section divider
	if t == "-" || t == "+" || t == "#" || t == "~" {
		return lineSkip
	}

	return lineAttr
}

