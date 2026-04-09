package comparator

import (
	"strings"
	"uml_compare/domain"
	"uml_compare/matcher"
)

type StandardComparator struct {
	fuzzyMatcher     matcher.IFuzzyMatcher
	typeAnalyzer     ITypeAnalyzer
	memberComparator IMemberComparator
	edgeComparator   IEdgeComparator
}

var _ IComparator = (*StandardComparator)(nil)

func NewStandardComparator(fz matcher.IFuzzyMatcher, ta ITypeAnalyzer, mc IMemberComparator, ec IEdgeComparator) *StandardComparator {
	return &StandardComparator{
		fuzzyMatcher:     fz,
		typeAnalyzer:     ta,
		memberComparator: mc,
		edgeComparator:   ec,
	}
}

func (c *StandardComparator) Compare(solution *domain.SolutionProcessedUMLGraph, student *domain.ProcessedUMLGraph, mapping domain.MappingTable) (*domain.DiffReport, error) {
	report := &domain.DiffReport{}

	if solution == nil || student == nil {
		return report, nil
	}

	// 1. Build TypeMap (Solution Name -> Student Name)
	typeMap := make(map[string]string)
	for _, solNode := range solution.Nodes {
		if mapped, ok := mapping[solNode.ID]; ok {
			for _, stuNode := range student.Nodes {
				if stuNode.ID == mapped.StudentID {
					typeMap[solNode.Name] = stuNode.Name
					break
				}
			}
		}
	}

	// 2. Node & Content check
	mappedStudentNodeIDs := make(map[string]bool)
	for _, solNode := range solution.Nodes {
		mapped, ok := mapping[solNode.ID]
		// check exist?

		if !ok {
			if strings.EqualFold(solNode.Type, "class") {
				report.MissingDetail.Class = append(report.MissingDetail.Class, domain.NodeDiff{Sol: &solNode, Stu: nil, Description: "Missing class"})
			} else {
				report.MissingDetail.Class = append(report.MissingDetail.Class, domain.NodeDiff{Sol: &solNode, Stu: nil, Description: "Missing node (" + solNode.Type + ")"})
			}
			continue
		}

		// Find student node
		var stuNode *domain.ProcessedNode
		for i := range student.Nodes {
			if student.Nodes[i].ID == mapped.StudentID {
				stuNode = &student.Nodes[i]
				mappedStudentNodeIDs[stuNode.ID] = true
				break
			}
		}

		if stuNode == nil {
			report.MissingDetail.Class = append(report.MissingDetail.Class, domain.NodeDiff{Sol: &solNode, Stu: nil, Description: "Mapped ID not found"})
			continue
		}

		// Type match?
		if !strings.EqualFold(solNode.Type, stuNode.Type) {
			report.WrongDetail.Class = append(report.WrongDetail.Class, domain.NodeDiff{Sol: &solNode, Stu: stuNode, Description: "Type mismatch (Solution: " + solNode.Type + ", Student: " + stuNode.Type + ")"})
		} else {
			report.CorrectDetail.Class = append(report.CorrectDetail.Class, domain.NodeDiff{Sol: &solNode, Stu: stuNode, Description: "Match"})
		}

		// Compare content inside the node via MemberComparator
		c.memberComparator.CompareAttributes(solNode, *stuNode, typeMap, report)
		c.memberComparator.CompareMethods(solNode, *stuNode, typeMap, report)
	}

	// Extra Nodes
	for i := range student.Nodes {
		stuNode := &student.Nodes[i]
		if !mappedStudentNodeIDs[stuNode.ID] {
			report.ExtraDetail.Class = append(report.ExtraDetail.Class, domain.NodeDiff{Sol: nil, Stu: stuNode, Description: "Extra node (" + stuNode.Type + ")"})
		}
	}

	// 3. Edge check via EdgeComparator
	c.edgeComparator.CompareEdges(solution, student, mapping, report)

	return report, nil
}
