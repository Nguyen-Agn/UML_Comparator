package AppBuilder

import (
	"fmt"
	"path/filepath"
	"time"
)

// ExamBuilder implements Builder specifically for the integrated solution GUI
type ExamBuilder struct {
	dirPrep      DirectoryPreparer
	assetCopier  AssetCopier
	taskBld      TaskBuilder
	solutionsDir string
}

// NewExamBuilder injects dependencies
func NewExamBuilder(dirPrep DirectoryPreparer, assetCopier AssetCopier, taskBld TaskBuilder, solutionsDir string) Builder {
	return &ExamBuilder{
		dirPrep:      dirPrep,
		assetCopier:  assetCopier,
		taskBld:      taskBld,
		solutionsDir: solutionsDir,
	}
}

func (e *ExamBuilder) Build() error {
	fmt.Println("========================================================")
	fmt.Println("       UML Comparator — Exam Version Build System       ")
	fmt.Println("========================================================")

	portableDir := "portable"
	if err := e.dirPrep.Prepare(portableDir); err != nil {
		return fmt.Errorf("failed preparing portable dir: %w", err)
	}

	embeddedDir := filepath.Join("cmd", "exam_gui", "embedded_solutions")
	if err := e.dirPrep.Clear(embeddedDir); err != nil {
		return fmt.Errorf("failed clearing embedded dir: %w", err)
	}

	fmt.Printf("📦 Copying exam solutions from: %s\n", e.solutionsDir)
	err := e.assetCopier.CopyAssets(e.solutionsDir, embeddedDir, ".drawio")
	if err != nil {
		return fmt.Errorf("failed copying assets: %w", err)
	}

	start := time.Now()
	fmt.Println("\n🔨 Building Exam GUI...")

	outPath := filepath.Join(portableDir, "exam_student_uml.exe")
	if err := e.taskBld.BuildTask("Exam Student GUI", outPath, []string{"./cmd/exam_gui/main.go"}, true); err != nil {
		return fmt.Errorf("exam build failed: %w", err)
	}

	fmt.Printf("✅ Success -> %s\n", outPath)
	fmt.Println("========================================================")
	fmt.Printf("✨ BUILD COMPLETED in %v\n", time.Since(start).Round(time.Millisecond))
	fmt.Println("   Distribute the 'portable/exam_student_uml.exe' file to your students.")
	fmt.Println("========================================================")

	return nil
}
