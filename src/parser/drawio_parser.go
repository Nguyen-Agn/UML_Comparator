package parser

import (
	"compress/flate"
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"uml_compare/domain"
)

// DrawioParser is a stateless concrete implementation of IFileParser.
// It handles both compressed (Base64+Deflate) and plain XML Draw.io files.
type DrawioParser struct{}

// Compile-time interface guarantee (from template.md)
var _ IFileParser = (*DrawioParser)(nil)

// NewDrawioParser creates a new parser that implements IFileParser.
func NewDrawioParser() IFileParser {
	return &DrawioParser{}
}

// Parse reads a .drawio file at the given path and returns:
// 1. The raw mxGraphModel XML string (cleaned).
// 2. The source type "drawio".
// 3. Error if parsing fails.
// It handles both compressed and uncompressed formats transparently.
func (p *DrawioParser) Parse(filePath string) (domain.RawModelData, string, error) {
	if filePath == "" {
		return "", "", fmt.Errorf("DrawioParser.Parse: filePath cannot be empty")
	}

	content, err := p.readFileContent(filePath)
	if err != nil {
		return "", "", fmt.Errorf("DrawioParser.Parse: failed to read file: %w", err)
	}

	// Extract the raw payload inside <diagram>...</diagram>
	payload, err := p.extractDiagramPayload(content)
	if err != nil {
		return "", "", fmt.Errorf("DrawioParser.Parse: failed to extract diagram payload: %w", err)
	}

	// Detect if compressed (Base64 content does NOT start with '<')
	if p.isCompressed(payload) {
		xmlStr, err := p.decodeBase64Deflate(payload)
		if err != nil {
			return "", "", fmt.Errorf("DrawioParser.Parse: failed to decompress diagram: %w", err)
		}
		return domain.RawModelData(xmlStr), "drawio", nil
	}

	// Plain uncompressed XML — return the payload itself
	return domain.RawModelData(payload), "drawio", nil
}

// ─────────────────────────────────────────────
// Private Helper Methods (SRP)
// ─────────────────────────────────────────────

// readFileContent safely reads the entire file into memory.
func (p *DrawioParser) readFileContent(filePath string) (string, error) {
	data, err := os.ReadFile(filePath) // os.ReadFile closes automatically; no leak
	if err != nil {
		return "", fmt.Errorf("os.ReadFile: %w", err)
	}
	return string(data), nil
}

// extractDiagramPayload pulls out the text content between <diagram> and </diagram> tags.
// Supports both <diagram ...>PAYLOAD</diagram> and direct <mxGraphModel ...> files.
func (p *DrawioParser) extractDiagramPayload(content string) (string, error) {
	start := strings.Index(content, "<diagram")
	if start == -1 {
		// File might be raw mxGraphModel without wrapper
		if strings.Contains(content, "<mxGraphModel") {
			return content, nil
		}
		return "", fmt.Errorf("extractDiagramPayload: no <diagram> or <mxGraphModel> tag found — unsupported format")
	}

	// Skip past the closing '>' of the opening <diagram ...> tag
	tagEnd := strings.Index(content[start:], ">")
	if tagEnd == -1 {
		return "", fmt.Errorf("extractDiagramPayload: malformed <diagram> opening tag")
	}
	payloadStart := start + tagEnd + 1

	end := strings.Index(content, "</diagram>")
	if end == -1 {
		return "", fmt.Errorf("extractDiagramPayload: missing </diagram> closing tag")
	}

	payload := strings.TrimSpace(content[payloadStart:end])
	return payload, nil
}

// isCompressed returns true when the diagram payload is Base64-encoded (compressed),
// which is detected by the fact that it does NOT start with '<' (an XML character).
func (p *DrawioParser) isCompressed(payload string) bool {
	trimmed := strings.TrimSpace(payload)
	return len(trimmed) > 0 && trimmed[0] != '<'
}

// decodeBase64Deflate decodes a Draw.io compressed payload:
//   Base64 Decode → Deflate Decompress → URL Decode → Raw mxGraphModel XML
func (p *DrawioParser) decodeBase64Deflate(payload string) (string, error) {
	// Step 1: Base64 decode
	compressed, err := base64.StdEncoding.DecodeString(strings.TrimSpace(payload))
	if err != nil {
		return "", fmt.Errorf("base64 decode: %w", err)
	}

	// Step 2: Inflate (raw Deflate, no zlib header — Draw.io uses flate.NewReader)
	reader := flate.NewReader(strings.NewReader(string(compressed)))
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("flate decompress: %w", err)
	}

	// Step 3: URL decode
	decoded, err := url.QueryUnescape(string(decompressed))
	if err != nil {
		return "", fmt.Errorf("url.QueryUnescape: %w", err)
	}

	return decoded, nil
}
