package domain

// ProcessedUMLGraph represents the enhanced, struct-based UML model
// intended for precise matching.
type ProcessedUMLGraph struct {
	Nodes []ProcessedNode
	Edges []ProcessedEdge
}

// ProcessedNode contains detailed structural info including bitwise architecture weight.
type ProcessedNode struct {
	ID         string
	Name       string
	IsBold     bool     // true if the student formatted the node name as bold
	Type       string   // Class, Interface, etc.
	ArchWeight uint32   // Bitwise description of architecture (e.g., has Singleton, has Factory)
	Shortcut   uint32   // Bitwise: Bit 0 (1): getters, Bit 1 (2): setters
	Inherits   string   // ID of the parent class
	Implements []string // IDs of implemented interfaces
	Attributes []ProcessedAttribute
	Methods    []ProcessedMethod
}

// ProcessedAttribute distinguishes name, scope, and type.
type ProcessedAttribute struct {
	Name  string
	Scope string // +, -, #
	Type  string
	Kind  string // static, final, normal, static-final,
}

// ProcessedMethod provides granular signature info.
type ProcessedMethod struct {
	Name   string
	Scope  string // +, -, #
	Type   string // Original full type string
	Output string // Formal return type
	Inputs []MethodParam
	Kind   string // static, normal, abstract
}

// MethodParam represents a single method parameter.
type MethodParam struct {
	Name string
	Type string
}

// ProcessedEdge (re-use or wrap UMLEdge if needed, keeping it simple for now)
type ProcessedEdge = UMLEdge
