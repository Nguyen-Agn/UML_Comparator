package comparator

import "uml_compare/domain"

// IComparator defines the contract for computing the detailed differences between two UML graphs.
type IComparator interface {
	// Compare performs a deep comparison between the solution and student graphs using the provided mapping.
	// It categorizes results into Missing (in solution only), Wrong (in both but mismatched),
	// Extra (in student only), and Correct (perfect match) details.
	// Labels for members/edges should follow the format: "Class 'ClassName': [Issue description]".
	Compare(solution *domain.ProcessedUMLGraph, student *domain.ProcessedUMLGraph, mapping domain.MappingTable) (*domain.DiffReport, error)
}
