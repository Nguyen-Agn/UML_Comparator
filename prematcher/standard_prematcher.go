package prematcher

import (
	"strings"
	"uml_compare/domain"
)

type StandardPreMatcher struct {
	Parser     IMemberParser
	Calc       IWeightCalculator
	Detector   ITypeDetector
}

var _ IPreMatcher = (*StandardPreMatcher)(nil)

// NewStandardPreMatcher creates a new StandardPreMatcher with default components.
func NewStandardPreMatcher() *StandardPreMatcher {
	detector := NewTypeDetector()
	return &StandardPreMatcher{
		Parser:   NewStandardMemberParser(),
		Calc:     NewWeightCalculator(detector),
		Detector: detector,
	}
}

// NewCustomStandardPreMatcher allows injecting custom components for testing or extension.
func NewCustomStandardPreMatcher(parser IMemberParser, calc IWeightCalculator, detector ITypeDetector) *StandardPreMatcher {
	return &StandardPreMatcher{
		Parser:   parser,
		Calc:     calc,
		Detector: detector,
	}
}

func (p *StandardPreMatcher) Process(graph *domain.UMLGraph) (*domain.ProcessedUMLGraph, error) {
	if graph == nil {
		return nil, nil
	}

	processed := &domain.ProcessedUMLGraph{
		Nodes: make([]domain.ProcessedNode, len(graph.Nodes)),
		Edges: make([]domain.ProcessedEdge, len(graph.Edges)),
	}

	// 1. Analyze relationships
	inheritsMap := make(map[string]string)
	implementsMap := make(map[string][]string)
	relatedCountMap := make(map[string]int)

	for i, edge := range graph.Edges {
		processed.Edges[i] = edge

		switch edge.RelationType {
		case "Inheritance", "Generalization":
			inheritsMap[edge.SourceID] = edge.TargetID
		case "Realization", "Implementation":
			implementsMap[edge.SourceID] = append(implementsMap[edge.SourceID], edge.TargetID)
		default:
			relatedCountMap[edge.SourceID]++
		}
	}

	// 2. Map nodes
	for i, node := range graph.Nodes {
		isEnum := p.Detector.IsEnumType(node.Type)
		pNode := domain.ProcessedNode{
			ID:         node.ID,
			Name:       cleanText(node.Name),
			IsBold:     node.IsBold,
			Type:       p.Detector.NormalizeNodeType(node.Type),
			Inherits:   inheritsMap[node.ID],
			Implements: implementsMap[node.ID],
			Attributes: make([]domain.ProcessedAttribute, 0, len(node.Attributes)),
			Methods:    make([]domain.ProcessedMethod, 0, len(node.Methods)),
		}

		staticMembersCount := 0
		customTypeCount := 0

		if strings.Contains(pNode.Name, "<") && strings.Contains(pNode.Name, ">") {
			customTypeCount++
		}

		// --- STEP A: Parse Attributes ---
		for _, attr := range node.Attributes {
			raw := cleanText(attr)
			if isPureShortcut(raw) {
				lower := strings.ToLower(raw)
				if strings.Contains(lower, "getter") {
					pNode.Shortcut |= 1
				}
				if strings.Contains(lower, "setter") {
					pNode.Shortcut |= 2
				}
				continue
			}

			parsedAttr := p.Parser.ParseAttribute(raw, isEnum)
			pNode.Attributes = append(pNode.Attributes, parsedAttr)

			lowerRaw := strings.ToLower(raw)
			if strings.Contains(lowerRaw, "getter") {
				pNode.Methods = append(pNode.Methods, p.Parser.GenerateGetter(parsedAttr))
			}
			if strings.Contains(lowerRaw, "setter") {
				pNode.Methods = append(pNode.Methods, p.Parser.GenerateSetter(parsedAttr))
			}

			customTypeCount += strings.Count(parsedAttr.Type, "<") + strings.Count(parsedAttr.Type, ",")
			if parsedAttr.Kind == "static" || parsedAttr.Kind == "static-final" {
				staticMembersCount++
			}
		}

		// --- STEP B: Parse Methods ---
		claimedGetters := make(map[string]bool)
		claimedSetters := make(map[string]bool)
		for _, m := range pNode.Methods {
			if m.Type == "getter" {
				claimedGetters[m.Name] = true
			}
			if m.Type == "setter" {
				claimedSetters[m.Name] = true
			}
		}

		for _, method := range node.Methods {
			raw := cleanText(method)
			if isPureShortcut(raw) {
				lower := strings.ToLower(raw)
				if strings.Contains(lower, "getter") {
					pNode.Shortcut |= 1
				}
				if strings.Contains(lower, "setter") {
					pNode.Shortcut |= 2
				}
				continue
			}
			lowerRaw := strings.ToLower(raw)

			if (strings.Contains(lowerRaw, "getter") || strings.Contains(lowerRaw, "setter")) && !strings.Contains(raw, "(") {
				attr := p.Parser.ParseAttribute(raw, false)
				if strings.Contains(lowerRaw, "getter") {
					pNode.Methods = append(pNode.Methods, p.Parser.GenerateGetter(attr))
				}
				if strings.Contains(lowerRaw, "setter") {
					pNode.Methods = append(pNode.Methods, p.Parser.GenerateSetter(attr))
				}
				continue
			}

			parsedMethod := p.Parser.ParseMethod(raw, pNode.Name, pNode.Attributes, claimedGetters, claimedSetters)
			pNode.Methods = append(pNode.Methods, parsedMethod)

			customTypeCount += strings.Count(parsedMethod.Output, "<") + strings.Count(parsedMethod.Output, ",")
			for _, param := range parsedMethod.Inputs {
				customTypeCount += strings.Count(param.Type, "<") + strings.Count(param.Type, ",")
			}

			if parsedMethod.Kind == "static" {
				staticMembersCount++
			}
		}

		var validMethodCount int
		for _, m := range pNode.Methods {
			if m.Type != "getter" && m.Type != "setter" {
				validMethodCount++
			}
		}

		pNode.ArchWeight = p.Calc.Calculate(
			pNode.Type,
			pNode.Inherits != "",
			len(pNode.Implements),
			validMethodCount,
			len(pNode.Attributes),
			relatedCountMap[node.ID],
			customTypeCount,
			staticMembersCount,
		)

		processed.Nodes[i] = pNode
	}

	return processed, nil
}


