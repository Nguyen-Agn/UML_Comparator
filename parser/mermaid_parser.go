package parser

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"uml_compare/domain"
)

// MermaidParser implements IFileParser for Mermaid class diagrams (.mmd, .mermaid).
type MermaidParser struct{}

// Compile-time interface guarantee.
var _ IFileParser = (*MermaidParser)(nil)

// NewMermaidParser creates a new instance of MermaidParser.
func NewMermaidParser() IFileParser {
	return &MermaidParser{}
}

// Parse reads a Mermaid file, filters out comments and empty lines,
// and returns the cleaned content and the type "mermaid".
func (p *MermaidParser) Parse(filePath string) (domain.RawModelData, string, error) {
	if filePath == "" {
		return "", "", fmt.Errorf("MermaidParser.Parse: filePath cannot be empty")
	}

	file, err := os.Open(filePath)
	if err != nil {
		return "", "", fmt.Errorf("MermaidParser.Parse: open file: %w", err)
	}
	defer file.Close()

	var cleanedLines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// Filter unnecessary information:
		// 1. Skip empty lines.
		// 2. Skip Mermaid comments (starting with %%).
		// 3. Skip header "classDiagram" as it's redundant for the builder.
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "%%") || strings.EqualFold(trimmedLine, "classDiagram") {
			continue
		}

		cleanedLines = append(cleanedLines, line)
	}

	if err := scanner.Err(); err != nil {
		return "", "", fmt.Errorf("MermaidParser.Parse: scan file: %w", err)
	}

	cleanedContent := strings.Join(cleanedLines, "\n")
	return domain.RawModelData(cleanedContent), "mermaid", nil
}
