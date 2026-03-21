package parser

import "uml_compare/domain"

// IFileParser defines the contract for reading/parsing UML files (e.g., Draw.io, Lucid).
type IFileParser interface {
	// Parse reads a file from the given path and returns the raw model data string.
	Parse(filePath string) (domain.RawModelData, error)
}
