package domain

// RawModelData represents the raw data (XML, JSON, etc.) from a source.
type RawModelData string

// RawXMLData is a legacy alias for RawModelData, specifically for XML sources.
type RawXMLData = RawModelData

// UMLGraph represents the structured UML model built from XML.
type UMLGraph struct {
	ID    string
	Nodes []UMLNode
	Edges []UMLEdge
}

// UMLNode represents a UML entity like a Class, Interface, etc.
type UMLNode struct {
	ID         string
	Name       string
	Type       string // e.g., "Class", "Interface", "Actor"
	Attributes []string
	Methods    []string
}

// UMLEdge represents a relationship between two nodes.
type UMLEdge struct {
	SourceID     string
	TargetID     string
	RelationType string // e.g., "Inheritance", "Association"
	SourceLabel  string
	TargetLabel  string
}

// DiffReport contains the partitioned differences and missing items found between two UMLGraphs.
type DiffReport struct {
	// Missed items (Solution has them, Student does not)
	MissedClass    []string
	MissingNodes   []string
	MissingEdges   []string
	MissingMembers []string

	// Wrong/Different items (Found but details mismatch)
	AttributeErrors []string
	MethodErrors    []string
	NodeEdgeErrors  []string
}

// GradeResult contains the final score and text feedbacks.
type GradeResult struct {
	TotalScore float64
	Feedbacks  []string
}

// MappedNode represents a connected student node with its similarity score.
type MappedNode struct {
	StudentID  string
	Similarity float64
}

// MappingTable holds the mapped nodes between Solution nodes and Student nodes.
type MappingTable map[string]MappedNode // Key: SolutionNodeID, Value: MappedNode
