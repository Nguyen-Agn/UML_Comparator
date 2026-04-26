package drawio_test

import (
	"strings"
	"testing"
	"uml_compare/domain"
	"uml_compare/src/builder/drawio"
)

// ─── Inline test XML (shared) ─────────────────────────────────────────────────

// sampleXML is a minimal Draw.io swimlane-based UML diagram with:
//   - Animal (class, 1 attr, 1 method)
//   - Dog    (class, 0 attrs, 1 method)
//   - Dog → Animal (Inheritance edge)
const sampleXML = `<mxGraphModel dx="1290" dy="687" grid="1" gridSize="10">
  <root>
    <mxCell id="0" />
    <mxCell id="1" parent="0" />
    <mxCell id="2" value="Animal" style="shape=umlClass;swimlaneHead=0;swimlaneBody=0;startSize=26;container=1;" vertex="1" parent="1">
      <mxGeometry x="100" y="100" width="180" height="90" as="geometry" />
    </mxCell>
    <mxCell id="3" value="- name : String" style="text;strokeColor=none;" vertex="1" parent="2">
      <mxGeometry y="26" width="180" height="30" as="geometry" />
    </mxCell>
    <mxCell id="4" value="+ getName() : String" style="text;strokeColor=none;" vertex="1" parent="2">
      <mxGeometry y="56" width="180" height="30" as="geometry" />
    </mxCell>
    <mxCell id="5" value="Dog" style="shape=umlClass;startSize=26;container=1;" vertex="1" parent="1">
      <mxGeometry x="100" y="260" width="180" height="60" as="geometry" />
    </mxCell>
    <mxCell id="6" value="+ bark() : void" style="text;strokeColor=none;" vertex="1" parent="5">
      <mxGeometry y="26" width="180" height="30" as="geometry" />
    </mxCell>
    <mxCell id="7" value="" style="endArrow=block;endFill=1;" edge="1" source="5" target="2" parent="1">
      <mxGeometry relative="1" as="geometry" />
    </mxCell>
  </root>
</mxGraphModel>`

// ─── Integration tests (Build pipeline) ──────────────────────────────────────

func TestBuild_NodeCount(t *testing.T) {
	b := drawio.NewDrawioModelBuilder()
	graph, err := b.Build(domain.RawModelData(sampleXML), "drawio")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(graph.Nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d: %+v", len(graph.Nodes), graph.Nodes)
	}
	t.Logf("✔ node count: %d", len(graph.Nodes))
}

func TestBuild_EdgeCount_And_RelationType(t *testing.T) {
	b := drawio.NewDrawioModelBuilder()
	graph, err := b.Build(domain.RawModelData(sampleXML), "drawio")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(graph.Edges) != 1 {
		t.Errorf("expected 1 edge, got %d", len(graph.Edges))
	}
	if got := graph.Edges[0].RelationType; got != "Inheritance" {
		t.Errorf("expected RelationType=Inheritance, got %q", got)
	}
	t.Logf("✔ edge: %s", graph.Edges[0].RelationType)
}

func TestBuild_EmptyInput_ReturnsError(t *testing.T) {
	b := drawio.NewDrawioModelBuilder()
	_, err := b.Build("", "drawio")
	if err == nil {
		t.Fatal("expected error for empty input, got nil")
	}
	t.Logf("✔ empty input rejected: %v", err)
}

func TestBuild_HTMLSanitize_SingleCell(t *testing.T) {
	// Class defined with all members in the container value (single-cell style)
	htmlXML := `<mxGraphModel><root>
	  <mxCell id="0" /><mxCell id="1" parent="0"/>
	  <mxCell id="10" value="&lt;b&gt;Account&lt;/b&gt;&lt;br/&gt;- id : int&lt;br/&gt;+ login() : bool" style="shape=umlClass;" vertex="1" parent="1">
	    <mxGeometry x="0" y="0" width="200" height="80" as="geometry" />
	  </mxCell>
	</root></mxGraphModel>`

	b := drawio.NewDrawioModelBuilder()
	graph, err := b.Build(domain.RawModelData(htmlXML), "drawio")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(graph.Nodes) == 0 {
		t.Fatal("expected at least one node")
	}
	node := graph.Nodes[0]
	if !strings.Contains(node.Name, "Account") {
		t.Errorf("expected name containing 'Account', got %q", node.Name)
	}
	t.Logf("✔ sanitize OK: Name=%s Attrs=%v Methods=%v", node.Name, node.Attributes, node.Methods)
}

// ─── Interface stereotype tests ───────────────────────────────────────────────

