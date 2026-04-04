package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type BuildConfig struct {
	Name    string
	Output  string
	Sources []string
	IsGUI   bool
}

func main() {
	fmt.Println("========================================================")
	fmt.Println("       UML Comparator — Universal Build System")
	fmt.Println("========================================================")

	// 1. Ensure portable directory exists
	portableDir := "portable"
	if _, err := os.Stat(portableDir); os.IsNotExist(err) {
		fmt.Printf("📁 Creating directory: %s\n", portableDir)
		_ = os.Mkdir(portableDir, 0755)
	}

	// 2. Define build tasks
	tasks := []BuildConfig{
		{
			Name:    "Student GUI (Morning Dawn)",
			Output:  filepath.Join(portableDir, "student_uml.exe"),
			Sources: []string{"./gui/main.go"},
			IsGUI:   true,
		},
		{
			Name:    "Student CLI (Smart Fallback)",
			Output:  filepath.Join(portableDir, "student_uml_cli.exe"),
			Sources: []string{"./cmd/visualize/main.go", "./cmd/visualize/interactive.go"},
			IsGUI:   false,
		},
		{
			Name:    "Lecture Parallel (Batch Grading)",
			Output:  filepath.Join(portableDir, "lecture_cli_parallel.exe"),
			Sources: []string{"./cmd/grade_batch/main.go", "./cmd/grade_batch/interactive.go"},
			IsGUI:   false,
		},
	}

	start := time.Now()

	// 3. Run go mod tidy first
	fmt.Print("📦 Tidying Go modules... ")
	if err := runCommand("go", "mod", "tidy"); err != nil {
		fmt.Printf("❌ FAILED\nError: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Done")

	// 4. Run builds
	for i, task := range tasks {
		fmt.Printf("[%d/%d] Building %s...\n", i+1, len(tasks), task.Name)
		
		args := []string{"build"}
		if task.IsGUI {
			args = append(args, "-ldflags=-H windowsgui")
		}
		args = append(args, "-o", task.Output)
		args = append(args, task.Sources...)

		if err := runCommand("go", args...); err != nil {
			fmt.Printf("   ❌ Build FAILED for %s\n   Error: %v\n", task.Name, err)
			continue
		}
		fmt.Printf("   ✅ Success -> %s\n", task.Output)
	}

	fmt.Println("========================================================")
	fmt.Printf("✨ ALL BUILDS COMPLETED in %v\n", time.Since(start).Round(time.Millisecond))
	fmt.Println("   Check the 'portable' folder for your executables.")
	fmt.Println("========================================================")
	
	// Keep window open if run via double-click
	fmt.Println("\nPress Enter to exit...")
	fmt.Scanln()
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
