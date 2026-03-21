package comparator

import "uml_compare/domain"

// IComparator defines the contract for computing the detailed differences using the MappingTable.
type IComparator interface {
	// Compare compares the detailed structure (attributes, methods, edges) of matched nodes.
	Compare(solution *domain.ProcessedUMLGraph, student *domain.ProcessedUMLGraph, mapping domain.MappingTable) (*domain.DiffReport, error)
}
