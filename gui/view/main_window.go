package view

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ncruces/zenity"
	"github.com/zserge/lorca"

	"uml_compare/domain"

	_ "embed"
	"uml_compare/src/visualizer"
)

//go:embed main_window.html
var htmlContent string

type lorcaMainView struct {
	ui         lorca.UI
	controller domain.MainController
	solPath    string
	stuPath    string

	dialogBusy bool // Prevent multiple simultaneous dialogs
}

// NewMainView instantiates the Lorca app
func NewMainView() (domain.MainView, error) {
	fmt.Println("Initializing Lorca UI...")

	// Convert HTML to Base64 to avoid encoding issues with special characters/newlines
	b64Content := base64.StdEncoding.EncodeToString([]byte(htmlContent))
	url := "data:text/html;base64," + b64Content

	// Start Lorca window with a blank page to avoid command-line length limits on Windows.
	// We then load the content immediately after.
	ui, err := lorca.New("", "", 1000, 800, "--remote-allow-origins=*")
	if err != nil {
		fmt.Printf("Lorca New Error: %v\n", err)
		return nil, err
	}
	fmt.Println("Lorca window opened successfully.")

	// Load the actual content
	ui.Load(url)

	// Optimize window title
	ui.Eval(`document.title = "UML Visual Grader - Student Edition"`)

	v := &lorcaMainView{
		ui: ui,
	}

	// Bind Go functions to JS

	v.ui.Bind("goChooseStu", v.chooseStu)
	v.ui.Bind("goChooseSol", v.chooseSol)
	v.ui.Bind("goSubmit", v.submit)
	v.ui.Bind("goExport", v.export)

	return v, nil
}

func (v *lorcaMainView) SetController(c domain.MainController) {
	v.controller = c
}

// GUI Binding methods called from JS
func (v *lorcaMainView) chooseSol() {
	if v.dialogBusy {
		return
	}
	v.dialogBusy = true
	defer func() { v.dialogBusy = false }()

	// Force focus to front before opening dialog
	v.ui.Eval(`window.focus()`)

	file, err := zenity.SelectFile(
		zenity.Title("Select Solution (.drawio)"),
		zenity.FileFilters{zenity.FileFilter{Name: "UML Diagrams", Patterns: []string{"*.drawio", "*.solution", "*.mmd", "*.mermaid", "*.drawio", "*.xml"}}},
	)
	if err == nil && file != "" {
		v.solPath = file
		v.ui.Eval(fmt.Sprintf(`setFile("sol", "%s")`, filepath.Base(file)))
	}
}

func (v *lorcaMainView) chooseStu() {
	if v.dialogBusy {
		return
	}
	v.dialogBusy = true
	defer func() { v.dialogBusy = false }()

	// Force focus to front before opening dialog
	v.ui.Eval(`window.focus()`)

	file, err := zenity.SelectFile(
		zenity.Title("Select Assignment (.drawio)"),
		zenity.FileFilters{zenity.FileFilter{Name: "UML Diagrams", Patterns: []string{"*.drawio", "*.mmd", "*.mermaid", "*.drawio.xml"}}},
	)
	if err == nil && file != "" {
		v.stuPath = file
		v.ui.Eval(fmt.Sprintf(`setFile("stu", "%s")`, filepath.Base(file)))
	}
}

func (v *lorcaMainView) submit() {
	if v.solPath == "" || v.stuPath == "" {
		zenity.Error("Please select both Solution and Assignment files first.", zenity.Title("Error"))
		return
	}

	if strings.HasSuffix(strings.ToLower(v.stuPath), ".solution") {
		zenity.Error("Student assignment cannot be a .solution file. Please use a .drawio file.", zenity.Title("Invalid File Format"))
		return
	}

	fmt.Printf("Submit triggered: Sol=%s, Stu=%s\n", v.solPath, v.stuPath)
	if v.controller != nil {
		v.controller.OnSubmit(v.solPath, v.stuPath)
	}
}

func (v *lorcaMainView) export() {
	file, err := zenity.SelectFileSave(
		zenity.Title("Export HTML Report"),
		zenity.Filename("student_report.html"),
		zenity.FileFilters{zenity.FileFilter{Name: "HTML Document", Patterns: []string{"*.html"}}},
	)
	if err == nil && file != "" {
		if v.controller != nil {
			v.controller.OnExport(file)
		}
	}
}

// domain.MainView Implementations
func (v *lorcaMainView) ShowError(err error) {
	v.ui.Eval(`document.getElementById("loading").style.display = "none"`)
	v.ui.Eval(`document.getElementById("btnSubmit").disabled = false`)
	zenity.Error(err.Error(), zenity.Title("Processing Error"))
}

func (v *lorcaMainView) ShowLoading() {
	v.ui.Eval(`
		document.getElementById("loading").style.display = "flex";
		const res = document.getElementById("results");
		res.classList.remove("ready");
	`)
}

func (v *lorcaMainView) ShowResult(result *domain.GradeResult) {
	fmt.Println("Result received, generating HTML...")
	// Re-use logic to generate the full student feedback HTML snippet via Temp File
	tmpPath := filepath.Join(os.TempDir(), "uml_tmp_student_report.html")
	vis := visualizer.NewHTMLVisualizer()
	err := vis.ExportStudentHTML(result, tmpPath)

	if err != nil {
		fmt.Printf("ExportStudentHTML Error: %v\n", err)
		v.ui.Eval(`document.getElementById("loading").style.display = "none"`)
		zenity.Error("Result Generation Error: "+err.Error(), zenity.Title("Error"))
		return
	}

	// Read it back
	b, err := os.ReadFile(tmpPath)
	if err != nil {
		fmt.Printf("ReadFile Error: %v\n", err)
		v.ui.Eval(`document.getElementById("loading").style.display = "none"`)
		zenity.Error("Read Output Error: "+err.Error(), zenity.Title("Error"))
		return
	}

	fmt.Printf("HTML read successfully (%d bytes). Injecting into UI...\n", len(b))

	// Hide loading via renderReport JS for better sync
	b64 := base64.StdEncoding.EncodeToString(b)
	v.ui.Eval(fmt.Sprintf("renderReport('%s')", b64))
	fmt.Println("Render script sent.")
}

func (v *lorcaMainView) EnableExport() {
	v.ui.Eval(`document.getElementById("btnExport").disabled = false`)
}

func (v *lorcaMainView) ShowAIStatus(available bool) {
	v.ui.Eval(fmt.Sprintf("showAIStatus(%v)", available))
}

// Window logic
func (v *lorcaMainView) Wait() {
	<-v.ui.Done()
}

func (v *lorcaMainView) Close() {
	v.ui.Close()
}
