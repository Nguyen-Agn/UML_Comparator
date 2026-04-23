package AppBuilder

import (
	"fmt"
	"path/filepath"
	"time"
)

type universalBuildConfig struct {
	Name    string
	Output  string
	Sources []string
	IsGUI   bool
}

// UniversalBuilder implements Builder specifically for the standard CLI/GUI suite
type UniversalBuilder struct {
	dirPrep DirectoryPreparer
	depMan  DependencyManager
	taskBld TaskBuilder
}

// NewUniversalBuilder injects dependencies
func NewUniversalBuilder(dirPrep DirectoryPreparer, depMan DependencyManager, taskBld TaskBuilder) Builder {
	return &UniversalBuilder{
		dirPrep: dirPrep,
		depMan:  depMan,
		taskBld: taskBld,
	}
}

func (u *UniversalBuilder) Build() error {
	fmt.Println("========================================================")
	fmt.Println("       UML Comparator — Universal Build System          ")
	fmt.Println("========================================================")

	portableDir := "portable"
	if err := u.dirPrep.Prepare(portableDir); err != nil {
		return fmt.Errorf("directory preparation failed: %w", err)
	}

	tasks := []universalBuildConfig{
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
		{
			Name:    "Teacher Cipher (Solution Encryptor) [OPTIONAL]",
			Output:  filepath.Join(portableDir, "teacher_cipher.exe"),
			Sources: []string{"./cmd/cipher/main.go", "./cmd/cipher/interactive.go"},
			IsGUI:   false,
		},
		{
			Name:    "Instructor Suite (All-in-one GUI)",
			Output:  filepath.Join(portableDir, "instructor_suite.exe"),
			Sources: []string{"./cmd/instructor/main.go"},
			IsGUI:   true,
		},
	}

	start := time.Now()

	fmt.Print("📦 Tidying Go modules... ")
	if err := u.depMan.Tidy(); err != nil {
		return fmt.Errorf("failed to tidy dependencies: %w", err)
	}
	fmt.Println("✅ Done")

	for i, task := range tasks {
		fmt.Printf("[%d/%d] Building %s...\n", i+1, len(tasks), task.Name)
		if err := u.taskBld.BuildTask(task.Name, task.Output, task.Sources, task.IsGUI); err != nil {
			fmt.Printf("   ❌ Build FAILED: %v\n", err)
			continue
		}
		fmt.Printf("   ✅ Success -> %s\n", task.Output)
	}

	fmt.Println("========================================================")
	fmt.Printf("✨ ALL BUILDS COMPLETED in %v\n", time.Since(start).Round(time.Millisecond))
	fmt.Println("   Check the 'portable' folder for your executables.")
	fmt.Println("========================================================")

	return nil
}
