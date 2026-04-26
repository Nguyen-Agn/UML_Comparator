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

	"uml_compare/src/visualizer"
)

// htmlContent is the gorgeous frontend embedded directly
const htmlContent = `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<title>UML Visual Grader - Student Edition</title>
	<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;600;700&display=swap" rel="stylesheet">
	<style>
		:root {
			--bg-base: #f8fafc;      /* Morning Sky */
			--bg-card: #ffffff;      
			--accent: #720f32;       /* Berry Wine */
			--accent-hover: #9c1545;
			--accent-soft: #114665;  /* Ocean Noir */
			--text-main: #16202b;    /* Indigo Night */
			--text-dim: #7b445a;     /* Muted Mauve */
			--glass-border: rgba(22, 32, 43, 0.08);
			--card-radius: 16px;
			--shadow-soft: 0 4px 20px rgba(0, 0, 0, 0.05);
			--shadow-medium: 0 10px 30px rgba(0, 0, 0, 0.08);
			--shadow-glow: 0 0 15px rgba(114, 15, 50, 0.15);
		}

		* { box-sizing: border-box; }
		body {
			font-family: 'Inter', -apple-system, sans-serif;
			background: var(--bg-base);
			color: var(--text-main);
			margin: 0;
			overflow: hidden;
			display: flex;
			flex-direction: column;
			height: 100vh;
		}

		header {
			padding: 28px 40px;
			background: rgba(255, 255, 255, 0.85);
			backdrop-filter: blur(25px);
			border-bottom: 1px solid var(--glass-border);
			box-shadow: var(--shadow-soft);
			z-index: 100;
		}

		.brand {
			font-size: 1.5rem;
			font-weight: 700;
			letter-spacing: -0.025em;
			margin-bottom: 20px;
			display: flex;
			align-items: center;
			gap: 10px;
		}
		.brand span { color: var(--accent); }

		.controls {
			display: grid;
			grid-template-columns: 1fr 1fr auto auto;
			gap: 16px;
			align-items: center;
		}

		.input-card {
			background: var(--bg-card);
			border: 1px solid var(--glass-border);
			padding: 12px 20px;
			border-radius: var(--card-radius);
			display: flex;
			align-items: center;
			justify-content: space-between;
			transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
			cursor: pointer;
			box-shadow: 0 2px 8px rgba(0,0,0,0.03);
		}
		.input-card:hover { 
			border-color: var(--accent-soft); 
			transform: translateY(-2px);
			box-shadow: var(--shadow-medium);
		}
		.input-card.selected { 
			border-color: var(--accent-soft); 
			background: rgba(17, 70, 101, 0.05);
			box-shadow: 0 0 15px rgba(17, 70, 101, 0.15);
		}
		
		.input-info { display: flex; flex-direction: column; gap: 2px; }
		.input-label { font-size: 0.75rem; font-weight: 600; text-transform: uppercase; color: var(--text-dim); }
		.file-name { font-size: 0.875rem; color: var(--text-main); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 180px; }

		button {
			font-family: inherit;
			font-weight: 600;
			font-size: 0.875rem;
			padding: 12px 24px;
			border-radius: var(--card-radius);
			border: none;
			cursor: pointer;
			transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
			display: flex;
			align-items: center;
			gap: 8px;
		}

		.btn-primary { 
			background: linear-gradient(135deg, var(--accent) 0%, #4d0a22 100%); 
			color: white; 
			box-shadow: 0 4px 15px rgba(114, 15, 50, 0.25);
			text-transform: uppercase;
			letter-spacing: 0.05em;
		}
		.btn-primary:hover { 
			background: linear-gradient(135deg, var(--accent-hover) 0%, var(--accent) 100%);
			transform: translateY(-2px) scale(1.02); 
			box-shadow: 0 10px 20px rgba(114, 15, 50, 0.35); 
		}
		.btn-primary:active { transform: translateY(0) scale(1); }
		.btn-primary:disabled { background: #cbd5e1; transform: none; box-shadow: none; cursor:not-allowed; opacity: 0.7; }

		.btn-success { 
			background: var(--bg-card); 
			color: var(--accent-soft);
			border: 1px solid var(--accent-soft);
		}
		.btn-success:hover { 
			background: var(--accent-soft);
			color: white;
			transform: translateY(-2px);
			box-shadow: 0 8px 20px rgba(17, 70, 101, 0.2);
		}

		main {
			flex: 1;
			position: relative;
			background: var(--bg-base);
			overflow: hidden;
		}

		#results {
			width: 100%;
			height: 100%;
			border: none;
			opacity: 0;
			transition: opacity 0.8s cubic-bezier(0.4, 0, 0.2, 1);
			background: var(--bg-base);
		}
		#results.ready { opacity: 1; }

		.status-dot {
			width: 10px;
			height: 10px;
			border-radius: 50%;
			background: #e2e8f0;
			border: 2px solid rgba(0, 0, 0, 0.05);
			transition: all 0.3s ease;
		}
		.selected .status-dot { 
			background: var(--accent-soft); 
			box-shadow: 0 0 12px var(--accent-soft);
			border-color: rgba(17, 70, 101, 0.2);
		}

		/* Spinner */
		.overlay {
			position: absolute; inset: 0; 
			background: rgba(248, 250, 252, 0.9);
			display: none; flex-direction: column; justify-content: center; align-items: center;
			z-index: 200; gap: 20px;
			animation: fadeIn 0.3s ease;
		}
		.spinner {
			width: 40px; height: 40px;
			border: 4px solid rgba(114, 15, 50, 0.1);
			border-top-color: var(--accent);
			border-radius: 50%;
			animation: spin 1s linear infinite;
		}
		@keyframes spin { to { transform: rotate(360deg); } }
		@keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
	</style>
</head>
<body>
	<div class="overlay" id="loading">
		<div class="spinner"></div>
		<p style="font-weight: 700; color: var(--accent); letter-spacing: 0.05em;">ANALYSTING UML...</p>
	</div>

	<header>
		<div class="brand">
			<span>UML</span> Visual Grader
		</div>
		<div class="controls">
			<div class="input-card" id="cardSol" onclick="goChooseSol()">
				<div class="input-info">
					<span class="input-label">Solution</span>
					<span class="file-name" id="solLabel">Select .drawio file</span>
				</div>
				<div class="status-dot"></div>
			</div>
			
			<div class="input-card" id="cardStu" onclick="goChooseStu()">
				<div class="input-info">
					<span class="input-label">Assignment</span>
					<span class="file-name" id="stuLabel">Select .drawio file</span>
				</div>
				<div class="status-dot"></div>
			</div>

			<button class="btn-primary" id="btnSubmit" onclick="goSubmit()">
				RUN ANALYSIS
			</button>
			
			<button class="btn-success" id="btnExport" disabled onclick="goExport()">
				SAVE HTML
			</button>
		</div>
	</header>

	<main>
		<iframe id="results"></iframe>
	</main>

	<script>
		function setFile(type, name) {
			const label = document.getElementById(type + "Label");
			const card = document.getElementById("card" + (type.charAt(0).toUpperCase() + type.slice(1)));
			label.innerText = name;
			card.classList.add("selected");
		}

		function renderReport(b64) {
			const bin = atob(b64);
			const blob = new Blob([bin], { type: 'text/html' });
			const url = URL.createObjectURL(blob);
			const iframe = document.getElementById("results");
			
			// Remove previous ready class to allow new transition
			iframe.classList.remove("ready");
			
			iframe.onload = () => {
				iframe.classList.add("ready");
				URL.revokeObjectURL(url); 
			};
			
			iframe.src = url;
			document.getElementById("loading").style.display = "none";
		}
	</script>
</body>
</html>`

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
		zenity.FileFilters{zenity.FileFilter{Name: "UML Diagrams", Patterns: []string{"*.drawio", "*.solution", "*.mmd", "*.mermaid", "*.drawio.xml"}}},
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

// Window logic
func (v *lorcaMainView) Wait() {
	<-v.ui.Done()
}

func (v *lorcaMainView) Close() {
	v.ui.Close()
}
