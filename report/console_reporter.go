package report

import (
	"fmt"
	"strings"
	"uml_compare/domain"
)

// ConsoleReporter is a simple implementation of IReporter that outputs to the terminal.
type ConsoleReporter struct{}

// NewConsoleReporter creates a new instance of ConsoleReporter.
func NewConsoleReporter() IReporter {
	return &ConsoleReporter{}
}

// GenerateReport generates a simple text report to the terminal.
func (c *ConsoleReporter) GenerateReport(batchResult *domain.BatchGradeResult) error {
	fmt.Println()
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println(" BATCH GRADING REPORT")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf(" Solution File: %s\n", batchResult.SolutionPath)
	fmt.Printf(" Total Submissions: %d\n", len(batchResult.StudentResults))
	fmt.Println(strings.Repeat("-", 80))

	for studentID, result := range batchResult.StudentResults {
		if result == nil {
			fmt.Printf(" %-30s | [ERROR] Result is nil\n", studentID)
			continue
		}

		status := "FAIL"
		colorCode := "\033[31m" // Red
		resetCode := "\033[0m"

		if result.CorrectPercent >= 60.0 {
			if result.CorrectPercent >= 90.0 {
				status = "EXCELLENT"
				colorCode = "\033[32m" // Green
			} else {
				status = "PASS"
				colorCode = "\033[33m" // Yellow
			}
		}

		fmt.Printf(" %-30s | Score: %5.2f/%-5.2f (%6.2f%%) | %s%s%s\n",
			studentID, result.TotalScore, result.MaxScore, result.CorrectPercent, colorCode, status, resetCode)

		if len(result.Feedbacks) > 0 {
			fmt.Println("   Feedbacks:")
			// Lấy 3 feedback đầu tiên để report không quá dài
			limit := len(result.Feedbacks)
			if limit > 3 {
				limit = 3
			}
			for i := 0; i < limit; i++ {
				fmt.Printf("    - %s\n", result.Feedbacks[i])
			}
			if len(result.Feedbacks) > 3 {
				fmt.Printf("    ... and %d more issues.\n", len(result.Feedbacks)-3)
			}
		}
		fmt.Println(strings.Repeat("-", 80))
	}

	return nil
}
