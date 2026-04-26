package domain

// InstructorController interface manages the application flow for the instructor suite.
type InstructorController interface {
	// Security Tab
	OnEncrypt(inputPath, outputPath string)

	// Distribute Tab
	OnBuildExam(solutionsDir, outputPath string)

	// Batch Tab
	OnGradeBatch(solutionPath, studentDir, outputPath string)

	// Live Tab
	OnLiveCompare(solutionPath, studentPath string)
}

// InstructorView interface abstract the GUI updating logic for the instructor suite.
type InstructorView interface {
	SetController(c InstructorController)
	ShowError(err error)
	ShowSuccess(msg string)
	ShowLoading()
	HideLoading()

	// Called by controller when live compare finishes
	ShowLiveCompareResult(result *CompareResult)

	Wait()
	Close()
}
