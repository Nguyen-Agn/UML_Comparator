package builder_test

// ─── Incorrect UML Test Suite ─────────────────────────────────────────────────
//
// Skill: UML-TestCase-Runner (Skills/UML-TestCase-Runner/SKILL.md)
//
// Mỗi test case trong file này:
//  1. Đọc 1 file .drawio từ ../UMLs_testcase/incorrect/ qua DrawioParser
//  2. Chạy qua Builder.Build() để tạo *domain.UMLGraph
//  3. Chạy domain.ValidateGraph() để lấy []IntegrityError
//  4. Assert: expected error code phải xuất hiện trong kết quả
//  5. Assert pipeline behavior: ERROR codes → pipeline stop | WARN codes → pipeline continue
//
// Expected results chi tiết: Skills/UML-TestCase-Runner/testcases.md
// Error code reference:       Skills/UML-TestCase-Runner/knowledge.md
// Verification checklist:     Skills/UML-TestCase-Runner/check.md

import (
	"testing"
	"uml_compare/builder"
	"uml_compare/domain"
	"uml_compare/parser"
)

// ─── Helpers ──────────────────────────────────────────────────────────────────

const incorrectDir = "../UMLs_testcase/incorrect/"

// loadAndBuild reads a .drawio file via DrawioParser then runs Builder.Build().
// Returns (graph, validationErrors, buildError).
// If buildError != nil the graph may be nil; callers must check.
func loadAndBuild(t *testing.T, filename string) (*domain.UMLGraph, []domain.IntegrityError) {
	t.Helper()
	p := parser.NewAutoParserDefault()
	raw, sourceType, err := p.Parse(incorrectDir + filename)
	if err != nil {
		t.Fatalf("[%s] Parser.Parse failed: %v", filename, err)
	}

	b, err := builder.GetBuilder(sourceType)
	if err != nil {
		t.Fatalf("[%s] builder.GetBuilder failed: %v", filename, err)
	}

	graph, buildErr := b.Build(raw, sourceType)
	if buildErr != nil {
		// Build error (e.g. empty raw) → use empty graph for Validate
		t.Logf("[%s] Builder.Build returned error (may be expected): %v", filename, buildErr)
		graph = &domain.UMLGraph{}
	}

	errs := domain.ValidateGraph(graph, filename)
	return graph, errs
}

// assertHasCode fails the test if the expected error code is not present.
func assertHasCode(t *testing.T, errs []domain.IntegrityError, code string) {
	t.Helper()
	for _, e := range errs {
		if e.Code == code {
			t.Logf("✔ Got expected code: %s [%s] — %s", e.Code, e.Severity, e.Message)
			return
		}
	}
	t.Errorf("✘ Expected code %q not found. Actual errors: %v", code, errs)
}

// assertHasAnyCode fails unless at least one of the given codes is present.
// Used when multiple outcomes are possible (e.g. INVALID_NODE_TYPE or EMPTY_GRAPH).
func assertHasAnyCode(t *testing.T, errs []domain.IntegrityError, codes ...string) {
	t.Helper()
	for _, e := range errs {
		for _, c := range codes {
			if e.Code == c {
				t.Logf("✔ Got expected code (any-of): %s [%s] — %s", e.Code, e.Severity, e.Message)
				return
			}
		}
	}
	t.Errorf("✘ Expected one of %v not found. Actual errors: %v", codes, errs)
}

// assertPipelineStop verifies that at least 1 ERROR-severity issue exists (pipeline must stop).
func assertPipelineStop(t *testing.T, errs []domain.IntegrityError) {
	t.Helper()
	blocking := domain.FilterErrors(errs)
	if len(blocking) == 0 {
		t.Errorf("✘ Pipeline should STOP (expected ≥1 ERROR), but FilterErrors returned 0")
		return
	}
	t.Logf("✔ Pipeline correctly STOPS: %d blocking ERROR(s) found", len(blocking))
}

// assertPipelineContinue verifies that NO ERROR-severity issues exist (WARN ok, pipeline continues).
func assertPipelineContinue(t *testing.T, errs []domain.IntegrityError) {
	t.Helper()
	blocking := domain.FilterErrors(errs)
	if len(blocking) > 0 {
		t.Errorf("✘ Pipeline should CONTINUE (expected 0 ERRORs), but got: %v", blocking)
		return
	}
	t.Logf("✔ Pipeline correctly CONTINUES: 0 blocking ERRORs")
}

// ─── TC-01: EMPTY_GRAPH ───────────────────────────────────────────────────────
// File: err_empty_graph.drawio
// No class nodes in <root> → ValidateGraph must return EMPTY_GRAPH ERROR
// Pipeline: STOP

func TestIncorrect_EmptyGraph(t *testing.T) {
	_, errs := loadAndBuild(t, "err_empty_graph.drawio")

	assertHasCode(t, errs, "EMPTY_GRAPH")
	assertPipelineStop(t, errs)
}

