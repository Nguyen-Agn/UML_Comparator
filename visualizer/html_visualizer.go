package visualizer

import (
	"fmt"
	"html/template"
	"os"
	"strings"
	"time"
	"uml_compare/domain"
)

// HTMLVisualizer implements IVisualizer by producing a self-contained HTML report.
type HTMLVisualizer struct{}

// NewHTMLVisualizer creates a new HTMLVisualizer instance.
func NewHTMLVisualizer() *HTMLVisualizer {
	return &HTMLVisualizer{}
}

// ── Template Data Structs ────────────────────────────────────────────────────

// templateData is the root data structure passed to the HTML template.
type templateData struct {
	Score      float64
	MaxScore   float64
	Percent    float64
	ScoreClass string // CSS class: score-green, score-yellow, score-red
	FillClass  string // CSS class: fill-green, fill-yellow, fill-red

	StuNodes  []nodeView
	SolNodes  []nodeView
	Relations []relationView

	Stats     statsView
	Feedbacks []string
	Timestamp string
}

// nodeView represents a single node card in the HTML.
type nodeView struct {
	Name       string
	Type       string
	BadgeClass string // class, interface, abstract, enum
	Attributes []memberView
	Methods    []memberView
}

// memberView represents a single attribute or method line.
type memberView struct {
	Display string
	Status  string // correct, wrong, missing, extra, neutral
}

// relationView represents a single relation row.
type relationView struct {
	Source  string
	Target string
	RelType string
	Status string // correct, missing, wrong, extra
	Icon   string
	Note   string
}

// statsView holds the summary counters.
type statsView struct {
	Correct int
	Missing int
	Wrong   int
	Extra   int
}

// ── Core Logic ───────────────────────────────────────────────────────────────

// ExportHTML renders the GradeResult into a self-contained HTML file.
func (v *HTMLVisualizer) ExportHTML(result *domain.GradeResult, outputPath string) error {
	data := v.buildTemplateData(result)

	tmpl, err := template.New("report").Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("template parse error: %w", err)
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("cannot create output file: %w", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("template execute error: %w", err)
	}

	return nil
}

