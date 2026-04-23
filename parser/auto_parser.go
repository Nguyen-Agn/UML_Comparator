package parser

import (
	"fmt"
	"path/filepath"
	"uml_compare/domain"
)

// AutoParser is a registry-based IFileParser that routes to the correct
// concrete parser based on file extension.
//
// It is Open for Extension (OCP): call Register() to add new formats.
// Closed for Modification: AutoParser.Parse() never changes.
//
// Usage — zero config (recommended):
//
//	p := parser.NewAutoParserDefault()
//	raw, _ := p.Parse("exam.drawio")    // ✅ plain file
//	raw, _ := p.Parse("exam.solution")  // ✅ encrypted file (default key)
//
// Usage — custom key:
//
//	p := parser.NewAutoParser()
//	p.Register(".drawio",   parser.NewDrawioParser())
//	p.Register(".solution", parser.NewSolutionParser(myDecryptor))
type AutoParser struct {
	registry map[string]IFileParser
}

// Compile-time interface guarantee.
var _ IFileParser = (*AutoParser)(nil)

// NewAutoParserDefault returns a ready-to-use AutoParser with supported extensions:
// .drawio, .solution (type "drawio") and .mmd, .mermaid (type "mermaid").
//
// Key resolution order for .solution: SOLUTION_KEY env var → built-in default.
func NewAutoParserDefault() *AutoParser {
	p := NewAutoParser()
	p.Register(".drawio", NewDrawioParser())
	p.Register(".drawio.xml", NewDrawioParser())
	p.Register(".solution", NewSolutionParserDefault())
	p.Register(".mmd", NewMermaidParser())
	p.Register(".mermaid", NewMermaidParser())
	return p
}

// NewAutoParser returns an empty AutoParser.
// Use Register() to add parsers for each file extension.
func NewAutoParser() *AutoParser {
	return &AutoParser{registry: make(map[string]IFileParser)}
}

// Register maps a file extension (e.g. ".drawio") to an IFileParser.
// Calling Register with an existing extension overwrites the previous mapping.
func (p *AutoParser) Register(ext string, f IFileParser) {
	p.registry[ext] = f
}

// Parse dispatches to the IFileParser registered for the file's extension.
// Returns cleaned raw data, detected source type, and any error.
func (p *AutoParser) Parse(filePath string) (domain.RawModelData, string, error) {
	if filePath == "" {
		return "", "", fmt.Errorf("AutoParser.Parse: filePath cannot be empty")
	}

	ext := filepath.Ext(filePath)
	f, ok := p.registry[ext]
	if !ok {
		return "", "", fmt.Errorf(
			"AutoParser.Parse: no parser registered for extension %q", ext,
		)
	}

	return f.Parse(filePath)
}
