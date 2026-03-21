package builder

import (
	"regexp"
	"strings"
)

// ─────────────────────────────────────────────────────────────────────────────
// htmlSanitizer — SRP: text cleaning only.
// DIP: all regexp patterns compiled ONCE at construction, not per call.
// ─────────────────────────────────────────────────────────────────────────────

type htmlSanitizer struct {
	tagRe     *regexp.Regexp // matches any HTML tag: <...>
	newlineRe *regexp.Regexp // collapses 2+ consecutive newlines
	stereoRe  *regexp.Regexp // matches <<...>> stereotype tokens
	spaceRe   *regexp.Regexp // collapses 2+ whitespace chars (for signatures)
}

// Compile-time interface satisfaction check.
var _ ITextSanitizer = (*htmlSanitizer)(nil)

// newHTMLSanitizer constructs a sanitizer with all regexes pre-compiled.
func newHTMLSanitizer() *htmlSanitizer {
	return &htmlSanitizer{
		tagRe:     regexp.MustCompile(`<[^>]+>`),
		newlineRe: regexp.MustCompile(`\n{2,}`),
		stereoRe:  regexp.MustCompile(`<<[^>]+>>`),
		spaceRe:   regexp.MustCompile(`\s{2,}`),
	}
}

// clean decodes HTML entities and strips all HTML tags from a raw Draw.io
// cell value string, returning a clean multi-line plain-text string.
//
// Decode order matters:
//  1. Double-encoded entities (&amp;lt; → &lt; → <)
//  2. Numeric newline entities (&#10; &#13; &#xA;) — BEFORE tag stripping
//  3. Strip <<...>> stereotype tokens — BEFORE HTML tag regex to prevent tearing
//     (the tag regex <[^>]+> would match <interface> inside <<interface>>, splitting it)
//  4. Strip HTML tags (replace with \n to preserve line structure)
//  5. Named entities (&lt; &gt; &amp; &nbsp; &quot;)
//  6. Collapse consecutive newlines
func (s *htmlSanitizer) clean(raw string) string {
	// Step 1: double-encoded entities
	raw = strings.ReplaceAll(raw, "&amp;lt;", "&lt;")
	raw = strings.ReplaceAll(raw, "&amp;gt;", "&gt;")
	raw = strings.ReplaceAll(raw, "&amp;amp;", "&amp;")

	// Step 2: numeric newline entities (must be BEFORE tag stripping)
	raw = strings.ReplaceAll(raw, "&#10;", "\n")
	raw = strings.ReplaceAll(raw, "&#13;", "\r")
	raw = strings.ReplaceAll(raw, "&#xA;", "\n")
	raw = strings.ReplaceAll(raw, "&#xa;", "\n")

	// Step 2b: decode &lt; / &gt; now so we can see << >> stereotypes
	//          (necessary before stereoRe can match actual < > chars)
	raw = strings.ReplaceAll(raw, "&lt;", "<")
	raw = strings.ReplaceAll(raw, "&gt;", ">")
	raw = strings.ReplaceAll(raw, "&amp;", "&")
	raw = strings.ReplaceAll(raw, "&nbsp;", " ")
	raw = strings.ReplaceAll(raw, "&quot;", "\"")

	// Step 3: protect << >> stereotypes BEFORE HTML tag stripping.
	// Replace <<Foo>> with a placeholder, strip tags, then restore.
	// Simpler: just strip stereotypes outright here — extractCleanName also strips,
	// but clean() must not TEAR them with the tag regex.
	raw = s.stereoRe.ReplaceAllString(raw, "\n")

	// Step 4: strip HTML tags (replace with newline)
	clean := s.tagRe.ReplaceAllString(raw, "\n")
	clean = strings.ReplaceAll(clean, "\r\n", "\n")
	clean = strings.ReplaceAll(clean, "\r", "\n")

	// Step 6: collapse consecutive newlines
	clean = s.newlineRe.ReplaceAllString(clean, "\n")
	return strings.TrimSpace(clean)
}

// decodeOnly decodes all HTML entities (including double-encoded) but does NOT
// strip HTML tags or stereotype tokens. Used solely for type detection so that
// typeDetector.nodeType() can see the raw << >> stereotype text before clean()
// removes it during tag-safe stripping.
func (s *htmlSanitizer) decodeOnly(raw string) string {
	// Double-encoded
	raw = strings.ReplaceAll(raw, "&amp;lt;", "&lt;")
	raw = strings.ReplaceAll(raw, "&amp;gt;", "&gt;")
	raw = strings.ReplaceAll(raw, "&amp;amp;", "&amp;")
	// Numeric newlines
	raw = strings.ReplaceAll(raw, "&#10;", "\n")
	raw = strings.ReplaceAll(raw, "&#13;", "\r")
	raw = strings.ReplaceAll(raw, "&#xA;", "\n")
	raw = strings.ReplaceAll(raw, "&#xa;", "\n")
	// Named entities
	raw = strings.ReplaceAll(raw, "&lt;", "<")
	raw = strings.ReplaceAll(raw, "&gt;", ">")
	raw = strings.ReplaceAll(raw, "&amp;", "&")
	raw = strings.ReplaceAll(raw, "&nbsp;", " ")
	raw = strings.ReplaceAll(raw, "&quot;", "\"")
	return raw
}

// extractCleanName returns the first non-empty, non-stereotype line from a
// sanitized multi-line string — i.e., strips <<interface>>, <<abstract>>,
// <<enum>> tokens whether standalone or inline with the class name.
func (s *htmlSanitizer) extractCleanName(sanitized string) string {
	for _, line := range strings.Split(sanitized, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		// Strip all <<...>> tokens anywhere in the line
		clean := strings.TrimSpace(s.stereoRe.ReplaceAllString(trimmed, ""))
		if clean == "" {
			continue // entire line was stereotype annotation
		}
		return clean
	}
	return sanitized // fallback: return as-is
}

// normalizeSignature collapses redundant internal whitespace in a method
// signature that was joined from multiple continuation lines.
func (s *htmlSanitizer) normalizeSignature(sig string) string {
	return s.spaceRe.ReplaceAllString(strings.TrimSpace(sig), " ")
}
