package report

import "uml_compare/domain"

// IReporter defines the contract for generating reports from batch grading results.
// Implementations of this interface should handle the output format (e.g., Console, HTML, CSV).
type IReporter interface {
	// GenerateReport takes a BatchGradeResult and generates the corresponding report.
	// It returns an error if the report generation fails.
	GenerateReport(batchResult *domain.BatchGradeResult) error
}