// ─── TC-02: EMPTY_NODE_NAME ───────────────────────────────────────────────────
// File: err_empty_node_name.drawio
// Swimlane container with value="" → node.Name == "" → EMPTY_NODE_NAME ERROR
// Pipeline: STOP

func TestIncorrect_EmptyNodeName(t *testing.T) {
	_, errs := loadAndBuild(t, "err_empty_node_name.drawio")

	assertHasCode(t, errs, "EMPTY_NODE_NAME")
	assertPipelineStop(t, errs)
}

// ─── TC-03: INVALID_NODE_TYPE ─────────────────────────────────────────────────
// File: err_invalid_node_type.drawio
// Rounded rect style (not swimlane/umlClass) → Builder cannot classify type
// Expected: INVALID_NODE_TYPE or EMPTY_GRAPH (if Builder skips the cell entirely)
// Pipeline: STOP

func TestIncorrect_InvalidNodeType(t *testing.T) {
	_, errs := loadAndBuild(t, "err_invalid_node_type.drawio")

	// Builder may: (a) create node with Type="" → INVALID_NODE_TYPE
	//           or (b) skip the cell entirely → EMPTY_GRAPH
	assertHasAnyCode(t, errs, "INVALID_NODE_TYPE", "EMPTY_GRAPH")
	assertPipelineStop(t, errs)
}

// ─── TC-04: DANGLING_EDGE_SOURCE ─────────────────────────────────────────────
// File: err_dangling_edge.drawio
// Edge with source="999" (non-existent ID) → DANGLING_EDGE_SOURCE ERROR
// Pipeline: STOP

func TestIncorrect_DanglingEdgeSource(t *testing.T) {
	_, errs := loadAndBuild(t, "err_dangling_edge.drawio")

	assertHasCode(t, errs, "DANGLING_EDGE_SOURCE")
	assertPipelineStop(t, errs)
}

// ─── TC-05: SELF_LOOP_EDGE ────────────────────────────────────────────────────
// File: err_self_loop.drawio
// Edge with source="2" and target="2" → SELF_LOOP_EDGE ERROR
// Pipeline: STOP

func TestIncorrect_SelfLoop(t *testing.T) {
	_, errs := loadAndBuild(t, "err_self_loop.drawio")

	assertHasCode(t, errs, "SELF_LOOP_EDGE")
	assertPipelineStop(t, errs)
}

// ─── TC-06: SUSPECT_NODE_NAME ─────────────────────────────────────────────────
// File: warn_suspect_name.drawio
// Node name contains "<<" or ">>" (HTML entity remnants) → SUSPECT_NODE_NAME WARN
// Pipeline: CONTINUE (WARN only, no ERRORs)

func TestIncorrect_SuspectName(t *testing.T) {
	_, errs := loadAndBuild(t, "warn_suspect_name.drawio")

	assertHasCode(t, errs, "SUSPECT_NODE_NAME")
	assertPipelineContinue(t, errs)

	warns := domain.FilterWarns(errs)
	t.Logf("✔ Total WARNs: %d — pipeline flags but continues", len(warns))
}

// ─── TC-07: TRIVIAL_NODE_NAME ─────────────────────────────────────────────────
// File: warn_trivial_name.drawio
// Node name "A" (1 rune, ≤ 2) → TRIVIAL_NODE_NAME WARN
// Pipeline: CONTINUE

func TestIncorrect_TrivialName(t *testing.T) {
	_, errs := loadAndBuild(t, "warn_trivial_name.drawio")

	assertHasCode(t, errs, "TRIVIAL_NODE_NAME")
	assertPipelineContinue(t, errs)
}

// ─── TC-08: INCOMPLETE_ATTRIBUTE ─────────────────────────────────────────────
// File: warn_incomplete_attr.drawio
// Attribute "- id" has no ":" → INCOMPLETE_ATTRIBUTE WARN
// Pipeline: CONTINUE

func TestIncorrect_IncompleteAttribute(t *testing.T) {
	_, errs := loadAndBuild(t, "warn_incomplete_attr.drawio")

	assertHasCode(t, errs, "INCOMPLETE_ATTRIBUTE")
	assertPipelineContinue(t, errs)
}

// ─── TC-09: INCOMPLETE_METHOD ────────────────────────────────────────────────
// File: warn_incomplete_method.drawio
// Method "connect(" ends with "(" → INCOMPLETE_METHOD WARN
// Pipeline: CONTINUE

func TestIncorrect_IncompleteMethod(t *testing.T) {
	_, errs := loadAndBuild(t, "warn_incomplete_method.drawio")

	assertHasCode(t, errs, "INCOMPLETE_METHOD")
	assertPipelineContinue(t, errs)
}
