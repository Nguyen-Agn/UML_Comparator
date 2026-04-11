package grader

import "uml_compare/domain"

// GradingRules holds configuration values for grading criteria.
// This could eventually be loaded from a JSON config file.
type GradingRules struct {
	Max_point    int32
	Miss_point   int32
	Wrong_point  int32
	Format_point float64 // Configurable penalty for formatting errors like missing bold
}

// IGrader defines the contract for calculating the final score based on the DiffReport.
type IGrader interface {
	// Grade computes the final GradeResult.
	Grade(report *domain.DiffReport, sol *domain.SolutionProcessedUMLGraph, stu *domain.ProcessedUMLGraph, rule *GradingRules) (*domain.GradeResult, error)
}
