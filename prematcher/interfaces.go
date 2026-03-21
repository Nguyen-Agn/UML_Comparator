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
