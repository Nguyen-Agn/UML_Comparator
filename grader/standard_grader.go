package grader

import (
	"fmt"
	"math"
	"strings"
	"uml_compare/domain"
)

// StandardGrader provides the implementation for the default grading rules.
// Rules:
// - Attribute or Method (non-getter/setter) = 1 point maximum.
// - Relation = 2 points maximum.
// - Node = Method pts + Attribute pts + 1 (if generic) + 1 (per Inheritance/Implementation).
// Deductions occur for items missing or wrong. Max score represents a perfect match.
type StandardGrader struct{}

func NewStandardGrader() *StandardGrader {
	return &StandardGrader{}
}

// roundData rounds float64 appropriately strictly following the template.
func roundData(val float64) float64 {
	return math.Round(val*100) / 100
}

func calculateNodeMaxScore(node *domain.SolutionProcessedNode) float64 {
	score := 0.0

	// Attribute points
	score += float64(len(node.Attributes))

	// Method points (exclude getters/setters)
	for _, m := range node.Methods {
		// Ensure it's not a getter/setter
		if m.Type != "getter" && m.Type != "setter" {
			score += 1.0
		}
	}

	// Generic type points
	if strings.Contains(node.Name, "<") && strings.Contains(node.Name, ">") {
		score += 1.0
	}

	// Inheritance
	if node.Inherits != "" {
		score += 1.0
	}

	// Implementations
	score += float64(len(node.Implements))

	return score
}

func (g *StandardGrader) Grade(report *domain.DiffReport, sol *domain.SolutionProcessedUMLGraph, stu *domain.ProcessedUMLGraph, rule *GradingRules) (*domain.GradeResult, error) {
	maxScore := 0.0

	// 1. Calculate MaxScore based on the Solution Graph
	for _, node := range sol.Nodes {
		maxScore += calculateNodeMaxScore(&node)
	}
	for range sol.Edges {
		maxScore += 2.0 // Each edge is worth 2 points
	}

	totalScore := maxScore
	var feedbacks []string

	// Helper to handle DetailError category deductions
	processDetail := func(detail *domain.DetailError, category string) {
		// Calculate edge omissions
		for _, e := range detail.Edge {
			totalScore -= 2.0
			feedbacks = append(feedbacks, fmt.Sprintf("%s Edge: %s", category, e.Description))
		}

		// Calculate attribute mismatches
		for _, a := range detail.Attribute {
			totalScore -= 1.0
			feedbacks = append(feedbacks, fmt.Sprintf("%s Attribute in %s: %s", category, a.ParentClassName, a.Description))
		}

		// Calculate method mismatches
		for _, m := range detail.Method {
			totalScore -= 1.0
			feedbacks = append(feedbacks, fmt.Sprintf("%s Method in %s: %s", category, m.ParentClassName, m.Description))
		}

		// Calculate full node omissions or structural wrong types
		for _, n := range detail.Class {
			if n.Sol != nil {
				penalty := calculateNodeMaxScore(n.Sol)
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
