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
