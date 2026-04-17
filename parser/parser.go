package parser

import (
	"uml_compare/domain"
)

// IFileParser defines the contract for reading/parsing UML files (e.g., Draw.io, Mermaid).
type IFileParser interface {
	// Parse reads a file from the given path and returns:
	// 1. Cleaned raw model data string.
	// 2. Detected source type (e.g., "drawio", "mermaid").
	// 3. Error if parsing fails.
	Parse(filePath string) (domain.RawModelData, string, error)
}

// GetParser is a factory function that returns the default AutoParser
// initialized with all supported formats. This implements the Strategy Pattern
// for selecting the correct parser based on file extension.
func GetParser(filePath string) (IFileParser, error) {
	return NewAutoParserDefault(), nil
}