// ExportStudentHTML renders a student-facing HTML report.
// Shows only the student's own nodes and relations with color-coded feedback.
// Does NOT reveal solution content or detailed deduction breakdown.
func (v *HTMLVisualizer) ExportStudentHTML(result *domain.GradeResult, outputPath string) error {
	data := v.buildStudentTemplateData(result)

	tmpl, err := template.New("student_report").Parse(studentHTMLTemplate)
	if err != nil {
		return fmt.Errorf("student template parse error: %w", err)
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("cannot create output file: %w", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("student template execute error: %w", err)
	}

	return nil
}

// buildTemplateData transforms a GradeResult into the template-friendly struct.
func (v *HTMLVisualizer) buildTemplateData(result *domain.GradeResult) templateData {
	d := templateData{
		Score:      result.TotalScore,
		MaxScore:   result.MaxScore,
		Percent:    result.CorrectPercent,
		ScoreClass: scoreColorClass(result.CorrectPercent),
		FillClass:  fillColorClass(result.CorrectPercent),
		Feedbacks:  result.Feedbacks,
		Timestamp:  time.Now().Format("2006-01-02 15:04:05"),
	}

	report := result.Report

	// ── Build Student Nodes ──────────────────────────────────────────────
	if result.StudentGraph != nil {
		for _, n := range result.StudentGraph.Nodes {
			nv := nodeView{
				Name:       n.Name,
				Type:       n.Type,
				BadgeClass: badgeClass(n.Type),
			}

			// Tag each student attribute with its diff status
			for _, a := range n.Attributes {
				status := memberStatus(n.Name, stuAttrDisplay(&a), report, "attribute", "student")
				nv.Attributes = append(nv.Attributes, memberView{
					Display: formatStuAttr(&a),
					Status:  status,
				})
			}

			// Tag each student method with diff status
			for _, m := range n.Methods {
				if m.Type == "getter" || m.Type == "setter" {
					continue
				}
				status := memberStatus(n.Name, stuMethDisplay(&m), report, "method", "student")
				nv.Methods = append(nv.Methods, memberView{
					Display: formatStuMethod(&m),
					Status:  status,
				})
			}

			d.StuNodes = append(d.StuNodes, nv)
		}
	}

	// ── Build Solution Nodes ─────────────────────────────────────────────
	if result.SolutionGraph != nil {
		for _, n := range result.SolutionGraph.Nodes {
			nv := nodeView{
				Name:       n.Name,
				Type:       n.Type,
				BadgeClass: badgeClass(n.Type),
			}

			for _, a := range n.Attributes {
				nv.Attributes = append(nv.Attributes, memberView{
					Display: formatSolAttr(&a),
					Status:  "neutral",
				})
			}

			for _, m := range n.Methods {
				if m.Type == "getter" || m.Type == "setter" {
					continue
				}
				nv.Methods = append(nv.Methods, memberView{
					Display: formatSolMethod(&m),
					Status:  "neutral",
				})
			}

			d.SolNodes = append(d.SolNodes, nv)
		}
	}

	// ── Build Relations ──────────────────────────────────────────────────
	d.Relations = v.buildRelations(result)

	// ── Build Stats ──────────────────────────────────────────────────────
	d.Stats = countStats(report)

	return d
}

// ── Relations Builder ────────────────────────────────────────────────────────

// buildStudentTemplateData builds template data for the student-facing report.
// Only includes student nodes and student-visible relations (no solution, no feedbacks).
func (v *HTMLVisualizer) buildStudentTemplateData(result *domain.GradeResult) templateData {
	d := templateData{
		Score:      result.TotalScore,
		MaxScore:   result.MaxScore,
		Percent:    result.CorrectPercent,
		ScoreClass: scoreColorClass(result.CorrectPercent),
		FillClass:  fillColorClass(result.CorrectPercent),
		Timestamp:  time.Now().Format("2006-01-02 15:04:05"),
		// Feedbacks intentionally omitted — student should not see deduction details
	}

	report := result.Report

	// Only student nodes
	if result.StudentGraph != nil {
		for _, n := range result.StudentGraph.Nodes {
			nv := nodeView{
				Name:       n.Name,
				Type:       n.Type,
				BadgeClass: badgeClass(n.Type),
			}

			for _, a := range n.Attributes {
				status := memberStatus(n.Name, stuAttrDisplay(&a), report, "attribute", "student")
				nv.Attributes = append(nv.Attributes, memberView{
					Display: formatStuAttr(&a),
					Status:  status,
				})
			}

			for _, m := range n.Methods {
				if m.Type == "getter" || m.Type == "setter" {
					continue
				}
				status := memberStatus(n.Name, stuMethDisplay(&m), report, "method", "student")
				nv.Methods = append(nv.Methods, memberView{
					Display: formatStuMethod(&m),
					Status:  status,
				})
			}

			d.StuNodes = append(d.StuNodes, nv)
		}
	}

	// No SolNodes — intentionally empty

	// Student-visible relations only (correct, wrong, extra — no missing)
	d.Relations = v.buildStudentRelations(result)

	return d
}

// buildStudentRelations returns only relations that the student actually drew.
// Missing relations are excluded because they would reveal solution info.
func (v *HTMLVisualizer) buildStudentRelations(result *domain.GradeResult) []relationView {
	var rels []relationView
	report := result.Report

	stuNames := make(map[string]string)
	solNames := make(map[string]string)
	if result.StudentGraph != nil {
		for _, n := range result.StudentGraph.Nodes {
			stuNames[n.ID] = n.Name
		}
	}
	if result.SolutionGraph != nil {
		for _, n := range result.SolutionGraph.Nodes {
			solNames[n.ID] = n.Name
		}
	}
	nameOf := func(id string) string {
		if n, ok := stuNames[id]; ok {
			return n
		}
		if n, ok := solNames[id]; ok {
			return n
		}
		return id
	}

	// Correct — student drew it and it matches
	for _, e := range report.CorrectDetail.Edge {
		edge := edgeFromDiff(&e)
		rels = append(rels, relationView{
			Source: nameOf(edge.SourceID), Target: nameOf(edge.TargetID),
			RelType: edge.RelationType, Status: "correct", Icon: "✅",
		})
	}

	// Wrong — student drew it but something is off
	for _, e := range report.WrongDetail.Edge {
		edge := edgeFromDiff(&e)
		rels = append(rels, relationView{
			Source: nameOf(edge.SourceID), Target: nameOf(edge.TargetID),
			RelType: edge.RelationType, Status: "wrong", Icon: "⚠️",
		})
	}

	// Extra — student drew it but it's not in solution
	for _, e := range report.ExtraDetail.Edge {
		if e.Stu != nil {
			rels = append(rels, relationView{
				Source: nameOf(e.Stu.SourceID), Target: nameOf(e.Stu.TargetID),
				RelType: e.Stu.RelationType, Status: "extra", Icon: "💡",
			})
		}
	}

	// Missing edges intentionally excluded — would reveal solution

	return rels
}

func (v *HTMLVisualizer) buildRelations(result *domain.GradeResult) []relationView {
	var rels []relationView
	report := result.Report

	// Helper name lookups
	solNames := make(map[string]string)
	if result.SolutionGraph != nil {
		for _, n := range result.SolutionGraph.Nodes {
			solNames[n.ID] = n.Name
		}
	}
	stuNames := make(map[string]string)
	if result.StudentGraph != nil {
		for _, n := range result.StudentGraph.Nodes {
			stuNames[n.ID] = n.Name
		}
	}
	nameOf := func(id string) string {
		if n, ok := solNames[id]; ok {
			return n
		}
		if n, ok := stuNames[id]; ok {
			return n
		}
		return id
	}

	// Correct edges
	for _, e := range report.CorrectDetail.Edge {
		edge := edgeFromDiff(&e)
		rels = append(rels, relationView{
			Source: nameOf(edge.SourceID), Target: nameOf(edge.TargetID),
			RelType: edge.RelationType, Status: "correct", Icon: "✅",
		})
	}

	// Wrong edges
	for _, e := range report.WrongDetail.Edge {
		edge := edgeFromDiff(&e)
		rels = append(rels, relationView{
			Source: nameOf(edge.SourceID), Target: nameOf(edge.TargetID),
			RelType: edge.RelationType, Status: "wrong", Icon: "⚠️",
			Note: e.Description,
		})
	}

	// Missing edges
	for _, e := range report.MissingDetail.Edge {
		if e.Sol != nil {
			rels = append(rels, relationView{
				Source: nameOf(e.Sol.SourceID), Target: nameOf(e.Sol.TargetID),
				RelType: e.Sol.RelationType, Status: "missing", Icon: "❌",
				Note: "Missing in student",
			})
		}
	}

	// Extra edges
	for _, e := range report.ExtraDetail.Edge {
		if e.Stu != nil {
			rels = append(rels, relationView{
				Source: nameOf(e.Stu.SourceID), Target: nameOf(e.Stu.TargetID),
				RelType: e.Stu.RelationType, Status: "extra", Icon: "➕",
				Note: "Extra in student",
			})
		}
	}

	return rels
}

// ── Formatting Helpers ───────────────────────────────────────────────────────

func formatStuAttr(a *domain.ProcessedAttribute) string {
	return fmt.Sprintf("%s %s : %s", a.Scope, a.Name, a.Type)
}

func formatStuMethod(m *domain.ProcessedMethod) string {
	params := make([]string, len(m.Inputs))
	for i, p := range m.Inputs {
		params[i] = p.Name + ":" + p.Type
	}
	return fmt.Sprintf("%s %s(%s) : %s", m.Scope, m.Name, strings.Join(params, ", "), m.Output)
}

func formatSolAttr(a *domain.SolutionProcessedAttribute) string {
	return fmt.Sprintf("%s %s : %s", a.Scope, strings.Join(a.Names, " | "), strings.Join(a.Types, " | "))
}

func formatSolMethod(m *domain.SolutionProcessedMethod) string {
	params := make([]string, len(m.Inputs))
	for i, p := range m.Inputs {
		params[i] = p.Name + ":" + p.Type
	}
	return fmt.Sprintf("%s %s(%s) : %s", m.Scope, strings.Join(m.Names, " | "), strings.Join(params, ", "), strings.Join(m.Outputs, " | "))
}

func stuAttrDisplay(a *domain.ProcessedAttribute) string {
	return a.Name
}

func stuMethDisplay(m *domain.ProcessedMethod) string {
	return m.Name
}

// ── Status Detection ─────────────────────────────────────────────────────────

// memberStatus determines the diff status of a student member by scanning the report.
func memberStatus(parentName, memberIdent string, report *domain.DiffReport, kind, side string) string {
	if report == nil {
		return "neutral"
	}

	if kind == "attribute" && side == "student" {
		for _, d := range report.CorrectDetail.Attribute {
			if d.ParentClassName == parentName && d.Stu != nil && d.Stu.Name == memberIdent {
				return "correct"
			}
		}
		for _, d := range report.WrongDetail.Attribute {
			if d.ParentClassName == parentName && d.Stu != nil && d.Stu.Name == memberIdent {
				return "wrong"
			}
		}
		for _, d := range report.ExtraDetail.Attribute {
			if d.ParentClassName == parentName && d.Stu != nil && d.Stu.Name == memberIdent {
				return "extra"
			}
		}
	}

	if kind == "method" && side == "student" {
		for _, d := range report.CorrectDetail.Method {
			if d.ParentClassName == parentName && d.Stu != nil && d.Stu.Name == memberIdent {
				return "correct"
			}
		}
		for _, d := range report.WrongDetail.Method {
			if d.ParentClassName == parentName && d.Stu != nil && d.Stu.Name == memberIdent {
				return "wrong"
			}
		}
		for _, d := range report.ExtraDetail.Method {
			if d.ParentClassName == parentName && d.Stu != nil && d.Stu.Name == memberIdent {
				return "extra"
			}
		}
	}

	return "neutral"
}

// ── Utility Helpers ──────────────────────────────────────────────────────────

func scoreColorClass(pct float64) string {
	if pct >= 90 {
		return "score-green"
	}
	if pct >= 60 {
		return "score-yellow"
	}
	return "score-red"
}

func fillColorClass(pct float64) string {
	if pct >= 90 {
		return "fill-green"
	}
	if pct >= 60 {
		return "fill-yellow"
	}
	return "fill-red"
}

func badgeClass(nodeType string) string {
	t := strings.ToLower(nodeType)
	switch {
	case strings.Contains(t, "interface"):
		return "interface"
	case strings.Contains(t, "abstract"):
		return "abstract"
	case strings.Contains(t, "enum"):
		return "enum"
	default:
		return "class"
	}
}

func countStats(report *domain.DiffReport) statsView {
	if report == nil {
		return statsView{}
	}
	count := func(d *domain.DetailError) int {
		return len(d.Class) + len(d.Method) + len(d.Attribute) + len(d.Edge)
	}
	return statsView{
		Correct: count(&report.CorrectDetail),
		Missing: count(&report.MissingDetail),
		Wrong:   count(&report.WrongDetail),
		Extra:   count(&report.ExtraDetail),
	}
}

func edgeFromDiff(d *domain.EdgeDiff) *domain.ProcessedEdge {
	if d.Sol != nil {
		return d.Sol
	}
	return d.Stu
}
