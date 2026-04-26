package comparator

import (
	"uml_compare/domain"
)

// StandardEdgeComparator implements IEdgeComparator for relationship comparison.
type StandardEdgeComparator struct{}

var _ IEdgeComparator = (*StandardEdgeComparator)(nil)

// NewStandardEdgeComparator creates a new instance of StandardEdgeComparator.
func NewStandardEdgeComparator() *StandardEdgeComparator {
	return &StandardEdgeComparator{}
}

// CompareEdges identifies differences in relationships between solution and student graphs.
func (v *StandardEdgeComparator) CompareEdges(solution *domain.SolutionProcessedUMLGraph, student *domain.ProcessedUMLGraph, mapping domain.MappingTable, report *domain.DiffReport) {
	matchedStuEdgeIdx := make(map[int]bool)

	for _, solEdge := range solution.Edges {
		mappedSrc, okSrc := mapping[solEdge.SourceID]
		mappedTgt, okTgt := mapping[solEdge.TargetID]

		if !okSrc || !okTgt {
			report.MissingDetail.Edge = append(report.MissingDetail.Edge, domain.EdgeDiff{Sol: &solEdge, Stu: nil, Description: "Missing relationship (" + solEdge.RelationType + ")"})
			continue
		}

		foundIdx := -1
		wrongTypeIdx := -1
		reverseIdx := -1

		for i, stuEdge := range student.Edges {
			if matchedStuEdgeIdx[i] { continue }
			
			// Exact match
			if stuEdge.SourceID == mappedSrc.StudentID && stuEdge.TargetID == mappedTgt.StudentID {
				if stuEdge.RelationType == solEdge.RelationType {
					foundIdx = i
					break
				} else {
					wrongTypeIdx = i
				}
			}
			// Reverse
			if stuEdge.SourceID == mappedTgt.StudentID && stuEdge.TargetID == mappedSrc.StudentID && stuEdge.RelationType == solEdge.RelationType {
				reverseIdx = i
			}
		}

		if foundIdx != -1 {
			matchedStuEdgeIdx[foundIdx] = true
			report.CorrectDetail.Edge = append(report.CorrectDetail.Edge, domain.EdgeDiff{Sol: &solEdge, Stu: &student.Edges[foundIdx], Description: "Relationship match"})
		} else if wrongTypeIdx != -1 {
			matchedStuEdgeIdx[wrongTypeIdx] = true
			report.WrongDetail.Edge = append(report.WrongDetail.Edge, domain.EdgeDiff{Sol: &solEdge, Stu: &student.Edges[wrongTypeIdx], Description: "Wrong relationship type"})
		} else if reverseIdx != -1 {
			matchedStuEdgeIdx[reverseIdx] = true
			report.WrongDetail.Edge = append(report.WrongDetail.Edge, domain.EdgeDiff{Sol: &solEdge, Stu: &student.Edges[reverseIdx], Description: "Reverse arrow"})
		} else {
			report.MissingDetail.Edge = append(report.MissingDetail.Edge, domain.EdgeDiff{Sol: &solEdge, Stu: nil, Description: "Missing relationship (" + solEdge.RelationType + ")"})
		}
	}

	// Extra Edges
	for i := range student.Edges {
		if !matchedStuEdgeIdx[i] {
			report.ExtraDetail.Edge = append(report.ExtraDetail.Edge, domain.EdgeDiff{Sol: nil, Stu: &student.Edges[i], Description: "Extra relationship (" + student.Edges[i].RelationType + ")"})
		}
	}
}
