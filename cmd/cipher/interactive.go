// cmd/cipher/interactive.go
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"uml_compare/cmd/share"
)

// runInteractiveLoop provides a terminal UI for teachers to encrypt solutions.
func runInteractiveLoop(app *EncryptorApp) {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("--------------------------------------------------------")
		fmt.Println("  UML Solution Encryptor — Interactive Mode")
		fmt.Println("--------------------------------------------------------")

		input := share.Prompt(scanner, "[1/2] Input (.drawio) file path")
		if input == "" {
			fmt.Println("❌ Path cannot be empty.")
			continue
		}
		
		if _, err := os.Stat(input); os.IsNotExist(err) {
			fmt.Printf("❌ File not found: %s\n", input)
			continue
		}

		app.InputPath = input
		app.OutputPath = share.Prompt(scanner, "[2/2] Output (.solution) path (Enter = auto)")

		fmt.Println("\n🚀 Processing...")

		if err := app.Run(); err != nil {
			fmt.Printf("\n❌ Encryption failed: %v\n", err)
		}

		fmt.Println("\n--------------------------------------------------------")
		ans := share.Prompt(scanner, "Encrypt another file? (Y/N)")
		if strings.ToLower(ans) != "y" {
			fmt.Println("\nGoodbye!")
			break
		}
		fmt.Println()
		
		// Reset paths for next loop
		app.InputPath = ""
		app.OutputPath = ""
	}
}
