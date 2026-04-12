// cmd/parse/main.go - Thử nghiệm parser: go run ./cmd/parse/main.go [<file.drawio>]
package main

import (
	"fmt"
	"os"
	"uml_compare/cmd/share"
	"uml_compare/domain"
	"uml_compare/parser"
)

func main() {
	filePath := "parser/testdata/plain_sample.drawio"
	if len(os.Args) > 1 {
		filePath = os.Args[1]
	}

	share.PrintBanner("UML Compare - Draw.io Parser Demo")
	fmt.Printf("📂 Input file : %s\n\n", filePath)

	data, err := run(filePath)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		os.Exit(1)
	}

	printResult(data)
}

// run thực hiện parse file và trả về raw model data.
func run(filePath string) (domain.RawModelData, error) {
	p, err := parser.GetParser(filePath)
	if err != nil {
		return "", err
	}
	return p.Parse(filePath)
}

// printResult in kết quả raw model ra stdout.
func printResult(data domain.RawModelData) {
	fmt.Println("✅ Parse thành công!")
	fmt.Printf("📏 Độ dài chuỗi dữ liệu sau parse: %d ký tự\n", len(data))
	fmt.Println("\n─── Raw Output (domain.RawModelData) ──────────────────")
	fmt.Println(string(data))
	fmt.Println("───────────────────────────────────────────────────────")
}
