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

// ITypeAnalyzer handles type translation and deep comparison including generics.
type ITypeAnalyzer interface {
	// TranslateType converts a solution type name to its mapped student name if it exists.
	TranslateType(typeName string, typeMap map[string]string) string
	// CompareTypes checks if two type strings are compatible, considering generics and type mapping.
	CompareTypes(solType, stuType string, typeMap map[string]string) bool
}

// IMemberComparator handles comparison of attributes and methods within a node.
type IMemberComparator interface {
	// CompareAttributes identifies differences in attributes between solution and student nodes.
	CompareAttributes(sol, stu domain.ProcessedNode, typeMap map[string]string, report *domain.DiffReport)
	// CompareMethods identifies differences in methods between solution and student nodes.
	CompareMethods(sol, stu domain.ProcessedNode, typeMap map[string]string, report *domain.DiffReport)
}

// IEdgeComparator handles comparison of relationships between nodes.
type IEdgeComparator interface {
	// CompareEdges identifies differences in relationships between solution and student graphs.
	CompareEdges(solution, student *domain.ProcessedUMLGraph, mapping domain.MappingTable, report *domain.DiffReport)
}
