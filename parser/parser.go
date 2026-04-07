package parser

import (
	"fmt"
	"path/filepath"
	"uml_compare/domain"
)

// IFileParser defines the contract for reading/parsing UML files (e.g., Draw.io, Lucid).
type IFileParser interface {
	// Parse reads a file from the given path and returns the raw model data string.
	Parse(filePath string) (domain.RawModelData, error)
}

func GetParser(filePath string) (IFileParser, error) {
	ext := filepath.Ext(filePath)
	switch ext {
	case ".drawio":
		return NewDrawioParser(), nil
	case ".solution":
		return NewSolutionParserDefault(), nil
	default:
		return nil, fmt.Errorf("parser: no parser registered for extension %q", ext)
	}
}
