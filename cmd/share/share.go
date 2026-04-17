// Package share cung cấp các tiện ích CLI dùng chung cho tất cả tool trong cmd/.
// Bao gồm: hằng số màu ANSI, load UML graph, mở file theo OS, prompt nhập liệu, in banner.
package share

import (
	"bufio"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"uml_compare/builder"
	"uml_compare/domain"
	"uml_compare/parser"
)

// ── ANSI Color Constants ───────────────────────────────────────────────────────

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Cyan   = "\033[36m"
	Bold   = "\033[1m"
)

// ── Banner ────────────────────────────────────────────────────────────────────

// PrintBanner in tiêu đề dạng box ASCII cho tool CLI.
// title nên ngắn gọn, vừa trong 50 ký tự.
func PrintBanner(title string) {
	border := strings.Repeat("═", 52)
	fmt.Printf("╔%s╗\n", border)
	fmt.Printf("║  %-50s║\n", title)
	fmt.Printf("╚%s╝\n", border)
}

// ── Graph Loading ─────────────────────────────────────────────────────────────

// LoadGraph thực hiện toàn bộ pipeline Parse → Build và trả về UMLGraph.
// Trả về lỗi mô tả rõ bước nào thất bại.
func LoadGraph(filePath string) (*domain.UMLGraph, error) {
	p, err := parser.GetParser(filePath)
	if err != nil {
		return nil, fmt.Errorf("parser factory: %w", err)
	}
	rawXML, sourceType, err := p.Parse(filePath)
	if err != nil {
		return nil, fmt.Errorf("parse %q: %w", filePath, err)
	}

	b, err := builder.GetBuilder(sourceType)
	if err != nil {
		return nil, fmt.Errorf("builder factory for %q: %w", sourceType, err)
	}

	graph, err := b.Build(rawXML, sourceType)
	if err != nil {
		return nil, fmt.Errorf("build graph from %q (%s): %w", filePath, sourceType, err)
	}
	return graph, nil
}

// ── File / Browser Opener ─────────────────────────────────────────────────────

// OpenFile mở file bằng ứng dụng mặc định của OS (hữu ích với .html, .csv).
// Lỗi được bỏ qua vì đây là tính năng tiện ích, không ảnh hưởng kết quả chính.
func OpenFile(path string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", path)
	case "darwin":
		cmd = exec.Command("open", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	_ = cmd.Start()
}

// ── Interactive Prompt ────────────────────────────────────────────────────────

// Prompt hiển thị nhãn và đọc một dòng nhập từ stdin.
// Trả về chuỗi đã trim whitespace, hoặc "" nếu không có input.
func Prompt(scanner *bufio.Scanner, label string) string {
	fmt.Printf("  %s: ", label)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text())
	}
	return ""
}
