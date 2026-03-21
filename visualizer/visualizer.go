package visualizer

import "uml_compare/domain"

// IVisualizer defines the contract for exporting or visualizing the differences.
type IVisualizer interface {
	// VisualizeDiff creates an output file (like a colored draw.io) highlighting the errors.
	VisualizeDiff(report *domain.DiffReport, solution *domain.UMLGraph, student *domain.UMLGraph, studentRawPath string) error

	// ExportReport creates a text/HTML report string summarizing the diff.
	ExportReport(report *domain.DiffReport) (string, error)
}