func TestBuild_Interface_FromValue_Stereotype(t *testing.T) {
	xml := `<mxGraphModel><root>
	  <mxCell id="0"/><mxCell id="1" parent="0"/>
	  <mxCell id="2" value="&lt;&lt;interface&gt;&gt;&#10;IShape" style="swimlane;" vertex="1" parent="1">
	    <mxGeometry width="140" height="60" as="geometry"/>
	  </mxCell>
	  <mxCell id="3" value="+ draw(): void" style="text;" vertex="1" parent="2">
	    <mxGeometry y="26" width="140" height="26" as="geometry"/>
	  </mxCell>
	</root></mxGraphModel>`

	b := drawio.NewDrawioModelBuilder()
	graph, err := b.Build(domain.RawModelData(xml), "drawio")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(graph.Nodes) == 0 {
		t.Fatal("expected 1 node")
	}
	n := graph.Nodes[0]
	if n.Type != "Interface" {
		t.Errorf("expected Type=Interface, got %q", n.Type)
	}
	if n.Name != "IShape" {
		t.Errorf("expected Name=IShape (stereotype stripped), got %q", n.Name)
	}
	t.Logf("✔ Interface from value stereotype: Name=%s Type=%s", n.Name, n.Type)
}

// ─── Abstract stereotype test ─────────────────────────────────────────────────

func TestBuild_Abstract_InlineStereotype(t *testing.T) {
	// Real .drawio: value="&amp;lt;&amp;lt;abstract&amp;gt;&amp;gt; BankAccount"
	// xml.Unmarshal decodes &amp; → & so value becomes "&lt;&lt;abstract&gt;&gt; BankAccount"
	// sanitizeHTML step-1 decodes &lt; → < giving "<<abstract>> BankAccount"
	xml := `<mxGraphModel><root>
	  <mxCell id="0"/><mxCell id="1" parent="0"/>
	  <mxCell id="2" value="&amp;lt;&amp;lt;abstract&amp;gt;&amp;gt; BankAccount" style="swimlane;" vertex="1" parent="1">
	    <mxGeometry width="200" height="60" as="geometry"/>
	  </mxCell>
	</root></mxGraphModel>`

	b := drawio.NewDrawioModelBuilder()
	graph, err := b.Build(domain.RawModelData(xml), "drawio")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(graph.Nodes) == 0 {
		t.Fatal("expected 1 node")
	}
	n := graph.Nodes[0]
	if n.Type != "Abstract" {
		t.Errorf("expected Type=Abstract, got %q", n.Type)
	}
	if n.Name != "BankAccount" {
		t.Errorf("expected Name=BankAccount (inline stereotype stripped), got %q", n.Name)
	}
	t.Logf("✔ Abstract inline: Name=%s Type=%s", n.Name, n.Type)
}

// ─── Numeric entity decode test ───────────────────────────────────────────────

func TestBuild_NumericEntity_Newline(t *testing.T) {
	// Verifies that when the stereotype and class name are on separate lines —
	// as commonly produced by Draw.io's HTML encoding — the name is extracted correctly.
	// We use &lt;br/&gt; (decoded by xml.Unmarshal to literal "<br/>") which
	// sanitizeHTML then replaces with "\n" before line-splitting.
	xmlStr := `<mxGraphModel><root>
	  <mxCell id="0"/><mxCell id="1" parent="0"/>
	  <mxCell id="2" value="&lt;&lt;interface&gt;&gt;&lt;br/&gt;IWidget" style="swimlane;" vertex="1" parent="1">
	    <mxGeometry width="140" height="60" as="geometry"/>
	  </mxCell>
	</root></mxGraphModel>`

	b := drawio.NewDrawioModelBuilder()
	graph, err := b.Build(domain.RawModelData(xmlStr), "drawio")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(graph.Nodes) == 0 {
		t.Fatal("expected 1 node")
	}
	n := graph.Nodes[0]
	if n.Name != "IWidget" {
		t.Errorf("expected Name=IWidget (br-separated stereotype stripped), got %q", n.Name)
	}
	if n.Type != "Interface" {
		t.Errorf("expected Type=Interface, got %q", n.Type)
	}
	t.Logf("✔ br-newline decode: Name=%s Type=%s", n.Name, n.Type)
}

// ─── Multi-line method signature test ────────────────────────────────────────

func TestBuild_MultilineMethodSignature(t *testing.T) {
	// Method signature split across multiple HTML lines in one cell value
	xml := `<mxGraphModel><root>
	  <mxCell id="0"/><mxCell id="1" parent="0"/>
	  <mxCell id="2" value="Police" style="swimlane;" vertex="1" parent="1">
	    <mxGeometry width="200" height="80" as="geometry"/>
	  </mxCell>
	  <mxCell id="3" value="+ issueTicket(&lt;br&gt;car: Car, meter: Meter&lt;br&gt;): Ticket" style="text;" vertex="1" parent="2">
	    <mxGeometry y="26" width="200" height="50" as="geometry"/>
	  </mxCell>
	</root></mxGraphModel>`

	b := drawio.NewDrawioModelBuilder()
	graph, err := b.Build(domain.RawModelData(xml), "drawio")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(graph.Nodes) == 0 {
		t.Fatal("expected 1 node")
	}
	n := graph.Nodes[0]
	if len(n.Methods) == 0 {
		t.Errorf("expected at least 1 method, got 0")
	}
	if len(n.Attributes) > 0 {
		// Continuation lines must NOT become attributes
		t.Errorf("expected 0 attributes (no fake attr from multi-line method), got: %v", n.Attributes)
	}
	t.Logf("✔ multi-line method: Methods=%v Attrs=%v", n.Methods, n.Attributes)
}

