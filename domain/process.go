package domain

import (
	"time"
)

type BatchResult struct {
	BatchResult *BatchGradeResult
	Duration    time.Duration
	TotalFiles  int
}

// CompareResult wraps the output of a Live Comparison check.
type CompareResult struct {
	SolProcessed *SolutionProcessedUMLGraph
	StuProcessed *ProcessedUMLGraph
	SolStd       *ProcessedUMLGraph // dùng cho edge comparison display
	Mapping      MappingTable
	DiffReport   *DiffReport
	GradeResult  *GradeResult
	Warnings     []IntegrityError
}

// BatchGradeResult holds the results of grading multiple student submissions against a single solution.
type BatchGradeResult struct {
	SolutionPath   string
	StudentResults map[string]*GradeResult // Key: Student file path or ID, Value: The grading result
}
