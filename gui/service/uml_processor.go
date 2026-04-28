package service

import (
	"fmt"
	"uml_compare/domain"
	"uml_compare/similarity"
	"uml_compare/src/builder"
	"uml_compare/src/comparator"
	"uml_compare/src/grader"
	"uml_compare/src/matcher"
	"uml_compare/src/parser"
	"uml_compare/src/prematcher"
	"uml_compare/src/visualizer"
	"sync"
)

type StandardUMLProcessor struct {
	matcher domain.IHybridMatcher
	mu      sync.Mutex
}

// NewStandardUMLProcessor provides a new StandardUMLProcessor
func NewStandardUMLProcessor() domain.UMLProcessor {
	return &StandardUMLProcessor{}
}

func (p *StandardUMLProcessor) getMatcher() (domain.IHybridMatcher, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.matcher != nil {
		return p.matcher, nil
	}
	m, err := similarity.GetHybridMatcher(domain.DefaultConfig)
	if err != nil {
		return nil, err
	}
	p.matcher = m
	return p.matcher, nil
}

// Process takes solution and assignment paths and returns the GradeResult
func (p *StandardUMLProcessor) Process(solutionPath, assignmentPath string) (*domain.GradeResult, error) {
	parserObjSol, err := parser.GetParser(solutionPath)
	if err != nil {
		return nil, fmt.Errorf("Get parser error: %v", err)
	}

	solRaw, solType, err := parserObjSol.Parse(solutionPath)
	if err != nil {
		return nil, fmt.Errorf("Parse solution error: %v", err)
	}

	parserObjStu, err := parser.GetParser(assignmentPath)
	if err != nil {
		return nil, fmt.Errorf("Get parser error: %v", err)
	}

	stuRaw, stuType, err := parserObjStu.Parse(assignmentPath)
	if err != nil {
		return nil, fmt.Errorf("Parse student error: %v", err)
	}
	// If Parse success
	// Initialize semantic matcher

	similar_component, err := p.getMatcher()
	if err != nil {
		return nil, fmt.Errorf("Initialize semantic matcher error: %v", err)
	}

	b := builder.NewStandardModelBuilder()
	solGraph, err := b.Build(solRaw, solType)
	if err != nil {
		return nil, fmt.Errorf("Build solution error: %v", err)
	}
	stuGraph, err := b.Build(stuRaw, stuType)
	if err != nil {
		return nil, fmt.Errorf("Build student error: %v", err)
	}

	allIssues := append(
		domain.ValidateGraph(solGraph, "Solution"),
		domain.ValidateGraph(stuGraph, "Student")...,
	)
	hardErrors := domain.FilterErrors(allIssues)
	if len(hardErrors) > 0 {
		return nil, fmt.Errorf("Integrity errors: %v", hardErrors[0])
	}

	stdPM := prematcher.NewStandardPreMatcher()
	solPM := prematcher.NewUMLSolutionPreMatcher()

	stuProc, _ := stdPM.Process(stuGraph)
	solForMatch, _ := solPM.ProcessSolution(solGraph)

	entityMatcher := matcher.NewStandardEntityMatcher(similar_component)
	mapping, _ := entityMatcher.Match(solForMatch, stuProc)

	comp := comparator.NewStandardComparator(similar_component)
	diffReport, _ := comp.Compare(solForMatch, stuProc, mapping)

	gr := grader.NewStandardGrader()
	rules := &grader.GradingRules{}
	gradeResult, _ := gr.Grade(diffReport, solForMatch, stuProc, rules)

	return gradeResult, nil
}

// ExportHTML saves the generated report to a specific file path
func (p *StandardUMLProcessor) ExportHTML(result *domain.GradeResult, outputPath string) error {
	vis := visualizer.NewHTMLVisualizer()
	// Using ExportStudentHTML since this offline GUI is for students.
	return vis.ExportStudentHTML(result, outputPath)
}

func (p *StandardUMLProcessor) IsAIAvailable() bool {
	m, err := p.getMatcher()
	if err != nil {
		return false
	}
	return m.IsAIAvailable()
}
