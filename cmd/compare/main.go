// cmd/compare/main.go
// Chạy: go run ./cmd/compare/main.go <solution.drawio> <student.drawio>
// So sánh UML mẫu vs bài sinh viên, hiển thị trực quan side-by-side
package main

import (
	"fmt"
	"os"
	"strings"
	"uml_compare/cmd/share"
	"uml_compare/domain"
	"uml_compare/src/comparator"
	"uml_compare/src/grader"
	"uml_compare/src/matcher"
	"uml_compare/src/prematcher"
)

// compareResult chứa toàn bộ dữ liệu kết quả pipeline để truyền vào print layer.
type compareResult struct {
	SolProcessed *domain.SolutionProcessedUMLGraph
	StuProcessed *domain.ProcessedUMLGraph
	SolStd       *domain.ProcessedUMLGraph // dùng cho edge comparison display
	Mapping      domain.MappingTable
	DiffReport   *domain.DiffReport
	GradeResult  *domain.GradeResult
	Warnings     []domain.IntegrityError
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run ./cmd/compare/main.go <solution.drawio> <student.drawio>")
		fmt.Println("  Example: go run ./cmd/compare/main.go UMLs_testcase/problem1.drawio UMLs_testcase/problem_1.drawio")
		os.Exit(1)
	}

	share.PrintBanner("UML Compare — Data Flow Integrity Check")

	result, err := run(os.Args[1], os.Args[2])
	if err != nil {
		fmt.Printf("%s❌ Pipeline error: %v%s\n", share.Red+share.Bold, err, share.Reset)
		os.Exit(1)
	}

	printCompareResult(result)
}

// run thực hiện toàn bộ pipeline: Load → Validate → PreMatch → Match → Compare → Grade.
// Dừng sớm và trả lỗi nếu có integrity error.
func run(solutionPath, studentPath string) (*compareResult, error) {
	// 1. Load graphs
	solutionGraph, err := share.LoadGraph(solutionPath)
	if err != nil {
		return nil, fmt.Errorf("load solution: %w", err)
	}
	studentGraph, err := share.LoadGraph(studentPath)
	if err != nil {
		return nil, fmt.Errorf("load student: %w", err)
	}

	// 2. Integrity Validation
	allIssues := append(
		domain.ValidateGraph(solutionGraph, "Solution"),
		domain.ValidateGraph(studentGraph, "Student")...,
	)
	hardErrors := domain.FilterErrors(allIssues)
	if len(hardErrors) > 0 {
		msgs := make([]string, len(hardErrors))
		for i, e := range hardErrors {
			msgs[i] = e.Error()
		}
		return nil, fmt.Errorf("integrity errors:\n   • %s", strings.Join(msgs, "\n   • "))
	}

	warnings := filterDisplayWarnings(domain.FilterWarns(allIssues))

	// 3. PreMatch
	stdPM := prematcher.NewStandardPreMatcher()
	solPM := prematcher.NewUMLSolutionPreMatcher()

	solStd, _ := stdPM.Process(solutionGraph) // dùng cho edge display
	stuProc, _ := stdPM.Process(studentGraph)
	solForMatch, _ := solPM.ProcessSolution(solutionGraph)

	// 4. Match
	entityMatcher := matcher.NewStandardEntityMatcher(0.8)
	mapping, _ := entityMatcher.Match(solForMatch, stuProc)

	// 5. Compare
	comp := comparator.NewStandardComparator()
	diffReport, _ := comp.Compare(solForMatch, stuProc, mapping)

	// 6. Grade
	gr := grader.NewStandardGrader()
	rules := &grader.GradingRules{}
	gradeResult, _ := gr.Grade(diffReport, solForMatch, stuProc, rules)

	return &compareResult{
		SolProcessed: solForMatch,
		StuProcessed: stuProc,
		SolStd:       solStd,
		Mapping:      mapping,
		DiffReport:   diffReport,
		GradeResult:  gradeResult,
		Warnings:     warnings,
	}, nil
}

// filterDisplayWarnings lọc bỏ các warning không cần hiển thị (getter/setter shortcuts).
func filterDisplayWarnings(warns []domain.IntegrityError) []domain.IntegrityError {
	out := warns[:0]
	for _, w := range warns {
		lower := strings.ToLower(w.Message)
		if w.Code == "INCOMPLETE_ATTRIBUTE" && (strings.Contains(lower, "getter") || strings.Contains(lower, "setter")) {
			continue
		}
		out = append(out, w)
	}
	return out
}

