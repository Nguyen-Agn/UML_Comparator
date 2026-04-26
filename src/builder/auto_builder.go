package builder

import (
	"fmt"
	"uml_compare/domain"
	"uml_compare/src/builder/drawio"
	"uml_compare/src/builder/mermaid"
)

// AutoBuilder implements IModelBuilder by delegating to a registered
// concrete builder based on the sourceType.
type AutoBuilder struct {
	registry map[string]IModelBuilder
}

// Compile-time interface check.
var _ IModelBuilder = (*AutoBuilder)(nil)

// NewAutoBuilder returns an empty AutoBuilder.
func NewAutoBuilder() *AutoBuilder {
	return &AutoBuilder{
		registry: make(map[string]IModelBuilder),
	}
}

// NewAutoBuilderDefault returns an AutoBuilder pre-populated with
// standard builders (Drawio, etc.).
func NewAutoBuilderDefault() *AutoBuilder {
	ab := NewAutoBuilder()
	ab.Register("drawio", drawio.NewDrawioModelBuilder())
	ab.Register("mermaid", mermaid.NewMermaidModelBuilder())
	return ab
}

// Register adds a builder for a specific source type.
func (ab *AutoBuilder) Register(sourceType string, b IModelBuilder) {
	ab.registry[sourceType] = b
}

// Build delegates to the registered builder for the given sourceType.
func (ab *AutoBuilder) Build(rawData domain.RawModelData, sourceType string) (*domain.UMLGraph, error) {
	b, ok := ab.registry[sourceType]
	if !ok {
		return nil, fmt.Errorf("AutoBuilder: no builder registered for sourceType %q", sourceType)
	}
	return b.Build(rawData, sourceType)
}
