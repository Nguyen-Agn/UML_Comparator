// cmd/parse/main.go - File thử nghiệm parser: chạy bằng: go run ./cmd/parse/main.go
package main

import (
	"fmt"
	"os"
	"uml_compare/parser"
)

func main() {
	filePath := "parser/testdata/plain_sample.drawio"
	if len(os.Args) > 1 {
		filePath = os.Args[1]
	}

	fmt.Println("╔══════════════════════════════════════════════════╗")
	fmt.Println("║        UML Compare - Draw.io Parser Demo         ║")
	fmt.Println("╚══════════════════════════════════════════════════╝")
	fmt.Printf("📂 Input file : %s\n\n", filePath)

	p := parser.NewDrawioParser()
	rawModel, err := p.Parse(filePath)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Parse thành công!")
	fmt.Printf("📏 Độ dài chuỗi dữ liệu sau parse: %d ký tự\n", len(rawModel))
	fmt.Println("\n─── Raw Output (domain.RawModelData) ──────────────────")
	fmt.Println(string(rawModel))
	fmt.Println("───────────────────────────────────────────────────────")
}
