package grader

import (
	"math"
	"testing"
	"uml_compare/domain"
)

func TestStandardGrader_Grade(t *testing.T) {
	grader := NewStandardGrader()

	solGraph := &domain.SolutionProcessedUMLGraph{
		Nodes: []domain.SolutionProcessedNode{
			{
				ID:   "N1",
				Name: "MyClass",
				Type: "Class",
				Attributes: []domain.SolutionProcessedAttribute{
					{Names: []string{"attr1"}, Scope: "-"},
				},
				Methods: []domain.SolutionProcessedMethod{
					{Names: []string{"doSomething"}, Type: "custom"},
					{Names: []string{"getAttr1"}, Type: "getter"},
				},
			},
			{
				ID:       "N2",
				Name:     "List<String>",
				Inherits: "N1",
			},
		},
		Edges: []domain.ProcessedEdge{
			{SourceID: "N2", TargetID: "N1", RelationType: "Inheritance"},
		},
	}

	// MaxScore:
	// N1 (1 attr + 1 custom method) = 2
	// N2 (generic + inherits) = 2
	// Edge (1) = 2
	// Total Max Score = 6

	report := &domain.DiffReport{
		MissingDetail: domain.DetailError{
			Method: []domain.MethodDiff{
				{ParentClassName: "MyClass", Description: "Missing doSomething"},
			},
		},
		WrongDetail: domain.DetailError{
			Edge: []domain.EdgeDiff{
				{Description: "Wrong edge type"},
			},
		},
	}

	stuGraph := &domain.ProcessedUMLGraph{}
	rules := &GradingRules{}

	result, err := grader.Grade(report, solGraph, stuGraph, rules)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.MaxScore != 6.0 {
		t.Errorf("expected max score 6.0, got %f", result.MaxScore)
	}

	// Score deductions:
	// 1 missing method: -1
	// 1 wrong edge: -2
	// TotalScore = 6 - 3 = 3.0
	if result.TotalScore != 3.0 {
		t.Errorf("expected total score 3.0, got %f", result.TotalScore)
	}

	if math.Abs(result.CorrectPercent-50.0) > 0.01 {
		t.Errorf("expected percent 50.0, got %f", result.CorrectPercent)
	}
}
