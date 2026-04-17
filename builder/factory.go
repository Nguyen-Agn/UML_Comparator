package builder



// GetBuilder returns the appropriate IModelBuilder based on the source type.
// This implements the Strategy Pattern for builder selection.
func GetBuilder(sourceType string) (IModelBuilder, error) {
	// For now, we return the AutoBuilder orchestrator.
	// In the future, we could also return specific builders directly if needed.
	return NewAutoBuilderDefault(), nil
}

// NewStandardModelBuilder is a compatibility factory for the main application flow.
// It returns an IModelBuilder that can handle multiple source types via AutoBuilder.
func NewStandardModelBuilder() IModelBuilder {
	return NewAutoBuilderDefault()
}
