// Package main implements the teacher-only solution encryption CLI.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"uml_compare/cipher"
)

func main() {
	app := &EncryptorApp{}
	// Đảm bảo luôn đợi người dùng nhấn phím trước khi đóng cửa sổ
	defer app.WaitIfInteractive()

	if len(os.Args) <= 1 {
		app.PrintBanner()
		fmt.Printf("   Tip: You can also drag & drop files here!\n\n")
		runInteractiveLoop(app)
		return
	}

	if !app.ParseArgs() {
		return
	}

	app.PrintBanner()

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "\n❌ Error: %v\n", err)
		return
	}
}

// ─── EncryptorApp ─────────────────────────────────────────────────────────────

// EncryptorApp encapsulates the state and logic for the solution encryption tool.
type EncryptorApp struct {
	InputPath  string
	OutputPath string
	CustomKey  string
}

// ParseArgs handles flag parsing and basic validation.
// Returns false if the application should terminate early (e.g., -help or missing args).
func (a *EncryptorApp) ParseArgs() bool {
	flag.StringVar(&a.OutputPath, "o", "", "Output .solution path")
	flag.StringVar(&a.CustomKey, "k", "", "Custom encryption key")

	flag.Usage = func() {
		exe := filepath.Base(os.Args[0])
		fmt.Fprintf(os.Stderr, "Usage: %s <input.drawio> [options]\n", exe)
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	// Positional argument check
	a.InputPath = flag.Arg(0)
	if a.InputPath == "" {
		fmt.Fprintln(os.Stderr, "Error: missing input .drawio file")
		flag.Usage()
		return false
	}

	return true
}

// PrintBanner displays the tool's header.
func (a *EncryptorApp) PrintBanner() {
	fmt.Println("╔══════════════════════════════════════╗")
	fmt.Println("║  UML Comparator — Solution Encryptor ║")
	fmt.Println("║  [Teacher Tool — Students skip this] ║")
	fmt.Println("╚══════════════════════════════════════╝")
}

// Run executes the core business logic.
func (a *EncryptorApp) Run() error {
	// 1. Validate Input
	if !strings.HasSuffix(a.InputPath, ".drawio") {
		return fmt.Errorf("input file must be a .drawio file")
	}

	if _, err := os.Stat(a.InputPath); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", a.InputPath)
	}

	// 2. Resolve output path
	if a.OutputPath == "" {
		a.OutputPath = strings.TrimSuffix(a.InputPath, ".drawio") + ".solution"
	}
	if !strings.HasSuffix(strings.ToLower(a.OutputPath), ".solution") {
		a.OutputPath += ".solution"
	}

	// 3. Initialize Cipher (Strategy selection)
	c := a.buildCipher()

	// 4. Perform Encryption
	fmt.Printf("\n⏳ Encrypting: %s → %s\n", a.InputPath, a.OutputPath)
	if err := c.Encrypt(a.InputPath, a.OutputPath); err != nil {
		return err
	}

	fmt.Printf("✅ Done! Encrypted file saved to: %s\n", a.OutputPath)
	fmt.Println("\nℹ️  Share the .solution file with students.")
	fmt.Println("   The grader will auto-decrypt it — no key needed on their end.")
	return nil
}

// buildCipher selects the cipher implementation based on user input.
func (a *EncryptorApp) buildCipher() cipher.ISolutionCipher {
	if a.CustomKey != "" {
		fmt.Printf("🔑 Using custom key\n")
		return cipher.NewWithKey([]byte(a.CustomKey))
	}

	if env := os.Getenv("SOLUTION_KEY"); env != "" {
		fmt.Printf("🔑 Using SOLUTION_KEY environment variable\n")
		return cipher.New()
	}

	fmt.Printf("🔑 Using built-in default key\n")
	return cipher.New()
}

// WaitIfInteractive keeps the console open on Windows when run via GUI or Drag-and-Drop.
func (a *EncryptorApp) WaitIfInteractive() {
	// Nếu chạy bằng cách bấm đúp (1 arg) hoặc kéo thả file (2 args), ta cần đợi người dùng xem kết quả
	if len(os.Args) <= 2 {
		fmt.Print("\nPress Enter to exit...")
		var tmp string
		fmt.Scanln(&tmp)
	}
}
