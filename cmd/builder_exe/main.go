package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"uml_compare/AppBuilder"
)

func main() {
	fmt.Println("=================================================================")
	fmt.Println("          UML Comparator — Central Build Orchestrator            ")
	fmt.Println("=================================================================")
	fmt.Println("Please select the build pipeline you wish to execute:")
	fmt.Println("  [1] Universal Build (Builds standard Student CLI/GUI & Teacher tools)")
	fmt.Println("  [2] Exam Build (Produces a specialized GUI with embedded solutions)")
	fmt.Print("\nEnter selection (1 or 2): ")

	reader := bufio.NewReader(os.Stdin)
	selection, _ := reader.ReadString('\n')
	selection = strings.TrimSpace(selection)

	var builder AppBuilder.Builder

	// Instantiate shared ISP tools
	dirPrep := &AppBuilder.StandardDirManager{}
	taskBld := &AppBuilder.GoTaskBuilder{}

	switch selection {
	case "1":
		// Only Universal build requires module tidy dependency
		depMan := &AppBuilder.GoDependencyManager{}
		builder = AppBuilder.NewUniversalBuilder(dirPrep, depMan, taskBld)

	case "2":
		var paths []string
		fmt.Println("\nLet's add .drawio solutions for the exam.")
		for {
			fmt.Print("Enter .drawio file path or folder:\n> ")
			path, _ := reader.ReadString('\n')
			path = strings.TrimSpace(path)
			
			if path != "" {
				// Remove quotes if present
				if len(path) > 2 && path[0] == '"' && path[len(path)-1] == '"' {
					path = path[1 : len(path)-1]
				}
				paths = append(paths, path)
			} else if len(paths) == 0 {
				fmt.Println("❌ Error: Path cannot be empty.")
				continue
			}

			fmt.Print("\nAdd more solution files? (y/N): ")
			more, _ := reader.ReadString('\n')
			more = strings.TrimSpace(strings.ToLower(more))
			if more != "y" && more != "yes" {
				break
			}
			fmt.Println()
		}

		solutionsPath := strings.Join(paths, ",")

		// Only Exam build requires asset copying dependency
		assetCopier := &AppBuilder.FileAssetCopier{}
		builder = AppBuilder.NewExamBuilder(dirPrep, assetCopier, taskBld, solutionsPath)

	default:
		fmt.Println("❌ Invalid selection. Exiting.")
		os.Exit(1)
	}

	err := builder.Build()
	if err != nil {
		fmt.Printf("❌ Fatal Build Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nPress Enter to exit...")
	reader.ReadString('\n')
}
