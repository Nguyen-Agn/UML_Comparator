package prematcher

import (
	"regexp"
	"strings"
	"uml_compare/domain"
)

type StandardPreMatcher struct {
	attrRegex   *regexp.Regexp
	methodRegex *regexp.Regexp
}

var _ IPreMatcher = (*StandardPreMatcher)(nil)

func NewStandardPreMatcher() *StandardPreMatcher {
	return &StandardPreMatcher{
		// Regex for: [Scope] Name : Type [= DefaultValue]
		// Scope: +, -, # or none. Name: before colon. Type: after colon.
		attrRegex: regexp.MustCompile(`^([+\-#~])?\s*([^:]+)\s*:\s*(.+)$`),

		// Regex for: [Scope] Name(params) : ReturnType
		methodRegex: regexp.MustCompile(`^([+\-#~])?\s*([^\(]+)\s*\((.*?)\)\s*(?::\s*(.+))?$`),
	}
}

func (p *StandardPreMatcher) Process(graph *domain.UMLGraph) (*domain.ProcessedUMLGraph, error) {
	if graph == nil {
		return nil, nil
	}

	processed := &domain.ProcessedUMLGraph{
		Nodes: make([]domain.ProcessedNode, len(graph.Nodes)),
		Edges: make([]domain.ProcessedEdge, len(graph.Edges)), // We'll just copy them
	}

	// 1. Analyze relationships to figure out Inherits and Implements
	// Also figure out related classes for each node (dependency/association)
	inheritsMap := make(map[string]string)     // childID -> parentID
	implementsMap := make(map[string][]string) // childID -> []interfaceIDs
	relatedCountMap := make(map[string]int)    // nodeID -> count of relationships to other nodes

	for i, edge := range graph.Edges {
		processed.Edges[i] = edge

		switch edge.RelationType {
		case "Inheritance", "Generalization":
			inheritsMap[edge.SourceID] = edge.TargetID
		case "Realization", "Implementation":
			implementsMap[edge.SourceID] = append(implementsMap[edge.SourceID], edge.TargetID)
		default:
			// Count as dependent/association for the source node
			relatedCountMap[edge.SourceID]++
		}
	}

	// 2. Map nodes and calculate ArchWeight
	for i, node := range graph.Nodes {
		pNode := domain.ProcessedNode{
			ID:         node.ID,
			Name:       cleanText(node.Name),
			Type:       p.normalizeNodeType(node.Type),
			Inherits:   inheritsMap[node.ID],
			Implements: implementsMap[node.ID],
			Attributes: make([]domain.ProcessedAttribute, 0, len(node.Attributes)),
			Methods:    make([]domain.ProcessedMethod, 0, len(node.Methods)),
		}

		staticMembersCount := 0
		customTypeCount := 0

		// Check for custom types in the Name (e.g. List<T>)
		if strings.Contains(pNode.Name, "<") && strings.Contains(pNode.Name, ">") {
			customTypeCount++
		}

		// --- STEP A: Parse Attributes First ---
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

			parsedAttr := p.parseAttribute(raw)
			// Enums: default missing type to "void"
			if parsedAttr.Type == "" && p.isEnumType(pNode.Type) {
				parsedAttr.Type = "void"
			}
			pNode.Attributes = append(pNode.Attributes, parsedAttr)

			// Proactively implement shortcuts: if attribute has {getter} or {setter}
			lowerRaw := strings.ToLower(raw)
			if strings.Contains(lowerRaw, "getter") {
				pNode.Methods = append(pNode.Methods, p.generateGetter(parsedAttr))
			}
			if strings.Contains(lowerRaw, "setter") {
				pNode.Methods = append(pNode.Methods, p.generateSetter(parsedAttr))
			}

			// Count generic type parameters (e.g., List<T> = 1, Map<K, V> = 2)
			customTypeCount += strings.Count(parsedAttr.Type, "<") + strings.Count(parsedAttr.Type, ",")
			// Kind identification
			if parsedAttr.Kind == "static" || parsedAttr.Kind == "static-final" {
				staticMembersCount++
			}
		}

		// --- STEP B: Parse Methods with Attribute Context ---
		claimedGetters := make(map[string]bool)
		claimedSetters := make(map[string]bool)
		// Mark methods generated from {getter}/{setter} as claimed
		for _, m := range pNode.Methods {
			if m.Type == "getter" {
				claimedGetters[m.Name] = true // Simplified marker
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

			// Check for shortcuts in the methods list as well
			if (strings.Contains(lowerRaw, "getter") || strings.Contains(lowerRaw, "setter")) && !strings.Contains(raw, "(") {
				attr := p.parseAttribute(raw)
				if strings.Contains(lowerRaw, "getter") {
					pNode.Methods = append(pNode.Methods, p.generateGetter(attr))
				}
				if strings.Contains(lowerRaw, "setter") {
					pNode.Methods = append(pNode.Methods, p.generateSetter(attr))
				}
				continue
			}

			parsedMethod := p.parseMethod(raw, pNode.Name, pNode.Attributes, claimedGetters, claimedSetters)
			pNode.Methods = append(pNode.Methods, parsedMethod)

			// Count generic output and parameters
			customTypeCount += strings.Count(parsedMethod.Output, "<") + strings.Count(parsedMethod.Output, ",")
			for _, param := range parsedMethod.Inputs {
				customTypeCount += strings.Count(param.Type, "<") + strings.Count(param.Type, ",")
			}

			if parsedMethod.Kind == "static" {
				staticMembersCount++
			}
		}

		// Calculate valid methods for ArchWeight (ignoring getters and setters)
		var validMethodCount int
		for _, m := range pNode.Methods {
			if m.Type != "getter" && m.Type != "setter" {
				validMethodCount++
			}
		}

		// Calculate ArchWeight
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

func (p *StandardPreMatcher) parseAttribute(raw string) domain.ProcessedAttribute {
	attr := domain.ProcessedAttribute{
		Scope: "+", // Default scope
		Kind:  "normal",
	}

	// 1. Identify Kind and handle annotations
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

	// 2. Clean the string for structural parsing (remove keywords and annotations)
	working := cleanMemberString(raw)

	// 3. Apply regex or simple fallback to the cleaned string
	matches := p.attrRegex.FindStringSubmatch(working)
	if len(matches) > 0 {
		if matches[1] != "" {
			attr.Scope = matches[1]
		}
		attr.Name = strings.TrimSpace(matches[2])
		// Remove default value from type if exists (e.g. Type = Default)
		typePart := strings.TrimSpace(matches[3])
		if idx := strings.Index(typePart, "="); idx != -1 {
			typePart = strings.TrimSpace(typePart[:idx])
		}
		attr.Type = typePart
	} else if idx := strings.Index(working, ":"); idx != -1 {
		namePart := strings.TrimSpace(working[:idx])
		if len(namePart) > 0 && isScopeChar(namePart[0]) {
			attr.Scope = string(namePart[0])
			attr.Name = strings.TrimSpace(namePart[1:])
		} else {
			attr.Name = namePart
		}
		attr.Type = strings.TrimSpace(working[idx+1:])
	} else {
		// Even simpler fallback: look for scope at start
		if len(working) > 0 && isScopeChar(working[0]) {
			attr.Scope = string(working[0])
			attr.Name = strings.TrimSpace(working[1:])
		} else {
			attr.Name = working
		}
	}

	return attr
}

func (p *StandardPreMatcher) parseMethod(raw string, className string, attributes []domain.ProcessedAttribute, claimedG, claimedS map[string]bool) domain.ProcessedMethod {
	method := domain.ProcessedMethod{
		Scope:  "+",
		Name:   raw,
		Type:   "",
		Output: "",
		Inputs: []domain.MethodParam{},
		Kind:   "normal",
	}

	lowerRaw := strings.ToLower(raw)

	// Identify Kind
	if strings.Contains(lowerRaw, "static") || strings.Contains(lowerRaw, "{static}") {
		method.Kind = "static"
	} else if strings.Contains(lowerRaw, "abstract") || strings.Contains(lowerRaw, "{abstract}") {
		method.Kind = "abstract"
	}

	// Shortcut check (no parentheses)
	if (strings.Contains(lowerRaw, "getter") || strings.Contains(lowerRaw, "setter")) && !strings.Contains(raw, "(") {
		attr := p.parseAttribute(raw)
		method.Scope = attr.Scope
		method.Name = attr.Name
		method.Output = attr.Type
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
		method.Name = strings.TrimSpace(matches[2])
		method.Type = raw // Store original string

		// Parse params
		paramStr := strings.TrimSpace(matches[3])
		if paramStr != "" {
			parts := strings.Split(paramStr, ",")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if part == "" {
					continue
				}
				param := domain.MethodParam{}
				if idx := strings.Index(part, ":"); idx != -1 {
					param.Name = strings.TrimSpace(part[:idx])
					param.Type = strings.TrimSpace(part[idx+1:])
				} else {
					param.Name = part
				}
				method.Inputs = append(method.Inputs, param)
			}
		}

		if len(matches) > 4 {
			method.Output = strings.TrimSpace(matches[4])
		}
	} else {
		// Simple fallback
		if openIdx := strings.Index(working, "("); openIdx != -1 {
			method.Name = strings.TrimSpace(working[:openIdx])
			if len(working) > 0 && isScopeChar(method.Name[0]) {
				method.Scope = string(method.Name[0])
				method.Name = strings.TrimSpace(method.Name[1:])
			}
		} else {
			if len(working) > 0 && isScopeChar(working[0]) {
				method.Scope = string(working[0])
				method.Name = strings.TrimSpace(working[1:])
			} else {
				method.Name = working
			}
		}
	}

	// Classify the Method Type
	lowerName := strings.ToLower(method.Name)

	if lowerName == strings.ToLower(className) || lowerName == "constructor" || lowerName == "init" {
		method.Type = "constructor"
	} else if strings.HasPrefix(lowerName, "get") && (len(method.Inputs) == 0 || (len(method.Inputs) == 1 && strings.EqualFold(method.Inputs[0].Type, "void"))) {
		// Potential getter: check against attributes
		baseName := method.Name[3:]
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
			method.Type = "getter"
			claimedG[foundAttr] = true
		} else {
			method.Type = "custom"
		}
	} else if strings.HasPrefix(lowerName, "set") && len(method.Inputs) == 1 {
		// Potential setter: check against attributes
		baseName := method.Name[3:]
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
			method.Type = "setter"
			claimedS[foundAttr] = true
		} else {
			method.Type = "custom"
		}
	} else {
		method.Type = "custom"
	}

	// Default return type to "void" if empty and not a constructor
	if method.Output == "" && method.Type != "constructor" {
		method.Output = "void"
	}

	return method
}

