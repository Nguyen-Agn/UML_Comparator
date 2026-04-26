package grader

import (
	"fmt"
	"math"
	"uml_compare/domain"
)

// StandardGrader provides the implementation for the default grading rules.
// Rules:
// - Uses the scores assigned during preprocessing (extracted from __d__).
// - If no score configuration exists, defaults to 1.
// - Node penalty includes the base node score plus all its attributes and methods scores.
type StandardGrader struct{}

func NewStandardGrader() *StandardGrader {
	return &StandardGrader{}
}

// roundData rounds float64 appropriately strictly following the template.
func roundData(val float64) float64 {
	return math.Round(val*100) / 100
}

func calculateNodePenalty(node *domain.SolutionProcessedNode) float64 {
	penalty := node.Score
	for _, a := range node.Attributes {
		penalty += a.Score
	}
	for _, m := range node.Methods {
		penalty += m.Score
	}
	return penalty
}

func getEdgeScore(edge *domain.ProcessedEdge, sol *domain.SolutionProcessedUMLGraph) float64 {
	if edge == nil || sol == nil {
		return 1.0
	}
	srcName, tgtName := edge.SourceID, edge.TargetID
	for _, n := range sol.Nodes {
		if n.ID == edge.SourceID {
			srcName = n.Name
		}
		if n.ID == edge.TargetID {
			tgtName = n.Name
		}
	}
	edgeKey := srcName + "::" + tgtName + "::" + edge.RelationType
	if val, ok := sol.GradingConfig.Edges[edgeKey]; ok {
		return val
	}
	return 1.0
}

func (g *StandardGrader) Grade(report *domain.DiffReport, sol *domain.SolutionProcessedUMLGraph, stu *domain.ProcessedUMLGraph, rule *GradingRules) (*domain.GradeResult, error) {
	maxScore := 0.0

	// 1. Calculate MaxScore based on the Solution Graph
	for _, node := range sol.Nodes {
		maxScore += calculateNodePenalty(&node)
	}
	for _, edge := range sol.Edges {
		maxScore += getEdgeScore(&edge, sol)
	}

	totalScore := maxScore
	var feedbacks []string

	// --- Static Validation Phase (Convention Checks) ---
	// Scan student nodes to enforce UML convention (Bold class names) independently of the Solution.
	if stu != nil {
		for _, stuNode := range stu.Nodes {
			if stuNode.Type == "Class" && !stuNode.IsBold {
				penalty := 0.1 // DEFAULT PENALTY POINT FOR UNBOLD CLASS NAME
				if rule != nil && rule.Format_Penalty > 0 {
					penalty = rule.Format_Penalty
				}
				totalScore -= penalty
				feedbacks = append(feedbacks, fmt.Sprintf("Formatting Penalty: Class '%s' is missing bold format (-%.1f)", stuNode.Name, penalty))
			}
		}
	}

	// Helper to handle DetailError category deductions
	processDetail := func(detail *domain.DetailError, category string) {
		// Calculate edge omissions
		for _, e := range detail.Edge {
			penalty := getEdgeScore(e.Sol, sol)
			totalScore -= penalty
			feedbacks = append(feedbacks, fmt.Sprintf("%s Edge: %s (Penalty: -%.1f)", category, e.Description, penalty))
		}

		// Calculate attribute mismatches
		for _, a := range detail.Attribute {
			penalty := 1.0
			if a.Sol != nil {
				penalty = a.Sol.Score
			}
			totalScore -= penalty
			feedbacks = append(feedbacks, fmt.Sprintf("%s Attribute in %s: %s (Penalty: -%.1f)", category, a.ParentClassName, a.Description, penalty))
		}

		// Calculate method mismatches
		for _, m := range detail.Method {
			penalty := 1.0
			if m.Sol != nil {
				penalty = m.Sol.Score
			}
			totalScore -= penalty
			feedbacks = append(feedbacks, fmt.Sprintf("%s Method in %s: %s (Penalty: -%.1f)", category, m.ParentClassName, m.Description, penalty))
		}

		// Calculate full node omissions or structural wrong types
		for _, n := range detail.Class {
			if n.Sol != nil {
				penalty := calculateNodePenalty(n.Sol)
				totalScore -= penalty
				feedbacks = append(feedbacks, fmt.Sprintf("%s Node '%s': %s (Penalty: -%.1f)", category, n.Sol.Name, n.Description, penalty))
			}
		}
	}

	// Subtract points for missing and wrong items
	processDetail(&report.MissingDetail, "Missing")
	processDetail(&report.WrongDetail, "Wrong")

	// Ensure score doesn't drop below 0
	totalScore = math.Max(0, totalScore)
	totalScore = roundData(totalScore)

	var percentage float64
	if maxScore > 0 {
		percentage = roundData((totalScore / maxScore) * 100)
	} else {
		percentage = 100.0 // Edge case if empty diagram
	}

	return &domain.GradeResult{
		TotalScore:     totalScore,
		MaxScore:       maxScore,
		CorrectPercent: percentage,
		Feedbacks:      feedbacks,
		Report:         report,
		SolutionGraph:  sol,
		StudentGraph:   stu,
	}, nil
}
