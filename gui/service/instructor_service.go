package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"uml_compare/AppBuilder"
	"uml_compare/cipher"

	"uml_compare/cmd/share"
	"uml_compare/domain"
	"uml_compare/report"
	"uml_compare/src/comparator"
	"uml_compare/src/grader"
	"uml_compare/src/matcher"
	"uml_compare/src/prematcher"
)

// InstructorService defines all logic capabilities exposed to the admin GUI.
type InstructorService interface {
	EncryptSolution(input, output string) error
	BuildExamTool(solutionsDir, outputPath string) error
	GradeBatch(solutionPath, studentDir, outputReport string) (*domain.BatchResult, error)
	CompareUML(solutionPath, studentPath string, isAdmin bool) (*domain.CompareResult, error)
}

// StandardInstructorService implements InstructorService.
type StandardInstructorService struct{}

var _ InstructorService = (*StandardInstructorService)(nil)

func NewStandardInstructorService() *StandardInstructorService {
	return &StandardInstructorService{}
}

func (s *StandardInstructorService) EncryptSolution(input, output string) error {
	c := cipher.New()
	return c.Encrypt(input, output)
}

func (s *StandardInstructorService) BuildExamTool(solutionsDir, outputPath string) error {
	// Reutilize the AppBuilder framework
	dirPrep := &AppBuilder.StandardDirManager{}
	assetCopier := &AppBuilder.FileAssetCopier{}
	taskBld := &AppBuilder.GoTaskBuilder{}

	// Create an ExamBuilder but override destination manually by building the task explicitly,
	// because AppBuilder's Build() is somewhat hardcoded to portable/.
	// But let's adapt it cleanly here.

	embeddedDir := filepath.Join("cmd", "exam_gui", "embedded_solutions")
	if err := dirPrep.Clear(embeddedDir); err != nil {
		return fmt.Errorf("failed clearing embedded dir: %w", err)
	}

	err := assetCopier.CopyAssets(solutionsDir, embeddedDir, ".drawio")
	if err != nil {
		return fmt.Errorf("failed copying assets: %w", err)
	}

	if err := taskBld.BuildTask("Exam Student GUI", outputPath, []string{"./cmd/exam_gui/main.go"}, true); err != nil {
		return fmt.Errorf("exam build failed: %w", err)
	}

	return nil
}

func (s *StandardInstructorService) GradeBatch(solutionPath, studentDir, outputReport string) (*domain.BatchResult, error) {
	solutionGraph, err := share.LoadGraph(solutionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load solution: %w", err)
	}

	if errs := domain.FilterErrors(domain.ValidateGraph(solutionGraph, "Solution")); len(errs) > 0 {
		return nil, fmt.Errorf("solution has integrity errors")
	}

	stdPM := prematcher.NewStandardPreMatcher()
	solPM := prematcher.NewUMLSolutionPreMatcher()
	solForMatch, _ := solPM.ProcessSolution(solutionGraph)
	entityMatcher := matcher.NewStandardEntityMatcher(0.8)
	comp := comparator.NewStandardComparator()
	gr := grader.NewStandardGrader()
	rules := &grader.GradingRules{}

	entries, err := os.ReadDir(studentDir)
	if err != nil {
		return nil, fmt.Errorf("cannot read directory: %w", err)
	}
	var studentFiles []string
	for _, e := range entries {
		if !e.IsDir() && (strings.HasSuffix(e.Name(), ".drawio") || strings.HasSuffix(e.Name(), ".xml") || strings.HasSuffix(e.Name(), ".mmd") || strings.HasSuffix(e.Name(), ".mermaid")) {
			studentFiles = append(studentFiles, e.Name())
		}
	}
	if len(studentFiles) == 0 {
		return nil, fmt.Errorf("no UML files (.drawio, .mmd, .mermaid, .xml) found in %s", studentDir)
	}

	batchResult := &domain.BatchGradeResult{
		SolutionPath:   solutionPath,
		StudentResults: make(map[string]*domain.GradeResult),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	startTime := time.Now()

	for _, filename := range studentFiles {
		wg.Add(1)
		go func(fname string) {
			defer wg.Done()
			stuGraph, err := share.LoadGraph(filepath.Join(studentDir, fname))
			if err != nil {
				return
			}
			stuProc, _ := stdPM.Process(stuGraph)
			mapping, _ := entityMatcher.Match(solForMatch, stuProc)
			diffReport, _ := comp.Compare(solForMatch, stuProc, mapping)
			res, _ := gr.Grade(diffReport, solForMatch, stuProc, rules)

			mu.Lock()
			sname := strings.TrimSuffix(fname, filepath.Ext(fname))
			batchResult.StudentResults[sname] = res
			mu.Unlock()
		}(filename)
	}

	wg.Wait()

	res := &domain.BatchResult{
		BatchResult: batchResult,
		Duration:    time.Since(startTime),
		TotalFiles:  len(studentFiles),
	}

	if outputReport != "" {
		if !strings.HasSuffix(strings.ToLower(outputReport), ".csv") {
			outputReport += ".csv"
		}
		csvRep := report.NewCSVReporter(outputReport)
		if err := csvRep.GenerateReport(res.BatchResult); err != nil {
			return res, fmt.Errorf("CSV export failed: %w", err)
		}
	}

	return res, nil
}

func (s *StandardInstructorService) CompareUML(solutionPath, studentPath string, isAdmin bool) (*domain.CompareResult, error) {
	// Replicates cmd/compare logic
	solutionGraph, err := share.LoadGraph(solutionPath)
	if err != nil {
		return nil, fmt.Errorf("load solution: %w", err)
	}
	studentGraph, err := share.LoadGraph(studentPath)
	if err != nil {
		return nil, fmt.Errorf("load student: %w", err)
	}

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

	stdPM := prematcher.NewStandardPreMatcher()
	solPM := prematcher.NewUMLSolutionPreMatcher()

	solStd, _ := stdPM.Process(solutionGraph)
	stuProc, _ := stdPM.Process(studentGraph)
	solForMatch, _ := solPM.ProcessSolution(solutionGraph)

	entityMatcher := matcher.NewStandardEntityMatcher(0.8)
	mapping, _ := entityMatcher.Match(solForMatch, stuProc)

	comp := comparator.NewStandardComparator()
	diffReport, _ := comp.Compare(solForMatch, stuProc, mapping)

	gr := grader.NewStandardGrader()
	rules := &grader.GradingRules{}
	gradeResult, _ := gr.Grade(diffReport, solForMatch, stuProc, rules)

	// Filtering warnings
	warns := domain.FilterWarns(allIssues)
	outWarns := warns[:0]
	for _, w := range warns {
		lower := strings.ToLower(w.Message)
		if w.Code == "INCOMPLETE_ATTRIBUTE" && (strings.Contains(lower, "getter") || strings.Contains(lower, "setter")) {
			continue
		}
		outWarns = append(outWarns, w)
	}

	return &domain.CompareResult{
		SolProcessed: solForMatch,
		StuProcessed: stuProc,
		SolStd:       solStd,
		Mapping:      mapping,
		DiffReport:   diffReport,
		GradeResult:  gradeResult,
		Warnings:     outWarns,
	}, nil
}
