package mermaid

import (
	"strings"
	"uml_compare/domain"
)

// MermaidModelBuilder satisfies the builder.IModelBuilder interface.
type MermaidModelBuilder struct {
	parser *regexParser
}

// NewMermaidModelBuilder creates a new instance of MermaidModelBuilder.
func NewMermaidModelBuilder() *MermaidModelBuilder {
	return &MermaidModelBuilder{
		parser: &regexParser{},
	}
}

// Build converts raw Mermaid text into a domain.UMLGraph.
func (b *MermaidModelBuilder) Build(rawData domain.RawModelData, sourceType string) (*domain.UMLGraph, error) {
	text := string(rawData)

	rawClasses := b.parser.parseClasses(text)
	rawRelations := b.parser.parseRelations(text)

	graph := &domain.UMLGraph{
		ID:    "mermaid_graph",
		Nodes: make([]domain.UMLNode, 0, len(rawClasses)),
		Edges: make([]domain.UMLEdge, 0, len(rawRelations)),
	}

	// 1. Process Nodes
	for _, rc := range rawClasses {
		nodeType := b.detectType(rc.Stereotype)
		attrs, methods := b.splitMembers(rc.Members)

		graph.Nodes = append(graph.Nodes, domain.UMLNode{
			ID:         rc.Name, // Using Name as ID for simplicity in Mermaid
			Name:       rc.Name,
			Type:       nodeType,
			IsBold:     true, // Mermaid DSL doesn't usually specify boldness
			Attributes: attrs,
			Methods:    methods,
		})
	}

	// 2. Process Edges
	for _, rr := range rawRelations {
		graph.Edges = append(graph.Edges, domain.UMLEdge{
			SourceID:     rr.Source,
			TargetID:     rr.Target,
			RelationType: rr.Type,
		})
	}

	return graph, nil
}

func (b *MermaidModelBuilder) detectType(stereotype string) string {
	s := strings.ToLower(stereotype)
	switch s {
	case "interface":
		return "Interface"
	case "abstract":
		return "Abstract"
	case "enum", "enumeration":
		return "Enum"
	default:
		return "Class"
	}
}

func (b *MermaidModelBuilder) splitMembers(members []string) (attrs, methods []string) {
	for _, m := range members {
		if strings.Contains(m, "(") {
			methods = append(methods, m)
		} else {
			attrs = append(attrs, m)
		}
	}
	return attrs, methods
}
