package comparator

import (
	"strings"
	"uml_compare/domain"
	"uml_compare/matcher"
)

type StandardComparator struct {
	fuzzyMatcher matcher.IFuzzyMatcher
}

func NewStandardComparator(fz matcher.IFuzzyMatcher) *StandardComparator {
	return &StandardComparator{
		fuzzyMatcher: fz,
	}
}

func (c *StandardComparator) Compare(solution *domain.ProcessedUMLGraph, student *domain.ProcessedUMLGraph, mapping domain.MappingTable) (*domain.DiffReport, error) {
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

		// Compare content inside the node
		c.compareNodeContent(solNode, *stuNode, typeMap, report)
	}

	// Extra Nodes
	for i := range student.Nodes {
		stuNode := &student.Nodes[i]
		if !mappedStudentNodeIDs[stuNode.ID] {
			report.ExtraDetail.Class = append(report.ExtraDetail.Class, domain.NodeDiff{Sol: nil, Stu: stuNode, Description: "Extra node (" + stuNode.Type + ")"})
		}
	}

	// 3. Edge check
	c.compareEdges(solution, student, mapping, report)

	return report, nil
}

func (c *StandardComparator) translateType(typeName string, typeMap map[string]string) string {
	if translated, ok := typeMap[typeName]; ok {
		return translated
	}
	return typeName
}

func (c *StandardComparator) compareNodeContent(sol domain.ProcessedNode, stu domain.ProcessedNode, typeMap map[string]string, report *domain.DiffReport) {
	// --- Attributes ---
	stuAttrs := make([]domain.ProcessedAttribute, len(stu.Attributes))
	copy(stuAttrs, stu.Attributes)
	matchedStuAttrIdx := make(map[int]bool)

	for _, sAttr := range sol.Attributes {
		foundIdx := -1
		// Rule: Match by Type first, then Name(fuzzy)
		for i, stAttr := range stuAttrs {
			if matchedStuAttrIdx[i] {
				continue
			}
			if c.translateType(sAttr.Type, typeMap) == stAttr.Type {
				if c.fuzzyMatcher.Compare(sAttr.Name, stAttr.Name) >= 0.5 || 
				   strings.Contains(strings.ToLower(stAttr.Name), strings.ToLower(sAttr.Name)) ||
				   strings.Contains(strings.ToLower(sAttr.Name), strings.ToLower(stAttr.Name)) {
					foundIdx = i
					break
				}
			}
		}

		if foundIdx != -1 {
			matchedStuAttrIdx[foundIdx] = true
			matchingStu := stuAttrs[foundIdx]
			issues := []string{}
			if sAttr.Scope != matchingStu.Scope {
				issues = append(issues, "Scope mismatch ("+sAttr.Scope+" vs "+matchingStu.Scope+")")
			}
			if sAttr.Kind != matchingStu.Kind {
				issues = append(issues, "Kind mismatch ("+sAttr.Kind+" vs "+matchingStu.Kind+")")
			}

			if len(issues) > 0 {
				report.WrongDetail.Attribute = append(report.WrongDetail.Attribute, domain.AttributeDiff{ParentClassName: sol.Name, Sol: &sAttr, Stu: &matchingStu, Description: strings.Join(issues, ", ")})
			} else {
				report.CorrectDetail.Attribute = append(report.CorrectDetail.Attribute, domain.AttributeDiff{ParentClassName: sol.Name, Sol: &sAttr, Stu: &matchingStu, Description: "Match"})
			}
		} else {
			report.MissingDetail.Attribute = append(report.MissingDetail.Attribute, domain.AttributeDiff{ParentClassName: sol.Name, Sol: &sAttr, Stu: nil, Description: "Missing attribute (" + sAttr.Scope + " " + sAttr.Type + ")"})
		}
	}

	for i := range stuAttrs {
		stAttr := &stuAttrs[i]
		if !matchedStuAttrIdx[i] {
			report.ExtraDetail.Attribute = append(report.ExtraDetail.Attribute, domain.AttributeDiff{ParentClassName: stu.Name, Sol: nil, Stu: stAttr, Description: "Extra attribute (" + stAttr.Scope + " " + stAttr.Type + ")"})
		}
	}

	// --- Methods ---
	solG, solS, solNormal := c.splitMethods(sol.Methods)
	stuG, stuS, stuNormal := c.splitMethods(stu.Methods)

	// Getter/Setter Count logic
	if (sol.Shortcut&1) == 0 && (stu.Shortcut&1) == 0 {
		if len(solG) != len(stuG) {
			report.WrongDetail.Method = append(report.WrongDetail.Method, domain.MethodDiff{ParentClassName: sol.Name, Sol: nil, Stu: nil, Description: strings.Join([]string{"Expected", itoa(len(solG)), "getter(s), got", itoa(len(stuG))}, " ")})
		} else if len(solG) > 0 {
			report.CorrectDetail.Method = append(report.CorrectDetail.Method, domain.MethodDiff{ParentClassName: sol.Name, Sol: nil, Stu: nil, Description: itoa(len(solG)) + " getter(s) match"})
		}
	}
	if (sol.Shortcut&2) == 0 && (stu.Shortcut&2) == 0 {
		if len(solS) != len(stuS) {
			report.WrongDetail.Method = append(report.WrongDetail.Method, domain.MethodDiff{ParentClassName: sol.Name, Sol: nil, Stu: nil, Description: strings.Join([]string{"Expected", itoa(len(solS)), "setter(s), got", itoa(len(stuS))}, " ")})
		} else if len(solS) > 0 {
			report.CorrectDetail.Method = append(report.CorrectDetail.Method, domain.MethodDiff{ParentClassName: sol.Name, Sol: nil, Stu: nil, Description: itoa(len(solS)) + " setter(s) match"})
		}
	}

	// Normal Methods
	matchedStuMethIdx := make(map[int]bool)
	for _, sMethod := range solNormal {
		isCtor := c.isConstructor(sMethod, sol.Name)
		foundIdx := -1
		
		// Rule: Match by ReturnType & NumParams(+-1 if >= 2), then Name(fuzzy)
		for i, stMethod := range stuNormal {
			if matchedStuMethIdx[i] {
				continue
			}
			
			// Return type check
			retMatch := true
			if !isCtor {
				retMatch = (c.translateType(sMethod.Output, typeMap) == stMethod.Output)
			}
			
			if retMatch {
				// Param count check
				solPLen := len(sMethod.Inputs)
				stuPLen := len(stMethod.Inputs)
				paramCountMatch := (solPLen == stuPLen)
				
				// User rule: +-1 if both >= 2
				if solPLen >= 2 && stuPLen >= 2 {
					diff := solPLen - stuPLen
					if diff < 0 { diff = -diff }
					if diff <= 1 {
						paramCountMatch = true
					}
				}
				
				if paramCountMatch {
					if c.matchMethodName(sMethod, stMethod, isCtor, stu.Name) {
						foundIdx = i
						break
					}
				}
			}
		}

		if foundIdx != -1 {
			matchedStuMethIdx[foundIdx] = true
			matchingStu := stuNormal[foundIdx]
			issues := []string{}
			
			// Detailed check
			if sMethod.Scope != matchingStu.Scope {
				issues = append(issues, "Scope mismatch ("+sMethod.Scope+" vs "+matchingStu.Scope+")")
			}
			if sMethod.Kind != matchingStu.Kind {
				issues = append(issues, "Kind mismatch ("+sMethod.Kind+" vs "+matchingStu.Kind+")")
			}
			// Exact params check
			if len(sMethod.Inputs) != len(matchingStu.Inputs) {
				issues = append(issues, "Param count mismatch ("+itoa(len(sMethod.Inputs))+" vs "+itoa(len(matchingStu.Inputs))+")")
			} else {
				for j := range sMethod.Inputs {
					if c.translateType(sMethod.Inputs[j].Type, typeMap) != matchingStu.Inputs[j].Type {
						issues = append(issues, "Param "+itoa(j+1)+" type mismatch")
						break
					}
				}
			}

			if len(issues) > 0 {
				report.WrongDetail.Method = append(report.WrongDetail.Method, domain.MethodDiff{ParentClassName: sol.Name, Sol: &sMethod, Stu: &matchingStu, Description: strings.Join(issues, ", ")})
			} else {
				report.CorrectDetail.Method = append(report.CorrectDetail.Method, domain.MethodDiff{ParentClassName: sol.Name, Sol: &sMethod, Stu: &matchingStu, Description: "Match"})
			}
		} else {
			if isCtor {
				report.MissingDetail.Method = append(report.MissingDetail.Method, domain.MethodDiff{ParentClassName: sol.Name, Sol: &sMethod, Stu: nil, Description: "Missing constructor"})
			} else {
				report.MissingDetail.Method = append(report.MissingDetail.Method, domain.MethodDiff{ParentClassName: sol.Name, Sol: &sMethod, Stu: nil, Description: "Missing method (" + sMethod.Scope + " " + sMethod.Output + ")"})
			}
		}
	}

	for i := range stuNormal {
		stMethod := &stuNormal[i]
		if !matchedStuMethIdx[i] {
			report.ExtraDetail.Method = append(report.ExtraDetail.Method, domain.MethodDiff{ParentClassName: stu.Name, Sol: nil, Stu: stMethod, Description: "Extra method (" + stMethod.Scope + " " + stMethod.Output + ")"})
		}
	}
}

