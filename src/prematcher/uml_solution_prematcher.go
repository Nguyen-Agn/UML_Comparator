package prematcher

import (
	"strings"
	"uml_compare/domain"
)

// UMLSolutionPreMatcher implements IUMLSolutionPreMatcher.
type UMLSolutionPreMatcher struct {
	Parser   ISolutionMemberParser
	Calc     IWeightCalculator
	Detector ITypeDetector
	ScoreExc IScoreExtractor
}

// Compile-time interface check.
var _ IUMLSolutionPreMatcher = (*UMLSolutionPreMatcher)(nil)

// NewUMLSolutionPreMatcher creates a ready-to-use UMLSolutionPreMatcher with default components.
func NewUMLSolutionPreMatcher() *UMLSolutionPreMatcher {
	detector := NewTypeDetector()
	return &UMLSolutionPreMatcher{
		Parser:   NewSolutionMemberParser(),
		Calc:     NewWeightCalculator(detector),
		Detector: detector,
		ScoreExc: NewScoreExtractor(),
	}
}

// NewCustomUMLSolutionPreMatcher allows injecting custom components for testing or extension.
func NewCustomUMLSolutionPreMatcher(parser ISolutionMemberParser, calc IWeightCalculator, detector ITypeDetector, scoreExc IScoreExtractor) *UMLSolutionPreMatcher {
	return &UMLSolutionPreMatcher{
		Parser:   parser,
		Calc:     calc,
		Detector: detector,
		ScoreExc: scoreExc,
	}
}

