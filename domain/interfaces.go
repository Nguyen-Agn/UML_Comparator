package domain

// UMLProcessor interface handles parsing and comparing of UML diagrams
type UMLProcessor interface {
	// Process takes solution and assignment paths and returns the GradeResult
	Process(solutionPath, assignmentPath string) (*GradeResult, error)

	// ExportHTML saves the generated report to a specific file path
	ExportHTML(result *GradeResult, outputPath string) error

	// IsAIAvailable returns true if the AI matcher is ready
	IsAIAvailable() bool
}

// MainController interface manages the application flow
type MainController interface {
	// OnSubmit is called when the user submits their files
	OnSubmit(solutionPath string, assignmentPath string)

	// OnExport is called when the user wants to export HTML
	OnExport(saveFilePath string)
}

// MainView interface abstract the GUI updating logic
type MainView interface {
	// SetController injects the controller dependency
	SetController(c MainController)

	// ShowError shows an error dialog or message
	ShowError(err error)

	// ShowLoading indicates the app is processing
	ShowLoading()

	// ShowResult renders the GradeResult in a beautiful layout
	ShowResult(result *GradeResult)

	// EnableExport turns on the export button
	EnableExport()

	// ShowAIStatus indicates if AI matching is active
	ShowAIStatus(available bool)

	// Wait blockingly until the window is closed
	Wait()

	// Close terminates the window
	Close()
}

type IHybridMatcher interface {
	Compare(s1, s2 string) float64
	CompareMultiple(candidate string, optionals []string) (float64, string)
	// GetThreshold returns the similarity threshold from configuration
	GetThreshold() float64
	// IsAIAvailable returns true if the AI model was loaded successfully
	IsAIAvailable() bool
	Close() error
}