// ── Print Layer ───────────────────────────────────────────────────────────────

// printCompareResult điều phối toàn bộ việc in kết quả ra stdout.
func printCompareResult(r *compareResult) {
	printIntegrityStatus(r.Warnings)
	printSideBySideNodes(r.SolProcessed, r.StuProcessed, r.Mapping, r.DiffReport)
	printEdgeComparison(r.SolStd, r.StuProcessed, r.Mapping)
	printDiffReport(r.DiffReport)
	printSummary(r.SolStd, r.StuProcessed, r.Mapping)
	printGradeResult(r.GradeResult)
}

// printIntegrityStatus in kết quả kiểm tra integrity (warnings nếu có).
func printIntegrityStatus(warnings []domain.IntegrityError) {
	fmt.Printf("\n%s── [GATE] Data Integrity Check ──────────────────────────────%s\n", share.Cyan+share.Bold, share.Reset)
	if len(warnings) > 0 {
		fmt.Printf("%s⚠️  UML Quality Warnings (%d) — comparison continues:%s\n", share.Yellow+share.Bold, len(warnings), share.Reset)
		for _, w := range warnings {
			fmt.Printf("   • %s\n", w.Error())
		}
		fmt.Printf("%s⚠️  Continuing with warnings — results may be partially inaccurate%s\n", share.Yellow, share.Reset)
	} else {
		fmt.Printf("%s✅ Both graphs pass — proceeding to comparison%s\n", share.Green, share.Reset)
	}
}

// printGradeResult in điểm số cuối cùng và log khấu trừ.
func printGradeResult(gr *domain.GradeResult) {
	fmt.Printf("\n%s── [GRADER] Final Score ───────────────────────────────────────%s\n", share.Cyan+share.Bold, share.Reset)

	scoreColor := share.Green
	if gr.CorrectPercent < 90 {
		scoreColor = share.Yellow
	}
	if gr.CorrectPercent < 60 {
		scoreColor = share.Red
	}
	fmt.Printf("  %sScore: %.2f / %.2f (%.2f%%)%s\n",
		scoreColor+share.Bold, gr.TotalScore, gr.MaxScore, gr.CorrectPercent, share.Reset)

	if len(gr.Feedbacks) > 0 {
		fmt.Printf("\n  %sDeductions Log:%s\n", share.Yellow, share.Reset)
		for _, f := range gr.Feedbacks {
			fmt.Printf("   • %s\n", f)
		}
	}
}

func printSideBySideNodes(sol *domain.SolutionProcessedUMLGraph, stu *domain.ProcessedUMLGraph, mapping domain.MappingTable, report *domain.DiffReport) {
	fmt.Printf("\n%s── [COMPARE] Nodes Side-by-Side ────────────────────────────%s\n", share.Cyan+share.Bold, share.Reset)

	const col = 60
	header := fmt.Sprintf("%s  %-*s│  %-*s%s", share.Bold+share.Blue, col, "SOLUTION (đáp án)", col, "STUDENT (bài nộp)", share.Reset)
	fmt.Println(strings.Repeat("─", col*2+4))
	fmt.Println(header)
	fmt.Println(strings.Repeat("─", col*2+4))

	mappedStu := make(map[string]bool)

	for i := range sol.Nodes {
		solNode := &sol.Nodes[i]
		solPart := cleanStr(fmt.Sprintf("[%s] %s (%dA/%dM)", solNode.Type[:1], solNode.Name, len(solNode.Attributes), len(solNode.Methods)))

		var stuNode *domain.ProcessedNode
		stuPart := ""
		matchMark := "✗ "

		if mapped, ok := mapping[solNode.ID]; ok {
			for j := range stu.Nodes {
				if stu.Nodes[j].ID == mapped.StudentID {
					stuNode = &stu.Nodes[j]
					mappedStu[stuNode.ID] = true
					break
				}
			}
			if stuNode != nil {
				stuPart = cleanStr(fmt.Sprintf("[%s] %s (%dA/%dM)", stuNode.Type[:1], stuNode.Name, len(stuNode.Attributes), len(stuNode.Methods)))
				if mapped.Similarity == 1.0 {
					matchMark = "✔ "
				} else {
					matchMark = "≈ "
				}
			}
		}

		color := share.Reset
		switch matchMark {
		case "✔ ":
			color = share.Green
		case "≈ ":
			color = share.Yellow
		default:
			color = share.Red
		}

		fmt.Printf("%-*s│ %s%s%-*s%s\n", col, "  "+solPart, color, matchMark, col, stuPart, share.Reset)

		if stuNode != nil {
			printMemberDiffs(solNode, stuNode, report)
		}
	}

	// Remaining unmatched student nodes
	for i := range stu.Nodes {
		stuNode := &stu.Nodes[i]
		if !mappedStu[stuNode.ID] {
			stuPart := cleanStr(fmt.Sprintf("[%s] %s (%dA/%dM)", stuNode.Type[:1], stuNode.Name, len(stuNode.Attributes), len(stuNode.Methods)))
			fmt.Printf("%-*s│ %s%-*s%s\n", col, "  ", share.Red+"✗ ", col, limitStr(stuPart, col), share.Reset)
		}
	}

	fmt.Println(strings.Repeat("─", col*2+4))
	fmt.Printf("  %sLegend:%s [C]=Class [I]=Interface [A]=Actor  (A=attrs M=methods)\n", share.Bold, share.Reset)
	fmt.Printf("          %s✔%s exact/perfect match  %s≈%s fuzzy/arch match  %s✗%s missing in one side\n",
		share.Green, share.Reset, share.Yellow, share.Reset, share.Red, share.Reset)
}