func (p *StandardPreMatcher) generateGetter(attr domain.ProcessedAttribute) domain.ProcessedMethod {
	capitalized := strings.ToUpper(attr.Name[:1]) + attr.Name[1:]
	return domain.ProcessedMethod{
		Scope:  "+",
		Name:   "get" + capitalized,
		Type:   "getter",
		Output: attr.Type,
		Inputs: []domain.MethodParam{},
		Kind:   "normal",
	}
}

func (p *StandardPreMatcher) generateSetter(attr domain.ProcessedAttribute) domain.ProcessedMethod {
	capitalized := strings.ToUpper(attr.Name[:1]) + attr.Name[1:]
	return domain.ProcessedMethod{
		Scope:  "+",
		Name:   "set" + capitalized,
		Type:   "setter",
		Output: "void",
		Inputs: []domain.MethodParam{
			{Name: attr.Name, Type: attr.Type},
		},
		Kind: "normal",
	}
}

// calculateArchWeight uses bitwise shifting to pack structural info into a single uint32
//
// Bit 29-31: Loại Class (3 bit - 0: Unknown, 1: Class, 2: Interface, 3: Abstract, 4: Enum)
// Bit 28: Có thừa kế không? (1 bit - 1: Có, 0: Không)
// Bit 24-27: Số lượng Interface thực thi (4 bit - Max 15)
// Bit 18-23: Số lượng Method (6 bit - Max 63)
// Bit 13-17: Số lượng Attribute (5 bit - Max 31)
// Bit 9-12: Số lượng Class liên quan/phụ thuộc (4 bit - Max 15)
// Bit 6-8: Số lượng Custom Type <T> (3 bit - Max 7)
// Bit 2-5: Số lượng Static members (4 bit - Max 15)
// Bit 0-1: Dự phòng (Not used)
func (p *StandardPreMatcher) calculateArchWeight(
	nodeType string,
	hasInheritance bool,
	numInterfaces int,
	numMethods int,
	numAttributes int,
	numRelated int,
	numCustomTypes int,
	numStaticMembers int,
) uint32 {
	var weight uint32 = 0

	// 1. Loại Class (Bit 29-31)
	var typeVal uint32 = 0
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

	// 2. Thừa kế (Bit 28)
	if hasInheritance {
		weight |= (1 & 0x1) << 28
	}

	// 3. Số lượng Interface (Bit 24-27)
	weight |= (minU32(uint32(numInterfaces), 15) & 0xF) << 24

	// 4. Số lượng Method (Bit 18-23)
	weight |= (minU32(uint32(numMethods), 63) & 0x3F) << 18

	// 5. Số lượng Attribute (Bit 13-17)
	weight |= (minU32(uint32(numAttributes), 31) & 0x1F) << 13

	// 6. Số lượng Class liên quan (Bit 9-12)
	weight |= (minU32(uint32(numRelated), 15) & 0xF) << 9

	// 7. Số lượng Custom Type (Bit 6-8)
	weight |= (minU32(uint32(numCustomTypes), 7) & 0x7) << 6

	// 8. Số lượng Static members (Bit 2-5)
	weight |= (minU32(uint32(numStaticMembers), 15) & 0xF) << 2

	return weight
}

// normalizeNodeType maps various stereotype formats to a standard representation.
func (p *StandardPreMatcher) normalizeNodeType(t string) string {
	if p.isEnumType(t) {
		return "Enum"
	}
	// Add other normalizations if needed (e.g. interface, abstract)
	return t
}

// isEnumType checks if a given type string represents an enumeration (including stereotypes).
func (p *StandardPreMatcher) isEnumType(t string) bool {
	lower := strings.ToLower(t)
	// Match: enum, enumeration, <<enum>>, <<enumeration>>, «enum», «enumeration»
	return strings.Contains(lower, "enum") ||
		strings.Contains(lower, "enumeration") ||
		(strings.Contains(lower, "«") && strings.Contains(lower, "enu")) ||
		(strings.Contains(lower, "<<") && strings.Contains(lower, "enu"))
}


