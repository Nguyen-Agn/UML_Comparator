package controller

import "uml_compare/src/uml_generator"

// exportMermaid is a thin bridge that calls SolutionExporter from the controller package.
// This avoids importing the uml_generator package directly in generator_controller.go
// and keeps the dependency explicit.
func exportMermaid(mermaidCode, outputPath string) error {
	return uml_generator.NewSolutionExporter().Export(mermaidCode, outputPath)
}