// printMemberDiffs in attributes và methods so sánh giữa solution node và student node.
func printMemberDiffs(solNode *domain.SolutionProcessedNode, stuNode *domain.ProcessedNode, report *domain.DiffReport) {
	const col = 60

	for _, d := range report.CorrectDetail.Attribute {
		if d.ParentClassName == solNode.Name {
			fmt.Printf("%s%-*s%s│ %s%-*s%s\n", share.Green, col, limitStr("    ✔ [Attr] "+attrString(d.Sol), col), share.Reset, share.Green, col, limitStr(" ✔ [Attr] "+stuAttrString(d.Stu), col), share.Reset)
		}
	}
	for _, d := range report.WrongDetail.Attribute {
		if d.ParentClassName == solNode.Name {
			fmt.Printf("%s%-*s%s│ %s%-*s%s\n", share.Yellow, col, limitStr("    ≈ [Attr] "+attrString(d.Sol), col), share.Reset, share.Yellow, col, limitStr(" ≈ [Attr] "+stuAttrString(d.Stu), col), share.Reset)
		}
	}
	for _, d := range report.MissingDetail.Attribute {
		if d.ParentClassName == solNode.Name {
			fmt.Printf("%s%-*s%s│  %-*s\n", share.Red, col, limitStr("    ✗ [Attr] "+attrString(d.Sol), col), share.Reset, col, "")
		}
	}
	for _, d := range report.ExtraDetail.Attribute {
		if d.ParentClassName == stuNode.Name {
			fmt.Printf("%-*s│ %s%-*s%s\n", col, "", share.Red, col, limitStr(" ✗ [Attr] "+stuAttrString(d.Stu), col), share.Reset)
		}
	}

	for _, d := range report.CorrectDetail.Method {
		if d.ParentClassName == solNode.Name && d.Sol != nil {
			fmt.Printf("%s%-*s%s│ %s%-*s%s\n", share.Green, col, limitStr("    ✔ [Meth] "+methString(d.Sol), col), share.Reset, share.Green, col, limitStr(" ✔ [Meth] "+stuMethString(d.Stu), col), share.Reset)
		}
	}
	for _, d := range report.WrongDetail.Method {
		if d.ParentClassName == solNode.Name && d.Sol != nil {
			fmt.Printf("%s%-*s%s│ %s%-*s%s\n", share.Yellow, col, limitStr("    ≈ [Meth] "+methString(d.Sol), col), share.Reset, share.Yellow, col, limitStr(" ≈ [Meth] "+stuMethString(d.Stu), col), share.Reset)
		}
	}
	for _, d := range report.MissingDetail.Method {
		if d.ParentClassName == solNode.Name && d.Sol != nil {
			fmt.Printf("%s%-*s%s│  %-*s\n", share.Red, col, limitStr("    ✗ [Meth] "+methString(d.Sol), col), share.Reset, col, "")
		}
	}
	for _, d := range report.ExtraDetail.Method {
		if d.ParentClassName == stuNode.Name && d.Stu != nil {
			fmt.Printf("%-*s│ %s%-*s%s\n", col, "", share.Red, col, limitStr(" ✗ [Meth] "+stuMethString(d.Stu), col), share.Reset)
		}
	}
}

