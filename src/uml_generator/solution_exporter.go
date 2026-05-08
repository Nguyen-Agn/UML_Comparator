package uml_generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SolutionExporter writes the (possibly user-edited) Mermaid code to a .mmd file.
// Score tags (__d__) are expected to already be present in the mermaidCode string
// (either from the AI or from the user's edits in the textarea).
type SolutionExporter struct{}

// NewSolutionExporter creates a new SolutionExporter.
func NewSolutionExporter() *SolutionExporter {
	return &SolutionExporter{}
}

// Export validates the mermaid code and writes it to outputPath.
// outputPath must have a .mmd extension. If the directory does not exist it is created.
// Returns an error if the code does not contain a valid classDiagram declaration.
func (e *SolutionExporter) Export(mermaidCode, outputPath string) error {
	if err := validate(mermaidCode); err != nil {
		return fmt.Errorf("solution_exporter: %w", err)
	}
	if filepath.Ext(outputPath) != ".mmd" {
		return fmt.Errorf("solution_exporter: output path must end with .mmd, got %q", outputPath)
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("solution_exporter: create dir: %w", err)
	}
	content := strings.TrimSpace(mermaidCode) + "\n"
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("solution_exporter: write file: %w", err)
	}
	return nil
}

// validate checks that mermaidCode is a non-empty classDiagram block.
func validate(mermaidCode string) error {
	trimmed := strings.TrimSpace(mermaidCode)
	if trimmed == "" {
		return fmt.Errorf("mermaid code is empty")
	}
	if !strings.Contains(trimmed, "classDiagram") {
		return fmt.Errorf("mermaid code does not contain 'classDiagram' declaration")
	}
	return nil
}
