package grader

import "uml_compare/domain"

// GradingRules holds configuration values for grading criteria.
// This could eventually be loaded from a JSON config file.
type GradingRules struct {
	MissingNodePenalty float64
	MissingEdgePenalty float64
	WrongAttrPenalty   float64
}

// IGrader defines the contract for calculating the final score based on the DiffReport.
type IGrader interface {
	// Grade computes the final GradeResult.
	Grade(report *domain.DiffReport, criteria GradingRules) (*domain.GradeResult, error)
}
