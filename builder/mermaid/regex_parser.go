package mermaid

import (
	"regexp"
	"strings"
)

type rawClass struct {
	Name       string
	Stereotype string
	Members    []string
}

type rawRelation struct {
	Source string
	Target string
	Type   string
}

// regexParser extracts raw components from Mermaid text using regex.
type regexParser struct{}

func (p *regexParser) normalizeGenerics(s string) string {
	// Mermaid uses ~ as a substitute for < > in generics.
	// We use an iterative approach to handle nested generics like List~Set~T~~.
	// The regex look for an identifier followed by ~...~.
	re := regexp.MustCompile(`([a-zA-Z0-9_]*)~([^~]+)~`)
	for {
		next := re.ReplaceAllString(s, "$1<$2>")
		if next == s {
			break
		}
		s = next
	}
	return s
}

func (p *regexParser) parseClasses(text string) []rawClass {
	var classes []rawClass

	// 1. Matches class Name { ... }
	// Match until a closing brace that is followed by a newline or end of string.
	// This helps avoid stopping at internal braces like {static}.
	blockRegex := regexp.MustCompile(`class\s+([a-zA-Z0-9_~]+)\s*\{([\s\S]*?)\n\s*\}`)
	blocks := blockRegex.FindAllStringSubmatch(text, -1)
	for _, match := range blocks {
		name := p.normalizeGenerics(match[1])
		content := match[2]
		
		var stereotype string
		var members []string
		
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if strings.HasPrefix(line, "<<") && strings.HasSuffix(line, ">>") {
				stereotype = strings.Trim(line, "<>")
				continue
			}
			members = append(members, p.normalizeGenerics(line))
		}
		
		classes = append(classes, rawClass{
			Name:       name,
			Stereotype: stereotype,
			Members:    members,
		})
	}

	// 2. Matches class Name <<stereotype>> (if not already matched in block)
	standaloneRegex := regexp.MustCompile(`class\s+([a-zA-Z0-9_~]+)\s*(?:<<([a-zA-Z0-9_]+)>>)?\s*($|\n)`)
	standalones := standaloneRegex.FindAllStringSubmatch(text, -1)
	for _, match := range standalones {
		name := p.normalizeGenerics(match[1])
		stereo := match[2]
		
		// Avoid duplicate if already found in block
		found := false
		for _, c := range classes {
			if c.Name == name {
				found = true
				break
			}
		}
		if !found {
			classes = append(classes, rawClass{
				Name:       name,
				Stereotype: stereo,
			})
		}
	}

	return classes
}

func (p *regexParser) parseRelations(text string) []rawRelation {
	var relations []rawRelation

	// Improved regex for relationships: ensures it's a standalone line and supports generics (~).
	// Regex for relationships like: A <|-- B or A --> B
	relRegex := regexp.MustCompile(`(?m)^\s*([a-zA-Z0-9_~]+)\s*([<|o\*]*[-.]+([>|o\*|]*))\s*([a-zA-Z0-9_~]+)(?:\s*:\s*(.*))?$`)
	matches := relRegex.FindAllStringSubmatch(text, -1)
	
	for _, match := range matches {
		src := p.normalizeGenerics(match[1])
		arrow := match[2]
		tgt := p.normalizeGenerics(match[4])
		// match[5] is the label if present
		
		relType := p.mapArrowToRelation(arrow)
		relations = append(relations, rawRelation{
			Source: src,
			Target: tgt,
			Type:   relType,
		})
	}

	return relations
}

func (p *regexParser) mapArrowToRelation(arrow string) string {
	switch {
	case strings.Contains(arrow, "<|--") || strings.Contains(arrow, "--|>"):
		return "Inheritance"
	case strings.Contains(arrow, "..|>") || strings.Contains(arrow, "<|.."):
		return "Realization"
	case strings.Contains(arrow, "*--") || strings.Contains(arrow, "--*"):
		return "Composition"
	case strings.Contains(arrow, "o--") || strings.Contains(arrow, "--o"):
		return "Aggregation"
	case strings.Contains(arrow, "..>") || strings.Contains(arrow, "<.."):
		return "Dependency"
	case strings.Contains(arrow, "-->") || strings.Contains(arrow, "<--") || strings.Contains(arrow, "--"):
		return "Association"
	default:
		return "Association"
	}
}
