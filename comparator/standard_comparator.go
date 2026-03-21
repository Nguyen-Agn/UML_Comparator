package comparator

import (
	"fmt"
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
	report := &domain.DiffReport{
		MissedClass:     []string{},
		MissingNodes:    []string{},
		MissingEdges:    []string{},
		MissingMembers:  []string{},
		AttributeErrors: []string{},
		MethodErrors:    []string{},
		NodeEdgeErrors:  []string{},
	}

	if solution == nil || student == nil {
		return report, nil
	}

	// 1. Build TypeMap (Solution Name -> Student Name)
	typeMap := make(map[string]string)
	solIDToName := make(map[string]string)
	for _, solNode := range solution.Nodes {
		solIDToName[solNode.ID] = solNode.Name
		if mapped, ok := mapping[solNode.ID]; ok {
			// Find the student name
			for _, stuNode := range student.Nodes {
				if stuNode.ID == mapped.StudentID {
					typeMap[solNode.Name] = stuNode.Name
					break
				}
			}
		}
	}

	// 2. Node & Content check
	for _, solNode := range solution.Nodes {
		mapped, ok := mapping[solNode.ID]
		if !ok {
			if strings.EqualFold(solNode.Type, "class") {
				report.MissedClass = append(report.MissedClass, fmt.Sprintf("Missing Class: %s", solNode.Name))
			} else {
				report.MissingNodes = append(report.MissingNodes, fmt.Sprintf("Missing Node (%s): %s", solNode.Type, solNode.Name))
			}
			continue
		}

		// Find student node
		var stuNode *domain.ProcessedNode
		for i := range student.Nodes {
			if student.Nodes[i].ID == mapped.StudentID {
				stuNode = &student.Nodes[i]
				break
			}
		}

		if stuNode == nil {
			continue
		}

		// Compare content inside the node
		c.compareNodeContent(solNode, *stuNode, typeMap, report)
	}

	// 3. Edge check (Reverse Arrow & Missing)
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
	solAttrs := make([]domain.ProcessedAttribute, len(sol.Attributes))
	copy(solAttrs, sol.Attributes)

	stuAttrs := make([]domain.ProcessedAttribute, len(stu.Attributes))
	copy(stuAttrs, stu.Attributes)

	for _, sAttr := range solAttrs {
		foundIdx := -1
		for i, stAttr := range stuAttrs {
			if c.matchAttribute(sAttr, stAttr, typeMap) {
				foundIdx = i
				break
			}
		}

		if foundIdx != -1 {
			matchingStu := stuAttrs[foundIdx]
			// Check scope mismatch
			if sAttr.Scope != matchingStu.Scope {
				report.AttributeErrors = append(report.AttributeErrors, fmt.Sprintf("Class %s, Attribute %s: Scope mismatch (Solution: %s, Student: %s)", sol.Name, sAttr.Name, sAttr.Scope, matchingStu.Scope))
			}
			// Remove found to avoid duplicates
			stuAttrs = append(stuAttrs[:foundIdx], stuAttrs[foundIdx+1:]...)
		} else {
			report.MissingMembers = append(report.MissingMembers, fmt.Sprintf("Class %s: Missing attribute %s %s %s", sol.Name, sAttr.Scope, sAttr.Type, sAttr.Name))
		}
	}

	// --- Methods ---
	// Split into Getters, Setters, vs Normal/Constructors
	solG, solS, solNormal := c.splitMethods(sol.Methods)
	stuG, stuS, stuNormal := c.splitMethods(stu.Methods)

	// Check Getter Count (skip if either side has shortcut Bit 0)
	if (sol.Shortcut&1) == 0 && (stu.Shortcut&1) == 0 {
		if len(solG) != len(stuG) {
			report.MethodErrors = append(report.MethodErrors, fmt.Sprintf("Class %s: Expected %d Getter methods, found %d", sol.Name, len(solG), len(stuG)))
		}
	}

	// Check Setter Count (skip if either side has shortcut Bit 1)
	if (sol.Shortcut&2) == 0 && (stu.Shortcut&2) == 0 {
		if len(solS) != len(stuS) {
			report.MethodErrors = append(report.MethodErrors, fmt.Sprintf("Class %s: Expected %d Setter methods, found %d", sol.Name, len(solS), len(stuS)))
		}
	}

	// Compare Constructors & Normal Methods
	for _, sMethod := range solNormal {
		isCtor := c.isConstructor(sMethod, sol.Name)
		foundIdx := -1
		for i, stMethod := range stuNormal {
			if c.matchMethod(sMethod, stMethod, typeMap, isCtor, stu.Name) {
				foundIdx = i
				break
			}
		}

		if foundIdx != -1 {
			matchingStu := stuNormal[foundIdx]
			// Check scope mismatch
			if sMethod.Scope != matchingStu.Scope {
				report.MethodErrors = append(report.MethodErrors, fmt.Sprintf("Class %s, Method %s: Scope mismatch (Solution: %s, Student: %s)", sol.Name, sMethod.Name, sMethod.Scope, matchingStu.Scope))
			}
			stuNormal = append(stuNormal[:foundIdx], stuNormal[foundIdx+1:]...)
		} else {
			if isCtor {
				report.MissingMembers = append(report.MissingMembers, fmt.Sprintf("Class %s: Missing constructor with matching params", sol.Name))
			} else {
				report.MissingMembers = append(report.MissingMembers, fmt.Sprintf("Class %s: Missing method %s %s %s(...)", sol.Name, sMethod.Scope, sMethod.Name, sMethod.Output))
			}
		}
	}
}

