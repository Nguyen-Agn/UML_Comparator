package prematcher

import (
	"regexp"
	"strings"
	"uml_compare/domain"
)

// SolutionMemberParser implements ISolutionMemberParser.
type SolutionMemberParser struct {
	attrRegex   *regexp.Regexp
	methodRegex *regexp.Regexp
}

// NewSolutionMemberParser creates a new instance of SolutionMemberParser.
func NewSolutionMemberParser() *SolutionMemberParser {
	return &SolutionMemberParser{
		attrRegex:   regexp.MustCompile(`^([+\-#~])?\s*([^:]+)\s*:\s*(.+)$`),
		methodRegex: regexp.MustCompile(`^([+\-#~])?\s*([^\(]+)\s*\((.*?)\)\s*(?::\s*(.+))?$`),
	}
}

// ParseAttribute transforms a raw string into a SolutionProcessedAttribute.
func (p *SolutionMemberParser) ParseAttribute(raw string, isEnumType bool) domain.SolutionProcessedAttribute {
	attr := domain.SolutionProcessedAttribute{
		Scope: "+",
		Kind:  "normal",
	}

	lowerRaw := strings.ToLower(raw)
	isStatic := strings.Contains(lowerRaw, "{static}") || strings.Contains(lowerRaw, "static")
	isFinal := strings.Contains(lowerRaw, "final") || strings.Contains(lowerRaw, "const") || strings.Contains(lowerRaw, "{readonly}")
	isAbstract := strings.Contains(lowerRaw, "{abstract}") || strings.Contains(lowerRaw, "abstract")

	working := cleanMemberString(raw)

	matches := p.attrRegex.FindStringSubmatch(working)
	if len(matches) > 0 {
		if matches[1] != "" {
			attr.Scope = matches[1]
		}
		attr.Names = splitOR(matches[2])
		typePart := strings.TrimSpace(matches[3])
		if idx := strings.Index(typePart, "="); idx != -1 {
			typePart = strings.TrimSpace(typePart[:idx])
		}
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
		if len(working) > 0 && isScopeChar(working[0]) {
			attr.Scope = string(working[0])
			working = strings.TrimSpace(working[1:])
		}
		attr.Names = splitOR(working)
		attr.Types = []string{}
	}

	// --- Logic: Detect final by Naming (ALL_CAPS or name=value) ---
	processedNames := make([]string, 0, len(attr.Names))
	for _, name := range attr.Names {
		cleanName := name
		// Case: name = value
		if idx := strings.Index(name, "="); idx != -1 {
			isFinal = true
			cleanName = strings.TrimSpace(name[:idx])
		}
		// Case: ALL_CAPS (excluding visibility symbols)
		if isAllUpperCase(strings.TrimLeft(cleanName, "+-#~ ")) {
			isFinal = true
		}
		processedNames = append(processedNames, cleanName)
	}
	attr.Names = processedNames

	if isStatic && isFinal {
		attr.Kind = "static-final"
	} else if isStatic {
		attr.Kind = "static"
	} else if isFinal {
		attr.Kind = "final"
	} else if isAbstract {
		attr.Kind = "abstract" // Although Kind usually only holds static/final for attributes
	}

	if (len(attr.Types) == 0 || (len(attr.Types) == 1 && attr.Types[0] == "")) && isEnumType {
		attr.Types = []string{"void"}
	}

	return attr
}

// ParseMethod transforms a raw string into a SolutionProcessedMethod.
func (p *SolutionMemberParser) ParseMethod(raw string, className string, attributes []domain.ProcessedAttribute, claimedG, claimedS map[string]bool) domain.SolutionProcessedMethod {
	method := domain.SolutionProcessedMethod{
		Scope:   "+",
		Names:   []string{raw},
		Type:    "",
		Outputs: []string{},
		Inputs:  []domain.MethodParam{},
		Kind:    "normal",
	}

	lowerRaw := strings.ToLower(raw)

	if strings.Contains(lowerRaw, "{static}") || strings.Contains(lowerRaw, "static") {
		method.Kind = "static"
	} else if strings.Contains(lowerRaw, "{abstract}") || strings.Contains(lowerRaw, "abstract") {
		method.Kind = "abstract"
	}

	if (strings.Contains(lowerRaw, "getter") || strings.Contains(lowerRaw, "setter")) && !strings.Contains(raw, "(") {
		attr := p.ParseAttribute(raw, false)
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

	working := cleanMemberString(raw)

	matches := p.methodRegex.FindStringSubmatch(working)
	if len(matches) > 0 {
		if matches[1] != "" {
			method.Scope = matches[1]
		}
		method.Names = splitOR(matches[2])
		paramStr := strings.TrimSpace(matches[3])
		if paramStr != "" {
			for _, part := range splitParams(paramStr) {
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
		if len(matches) > 4 && matches[4] != "" {
			method.Outputs = splitOR(matches[4])
		}
	} else {
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

	primaryName := ""
	if len(method.Names) > 0 {
		primaryName = method.Names[0]
	}
	lowerName := strings.ToLower(primaryName)

	if lowerName == strings.ToLower(className) || lowerName == "constructor" || lowerName == "init" {
		method.Type = "constructor"
		return method
	} else if len(method.Names) == 1 && strings.HasPrefix(lowerName, "get") &&
		(len(method.Inputs) == 0 || (len(method.Inputs) == 1 && strings.EqualFold(method.Inputs[0].Type, "void"))) {
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

	if len(method.Outputs) == 0 && method.Type != "constructor" {
		method.Outputs = []string{"void"}
	}

	return method
}

// GenerateGetter creates a solution getter method for a solution attribute.
func (p *SolutionMemberParser) GenerateGetter(attr domain.SolutionProcessedAttribute) domain.SolutionProcessedMethod {
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

// GenerateSetter creates a solution setter method for a solution attribute.
func (p *SolutionMemberParser) GenerateSetter(attr domain.SolutionProcessedAttribute) domain.SolutionProcessedMethod {
	baseName := ""
	if len(attr.Names) > 0 {
		baseName = attr.Names[0]
	}
	paramType := ""
	if len(attr.Types) > 0 {
		paramType = attr.Types[0]
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
