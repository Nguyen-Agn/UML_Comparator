package prematcher

import (
	"regexp"
	"strings"
	"uml_compare/domain"
)

// StandardMemberParser implements IMemberParser.
type StandardMemberParser struct {
	attrRegex   *regexp.Regexp
	methodRegex *regexp.Regexp
}

// NewStandardMemberParser creates a new instance of StandardMemberParser.
func NewStandardMemberParser() *StandardMemberParser {
	return &StandardMemberParser{
		// Regex for: [Scope] Name : Type [= DefaultValue]
		attrRegex: regexp.MustCompile(`^([+\-#~])?\s*([^:]+)\s*:\s*(.+)$`),
		// Regex for: [Scope] Name(params) : ReturnType
		methodRegex: regexp.MustCompile(`^([+\-#~])?\s*([^\(]+)\s*\((.*?)\)\s*(?::\s*(.+))?$`),
	}
}

// ParseAttribute transforms a raw string into a ProcessedAttribute.
func (p *StandardMemberParser) ParseAttribute(raw string, isEnumType bool) domain.ProcessedAttribute {
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
		// Remove default value from type if exists
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
		if len(working) > 0 && isScopeChar(working[0]) {
			attr.Scope = string(working[0])
			attr.Name = strings.TrimSpace(working[1:])
		} else {
			attr.Name = working
		}
	}

	// Enums: default missing type to "void"
	if attr.Type == "" && isEnumType {
		attr.Type = "void"
	}

	return attr
}

// ParseMethod transforms a raw string into a ProcessedMethod.
func (p *StandardMemberParser) ParseMethod(raw string, className string, attributes []domain.ProcessedAttribute, claimedG, claimedS map[string]bool) domain.ProcessedMethod {
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
		attr := p.ParseAttribute(raw, false)
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

	working := cleanMemberString(raw)

	matches := p.methodRegex.FindStringSubmatch(working)
	if len(matches) > 0 {
		if matches[1] != "" {
			method.Scope = matches[1]
		}
		method.Name = strings.TrimSpace(matches[2])
		method.Type = raw // Store original string

		paramStr := strings.TrimSpace(matches[3])
		if paramStr != "" {
			parts := splitParams(paramStr)
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
		if openIdx := strings.Index(working, "("); openIdx != -1 {
			method.Name = strings.TrimSpace(working[:openIdx])
			if len(method.Name) > 0 && isScopeChar(method.Name[0]) {
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

	// Classify Method Type
	lowerName := strings.ToLower(method.Name)
	if lowerName == strings.ToLower(className) || lowerName == "constructor" || lowerName == "init" {
		method.Type = "constructor"
	} else if strings.HasPrefix(lowerName, "get") && (len(method.Inputs) == 0 || (len(method.Inputs) == 1 && strings.EqualFold(method.Inputs[0].Type, "void"))) {
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

	if method.Output == "" && method.Type != "constructor" {
		method.Output = "void"
	}

	return method
}

// GenerateGetter creates a synthetic getter method for an attribute.
func (p *StandardMemberParser) GenerateGetter(attr domain.ProcessedAttribute) domain.ProcessedMethod {
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

// GenerateSetter creates a synthetic setter method for an attribute.
func (p *StandardMemberParser) GenerateSetter(attr domain.ProcessedAttribute) domain.ProcessedMethod {
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