func TestBuild_GenericsAndStyling(t *testing.T) {
	// 1. Generic types: List<String>
	// 2. Italics <i>: abstract
	// 3. Underline <u>: static
	xml := `<mxGraphModel><root>
	  <mxCell id="0"/><mxCell id="1" parent="0"/>
	  <mxCell id="2" value="&lt;i&gt;GenericService&amp;lt;T&amp;gt;&lt;/i&gt;" style="swimlane;" vertex="1" parent="1">
	    <mxGeometry width="200" height="80" as="geometry"/>
	  </mxCell>
	  <mxCell id="3" value="- items : List&amp;lt;String&amp;gt;" style="text;" vertex="1" parent="2">
	    <mxGeometry y="26" width="200" height="26" as="geometry"/>
	  </mxCell>
	  <mxCell id="4" value="+ &lt;u&gt;getInstance()&lt;/u&gt; : T" style="text;" vertex="1" parent="2">
	    <mxGeometry y="52" width="200" height="26" as="geometry"/>
	  </mxCell>
	</root></mxGraphModel>`

	b := drawio.NewDrawioModelBuilder()
	graph, err := b.Build(domain.RawModelData(xml), "drawio")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(graph.Nodes) == 0 {
		t.Fatal("expected 1 node")
	}
	n := graph.Nodes[0]

	// Verify Name (should have generic preserved AND abstract keyword added from <i>)
	if !strings.Contains(n.Name, "GenericService<T>") {
		t.Errorf("expected Name to contain 'GenericService<T>', got %q", n.Name)
	}
	if !strings.Contains(n.Name, "{abstract}") {
		t.Errorf("expected Name to contain '{abstract}' from <i> tag, got %q", n.Name)
	}

	// Verify Attribute
	if len(n.Attributes) == 0 || !strings.Contains(n.Attributes[0], "List<String>") {
		t.Errorf("expected attribute with 'List<String>', got %v", n.Attributes)
	}

	// Verify Method (should have {static} from <u>)
	if len(n.Methods) == 0 || !strings.Contains(n.Methods[0], "{static}") {
		t.Fatalf("expected method with '{static}', got %v", n.Methods)
	}
	if !strings.Contains(n.Methods[0], "getInstance") {
		t.Errorf("expected method name 'getInstance', got %q", n.Methods[0])
	}

	t.Logf("✔ generics and styling: Name=%s Attrs=%v Methods=%v", n.Name, n.Attributes, n.Methods)
}

// ─── Relation Note test ───────────────────────────────────────────────────────

func TestBuild_EdgeNoteExtraction(t *testing.T) {
	xml := `<mxGraphModel><root>
	  <mxCell id="0"/><mxCell id="1" parent="0"/>
	  <mxCell id="2" value="ClassA" style="swimlane;" vertex="1" parent="1">
		<mxGeometry x="0" y="0" width="100" height="100" as="geometry" />
	  </mxCell>
	  <mxCell id="3" value="ClassB" style="swimlane;" vertex="1" parent="1">
	    <mxGeometry x="200" y="0" width="100" height="100" as="geometry" />
	  </mxCell>
	  <mxCell id="4" value="__1__&lt;br&gt;" style="edgeStyle=none;" edge="1" parent="1" source="2" target="3">
	    <mxGeometry relative="1" as="geometry" />
      </mxCell>
	</root></mxGraphModel>`

	b := drawio.NewDrawioModelBuilder()
	graph, err := b.Build(domain.RawModelData(xml), "drawio")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(graph.Edges) == 0 {
		t.Fatal("expected 1 edge")
	}
	edge := graph.Edges[0]
	if edge.Note != "__1__" {
		t.Errorf("expected Edge.Note='__1__', got %q", edge.Note)
	}
	t.Logf("✔ edge note extracted: %s", edge.Note)
}

func TestBuild_StaticByStyleFontStyle4(t *testing.T) {
	// fontStyle=4 is Underline -> represents static in UML
	xml := `<mxGraphModel><root>
	  <mxCell id="0"/><mxCell id="1" parent="0"/>
	  <mxCell id="2" value="Config" style="swimlane;" vertex="1" parent="1">
	    <mxGeometry width="140" height="60" as="geometry"/>
	  </mxCell>
	  <mxCell id="3" value="+ VERSION : String" style="text;fontStyle=4" vertex="1" parent="2">
	    <mxGeometry y="26" width="140" height="26" as="geometry"/>
	  </mxCell>
	</root></mxGraphModel>`

	b := drawio.NewDrawioModelBuilder()
	graph, err := b.Build(domain.RawModelData(xml), "drawio")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(graph.Nodes) == 0 {
		t.Fatal("expected 1 node")
	}
	n := graph.Nodes[0]

	// Verify Attribute has {static} injected
	found := false
	for _, attr := range n.Attributes {
		if strings.Contains(strings.ToLower(attr), "{static}") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected attribute to contain '{static}', got %v", n.Attributes)
	}
	t.Logf("✔ static by fontStyle=4 detection OK: Attrs=%v", n.Attributes)
}
