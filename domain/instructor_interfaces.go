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

	// Config Tab
	OnUpdateConfig(threshold float64, useAI bool)
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

	// UpdateConfigUI updates the config tab with current values
	UpdateConfigUI(threshold float64, useAI bool, aiAvailable bool)

	Wait()
	Close()
}
