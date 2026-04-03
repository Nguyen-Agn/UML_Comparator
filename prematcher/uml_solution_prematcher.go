package prematcher

import (
	"regexp"
	"strconv"
	"strings"
	"uml_compare/domain"
)

// UMLSolutionPreMatcher implements IUMLSolutionPreMatcher.
// It parses a UMLGraph into a SolutionProcessedUMLGraph where OR-patterns ("|")
// are supported for attribute names, attribute types, custom method names, and
// method return types.
//
// All helper functions are shared via prematch_helpers.go (splitOR, cleanText, etc.).
type UMLSolutionPreMatcher struct {
	attrRegex   *regexp.Regexp
	methodRegex *regexp.Regexp
}

// Compile-time interface check.
var _ IUMLSolutionPreMatcher = (*UMLSolutionPreMatcher)(nil)

// NewUMLSolutionPreMatcher creates a ready-to-use UMLSolutionPreMatcher.
func NewUMLSolutionPreMatcher() *UMLSolutionPreMatcher {
	return &UMLSolutionPreMatcher{
		// Regex for: [Scope] Name : Type [= DefaultValue]
		// Name may contain "|" (e.g. "x | y") — captured as-is, split later.
		attrRegex: regexp.MustCompile(`^([+\-#~])?\s*([^:]+)\s*:\s*(.+)$`),

		// Regex for: [Scope] Name(params) : ReturnType
		// Name may contain "|" (e.g. "doA | doB") — captured as-is, split later.
		methodRegex: regexp.MustCompile(`^([+\-#~])?\s*([^\(]+)\s*\((.*?)\)\s*(?::\s*(.+))?$`),
	}
}

var scoreRegex = regexp.MustCompile(`__(\d+(?:\.\d+)?)__\s*$`)