// ProcessSolution transforms a raw UMLGraph into an OR-aware SolutionProcessedUMLGraph.
func (p *UMLSolutionPreMatcher) ProcessSolution(graph *domain.UMLGraph) (*domain.SolutionProcessedUMLGraph, error) {
	if graph == nil {
		return nil, nil
	}

	processed := &domain.SolutionProcessedUMLGraph{
		Nodes: make([]domain.SolutionProcessedNode, len(graph.Nodes)),
		Edges: make([]domain.ProcessedEdge, len(graph.Edges)),
		GradingConfig: domain.ScoreConfig{
			Nodes:      make(map[string]float64),
			Attributes: make(map[string]float64),
			Methods:    make(map[string]float64),
			Edges:      make(map[string]float64),
		},
	}

	// --- Step 1: Analyze edges ---
	inheritsMap := make(map[string]string)
	implementsMap := make(map[string][]string)
	relatedCountMap := make(map[string]int)

	getNodeName := func(id string) string {
		for _, n := range graph.Nodes {
			if n.ID == id {
				cleaned, _ := p.ScoreExc.ExtractScore(n.Name)
				return cleanText(cleaned)
			}
		}
		return id
	}

	for i, edge := range graph.Edges {
		edgeScore := 1.0

		if cleaned, sc := p.ScoreExc.ExtractScore(edge.Note); cleaned != edge.Note {
			edge.Note = cleaned
			edgeScore = sc
		} else if cleaned, sc := p.ScoreExc.ExtractScore(edge.SourceLabel); cleaned != edge.SourceLabel {
			edge.SourceLabel = cleaned
			edgeScore = sc
		} else if cleaned, sc := p.ScoreExc.ExtractScore(edge.TargetLabel); cleaned != edge.TargetLabel {
			edge.TargetLabel = cleaned
			edgeScore = sc
		}

		processed.Edges[i] = edge

		srcName := getNodeName(edge.SourceID)
		tgtName := getNodeName(edge.TargetID)
		edgeKey := srcName + "::" + tgtName + "::" + edge.RelationType
		processed.GradingConfig.Edges[edgeKey] = edgeScore

		switch edge.RelationType {
		case "Inheritance", "Generalization":
			inheritsMap[edge.SourceID] = edge.TargetID
		case "Realization", "Implementation":
			implementsMap[edge.SourceID] = append(implementsMap[edge.SourceID], edge.TargetID)
		default:
			relatedCountMap[edge.SourceID]++
		}
	}

	// --- Step 2: Process each node ---
	for i, node := range graph.Nodes {
		cleanedNodeName, nodeScore := p.ScoreExc.ExtractScore(node.Name)
		cleanedNodeName = cleanText(cleanedNodeName)
		isEnum := p.Detector.IsEnumType(node.Type)

		pNode := domain.SolutionProcessedNode{
			ID:         node.ID,
			Name:       cleanedNodeName,
			Type:       p.Detector.NormalizeNodeType(node.Type),
			Inherits:   inheritsMap[node.ID],
			Implements: implementsMap[node.ID],
			Attributes: make([]domain.SolutionProcessedAttribute, 0, len(node.Attributes)),
			Methods:    make([]domain.SolutionProcessedMethod, 0, len(node.Methods)),
			Score:      nodeScore,
		}
		processed.GradingConfig.Nodes[cleanedNodeName] = nodeScore

		staticMembersCount := 0
		customTypeCount := 0

		if strings.Contains(pNode.Name, "<") && strings.Contains(pNode.Name, ">") {
			customTypeCount++
		}

		// --- Step A: Parse Attributes ---
		for _, attr := range node.Attributes {
			rawAttr, attrScore := p.ScoreExc.ExtractScore(attr)
			raw := cleanText(rawAttr)
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
			parsedAttr.Score = attrScore
			pNode.Attributes = append(pNode.Attributes, parsedAttr)

			for _, n := range parsedAttr.Names {
				processed.GradingConfig.Attributes[cleanedNodeName+"::"+n] = attrScore
			}

			lowerRaw := strings.ToLower(raw)
			if strings.Contains(lowerRaw, "getter") {
				getter := p.Parser.GenerateGetter(parsedAttr)
				getter.Score = attrScore
				pNode.Methods = append(pNode.Methods, getter)
				if len(getter.Names) > 0 {
					processed.GradingConfig.Methods[cleanedNodeName+"::"+getter.Names[0]] = attrScore
				}
			}
			if strings.Contains(lowerRaw, "setter") {
				setter := p.Parser.GenerateSetter(parsedAttr)
				setter.Score = attrScore
				pNode.Methods = append(pNode.Methods, setter)
				if len(setter.Names) > 0 {
					processed.GradingConfig.Methods[cleanedNodeName+"::"+setter.Names[0]] = attrScore
				}
			}

			for _, t := range parsedAttr.Types {
				customTypeCount += strings.Count(t, "<") + strings.Count(t, ",")
			}
			if parsedAttr.Kind == "static" || parsedAttr.Kind == "static-final" {
				staticMembersCount++
			}
		}

		// --- Step B: Parse Methods ---
		claimedGetters := make(map[string]bool)
		claimedSetters := make(map[string]bool)
		for _, m := range pNode.Methods {
			if m.Type == "getter" {
				claimedGetters[strings.Join(m.Names, "|")] = true
			}
			if m.Type == "setter" {
				claimedSetters[strings.Join(m.Names, "|")] = true
			}
		}

		stdAttrs := p.toStdAttributes(pNode.Attributes)

		for _, method := range node.Methods {
			rawMethod, methodScore := p.ScoreExc.ExtractScore(method)
			raw := cleanText(rawMethod)
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
					getter := p.Parser.GenerateGetter(attr)
					getter.Score = methodScore
					pNode.Methods = append(pNode.Methods, getter)
					if len(getter.Names) > 0 {
						processed.GradingConfig.Methods[cleanedNodeName+"::"+getter.Names[0]] = methodScore
					}
				}
				if strings.Contains(lowerRaw, "setter") {
					setter := p.Parser.GenerateSetter(attr)
					setter.Score = methodScore
					pNode.Methods = append(pNode.Methods, setter)
					if len(setter.Names) > 0 {
						processed.GradingConfig.Methods[cleanedNodeName+"::"+setter.Names[0]] = methodScore
					}
				}
				continue
			}

			parsedMethod := p.Parser.ParseMethod(raw, pNode.Name, stdAttrs, claimedGetters, claimedSetters)
			parsedMethod.Score = methodScore
			pNode.Methods = append(pNode.Methods, parsedMethod)

			for _, n := range parsedMethod.Names {
				processed.GradingConfig.Methods[cleanedNodeName+"::"+n] = methodScore
			}

			for _, out := range parsedMethod.Outputs {
				customTypeCount += strings.Count(out, "<") + strings.Count(out, ",")
			}
			for _, param := range parsedMethod.Inputs {
				for _, t := range param.Types {
					customTypeCount += strings.Count(t, "<") + strings.Count(t, ",")
				}
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

func (p *UMLSolutionPreMatcher) toStdAttributes(attrs []domain.SolutionProcessedAttribute) []domain.ProcessedAttribute {
	result := make([]domain.ProcessedAttribute, 0, len(attrs))
	for _, a := range attrs {
		name := ""
		if len(a.Names) > 0 {
			name = a.Names[0]
		}
		typ := ""
		if len(a.Types) > 0 {
			typ = a.Types[0]
		}
		result = append(result, domain.ProcessedAttribute{
			Name:  name,
			Scope: a.Scope,
			Type:  typ,
			Kind:  a.Kind,
		})
	}
	return result
}
