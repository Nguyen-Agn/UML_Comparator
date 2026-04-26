package builder

import "uml_compare/domain"

// IModelBuilder defines the contract for converting raw source data into a standardized UMLGraph.
type IModelBuilder interface {
	// Build parses the structural raw data into a UMLGraph.
	// sourceType provides additional context (e.g., "drawio", "mermaid").
	Build(rawData domain.RawModelData, sourceType string) (*domain.UMLGraph, error)
}
