package visualizer

import "uml_compare/domain"

// IVisualizer defines the contract for exporting visual reports of UML grading results.
type IVisualizer interface {
	// ExportHTML renders the full GradeResult into a self-contained HTML file
	// intended for the GRADER/INSTRUCTOR. Shows both Student and Solution
	// side-by-side, plus detailed summary and deduction feedbacks.
	// outputPath is the destination file path for the generated .html file.
	// Returns an error if the file cannot be written.
	ExportHTML(result *domain.GradeResult, outputPath string) error

	// ExportStudentHTML renders a student-facing HTML report.
	// Shows ONLY the student's own nodes and relations with color-coded
	// feedback (correct/wrong/extra), score, and progress bar.
	// Does NOT reveal the solution content or detailed deduction breakdown.
	// outputPath is the destination file path for the generated .html file.
	// Returns an error if the file cannot be written.
	ExportStudentHTML(result *domain.GradeResult, outputPath string) error
}