func (c *StandardComparator) matchMethodName(sol domain.ProcessedMethod, stu domain.ProcessedMethod, solIsCtor bool, stuClassName string) bool {
	if solIsCtor {
		return c.isConstructor(stu, stuClassName)
	}
	// Fuzzy Name >= 0.5
	return c.fuzzyMatcher.Compare(sol.Name, stu.Name) >= 0.5
}

// itoa converts an integer to its string representation.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	buf := [20]byte{}
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte(n%10) + '0'
		n /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}

func (c *StandardComparator) isConstructor(m domain.ProcessedMethod, className string) bool {
	return strings.EqualFold(m.Name, className) || strings.EqualFold(m.Name, "init") || strings.EqualFold(m.Name, "<<create>>")
}

func (c *StandardComparator) splitMethods(methods []domain.ProcessedMethod) (g []domain.ProcessedMethod, s []domain.ProcessedMethod, normal []domain.ProcessedMethod) {
	for _, m := range methods {
		switch m.Type {
		case "getter":
			g = append(g, m)
		case "setter":
			s = append(s, m)
		default:
			normal = append(normal, m)
		}
	}
	return
}

func (c *StandardComparator) compareEdges(solution *domain.ProcessedUMLGraph, student *domain.ProcessedUMLGraph, mapping domain.MappingTable, report *domain.DiffReport) {
	solIDToNode := make(map[string]domain.ProcessedNode)
	for _, n := range solution.Nodes {
		solIDToNode[n.ID] = n
	}

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
			if matchedStuEdgeIdx[i] {
				continue
			}
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
