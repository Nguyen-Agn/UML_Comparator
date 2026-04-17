package drawio

import (
	"regexp"
	"strings"
)

// ─────────────────────────────────────────────────────────────────────────────
// htmlSanitizer — SRP: text cleaning only.
// DIP: all regexp patterns compiled ONCE at construction, not per call.
// ─────────────────────────────────────────────────────────────────────────────

type htmlSanitizer struct {
	tagStructuralRe *regexp.Regexp // matches structural tags: <br>, <div>, etc. (replace with \n)
	tagStylingRe    *regexp.Regexp // matches styling tags: <i>, <b>, etc. (replace with "")
	newlineRe       *regexp.Regexp // collapses 2+ consecutive newlines
	stereoRe        *regexp.Regexp // matches <<...>> stereotype tokens
	spaceRe         *regexp.Regexp // collapses 2+ whitespace chars (for signatures)
}

// Compile-time interface satisfaction check.
var _ ITextSanitizer = (*htmlSanitizer)(nil)

// newHTMLSanitizer constructs a sanitizer with all regexes pre-compiled.
func newHTMLSanitizer() *htmlSanitizer {
	return &htmlSanitizer{
		// Structural tags that represent line breaks or container boundaries.
		tagStructuralRe: regexp.MustCompile(`(?i)</?(br|div|p|ul|li|ol|table|tr|td|thead|tbody|tfoot|hr|mxCell|mxGraphModel|root)\b[^>]*>`),
		// Styling tags that should not cause line breaks.
		tagStylingRe: regexp.MustCompile(`(?i)</?(b|i|u|font|span|strong|em|small|big|sub|sup)\b[^>]*>`),
		newlineRe:    regexp.MustCompile(`\n{2,}`),
		stereoRe:     regexp.MustCompile(`(<<[^>]+>>|«[^»]+»)`),
		spaceRe:      regexp.MustCompile(`\s{2,}`),
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
	raw = s.stereoRe.ReplaceAllString(raw, "\n")

	// Step 3.5: Translate UML semantic styling tags (<i>, <u>) to keywords
	// <i> -> italic -> abstract in UML
	// <u> -> underline -> static in UML
	raw = regexp.MustCompile(`(?i)<i\b[^>]*>`).ReplaceAllString(raw, " {abstract} ")
	raw = regexp.MustCompile(`(?i)<u\b[^>]*>`).ReplaceAllString(raw, " {static} ")

	// Step 4: strip HTML tags
	// Structural tags -> newline; Styling tags -> nothing (prevents name fragmentation)
	clean := s.tagStructuralRe.ReplaceAllString(raw, "\n")
	clean = s.tagStylingRe.ReplaceAllString(clean, "")

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

// extractNameAndFormat takes the raw decoded string (which still has HTML tags),
// isolates the first non-empty line representing the class name, checks if it is
// wrapped in bold tags (<b> or <strong>), and returns the clean name along with its bold status.
func (s *htmlSanitizer) extractNameAndFormat(rawDecoded string) (string, bool) {
	// Step 3.5 equivalent: Translate UML semantic styling tags (<i>, <u>) to keywords
	// before structural/styling tags are stripped so they appear in the name.
	rawDecoded = regexp.MustCompile(`(?i)<i\b[^>]*>`).ReplaceAllString(rawDecoded, " {abstract} ")
	rawDecoded = regexp.MustCompile(`(?i)<u\b[^>]*>`).ReplaceAllString(rawDecoded, " {static} ")

	// Structural tags -> newline
	cleanStr := s.tagStructuralRe.ReplaceAllString(rawDecoded, "\n")

	for _, line := range strings.Split(cleanStr, "\n") {
		trimmed := strings.TrimSpace(line)

		// Strip all <<...>> tokens anywhere in the line
		stereoStripped := strings.TrimSpace(s.stereoRe.ReplaceAllString(trimmed, ""))

		pureText := s.tagStylingRe.ReplaceAllString(stereoStripped, "")
		pureText = strings.TrimSpace(pureText)

		if pureText == "" || pureText == "{abstract}" || pureText == "{static}" {
			continue // entire line was stereotype annotation or just a marker
		}

		// This line is the name line. Check if it contains bold tags (case insensitive)
		lowerLine := strings.ToLower(stereoStripped)
		isBold := strings.Contains(lowerLine, "<b>") || strings.Contains(lowerLine, "<strong>")

		return pureText, isBold
	}
	
	cleanFallback := s.tagStylingRe.ReplaceAllString(cleanStr, "")
	return strings.TrimSpace(cleanFallback), false
}

// normalizeSignature collapses redundant internal whitespace in a method
// signature that was joined from multiple continuation lines.
func (s *htmlSanitizer) normalizeSignature(sig string) string {
	return s.spaceRe.ReplaceAllString(strings.TrimSpace(sig), " ")
}
