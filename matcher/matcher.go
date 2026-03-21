package matcher

import "uml_compare/domain"

// IEntityMatcher defines the contract for matching nodes between the solution and the student's diagram.
type IEntityMatcher interface {
	// Match compares two graphs and produces a dictionary mapping solution node IDs to student node IDs.
	Match(solution *domain.ProcessedUMLGraph, student *domain.ProcessedUMLGraph) (domain.MappingTable, error)
}
