package report

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"uml_compare/domain"
)

// CSVReporter implements IReporter to export batch results to a CSV/Excel file
type CSVReporter struct {
	OutputPath string
}

func NewCSVReporter(path string) IReporter {
	return &CSVReporter{OutputPath: path}
}

func (c *CSVReporter) GenerateReport(batchResult *BatchGradeResult) error {
	file, err := os.Create(c.OutputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 1. Trích xuất dàn Columns header từ bất kỳ bài nộp nào có SolutionGraph
	var solGraph *domain.SolutionProcessedUMLGraph
	for _, res := range batchResult.StudentResults {
		if res != nil && res.SolutionGraph != nil {
			solGraph = res.SolutionGraph
			break
		}
	}

	if solGraph == nil {
		return fmt.Errorf("no valid SolutionGraph found in batch results to build headers")
	}

	headers := []string{"Student ID", "Status"}
	var colKeys []string

	// Node (Classes)
	for _, node := range solGraph.Nodes {
		headers = append(headers, fmt.Sprintf("[Class] %s", node.Name))
		colKeys = append(colKeys, "class:"+node.Name)

		// Attributes
		for _, attr := range node.Attributes {
			attrName := strings.Join(attr.Names, "|")
			headers = append(headers, fmt.Sprintf("[Attr] %s.%s", node.Name, attrName))
			colKeys = append(colKeys, "attr:"+node.Name+":"+attrName)
		}

		// Methods
		for _, meth := range node.Methods {
			methName := strings.Join(meth.Names, "|")
			headers = append(headers, fmt.Sprintf("[Meth] %s.%s", node.Name, methName))
			colKeys = append(colKeys, "meth:"+node.Name+":"+methName)
		}
	}

	// Relations (Edges)
	for _, edge := range solGraph.Edges {
		srcName := edge.SourceID
		tgtName := edge.TargetID
		for _, n := range solGraph.Nodes {
			if n.ID == edge.SourceID {
				srcName = n.Name
			}
			if n.ID == edge.TargetID {
				tgtName = n.Name
			}
		}
		// edge header doesn't strictly need to be edge ID if we can reconstruct it, but edge diff uses relationType
		headers = append(headers, fmt.Sprintf("[Rel] %s -(%s)-> %s", srcName, edge.RelationType, tgtName))
		colKeys = append(colKeys, "edge:"+edge.SourceID+":"+edge.TargetID+":"+edge.RelationType)
	}

	if err := writer.Write(headers); err != nil {
		return err
	}

	// 2. Điền số (1/0)
	for studentID, result := range batchResult.StudentResults {
		if result == nil || result.Report == nil {
			row := []string{studentID, "FAIL"}
			for i := 0; i < len(colKeys); i++ {
				row = append(row, "0")
			}
			writer.Write(row)
			continue
		}

		status := "FAIL"
		if result.CorrectPercent >= 60.0 {
			if result.CorrectPercent >= 90.0 {
				status = "EXCELLENT"
			} else {
				status = "PASS"
			}
		}

		row := []string{studentID, status}

		report := result.Report

		// Create lookup maps specific to CorrectDetail
		classLookup := make(map[string]bool)
		for _, cDiff := range report.CorrectDetail.Class {
			if cDiff.Sol != nil {
				classLookup[cDiff.Sol.Name] = true
			}
		}

		attrLookup := make(map[string]bool)
		for _, aDiff := range report.CorrectDetail.Attribute {
			if aDiff.Sol != nil {
				attrName := strings.Join(aDiff.Sol.Names, "|")
				attrLookup[aDiff.ParentClassName+":"+attrName] = true
			}
		}

		methLookup := make(map[string]bool)
		for _, mDiff := range report.CorrectDetail.Method {
			if mDiff.Sol != nil {
				methName := strings.Join(mDiff.Sol.Names, "|")
				methLookup[mDiff.ParentClassName+":"+methName] = true
			} else {
				// Getters và Setters được nhóm chung và trả về nil trong Comparator
				if strings.Contains(mDiff.Description, "getter(s) match") {
					for _, node := range solGraph.Nodes {
						if node.Name == mDiff.ParentClassName {
							for _, m := range node.Methods {
								if m.Type == "getter" {
									methLookup[mDiff.ParentClassName+":"+strings.Join(m.Names, "|")] = true
								}
							}
						}
					}
				} else if strings.Contains(mDiff.Description, "setter(s) match") {
					for _, node := range solGraph.Nodes {
						if node.Name == mDiff.ParentClassName {
							for _, m := range node.Methods {
								if m.Type == "setter" {
									methLookup[mDiff.ParentClassName+":"+strings.Join(m.Names, "|")] = true
								}
							}
						}
					}
				}
			}
		}

		edgeLookup := make(map[string]bool)
		for _, eDiff := range report.CorrectDetail.Edge {
			if eDiff.Sol != nil {
				edgeLookup[eDiff.Sol.SourceID+":"+eDiff.Sol.TargetID+":"+eDiff.Sol.RelationType] = true
			}
		}

		// Duyệt từng cột động để đính giá trị 1/0
		for _, col := range colKeys {
			parts := strings.SplitN(col, ":", 2)
			kind := parts[0]
			key := parts[1]

			val := "0"
			switch kind {
			case "class":
				if classLookup[key] {
					val = "1"
				}
			case "attr":
				if attrLookup[key] {
					val = "1"
				}
			case "meth":
				if methLookup[key] {
					val = "1"
				}
			case "edge":
				if edgeLookup[key] {
					val = "1"
				}
			}
			row = append(row, val)
		}

		writer.Write(row)
	}

	return nil
}
