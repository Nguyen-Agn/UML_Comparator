package builder

import (
	"fmt"
	"uml_compare/domain"
)

// ─────────────────────────────────────────────────────────────────────────────
// StandardModelBuilder — IModelBuilder concrete implementation.
//
// SOLID responsibilities:
//   - Single Responsibility: Build() is a pure orchestrator. All domain logic
//     lives in the dedicated helper structs below.
//   - Open/Closed: new UML element types extend typeDetector, not this file.
//   - Liskov: satisfies IModelBuilder contract (compile-time check below).
//   - Interface Segregation: clients depend only on IModelBuilder.
//   - Dependency Inversion: Build() depends on the sub-component interfaces,
//     not on concrete regex or XML details.
// ─────────────────────────────────────────────────────────────────────────────

// Compile-time interface guarantee (template.md rule #1)
var _ IModelBuilder = (*DrawioModelBuilder)(nil)

// DrawioModelBuilder orchestrates the XML → UMLGraph transformation for Draw.io files.
// Each field is an interface (DIP): the orchestrator depends only on
// abstractions, never on concrete implementations.
type DrawioModelBuilder struct {
	parser  IXMLParser     // XML parsing + structural cell queries
	san     ITextSanitizer // text cleaning + name extraction
	types   ITypeDetector  // node type + relation type classification
	members IMemberParser  // attribute vs method classification
	style   IStyleHelper   // style extraction
}

// NewDrawioModelBuilder (formerly NewStandardModelBuilder) wires all sub-components
// and returns the builder as its IModelBuilder interface.
func NewDrawioModelBuilder() IModelBuilder {
	san := newHTMLSanitizer()
	style := NewStyleHelper()
	return &DrawioModelBuilder{
		parser:  &cellParser{},
		san:     san,
		types:   &typeDetector{san: san},
		members: &memberParser{san: san, style: style},
		style:   style,
	}
}

// NewStandardModelBuilder is a compatibility alias for NewDrawioModelBuilder.
func NewStandardModelBuilder() IModelBuilder {
	return NewDrawioModelBuilder()
}

// Build converts RawModelData into a fully structured *domain.UMLGraph.
//
// Pipeline:
//  1. Parse XML → []mxCell
//  2. Build structural indexes (cell map, root layer, children groups)
//  3. Build Nodes from top-level container cells
//  4. Build Edges, resolving child-cell endpoints to top-level class IDs
func (b *DrawioModelBuilder) Build(rawData domain.RawModelData) (*domain.UMLGraph, error) {
	if rawData == "" {
		return nil, fmt.Errorf("DrawioModelBuilder.Build: rawData cannot be empty")
	}

	// Step 1: XML → cells
	cells, err := b.parser.parse(string(rawData))
	if err != nil {
		return nil, fmt.Errorf("DrawioModelBuilder.Build: %w", err)
	}

	// Step 2: indexes
	rootLayerID := b.parser.findRootLayerID(cells)
	cellMap := b.parser.buildCellMap(cells)
	childrenByParent := b.parser.groupChildrenByParent(cells, rootLayerID)

	graph := &domain.UMLGraph{
		ID:    "graph",
		Nodes: []domain.UMLNode{},
		Edges: []domain.UMLEdge{},
	}

	// Step 3: nodes
	classIDSet := make(map[string]bool)
	for _, cell := range cells {
		if !b.parser.isTopLevelNode(cell, rootLayerID) {
			continue
		}
		children := childrenByParent[cell.ID]
		node := b.buildNode(cell, children)
		graph.Nodes = append(graph.Nodes, node)
		classIDSet[cell.ID] = true
	}

	// Step 4: edges
	for _, cell := range cells {
		if !b.parser.isEdge(cell) {
			continue
		}
		srcID := b.parser.resolveToClassID(cell.Source, cellMap, classIDSet, rootLayerID)
		tgtID := b.parser.resolveToClassID(cell.Target, cellMap, classIDSet, rootLayerID)
		graph.Edges = append(graph.Edges, domain.UMLEdge{
			SourceID:     srcID,
			TargetID:     tgtID,
			RelationType: b.types.relationType(cell.Style),
			Note:         b.san.clean(cell.Value),
		})
	}

	return graph, nil
}

// buildNode assembles a single UMLNode from a container cell and its children.
func (b *DrawioModelBuilder) buildNode(container mxCell, children []mxCell) domain.UMLNode {
	// decodeOnly: entities decoded but HTML tags + stereotypes still present.
	// Used ONLY for type detection and name/format extraction.
	rawDecoded := b.san.decodeOnly(container.Value)

	// properly extract the name from the first line and check if it was bolded
	name, isHtmlBold := b.san.extractNameAndFormat(rawDecoded)
	isBold := isHtmlBold || b.style.IsStyleBitSet(container.Style, "fontStyle", 1)

	// clean: fully sanitized text with HTML tags + stereotypes removed.
	// Now used primarily for type detection (if it relies on clean text) or child detection,
	// though member parser cleans child.Value on its own.
	
	// Type: detected from raw decoded value (stereotype visible), fallback child pattern
	nodeType := b.types.nodeType(container.Style, rawDecoded)
	if nodeType == "Class" && b.types.isEnumByChildPattern(children) {
		nodeType = "Enum"
	}

	attrs, methods := b.members.parseChildren(children)

	return domain.UMLNode{
		ID:         container.ID,
		Name:       name,
		IsBold:     isBold,
		Type:       nodeType,
		Attributes: attrs,
		Methods:    methods,
	}
}