func printEdgeComparison(sol, stu *domain.ProcessedUMLGraph, mapping domain.MappingTable) {
	fmt.Printf("\n%s── [COMPARE] Edges (Relations) ──────────────────────────────%s\n", share.Cyan+share.Bold, share.Reset)

	solNames := buildNameMap(sol.Nodes)
	stuNames := buildNameMap(stu.Nodes)

	fmt.Println("  Solution edges (mapped to student names if available):")
	for _, se := range sol.Edges {
		status := "  "
		var mappedSrc, mappedTgt string

		if mSrc, ok1 := mapping[se.SourceID]; ok1 {
			if mTgt, ok2 := mapping[se.TargetID]; ok2 {
				mappedSrc = mSrc.StudentID
				mappedTgt = mTgt.StudentID
				for _, ste := range stu.Edges {
					if ste.SourceID == mappedSrc && ste.TargetID == mappedTgt && ste.RelationType == se.RelationType {
						status = "✔ "
						break
					}
				}
			}
		}

		color := share.Reset
		if status == "✔ " {
			color = share.Green
		}
		fmt.Printf("    %s%s%s -[%s]-> %s%s\n", color, status, solNames[se.SourceID], se.RelationType, solNames[se.TargetID], share.Reset)
	}

	fmt.Println("  Extra/different edges in Student:")
	for _, ste := range stu.Edges {
		if !edgeFoundInSolution(ste, sol.Edges, mapping) {
			fmt.Printf("    %s✗ %s -[%s]-> %s (not in solution)%s\n",
				share.Red, stuNames[ste.SourceID], ste.RelationType, stuNames[ste.TargetID], share.Reset)
		}
	}
}

// edgeFoundInSolution kiểm tra xem một student edge có tương ứng với solution edge nào không.
func edgeFoundInSolution(ste domain.ProcessedEdge, solEdges []domain.ProcessedEdge, mapping domain.MappingTable) bool {
	for solID, m := range mapping {
		if m.StudentID != ste.SourceID {
			continue
		}
		for solTgtID, m2 := range mapping {
			if m2.StudentID != ste.TargetID {
				continue
			}
			for _, se := range solEdges {
				if se.SourceID == solID && se.TargetID == solTgtID && se.RelationType == ste.RelationType {
					return true
				}
			}
		}
	}
	return false
}

func printSummary(sol, stu *domain.ProcessedUMLGraph, mapping domain.MappingTable) {
	fmt.Printf("\n%s── [SUMMARY] Quick Stats ────────────────────────────────────%s\n", share.Cyan+share.Bold, share.Reset)

	nodeHit := len(mapping)
	edgeHit := countMatchedEdges(sol, stu, mapping)

	nodePct := pct(nodeHit, len(sol.Nodes))
	edgePct := pct(edgeHit, len(sol.Edges))
	overall := (nodePct + edgePct) / 2

	color := share.Cyan
	switch {
	case overall >= 90:
		color = share.Green + share.Bold
	case overall >= 60:
		color = share.Yellow
	default:
		color = share.Red
	}

	fmt.Printf("  Stats: Nodes %d/%d (%.0f%%), Edges %d/%d (%.0f%%) → %sOverall: %.1f%%%s\n",
		nodeHit, len(sol.Nodes), nodePct, edgeHit, len(sol.Edges), edgePct, color, overall, share.Reset)
}

func printDiffReport(report *domain.DiffReport) {
	fmt.Printf("\n%s── [REPORT] Detailed Diff ─────────────────────────────────%s\n", share.Cyan+share.Bold, share.Reset)

	printDetailSection("🚨 [MISSING DETAILS]", report.MissingDetail, share.Red+share.Bold)
	printDetailSection("⚠️ [WRONG/MISMATCHED DETAILS]", report.WrongDetail, share.Yellow+share.Bold)
	printDetailSection("➕ [EXTRA DETAILS]", report.ExtraDetail, share.Cyan+share.Bold)
	printDetailSection("✅ [CORRECT DETAILS]", report.CorrectDetail, share.Green+share.Bold)

	hasIssues := len(report.MissingDetail.Class) > 0 || len(report.MissingDetail.Method) > 0 ||
		len(report.MissingDetail.Attribute) > 0 || len(report.MissingDetail.Edge) > 0 ||
		len(report.WrongDetail.Class) > 0 || len(report.WrongDetail.Method) > 0 ||
		len(report.WrongDetail.Attribute) > 0 || len(report.WrongDetail.Edge) > 0

	if !hasIssues {
		fmt.Printf("\n%s✅ NO ISSUES FOUND — Perfect structural match!%s\n", share.Green+share.Bold, share.Reset)
	}
}

