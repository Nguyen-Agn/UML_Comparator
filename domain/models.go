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
	Note         string // __1__
}

// NodeDiff represents a difference at the node level.
type NodeDiff struct {
	Sol         *SolutionProcessedNode
	Stu         *ProcessedNode
	Description string // Human-readable summary of the mismatch
}

// AttributeDiff represents a difference for a specific attribute of a node.
type AttributeDiff struct {
	ParentClassName string
	Sol             *SolutionProcessedAttribute
	Stu             *ProcessedAttribute
	Description     string
}

// MethodDiff represents a difference for a specific method of a node.
type MethodDiff struct {
	ParentClassName string
	Sol             *SolutionProcessedMethod
	Stu             *ProcessedMethod
	Description     string
}

// EdgeDiff represents a difference in a relationship.
type EdgeDiff struct {
	Sol         *ProcessedEdge
	Stu         *ProcessedEdge
	Description string
}

// DetailError groups differences by their entity type.
type DetailError struct {
	Class     []NodeDiff
	Method    []MethodDiff
	Attribute []AttributeDiff
	Edge      []EdgeDiff
}

// DiffReport contains the partitioned differences found between two UMLGraphs.
type DiffReport struct {
	// MissingDetail: items in solution but not in student (Stu will be nil).
	MissingDetail DetailError
	// WrongDetail: items in both but with mismatches.
	WrongDetail DetailError
	// ExtraDetail: items in student but not in solution (Sol will be nil).
	ExtraDetail DetailError
	// CorrectDetail: items that match perfectly (Described as "Match").
	CorrectDetail DetailError
}

// GradeResult contains the final score and text feedbacks.
type GradeResult struct {
	TotalScore     float64
	MaxScore       float64
	CorrectPercent float64
	Feedbacks      []string

	// Inputs retained for subsequent visualization pipeline
	Report        *DiffReport
	SolutionGraph *SolutionProcessedUMLGraph
	StudentGraph  *ProcessedUMLGraph
}

// MappedNode represents a connected student node with its similarity score.
type MappedNode struct {
	StudentID  string
	Similarity float64
}

// MappingTable holds the mapped nodes between Solution nodes and Student nodes.
type MappingTable map[string]MappedNode // Key: SolutionNodeID, Value: MappedNode