// extractScore pulls out the __d__ or __d.d__ point value from the end of a string.
func extractScore(raw string) (string, float64) {
	matches := scoreRegex.FindStringSubmatch(raw)
	if len(matches) > 0 {
		scoreStr := matches[1]
		score, err := strconv.ParseFloat(scoreStr, 64)
		if err == nil {
			cleaned := scoreRegex.ReplaceAllString(raw, "")
			return strings.TrimSpace(cleaned), score
		}
	}
	return raw, 1.0 // Default score is 1.0
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

	// --- Step 1: Analyze edges for Inherits / Implements / RelatedCount ---
	inheritsMap := make(map[string]string)
	implementsMap := make(map[string][]string)
	relatedCountMap := make(map[string]int)

	getNodeName := func(id string) string {
		for _, n := range graph.Nodes {
			if n.ID == id {
				cleaned, _ := extractScore(n.Name)
				return cleanText(cleaned)
			}
		}
		return id
	}

	for i, edge := range graph.Edges {
		edgeScore := 1.0

		if cleaned, sc := extractScore(edge.Note); cleaned != edge.Note {
			edge.Note = cleaned
			edgeScore = sc
		} else if cleaned, sc := extractScore(edge.SourceLabel); cleaned != edge.SourceLabel {
			edge.SourceLabel = cleaned
			edgeScore = sc
		} else if cleaned, sc := extractScore(edge.TargetLabel); cleaned != edge.TargetLabel {
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
		cleanedNodeName, nodeScore := extractScore(node.Name)
		cleanedNodeName = cleanText(cleanedNodeName)

		pNode := domain.SolutionProcessedNode{
			ID:         node.ID,
			Name:       cleanedNodeName,
			Type:       p.normalizeNodeType(node.Type),
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
			rawAttr, attrScore := extractScore(attr)
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

			parsedAttr := p.parseSolutionAttribute(raw)
			parsedAttr.Score = attrScore
			// Enums: default missing type to "void"
			if len(parsedAttr.Types) == 0 || (len(parsedAttr.Types) == 1 && parsedAttr.Types[0] == "") {
				if p.isEnumType(pNode.Type) {
					parsedAttr.Types = []string{"void"}
				}
			}
			pNode.Attributes = append(pNode.Attributes, parsedAttr)

			for _, n := range parsedAttr.Names {
				processed.GradingConfig.Attributes[cleanedNodeName+"::"+n] = attrScore
			}

			// Proactively generate getters/setters from {getter}/{setter} annotations
			lowerRaw := strings.ToLower(raw)
			if strings.Contains(lowerRaw, "getter") {
				getter := p.generateGetter(parsedAttr)
				getter.Score = attrScore
				pNode.Methods = append(pNode.Methods, getter)
				if len(getter.Names) > 0 {
					processed.GradingConfig.Methods[cleanedNodeName+"::"+getter.Names[0]] = attrScore
				}
			}
			if strings.Contains(lowerRaw, "setter") {
				setter := p.generateSetter(parsedAttr)
				setter.Score = attrScore
				pNode.Methods = append(pNode.Methods, setter)
				if len(setter.Names) > 0 {
					processed.GradingConfig.Methods[cleanedNodeName+"::"+setter.Names[0]] = attrScore
				}
			}

			// Count generic type parameters for ArchWeight
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

		// Convert []SolutionProcessedAttribute to []ProcessedAttribute for getter/setter matching
		stdAttrs := p.toStdAttributes(pNode.Attributes)

		for _, method := range node.Methods {
			rawMethod, methodScore := extractScore(method)
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

			// Handle standalone shortcut attribute lines in method section
			if (strings.Contains(lowerRaw, "getter") || strings.Contains(lowerRaw, "setter")) && !strings.Contains(raw, "(") {
				attr := p.parseSolutionAttribute(raw)
				if strings.Contains(lowerRaw, "getter") {
					getter := p.generateGetter(attr)
					getter.Score = methodScore
					pNode.Methods = append(pNode.Methods, getter)
					if len(getter.Names) > 0 {
						processed.GradingConfig.Methods[cleanedNodeName+"::"+getter.Names[0]] = methodScore
					}
				}
				if strings.Contains(lowerRaw, "setter") {
					setter := p.generateSetter(attr)
					setter.Score = methodScore
					pNode.Methods = append(pNode.Methods, setter)
					if len(setter.Names) > 0 {
						processed.GradingConfig.Methods[cleanedNodeName+"::"+setter.Names[0]] = methodScore
					}
				}
				continue
			}

			parsedMethod := p.parseSolutionMethod(raw, pNode.Name, stdAttrs, claimedGetters, claimedSetters)
			parsedMethod.Score = methodScore
			pNode.Methods = append(pNode.Methods, parsedMethod)

			for _, n := range parsedMethod.Names {
				processed.GradingConfig.Methods[cleanedNodeName+"::"+n] = methodScore
			}

			for _, out := range parsedMethod.Outputs {
				customTypeCount += strings.Count(out, "<") + strings.Count(out, ",")
			}
			for _, param := range parsedMethod.Inputs {
				customTypeCount += strings.Count(param.Type, "<") + strings.Count(param.Type, ",")
			}
			if parsedMethod.Kind == "static" {
				staticMembersCount++
			}
		}

		// Calculate valid method count (exclude getters/setters for ArchWeight)
		var validMethodCount int
		for _, m := range pNode.Methods {
			if m.Type != "getter" && m.Type != "setter" {
				validMethodCount++
			}
		}

		pNode.ArchWeight = p.calculateArchWeight(
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

// parseSolutionAttribute parses a raw attribute string with OR support on Names and Types.
// Input examples:
//   - "- id : int|long"          -> Names=["id"],   Types=["int","long"]
//   - "+ x | y : String"         -> Names=["x","y"], Types=["String"]
//   - "{static} + count : int"   -> Names=["count"], Types=["int"], Kind="static"
func (p *UMLSolutionPreMatcher) parseSolutionAttribute(raw string) domain.SolutionProcessedAttribute {
	attr := domain.SolutionProcessedAttribute{
		Scope: "+",
		Kind:  "normal",
	}

	// Identify kind from keywords
	lowerRaw := strings.ToLower(raw)
	isStatic := strings.Contains(lowerRaw, "static") || strings.Contains(lowerRaw, "{static}")
	isFinal := strings.Contains(lowerRaw, "final") || strings.Contains(lowerRaw, "const") || strings.Contains(lowerRaw, "{readonly}")
	if isStatic && isFinal {
		attr.Kind = "static-final"
	} else if isStatic {
		attr.Kind = "static"
	} else if isFinal {
		attr.Kind = "final"
	}

	// Clean the string for structural parsing
	working := cleanMemberString(raw)

	// Apply regex
	matches := p.attrRegex.FindStringSubmatch(working)
	if len(matches) > 0 {
		if matches[1] != "" {
			attr.Scope = matches[1]
		}
		// Split Name on "|"
		attr.Names = splitOR(matches[2])

		// Remove default value from type if present (e.g. "Type = Default")
		typePart := strings.TrimSpace(matches[3])
		if idx := strings.Index(typePart, "="); idx != -1 {
			typePart = strings.TrimSpace(typePart[:idx])
		}
		// Split Type on "|"
		attr.Types = splitOR(typePart)
	} else if idx := strings.Index(working, ":"); idx != -1 {
		namePart := strings.TrimSpace(working[:idx])
		if len(namePart) > 0 && isScopeChar(namePart[0]) {
			attr.Scope = string(namePart[0])
			namePart = strings.TrimSpace(namePart[1:])
		}
		attr.Names = splitOR(namePart)
		attr.Types = splitOR(strings.TrimSpace(working[idx+1:]))
	} else {
		// Fallback: no colon
		if len(working) > 0 && isScopeChar(working[0]) {
			attr.Scope = string(working[0])
			working = strings.TrimSpace(working[1:])
		}
		attr.Names = splitOR(working)
		attr.Types = []string{}
	}

	return attr
}

// parseSolutionMethod parses a raw method string with OR support on Names and Outputs.
// Input examples:
//   - "doA | doB(a:int): void|boolean" -> Names=["doA","doB"], Outputs=["void","boolean"]
//   - "+ calculate(x:int): int"         -> Names=["calculate"],  Outputs=["int"]
//   - "MyClass()"                        -> Names=["MyClass"],    Type="constructor"
func (p *UMLSolutionPreMatcher) parseSolutionMethod(
	raw string,
	className string,
	attributes []domain.ProcessedAttribute,
	claimedG, claimedS map[string]bool,
) domain.SolutionProcessedMethod {

	method := domain.SolutionProcessedMethod{
		Scope:   "+",
		Names:   []string{raw},
		Type:    "",
		Outputs: []string{},
		Inputs:  []domain.MethodParam{},
		Kind:    "normal",
	}

	lowerRaw := strings.ToLower(raw)

	// Identify kind
	if strings.Contains(lowerRaw, "static") || strings.Contains(lowerRaw, "{static}") {
		method.Kind = "static"
	} else if strings.Contains(lowerRaw, "abstract") || strings.Contains(lowerRaw, "{abstract}") {
		method.Kind = "abstract"
	}

	// Shortcut check (no parentheses — treat as attribute-style getter/setter)
	if (strings.Contains(lowerRaw, "getter") || strings.Contains(lowerRaw, "setter")) && !strings.Contains(raw, "(") {
		attr := p.parseSolutionAttribute(raw)
		method.Scope = attr.Scope
		method.Names = attr.Names
		method.Outputs = attr.Types
		if strings.Contains(lowerRaw, "getter") {
			method.Type = "getter"
		} else {
			method.Type = "setter"
		}
		return method
	}

	// Clean string for structural parsing
	working := cleanMemberString(raw)

	matches := p.methodRegex.FindStringSubmatch(working)
	if len(matches) > 0 {
		if matches[1] != "" {
			method.Scope = matches[1]
		}

		// Split method name on "|" — OR on custom method names
		method.Names = splitOR(matches[2])

		// Parse params (comma-separated, values as plain strings — no OR split)
		paramStr := strings.TrimSpace(matches[3])
		if paramStr != "" {
			// Split params carefully, respecting angle brackets for generics
			for _, part := range splitParams(paramStr) {
				part = strings.TrimSpace(part)
				if part == "" {
					continue
				}
				param := domain.MethodParam{}
				if idx := strings.Index(part, ":"); idx != -1 {
					param.Name = strings.TrimSpace(part[:idx])
					param.Type = strings.TrimSpace(part[idx+1:]) // kept as-is, may include "|"
				} else {
					param.Name = part
				}
				method.Inputs = append(method.Inputs, param)
			}
		}

		// Split return type on "|"
		if len(matches) > 4 && matches[4] != "" {
			method.Outputs = splitOR(matches[4])
		}
	} else {
		// Fallback: simple parse
		if openIdx := strings.Index(working, "("); openIdx != -1 {
			namePart := strings.TrimSpace(working[:openIdx])
			if len(namePart) > 0 && isScopeChar(namePart[0]) {
				method.Scope = string(namePart[0])
				namePart = strings.TrimSpace(namePart[1:])
			}
			method.Names = splitOR(namePart)
		} else {
			if len(working) > 0 && isScopeChar(working[0]) {
				method.Scope = string(working[0])
				working = strings.TrimSpace(working[1:])
			}
			method.Names = splitOR(working)
		}
	}

	// --- Classify method type ---
	// Use first name for classification (primary name)
	primaryName := ""
	if len(method.Names) > 0 {
		primaryName = method.Names[0]
	}
	lowerName := strings.ToLower(primaryName)

	if lowerName == strings.ToLower(className) || lowerName == "constructor" || lowerName == "init" {
		method.Type = "constructor"
		// Constructors have no return type
		return method
	} else if len(method.Names) == 1 && strings.HasPrefix(lowerName, "get") &&
		(len(method.Inputs) == 0 || (len(method.Inputs) == 1 && strings.EqualFold(method.Inputs[0].Type, "void"))) {
		// Potential getter — only classify single-named methods as getters
		baseName := primaryName[3:]
		foundAttr := ""
		for _, attr := range attributes {
			if fuzzySimilarity(baseName, attr.Name) >= 0.8 {
				if !claimedG[attr.Name] {
					foundAttr = attr.Name
					break
				}
			}
		}
		if foundAttr != "" {
			method.Type = "custom"
			claimedG[foundAttr] = true
		} else {
			method.Type = "custom"
		}
	} else if len(method.Names) == 1 && strings.HasPrefix(lowerName, "set") && len(method.Inputs) == 1 {
		// Potential setter — only classify single-named methods as setters
		baseName := primaryName[3:]
		foundAttr := ""
		for _, attr := range attributes {
			if fuzzySimilarity(baseName, attr.Name) >= 0.8 {
				if !claimedS[attr.Name] {
					foundAttr = attr.Name
					break
				}
			}
		}
		if foundAttr != "" {
			method.Type = "custom"
			claimedS[foundAttr] = true
		} else {
			method.Type = "custom"
		}
	} else {
		method.Type = "custom"
	}

	// Default return type to "void" if not specified and not a constructor
	if len(method.Outputs) == 0 && method.Type != "constructor" {
		method.Outputs = []string{"void"}
	}

	return method
}

// generateGetter creates a solution getter method from a SolutionProcessedAttribute.
// Uses the first name in the attribute's Names slice as the base name.
func (p *UMLSolutionPreMatcher) generateGetter(attr domain.SolutionProcessedAttribute) domain.SolutionProcessedMethod {
	baseName := ""
	if len(attr.Names) > 0 {
		baseName = attr.Names[0]
	}
	capitalized := ""
	if len(baseName) > 0 {
		capitalized = strings.ToUpper(baseName[:1]) + baseName[1:]
	}
	return domain.SolutionProcessedMethod{
		Scope:   "+",
		Names:   []string{"get" + capitalized},
		Type:    "getter",
		Outputs: attr.Types,
		Inputs:  []domain.MethodParam{},
		Kind:    "normal",
	}
}

// generateSetter creates a solution setter method from a SolutionProcessedAttribute.
// Uses the first name in the attribute's Names slice as the base name.
func (p *UMLSolutionPreMatcher) generateSetter(attr domain.SolutionProcessedAttribute) domain.SolutionProcessedMethod {
	baseName := ""
	if len(attr.Names) > 0 {
		baseName = attr.Names[0]
	}
	paramType := ""
	if len(attr.Types) > 0 {
		paramType = attr.Types[0] // Use primary type for setter param
	}
	capitalized := ""
	if len(baseName) > 0 {
		capitalized = strings.ToUpper(baseName[:1]) + baseName[1:]
	}
	return domain.SolutionProcessedMethod{
		Scope:   "+",
		Names:   []string{"set" + capitalized},
		Type:    "setter",
		Outputs: []string{"void"},
		Inputs:  []domain.MethodParam{{Name: baseName, Type: paramType}},
		Kind:    "normal",
	}
}

// toStdAttributes converts []SolutionProcessedAttribute to []ProcessedAttribute for
// fuzzy getter/setter matching (uses only the first Name and first Type).
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

// splitParams splits a parameter string on commas while respecting angle brackets
// (generics like "Map<String, int>" are not split at the inner comma).
func splitParams(paramStr string) []string {
	var parts []string
	depth := 0
	start := 0
	for i, ch := range paramStr {
		switch ch {
		case '<':
			depth++
		case '>':
			depth--
		case ',':
			if depth == 0 {
				parts = append(parts, paramStr[start:i])
				start = i + 1
			}
		}
	}
	parts = append(parts, paramStr[start:])
	return parts
}

// normalizeNodeType standardises node type strings.
func (p *UMLSolutionPreMatcher) normalizeNodeType(t string) string {
	if p.isEnumType(t) {
		return "Enum"
	}
	return t
}

// isEnumType returns true for any string that represents an enumeration.
func (p *UMLSolutionPreMatcher) isEnumType(t string) bool {
	lower := strings.ToLower(t)
	return strings.Contains(lower, "enum") ||
		strings.Contains(lower, "enumeration") ||
		(strings.Contains(lower, "«") && strings.Contains(lower, "enu")) ||
		(strings.Contains(lower, "<<") && strings.Contains(lower, "enu"))
}

// calculateArchWeight packs structural info into a uint32 bitmask.
// Semantics are identical to StandardPreMatcher.calculateArchWeight.
func (p *UMLSolutionPreMatcher) calculateArchWeight(
	nodeType string,
	hasInheritance bool,
	numInterfaces int,
	numMethods int,
	numAttributes int,
	numRelated int,
	numCustomTypes int,
	numStaticMembers int,
) uint32 {
	var weight uint32

	var typeVal uint32
	lowerType := strings.ToLower(nodeType)
	if (strings.Contains(lowerType, "class") || lowerType == "default") && !strings.Contains(lowerType, "abstract") {
		typeVal = 1
	} else if strings.Contains(lowerType, "interface") {
		typeVal = 2
	} else if strings.Contains(lowerType, "abstract") {
		typeVal = 3
	} else if p.isEnumType(nodeType) {
		typeVal = 4
	}
	weight |= (typeVal & 0x7) << 29

	if hasInheritance {
		weight |= (1 & 0x1) << 28
	}
	weight |= (minU32(uint32(numInterfaces), 15) & 0xF) << 24
	weight |= (minU32(uint32(numMethods), 63) & 0x3F) << 18
	weight |= (minU32(uint32(numAttributes), 31) & 0x1F) << 13
	weight |= (minU32(uint32(numRelated), 15) & 0xF) << 9
	weight |= (minU32(uint32(numCustomTypes), 7) & 0x7) << 6
	weight |= (minU32(uint32(numStaticMembers), 15) & 0xF) << 2

	return weight
}
