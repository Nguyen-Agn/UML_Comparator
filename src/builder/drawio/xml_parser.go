package drawio

import (
	"encoding/xml"
	"fmt"
)

// ─────────────────────────────────────────────────────────────────────────────
// Draw.io XML structs — encoding/xml unmarshalling targets
// Responsibility: purely structural, no UML or business logic.
// ─────────────────────────────────────────────────────────────────────────────

type mxGraphModel struct {
	Root mxRoot `xml:"root"`
}

type mxRoot struct {
	Cells []mxCell `xml:"mxCell"`
}

type mxCell struct {
	ID     string `xml:"id,attr"`
	Value  string `xml:"value,attr"`
	Style  string `xml:"style,attr"`
	Vertex string `xml:"vertex,attr"`
	Edge   string `xml:"edge,attr"`
	Source string `xml:"source,attr"`
	Target string `xml:"target,attr"`
	Parent string `xml:"parent,attr"`
}

// ─────────────────────────────────────────────────────────────────────────────
// cellParser — SRP: parse XML bytes into mxCell slices and provide
//              structural queries (root layer, cell map, children grouping).
// ─────────────────────────────────────────────────────────────────────────────

type cellParser struct{}

// Compile-time interface satisfaction check.
var _ IXMLParser = (*cellParser)(nil)

// parse unmarshals the raw XML string into a flat slice of mxCells.
func (p *cellParser) parse(raw string) ([]mxCell, error) {
	var model mxGraphModel
	if err := xml.Unmarshal([]byte(raw), &model); err != nil {
		return nil, fmt.Errorf("cellParser.parse: xml.Unmarshal: %w", err)
	}
	return model.Root.Cells, nil
}

// findRootLayerID returns the ID of the layer cell whose parent is "0".
// This is typically id="1" in all Draw.io files.
func (p *cellParser) findRootLayerID(cells []mxCell) string {
	for _, c := range cells {
		if c.Parent == "0" && c.ID != "0" {
			return c.ID
		}
	}
	return "1" // safe fallback
}

// buildCellMap creates an O(1) lookup from cell ID → mxCell.
func (p *cellParser) buildCellMap(cells []mxCell) map[string]mxCell {
	m := make(map[string]mxCell, len(cells))
	for _, c := range cells {
		m[c.ID] = c
	}
	return m
}

// groupChildrenByParent groups all non-root-layer cells by their parent ID.
// Used to collect attributes/methods that belong to each class container.
func (p *cellParser) groupChildrenByParent(cells []mxCell, rootLayerID string) map[string][]mxCell {
	groups := make(map[string][]mxCell)
	for _, c := range cells {
		if c.Parent == "0" || c.Parent == rootLayerID || c.ID == "0" || c.ID == rootLayerID {
			continue
		}
		groups[c.Parent] = append(groups[c.Parent], c)
	}
	return groups
}

// isTopLevelNode returns true for cells that are direct children of the root
// layer and represent Class/Interface/Actor/Enum containers.
func (p *cellParser) isTopLevelNode(c mxCell, rootLayerID string) bool {
	return c.Vertex == "1" && c.Parent == rootLayerID && c.Edge != "1"
}

// isEdge returns true for cells that are relationship arrows with valid endpoints.
func (p *cellParser) isEdge(c mxCell) bool {
	return c.Edge == "1" && c.Source != "" && c.Target != ""
}

// resolveToClassID walks up the parent chain (max 5 hops) to find the
// top-level class container ID for a given cell (handles edge → child cell).
func (p *cellParser) resolveToClassID(cellID string, cellMap map[string]mxCell, classIDSet map[string]bool, rootLayerID string) string {
	current := cellID
	for i := 0; i < 5; i++ {
		if classIDSet[current] {
			return current
		}
		cell, ok := cellMap[current]
		if !ok || cell.Parent == "0" || cell.Parent == rootLayerID {
			break
		}
		current = cell.Parent
	}
	return cellID // fallback: return original
}
