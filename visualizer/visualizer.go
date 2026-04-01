package visualizer

import "uml_compare/domain"

// IVisualizer defines the contract for exporting visual reports of UML grading results.
type IVisualizer interface {
	// ExportHTML renders the GradeResult into a self-contained HTML file.
	// The GradeResult carries the full DiffReport, both graphs, score, and feedbacks.
	// outputPath is the destination file path for the generated .html file.
	// Returns an error if the file cannot be written.
	ExportHTML(result *domain.GradeResult, outputPath string) error
}
