package builder

// ─────────────────────────────────────────────────────────────────────────────
// Internal sub-component interfaces for the Builder module.
//
// Principle: StandardModelBuilder depends ONLY on these abstractions (DIP).
// Each concrete implementation lives in its own file and is injected via
// NewStandardModelBuilder(). This enables independent testing with mocks.
// ─────────────────────────────────────────────────────────────────────────────

// IXMLParser is the contract for XML → mxCell structural parsing.
// Responsibility: read raw Draw.io XML and provide structural queries.
type IXMLParser interface {
	parse(raw string) ([]mxCell, error)
	findRootLayerID(cells []mxCell) string
	buildCellMap(cells []mxCell) map[string]mxCell
	groupChildrenByParent(cells []mxCell, rootLayerID string) map[string][]mxCell
	isTopLevelNode(c mxCell, rootLayerID string) bool
	isEdge(c mxCell) bool
	resolveToClassID(cellID string, cellMap map[string]mxCell, classIDSet map[string]bool, rootLayerID string) string
}

// ITextSanitizer is the contract for HTML/entity decoding and text cleaning.
// Responsibility: convert raw Draw.io cell value strings to plain text.
type ITextSanitizer interface {
	clean(raw string) string
	decodeOnly(raw string) string
	extractCleanName(sanitized string) string
	normalizeSignature(sig string) string
}

// ITypeDetector is the contract for UML node and relation type classification.
// Responsibility: map Draw.io style + value text to UML type strings.
type ITypeDetector interface {
	nodeType(style, decodedValue string) string
	relationType(style string) string
	isEnumByChildPattern(children []mxCell) bool
}

// IMemberParser is the contract for extracting Attributes and Methods
// from a list of child mxCell elements belonging to a class container.
type IMemberParser interface {
	parseChildren(children []mxCell) (attrs, methods []string)
}