func printDetailSection(title string, detail domain.DetailError, color string) {
	if len(detail.Class) == 0 && len(detail.Method) == 0 && len(detail.Attribute) == 0 && len(detail.Edge) == 0 {
		return
	}

	fmt.Printf("\n%s%s%s\n", color, title, share.Reset)

	if len(detail.Class) > 0 {
		fmt.Printf("   %sNodes:%s\n", share.Bold, share.Reset)
		for _, d := range detail.Class {
			solName, stuName := "?", "?"
			if d.Sol != nil {
				solName = d.Sol.Name
			}
			if d.Stu != nil {
				stuName = d.Stu.Name
			}
			fmt.Printf("    • [%s vs %s] %s\n", solName, stuName, d.Description)
		}
	}
	if len(detail.Attribute) > 0 {
		fmt.Printf("   %sAttributes:%s\n", share.Bold, share.Reset)
		for _, d := range detail.Attribute {
			solV, stuV := "?", "?"
			if d.Sol != nil {
				solV = attrString(d.Sol)
			}
			if d.Stu != nil {
				stuV = stuAttrString(d.Stu)
			}
			fmt.Printf("    • [%s] %s vs %s → %s\n", d.ParentClassName, solV, stuV, d.Description)
		}
	}
	if len(detail.Method) > 0 {
		fmt.Printf("   %sMethods:%s\n", share.Bold, share.Reset)
		for _, d := range detail.Method {
			solV, stuV := "?", "?"
			if d.Sol != nil {
				solV = methString(d.Sol)
			}
			if d.Stu != nil {
				stuV = stuMethString(d.Stu)
			}
			fmt.Printf("    • [%s] %s vs %s → %s\n", d.ParentClassName, solV, stuV, d.Description)
		}
	}
	if len(detail.Edge) > 0 {
		fmt.Printf("   %sRelationships:%s\n", share.Bold, share.Reset)
		for _, d := range detail.Edge {
			solV, stuV := "?", "?"
			if d.Sol != nil {
				solV = d.Sol.RelationType
			}
			if d.Stu != nil {
				stuV = d.Stu.RelationType
			}
			fmt.Printf("    • [%s vs %s] %s\n", solV, stuV, d.Description)
		}
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func buildNameMap(nodes []domain.ProcessedNode) map[string]string {
	m := make(map[string]string, len(nodes))
	for _, n := range nodes {
		m[n.ID] = n.Name
	}
	return m
}

func countMatchedEdges(sol, stu *domain.ProcessedUMLGraph, mapping domain.MappingTable) int {
	count := 0
	for _, se := range sol.Edges {
		if mSrc, ok1 := mapping[se.SourceID]; ok1 {
			if mTgt, ok2 := mapping[se.TargetID]; ok2 {
				for _, ste := range stu.Edges {
					if ste.SourceID == mSrc.StudentID && ste.TargetID == mTgt.StudentID && ste.RelationType == se.RelationType {
						count++
						break
					}
				}
			}
		}
	}
	return count
}

func pct(hit, total int) float64 {
	if total == 0 {
		return 100.0
	}
	return float64(hit) / float64(total) * 100
}

func cleanStr(s string) string {
	s = strings.ReplaceAll(s, "\r", "")
	return strings.ReplaceAll(s, "\n", " ")
}

func limitStr(s string, limit int) string {
	if len(s) > limit-1 {
		return s[:limit-4] + "..."
	}
	return s
}

func attrString(a *domain.SolutionProcessedAttribute) string {
	return cleanStr(fmt.Sprintf("%s %s %s", a.Scope, strings.Join(a.Types, "|"), strings.Join(a.Names, "|")))
}

func stuAttrString(a *domain.ProcessedAttribute) string {
	return cleanStr(fmt.Sprintf("%s %s %s", a.Scope, a.Type, a.Name))
}

func methString(m *domain.SolutionProcessedMethod) string {
	params := make([]string, len(m.Inputs))
	for i, p := range m.Inputs {
		params[i] = strings.Join(p.Types, "|")
	}
	return cleanStr(fmt.Sprintf("%s %s(%s): %s", m.Scope, strings.Join(m.Names, "|"), strings.Join(params, ", "), strings.Join(m.Outputs, "|")))
}

func stuMethString(m *domain.ProcessedMethod) string {
	params := make([]string, len(m.Inputs))
	for i, p := range m.Inputs {
		params[i] = p.Type
	}
	return cleanStr(fmt.Sprintf("%s %s(%s): %s", m.Scope, m.Name, strings.Join(params, ", "), m.Output))
}
