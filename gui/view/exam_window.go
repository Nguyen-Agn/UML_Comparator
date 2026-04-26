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
	coreDomain "uml_compare/domain"
	"uml_compare/src/visualizer"
)

type examLorcaMainView struct {
	ui            lorca.UI
	controller    domain.MainController
	embeddedFiles map[string][]byte
	solPath       string // Represents the path to the written temp solution file
	stuPath       string

	dialogBusy bool
}

// NewExamMainView instantiates the Lorca app using embedded solutions
func NewExamMainView(embeddedFiles map[string][]byte) (domain.MainView, error) {
	fmt.Println("Initializing Lorca UI in EXAM mode...")

	// Convert HTML to Base64 to avoid encoding issues with special characters/newlines
	b64Content := base64.StdEncoding.EncodeToString([]byte(htmlContent))
	url := "data:text/html;base64," + b64Content

	ui, err := lorca.New("", "", 1000, 800, "--remote-allow-origins=*")
	if err != nil {
		fmt.Printf("Lorca New Error: %v\n", err)
		return nil, err
	}
	fmt.Println("Lorca window opened in Exam Mode successfully.")

	// Load the actual content
	ui.Load(url)

	// Optimize window title
	ui.Eval(`document.title = "UML Visual Grader - EXAM Edition"`)

	v := &examLorcaMainView{
		ui:            ui,
		embeddedFiles: embeddedFiles,
	}

	// Wait! We need to dynamically inject the dropdown.
	var options []string
	if len(embeddedFiles) == 0 {
		options = append(options, `<option value="">No Embedded Solutions Found</option>`)
	} else {
		for name := range embeddedFiles {
			displayName := strings.TrimSuffix(name, filepath.Ext(name))
			options = append(options, fmt.Sprintf(`<option value="%s">%s</option>`, name, displayName))
		}
	}

	injectJS := fmt.Sprintf(`
		setTimeout(() => {
			const cardSol = document.getElementById("cardSol");
			if (cardSol) {
				cardSol.onclick = null;
				cardSol.innerHTML = `+"`"+`
					<div class="input-info" style="width: 100%%">
							<span class="input-label">Embedded Solution</span>
							<select id="embeddedSolSelect" style="width: 100%%; margin-top: 5px; border: 1px solid var(--glass-border); border-radius: 4px; padding: 2px; font-family: inherit; font-size: 0.875rem; background-color: inherit; color: var(--text-main); outline: none; cursor: pointer; box-shadow: none;">
								%s
							</select>
					</div>
					<div class="status-dot" style="background: var(--accent-soft);"></div>
				`+"`"+`;
				cardSol.classList.add("selected");
			}
		}, 100);
	`, strings.Join(options, ""))

	v.ui.Eval(injectJS)

	// Bind Go functions to JS
	v.ui.Bind("goChooseStu", v.chooseStu)
	// We bind goSubmit to our own customized v.submit logic
	v.ui.Bind("goSubmit", v.submit)
	v.ui.Bind("goExport", v.export)

	return v, nil
}

func (v *examLorcaMainView) SetController(c domain.MainController) {
	v.controller = c
}

func (v *examLorcaMainView) chooseStu() {
	if v.dialogBusy {
		return
	}
	v.dialogBusy = true
	defer func() { v.dialogBusy = false }()

	v.ui.Eval(`window.focus()`)

	file, err := zenity.SelectFile(
		zenity.Title("Select Assignment (.drawio)"),
		zenity.FileFilters{zenity.FileFilter{Name: "UML Diagrams", Patterns: []string{"*.drawio", "*.xml", "*.mmd", "*.mermaid"}}},
	)
	if err == nil && file != "" {
		v.stuPath = file
		v.ui.Eval(fmt.Sprintf(`setFile("stu", "%s")`, filepath.Base(file)))
	}
}

func (v *examLorcaMainView) submit() {
	if v.stuPath == "" {
		zenity.Error("Please select the Assignment file first.", zenity.Title("Error"))
		return
	}

	if strings.HasSuffix(strings.ToLower(v.stuPath), ".solution") {
		zenity.Error("Student assignment cannot be a .solution file. Please use a .drawio file.", zenity.Title("Invalid File Format"))
		return
	}

	// Get chosen embedded solution
	solFileName := v.ui.Eval(`document.getElementById("embeddedSolSelect") ? document.getElementById("embeddedSolSelect").value : ""`).String()
	if solFileName == "" {
		zenity.Error("No embedded solution selected.", zenity.Title("Error"))
		return
	}

	content, exists := v.embeddedFiles[solFileName]
	if !exists {
		zenity.Error("Selected embedded solution not found.", zenity.Title("Error"))
		return
	}

	// Write embedded content to temp file
	tmpPath := filepath.Join(os.TempDir(), "uml_tmp_embedded_sol_"+solFileName)
	err := os.WriteFile(tmpPath, content, 0644)
	if err != nil {
		zenity.Error("Failed to write temporary solution: "+err.Error(), zenity.Title("Error"))
		return
	}
	v.solPath = tmpPath

	fmt.Printf("Submit triggered (Exam Mode): Sol[Embedded]=%s, Stu=%s\n", solFileName, v.stuPath)
	if v.controller != nil {
		v.controller.OnSubmit(v.solPath, v.stuPath)
	}
}

func (v *examLorcaMainView) export() {
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

func (v *examLorcaMainView) ShowError(err error) {
	v.ui.Eval(`document.getElementById("loading").style.display = "none"`)
	v.ui.Eval(`document.getElementById("btnSubmit").disabled = false`)
	zenity.Error(err.Error(), zenity.Title("Processing Error"))
}

func (v *examLorcaMainView) ShowLoading() {
	v.ui.Eval(`
		document.getElementById("loading").style.display = "flex";
		const res = document.getElementById("results");
		if (res) { res.classList.remove("ready"); }
	`)
}

func (v *examLorcaMainView) ShowResult(result *coreDomain.GradeResult) {
	fmt.Println("Result received, generating HTML...")
	tmpPath := filepath.Join(os.TempDir(), "uml_tmp_student_report.html")
	vis := visualizer.NewHTMLVisualizer()
	err := vis.ExportStudentHTML(result, tmpPath)

	if err != nil {
		fmt.Printf("ExportStudentHTML Error: %v\n", err)
		v.ui.Eval(`document.getElementById("loading").style.display = "none"`)
		zenity.Error("Result Generation Error: "+err.Error(), zenity.Title("Error"))
		return
	}

	b, err := os.ReadFile(tmpPath)
	if err != nil {
		fmt.Printf("ReadFile Error: %v\n", err)
		v.ui.Eval(`document.getElementById("loading").style.display = "none"`)
		zenity.Error("Read Output Error: "+err.Error(), zenity.Title("Error"))
		return
	}

	fmt.Printf("HTML read successfully (%d bytes). Injecting into UI...\n", len(b))

	b64 := base64.StdEncoding.EncodeToString(b)
	v.ui.Eval(fmt.Sprintf("renderReport('%s')", b64))
	fmt.Println("Render script sent.")
}

func (v *examLorcaMainView) EnableExport() {
	v.ui.Eval(`document.getElementById("btnExport").disabled = false`)
}

func (v *examLorcaMainView) Wait() {
	<-v.ui.Done()
}

func (v *examLorcaMainView) Close() {
	// Clean up temporary solution file
	if v.solPath != "" {
		os.Remove(v.solPath)
	}
	v.ui.Close()
}
