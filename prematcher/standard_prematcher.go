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
			Type:       node.Type,
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

		for _, attr := range node.Attributes {
			parsedAttr := p.parseAttribute(cleanText(attr))
			pNode.Attributes = append(pNode.Attributes, parsedAttr)

			// Simple heuristic: if type contains <, it might be a custom type
			if strings.Contains(parsedAttr.Type, "<") {
				customTypeCount++
			}
			// In UML, static attributes are visually underlined. We might not have that info in string,
			// but we can check if it contains the word static or {static}
			if strings.Contains(strings.ToLower(parsedAttr.Name), "static") || strings.Contains(strings.ToLower(parsedAttr.Type), "static") {
				staticMembersCount++
			}
		}

		for _, method := range node.Methods {
			parsedMethod := p.parseMethod(cleanText(method), pNode.Name)
			pNode.Methods = append(pNode.Methods, parsedMethod)

			if strings.Contains(parsedMethod.Type, "<") || strings.Contains(parsedMethod.Output, "<") {
				customTypeCount++
			}
			for _, param := range parsedMethod.Inputs {
				if strings.Contains(param.Type, "<") {
					customTypeCount++
				}
			}

			if strings.Contains(strings.ToLower(parsedMethod.Name), "static") {
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
	// Fallback default
	attr := domain.ProcessedAttribute{
		Scope: "+", // Default scope
		Name:  raw, // Fallback if no colon
		Type:  "",
	}

	matches := p.attrRegex.FindStringSubmatch(raw)
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
	} else if idx := strings.Index(raw, ":"); idx != -1 { // simple fallback
		attr.Name = strings.TrimSpace(raw[:idx])
		attr.Type = strings.TrimSpace(raw[idx+1:])
	}

	return attr
}

func (p *StandardPreMatcher) parseMethod(raw string, className string) domain.ProcessedMethod {
	method := domain.ProcessedMethod{
		Scope:  "+",
		Name:   raw,
		Type:   "",
		Output: "",
		Inputs: []domain.MethodParam{},
	}

	matches := p.methodRegex.FindStringSubmatch(raw)
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
					param.Name = part // Only name or type provided
				}
				method.Inputs = append(method.Inputs, param)
			}
		}

		if len(matches) > 4 {
			method.Output = strings.TrimSpace(matches[4])
		}
	} else {
		// Simple fallback
		if openIdx := strings.Index(raw, "("); openIdx != -1 {
			method.Name = strings.TrimSpace(raw[:openIdx])
		}
	}

	// Classify the Method Type
	lowerName := strings.ToLower(method.Name)
	if lowerName == strings.ToLower(className) || lowerName == "constructor" || lowerName == "init" {
		method.Type = "constructor"
	} else if strings.HasPrefix(lowerName, "get") && len(method.Inputs) <= 1 {
		method.Type = "getter"
	} else if strings.HasPrefix(lowerName, "set") && len(method.Inputs) > 0 {
		method.Type = "setter"
	} else {
		method.Type = "custom"
	}

	return method
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
	if strings.Contains(lowerType, "class") && !strings.Contains(lowerType, "abstract") {
		typeVal = 1
	} else if strings.Contains(lowerType, "interface") {
		typeVal = 2
	} else if strings.Contains(lowerType, "abstract") {
		typeVal = 3
	} else if strings.Contains(lowerType, "enum") {
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

func minU32(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}

func cleanText(text string) string {
	// Strip HTML tags if any
	re := regexp.MustCompile(`<[^>]*>`)
	text = re.ReplaceAllString(text, "")
	// Also decode literal common entities that drawio leaves like &nbsp;
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	return strings.TrimSpace(text)
}
