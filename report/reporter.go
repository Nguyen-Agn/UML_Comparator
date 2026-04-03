package report

import "uml_compare/domain"

// BatchGradeResult holds the results of grading multiple student submissions against a single solution.
type BatchGradeResult struct {
	SolutionPath   string
	StudentResults map[string]*domain.GradeResult // Key: Student file path or ID, Value: The grading result
}

// IReporter defines the contract for generating reports from batch grading results.
// Implementations of this interface should handle the output format (e.g., Console, HTML, CSV).
type IReporter interface {
	// GenerateReport takes a BatchGradeResult and generates the corresponding report.
	// It returns an error if the report generation fails.
	GenerateReport(batchResult *BatchGradeResult) error
}