func (c *StandardComparator) matchAttribute(sol domain.ProcessedAttribute, stu domain.ProcessedAttribute, typeMap map[string]string) bool {
	// Scope check - REMOVED (checked after matching)
	// if sol.Scope != stu.Scope {
	// 	return false
	// }
	// Type check (Translated)
	if c.translateType(sol.Type, typeMap) != stu.Type {
		return false
	}
	// Kind check: static, final, static-final, normal must match
	if sol.Kind != stu.Kind {
		return false
	}
	// Name check: Similarity >= 0.5 OR Contains
	if c.fuzzyMatcher.Compare(sol.Name, stu.Name) >= 0.5 {
		return true
	}
	if strings.Contains(strings.ToLower(stu.Name), strings.ToLower(sol.Name)) ||
		strings.Contains(strings.ToLower(sol.Name), strings.ToLower(stu.Name)) {
		return true
	}
	return false
}

func (c *StandardComparator) isConstructor(m domain.ProcessedMethod, className string) bool {
	return strings.EqualFold(m.Name, className) || strings.EqualFold(m.Name, "init")
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

func (c *StandardComparator) matchMethod(sol domain.ProcessedMethod, stu domain.ProcessedMethod, typeMap map[string]string, isCtor bool, stuClassName string) bool {
	// 1. Name Check (skip if both are constructors)
	if isCtor {
		if !c.isConstructor(stu, stuClassName) {
			return false
		}
	} else {
		if c.fuzzyMatcher.Compare(sol.Name, stu.Name) < 0.5 {
			return false
		}
	}
	// 3. Kind check: static, abstract, normal must match
	if sol.Kind != stu.Kind {
		return false
	}

	// 4. Return Type (except for constructors)
	if !isCtor {
		if c.translateType(sol.Output, typeMap) != stu.Output {
			return false
		}
	}

	// 5. Params
	if len(sol.Inputs) != len(stu.Inputs) {
		return false
	}

	if isCtor {
		// Unordered check for constructor params
		stuParams := make([]domain.MethodParam, len(stu.Inputs))
		copy(stuParams, stu.Inputs)

		for _, sP := range sol.Inputs {
			foundIdx := -1
			sTypeMapped := c.translateType(sP.Type, typeMap)
			for i, stP := range stuParams {
				if sTypeMapped == stP.Type {
					foundIdx = i
					break
				}
			}
			if foundIdx != -1 {
				stuParams = append(stuParams[:foundIdx], stuParams[foundIdx+1:]...)
			} else {
				return false
			}
		}
		return true
	} else {
		// Ordered check for normal methods
		for i := range sol.Inputs {
			if c.translateType(sol.Inputs[i].Type, typeMap) != stu.Inputs[i].Type {
				return false
			}
		}
		return true
	}
}

func (c *StandardComparator) compareEdges(solution *domain.ProcessedUMLGraph, student *domain.ProcessedUMLGraph, mapping domain.MappingTable, report *domain.DiffReport) {
	solIDToNode := make(map[string]domain.ProcessedNode)
	for _, n := range solution.Nodes {
		solIDToNode[n.ID] = n
	}

	for _, solEdge := range solution.Edges {
		mappedSrc, okSrc := mapping[solEdge.SourceID]
		mappedTgt, okTgt := mapping[solEdge.TargetID]

		if !okSrc || !okTgt {
			// Already reported as missing nodes usually, but let's record missing edge
			report.MissingEdges = append(report.MissingEdges, fmt.Sprintf("Missing relationship: %s -> %s", solIDToNode[solEdge.SourceID].Name, solIDToNode[solEdge.TargetID].Name))
			continue
		}

		found := false
		reverse := false
		for _, stuEdge := range student.Edges {
			if stuEdge.SourceID == mappedSrc.StudentID && stuEdge.TargetID == mappedTgt.StudentID && stuEdge.RelationType == solEdge.RelationType {
				found = true
				break
			}
			if stuEdge.SourceID == mappedTgt.StudentID && stuEdge.TargetID == mappedSrc.StudentID && stuEdge.RelationType == solEdge.RelationType {
				reverse = true
				break
			}
		}

		if found {
			continue
		}

		if reverse {
			report.NodeEdgeErrors = append(report.NodeEdgeErrors, fmt.Sprintf("[Reverse Arrow] Relationship between %s and %s is inverted", solIDToNode[solEdge.SourceID].Name, solIDToNode[solEdge.TargetID].Name))
		} else {
			report.MissingEdges = append(report.MissingEdges, fmt.Sprintf("Missing %s relationship between %s and %s", solEdge.RelationType, solIDToNode[solEdge.SourceID].Name, solIDToNode[solEdge.TargetID].Name))
		}
	}
}
