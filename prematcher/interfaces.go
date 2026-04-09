package prematcher

import "uml_compare/domain"

// IPreMatcher defines the contract for transforming a string-based UMLGraph
// into a struct-based ProcessedUMLGraph.
//
// Responsibilities:
//   - Parse string attributes/methods into structured fields (Name, Scope, Type).
//   - Calculate ArchWeight (bitwise) for nodes.
//   - Resolve structural relations like Inherits and Implements.
type IPreMatcher interface {
	Process(graph *domain.UMLGraph) (*domain.ProcessedUMLGraph, error)
}

// IUMLSolutionPreMatcher processes a UMLGraph into an OR-aware SolutionProcessedUMLGraph.
// OR-pattern ("|") is supported for:
//   - Attribute Names  (e.g. "x | y : int")
//   - Attribute Types  (e.g. "- a : int|long")
//   - Custom Method Names (e.g. "doA | doB() : void")
//   - Method Return Types (e.g. "doing(a:int): void|boolean")
//
// Param types are NOT split on "|"; they are kept as-is in MethodParam.Type.
type IUMLSolutionPreMatcher interface {
	ProcessSolution(graph *domain.UMLGraph) (*domain.SolutionProcessedUMLGraph, error)
}

// ITypeDetector identifies the stereotype and standardizes the node type.
type ITypeDetector interface {
	// NormalizeNodeType returns a standardized type name (e.g., "Enum", "Interface").
	NormalizeNodeType(t string) string
	// IsEnumType returns true if the type string represents an enumeration.
	IsEnumType(t string) bool
}

// IWeightCalculator computes the bitwise architecture weight for a UML node.
type IWeightCalculator interface {
	// Calculate packs structural information into a 32-bit integer.
	Calculate(nodeType string, hasInheritance bool, numInterfaces, numMethods, numAttributes, numRelated, numCustomTypes, numStaticMembers int) uint32
}

// IScoreExtractor pulls out point values from raw UML strings.
type IScoreExtractor interface {
	// ExtractScore parses the __d__ pattern and returns the cleaned string and the score value.
	ExtractScore(raw string) (string, float64)
}

// IMemberParser parses standard member strings (attributes and methods).
type IMemberParser interface {
	// ParseAttribute transforms a raw string into a ProcessedAttribute.
	ParseAttribute(raw string, isEnumType bool) domain.ProcessedAttribute
	// ParseMethod transforms a raw string into a ProcessedMethod.
	ParseMethod(raw string, className string, attributes []domain.ProcessedAttribute, claimedG, claimedS map[string]bool) domain.ProcessedMethod
	// GenerateGetter creates a synthetic getter method for an attribute.
	GenerateGetter(attr domain.ProcessedAttribute) domain.ProcessedMethod
	// GenerateSetter creates a synthetic setter method for an attribute.
	GenerateSetter(attr domain.ProcessedAttribute) domain.ProcessedMethod
}

// ISolutionMemberParser parses solution member strings with support for OR-patterns.
type ISolutionMemberParser interface {
	// ParseAttribute transforms a raw string into a SolutionProcessedAttribute.
	ParseAttribute(raw string, isEnumType bool) domain.SolutionProcessedAttribute
	// ParseMethod transforms a raw string into a SolutionProcessedMethod.
	ParseMethod(raw string, className string, attributes []domain.ProcessedAttribute, claimedG, claimedS map[string]bool) domain.SolutionProcessedMethod
	// GenerateGetter creates a synthetic solution getter method for a solution attribute.
	GenerateGetter(attr domain.SolutionProcessedAttribute) domain.SolutionProcessedMethod
	// GenerateSetter creates a synthetic solution setter method for a solution attribute.
	GenerateSetter(attr domain.SolutionProcessedAttribute) domain.SolutionProcessedMethod
}
