package domain

// SolutionProcessedUMLGraph is an OR-aware UML graph produced by UMLSolutionPreMatcher.
// It is designed for solution files where polymorphism is expressed via "|" patterns.
type SolutionProcessedUMLGraph struct {
	Nodes []SolutionProcessedNode
	Edges []ProcessedEdge // Reuse existing UMLEdge
	GradingConfig ScoreConfig
}

// ScoreConfig maps the entity identifier to its points.
// Key format: "NodeName", "NodeName::AttributeName", or "NodeName::MethodName"
type ScoreConfig struct {
	Nodes      map[string]float64
	Attributes map[string]float64
	Methods    map[string]float64
	Edges      map[string]float64
}

// SolutionProcessedNode holds OR-aware attributes and methods for a single UML node.
type SolutionProcessedNode struct {
	ID         string
	Name       string
	Type       string   // Class, Interface, Abstract, Enum, etc.
	ArchWeight uint32   // Bitwise architecture descriptor (same semantics as ProcessedNode)
	Shortcut   uint32   // Bit 0=getters generated, Bit 1=setters generated
	Inherits   string   // ID of the parent class (from Generalization/Inheritance edge)
	Implements []string // IDs of implemented interfaces (from Realization/Implementation edges)
	Attributes []SolutionProcessedAttribute
	Methods    []SolutionProcessedMethod
	Score      float64  // Point value for this class/node
}

// SolutionProcessedAttribute supports OR-patterns for both name and type.
// Example: "+ x | y : int|long" -> Names=["x","y"], Types=["int","long"]
type SolutionProcessedAttribute struct {
	// Names holds one or more attribute name alternatives separated by "|" in the source.
	// For a normal attribute, len(Names)==1.
	Names []string

	// Scope is the UML visibility modifier: "+", "-", "#", "~".
	Scope string

	// Types holds one or more type alternatives separated by "|" in the source.
	// For a normal attribute, len(Types)==1.
	Types []string

	// Kind describes the attribute modifier: "normal", "static", "final", "static-final".
	Kind string

	// Score is the point value extracted from __d__
	Score float64
}

// SolutionMethodParam represents a single method parameter with OR-aware type alternatives.
// Example: "a: double|Double|float|Float" -> Name="a", Types=["double","Double","float","Float"]
type SolutionMethodParam struct {
	// Name is the parameter name.
	Name string

	// Types holds one or more type alternatives separated by "|" in the source.
	// For a single-type param, len(Types)==1.
	Types []string
}

// SolutionProcessedMethod supports OR-patterns for method name, return type, and param types.
//
// Example: "setPrice|setCost(a: double|Double): void|String"
//   -> Names=["setPrice","setCost"], Outputs=["void","String"],
//      Inputs=[{Name:"a", Types:["double","Double"]}]
type SolutionProcessedMethod struct {
	// Names holds one or more method name alternatives separated by "|" in the source.
	// For a normal method, len(Names)==1.
	Names []string

	// Scope is the UML visibility modifier: "+", "-", "#", "~".
	Scope string

	// Type classifies the method role: "constructor", "getter", "setter", "custom".
	Type string

	// Outputs holds one or more return type alternatives separated by "|" in the source.
	// For a normal method, len(Outputs)==1. Constructors have an empty Outputs slice.
	Outputs []string

	// Inputs are the method parameters with OR-aware type alternatives.
	// Each SolutionMethodParam.Types may have multiple alternatives split on "|".
	Inputs []SolutionMethodParam

	// Kind describes the method modifier: "normal", "static", "abstract".
	Kind string

	// Score is the point value extracted from __d__
	Score float64
}
