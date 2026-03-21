package builder

import "uml_compare/domain"

// IModelBuilder defines the contract for converting raw source data into a standardized UMLGraph.
type IModelBuilder interface {
	// Build parses the structural raw data into a UMLGraph.
	Build(rawData domain.RawModelData) (*domain.UMLGraph, error)
}
