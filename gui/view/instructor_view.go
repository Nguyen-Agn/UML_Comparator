package view

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ncruces/zenity"
	"github.com/zserge/lorca"

	"uml_compare/gui/domain"
	"uml_compare/instructor"
	"uml_compare/visualizer"
)

const instructorHTMLContent = `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<title>UML Visual Grader - Instructor Suite</title>
	<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;600;700;800&display=swap" rel="stylesheet">
	<style>
		:root {
			--bg-base: #f8fafc;
			--bg-card: #ffffff;
			--accent: #114665;
			--accent-hover: #08293d;
			--accent-secondary: #720f32;
			--text-main: #16202b;
			--text-dim: #64748b;
			--border: rgba(22, 32, 43, 0.08);
			--border-hover: rgba(17, 70, 101, 0.3);
			--success: #10b981;
			--error: #720f32;
			--card-radius: 16px;
			--shadow-soft: 0 4px 20px rgba(0, 0, 0, 0.03);
			--shadow-medium: 0 10px 30px rgba(0, 0, 0, 0.06);
		}

		* { box-sizing: border-box; }
		body {
			font-family: 'Inter', -apple-system, sans-serif;
			background: var(--bg-base);
			color: var(--text-main);
			margin: 0;
			overflow: hidden;
			display: flex;
			height: 100vh;
		}

		/* Sidebar / Tabs */
		aside {
			width: 260px;
			background: var(--bg-card);
			border-right: 1px solid var(--border);
			display: flex;
			flex-direction: column;
			z-index: 100;
			box-shadow: 2px 0 10px rgba(0,0,0,0.02);
		}

		.brand {
			padding: 30px 20px;
			font-size: 1.4rem;
			font-weight: 800;
			letter-spacing: -0.02em;
			border-bottom: 1px solid var(--border);
			color: var(--text-main);
			position: relative;
		}
		.brand::before {
			content: '';
			position: absolute;
			top: 0; left: 0; right: 0;
			height: 4px;
			background: linear-gradient(90deg, var(--accent-secondary), var(--accent));
		}
		.brand span { color: var(--accent); }

		.tab-nav {
			display: flex;
			flex-direction: column;
			gap: 8px;
			padding: 20px 10px;
		}

		.tab-btn {
			background: transparent;
			border: none;
			color: var(--text-dim);
			text-align: left;
			padding: 14px 20px;
			font-size: 0.95rem;
			font-weight: 600;
			font-family: inherit;
			border-radius: 10px;
			cursor: pointer;
			transition: all 0.2s ease;
		}
		.tab-btn:hover {
			background: #f1f5f9;
			color: var(--accent);
		}
		.tab-btn.active {
			background: rgba(17, 70, 101, 0.06);
			color: var(--accent);
			box-shadow: inset 3px 0 0 var(--accent);
		}

		/* Main Content Area */
		main {
			flex: 1;
			position: relative;
			display: flex;
			flex-direction: column;
			background: var(--bg-base);
		}

		.tab-content {
			display: none;
			padding: 40px;
			flex-direction: column;
			gap: 20px;
			height: 100%;
			overflow-y: auto;
		}
		.tab-content.active { display: flex; animation: fadeIn 0.3s ease; }

		.header-title {
			font-size: 2rem;
			font-weight: 800;
			margin-bottom: 5px;
			color: var(--accent);
			letter-spacing: -0.02em;
		}
		.header-sub {
			color: var(--text-dim);
			margin-bottom: 20px;
		}

		.card {
			background: var(--bg-card);
			border: 1px solid var(--border);
			border-radius: var(--card-radius);
			padding: 24px;
			box-shadow: var(--shadow-soft);
		}

		.field-group {
			display: flex;
			flex-direction: column;
			gap: 8px;
			margin-bottom: 20px;
		}
		.field-group label {
			font-size: 0.8rem;
			font-weight: 700;
			text-transform: uppercase;
			color: var(--accent-secondary);
			letter-spacing: 0.05em;
		}

		.input-row {
			display: flex;
			gap: 10px;
		}
		.input-text {
			flex: 1;
			background: #f8fafc;
			border: 1px solid var(--border);
			color: var(--text-main);
			padding: 12px 16px;
			border-radius: 8px;
			font-family: inherit;
			font-size: 0.95rem;
			outline: none;
			transition: border-color 0.2s;
		}
		.input-text:focus { border-color: var(--accent); background: #ffffff; }

		.btn {
			font-family: inherit;
			font-weight: 600;
			font-size: 0.9rem;
			padding: 12px 24px;
			border-radius: 8px;
			border: none;
			cursor: pointer;
			transition: all 0.2s ease;
		}
		.btn-outline {
			background: #ffffff;
			border: 1px solid var(--border);
			color: var(--text-main);
		}
		.btn-outline:hover { background: #f1f5f9; border-color: var(--border-hover); color: var(--accent); }
		
		.btn-primary { 
			background: var(--accent);
			color: white; 
			box-shadow: 0 4px 10px rgba(17, 70, 101, 0.15);
		}
		.btn-primary:hover { 
			background: var(--accent-hover);
			transform: translateY(-1px); 
			box-shadow: 0 6px 15px rgba(17, 70, 101, 0.2);
		}

		iframe {
			width: 100%;
			height: 100%;
			border: 1px solid var(--border);
			border-radius: var(--card-radius);
			background: #fff;
			box-shadow: var(--shadow-soft);
		}

		/* Notifications */
		#notification {
			position: fixed;
			top: 20px; right: 20px;
			min-width: 300px;
			padding: 16px 20px;
			border-radius: 10px;
			background: white;
			box-shadow: var(--shadow-medium);
			display: flex;
			align-items: center;
			gap: 12px;
			transform: translateX(120%);
			transition: transform 0.4s cubic-bezier(0.4, 0, 0.2, 1);
			z-index: 1000;
			border-left: 5px solid gray;
		}
		#notification.show { transform: translateX(0); }
		#notification.success { border-left-color: var(--success); }
		#notification.error { border-left-color: var(--error); }
		#notif-msg { font-weight: 600; font-size: 0.9rem; color: var(--text-main); }

		/* Spinner */
		.overlay {
			position: absolute; inset: 0; 
			background: rgba(248, 250, 252, 0.85);
			backdrop-filter: blur(4px);
			display: none; flex-direction: column; justify-content: center; align-items: center;
			z-index: 200; gap: 15px;
		}
		.spinner {
			width: 40px; height: 40px;
			border: 4px solid rgba(17, 70, 101, 0.1);
			border-top-color: var(--accent);
			border-radius: 50%;
			animation: spin 1s linear infinite;
		}
		@keyframes spin { to { transform: rotate(360deg); } }
		@keyframes fadeIn { from { opacity: 0; transform: translateY(5px); } to { opacity: 1; transform: translateY(0); } }
	</style>
</head>
<body>
	<div id="notification">
		<span id="notif-msg">Notification message</span>
	</div>

	<div class="overlay" id="loading">
		<div class="spinner"></div>
		<p style="font-weight: 700; color: var(--accent); letter-spacing: 1px;">PROCESSING...</p>
	</div>

	<aside>
		<div class="brand">
			<span>UML</span> Instructor
		</div>
		<div class="tab-nav">
			<button class="tab-btn active" onclick="switchTab(event, 'tab-live')">Live Compare</button>
			<button class="tab-btn" onclick="switchTab(event, 'tab-batch')">Batch Grader</button>
			<button class="tab-btn" onclick="switchTab(event, 'tab-security')">Solution Encrypt</button>
			<button class="tab-btn" onclick="switchTab(event, 'tab-exam')">Exam Builder</button>
		</div>
	</aside>

	<main>
		<!-- 1. LIVE COMPARE -->
		<div id="tab-live" class="tab-content active">
			<div class="header-title">Live Comparison</div>
			<div class="header-sub">Run detailed diagram comparison with admin features</div>
			<div class="card" style="margin-bottom: 0;">
				<div class="field-group">
					<label>Solution File (.drawio)</label>
					<div class="input-row">
						<input type="text" id="live-sol" class="input-text" readonly placeholder="Select a solution file...">
						<button class="btn btn-outline" onclick="goSelectFile('live-sol', '*.drawio;*.solution')">Browse</button>
					</div>
				</div>
				<div class="field-group">
					<label>Student File (.drawio)</label>
					<div class="input-row">
						<input type="text" id="live-stu" class="input-text" readonly placeholder="Select a student file...">
						<button class="btn btn-outline" onclick="goSelectFile('live-stu', '*.drawio')">Browse</button>
					</div>
				</div>
				<div style="margin-top: 10px;">
					<button class="btn btn-primary" onclick="goRunLive()">Run Comparison</button>
				</div>
			</div>
			<iframe id="live-iframe" style="display: none; flex: 1; margin-top: 20px;"></iframe>
		</div>

		<!-- 2. BATCH GRADER -->
		<div id="tab-batch" class="tab-content">
			<div class="header-title">Batch Grader</div>
			<div class="header-sub">Grade multiple submissions automatically to CSV</div>
			<div class="card">
				<div class="field-group">
					<label>Solution File (.drawio)</label>
					<div class="input-row">
						<input type="text" id="batch-sol" class="input-text" readonly placeholder="Select a solution file...">
						<button class="btn btn-outline" onclick="goSelectFile('batch-sol', '*.drawio;*.solution')">Browse</button>
					</div>
				</div>
				<div class="field-group">
					<label>Student Directory</label>
					<div class="input-row">
						<input type="text" id="batch-dir" class="input-text" readonly placeholder="Select directory with submissions...">
						<button class="btn btn-outline" onclick="goSelectDir('batch-dir')">Browse</button>
					</div>
				</div>
				<div class="field-group">
					<label>Output Directory</label>
					<div class="input-row">
						<input type="text" id="batch-out" class="input-text" readonly placeholder="Select destination folder...">
						<button class="btn btn-outline" onclick="goSelectDir('batch-out')">Browse</button>
					</div>
					<small style="color:var(--text-dim); margin-top:4px;">Result will be saved as batch_result.csv</small>
				</div>
				<button class="btn btn-primary" onclick="goRunBatch()">Run Batch Grading</button>
			</div>
		</div>

		<!-- 3. SOLUTION ENCRYPT -->
		<div id="tab-security" class="tab-content">
			<div class="header-title">Solution Encryptor</div>
			<div class="header-sub">Securely encrypt solution files for offline distribution</div>
			<div class="card">
				<div class="field-group">
					<label>Input File (.drawio)</label>
					<div class="input-row">
						<input type="text" id="sec-in" class="input-text" readonly placeholder="Select a diagram to encrypt...">
						<button class="btn btn-outline" onclick="goSelectFile('sec-in', '*.drawio')">Browse</button>
					</div>
				</div>
				<div class="field-group">
					<label>Output Directory</label>
					<div class="input-row">
						<input type="text" id="sec-out" class="input-text" readonly placeholder="Select destination folder...">
						<button class="btn btn-outline" onclick="goSelectDir('sec-out')">Browse</button>
					</div>
					<small style="color:var(--text-dim); margin-top:4px;">Result will be named automatically based on input file (.solution)</small>
				</div>
				<button class="btn btn-primary" onclick="goRunEncrypt()">Encrypt File</button>
			</div>
		</div>

		<!-- 4. EXAM BUILDER -->
		<div id="tab-exam" class="tab-content">
			<div class="header-title">Exam Builder</div>
			<div class="header-sub">Compile a custom executable embedding specific solutions</div>
			<div class="card">
				<div class="field-group">
					<label>Solutions Directory</label>
					<div class="input-row">
						<input type="text" id="exam-dir" class="input-text" readonly placeholder="Select directory with solutions...">
						<button class="btn btn-outline" onclick="goSelectDir('exam-dir')">Browse</button>
					</div>
				</div>
				<div class="field-group">
					<label>Output Directory</label>
					<div class="input-row">
						<input type="text" id="exam-out" class="input-text" readonly placeholder="Select destination folder...">
						<button class="btn btn-outline" onclick="goSelectDir('exam-out')">Browse</button>
					</div>
					<small style="color:var(--text-dim); margin-top:4px;">Result will be saved as exam_student_uml.exe</small>
				</div>
				<button class="btn btn-primary" onclick="goRunExamBuild()">Build Application</button>
			</div>
		</div>
	</main>

	<script>
		let notifTimeout;
		function showNotification(msg, type) {
			const notif = document.getElementById('notification');
			const msgEl = document.getElementById('notif-msg');
			
			clearTimeout(notifTimeout);
			notif.className = ''; 
			notif.classList.add(type);
			notif.classList.add('show');
			msgEl.innerText = msg;

			notifTimeout = setTimeout(() => {
				notif.classList.remove('show');
			}, 4000);
		}

		function switchTab(e, tabId) {
			document.querySelectorAll('.tab-content').forEach(t => t.classList.remove('active'));
			document.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
			
			document.getElementById(tabId).classList.add('active');
			if (e && e.currentTarget) {
				e.currentTarget.classList.add('active');
			}
		}

		function setValue(id, val) { document.getElementById(id).value = val; }
		function getValue(id) { return document.getElementById(id).value; }

		function renderLiveResult(b64) {
			const bin = atob(b64);
			const blob = new Blob([bin], { type: 'text/html' });
			const url = URL.createObjectURL(blob);
			const iframe = document.getElementById("live-iframe");
			
			iframe.style.display = "block";
			iframe.onload = () => { URL.revokeObjectURL(url); };
			iframe.src = url;
			document.getElementById("loading").style.display = "none";
		}
		
		// Event forwarders
		function goRunLive() { goExecLive(getValue('live-sol'), getValue('live-stu')); }
		function goRunBatch() { goExecBatch(getValue('batch-sol'), getValue('batch-dir'), getValue('batch-out')); }
		function goRunEncrypt() { goExecEncrypt(getValue('sec-in'), getValue('sec-out')); }
		function goRunExamBuild() { goExecExam(getValue('exam-dir'), getValue('exam-out')); }
	</script>
</body>
</html>`

type instructorLorcaView struct {
	ui         lorca.UI
	controller domain.InstructorController
	dialogBusy bool
}

func NewInstructorView() (domain.InstructorView, error) {
	fmt.Println("Initializing Lorca UI for Instructor...")

	b64Content := base64.StdEncoding.EncodeToString([]byte(instructorHTMLContent))
	url := "data:text/html;base64," + b64Content

	ui, err := lorca.New("", "", 1150, 800, "--remote-allow-origins=*")
	if err != nil {
		fmt.Printf("Lorca New Error: %v\n", err)
		return nil, err
	}

	ui.Load(url)

	v := &instructorLorcaView{
		ui: ui,
	}

	v.bindFunctions()
	return v, nil
}

func (v *instructorLorcaView) SetController(c domain.InstructorController) {
	v.controller = c
}

func (v *instructorLorcaView) bindFunctions() {
	v.ui.Bind("goSelectFile", v.selectFile)
	v.ui.Bind("goSelectDir", v.selectDir)

	v.ui.Bind("goExecLive", func(sol, stu string) {
		if sol != "" && stu != "" && v.controller != nil {
			v.controller.OnLiveCompare(sol, stu)
		} else {
			v.ShowError(fmt.Errorf("Missing input files"))
		}
	})
	
	v.ui.Bind("goExecBatch", func(sol, dir, outFolder string) {
		if sol != "" && dir != "" && outFolder != "" && v.controller != nil {
			outPath := filepath.Join(outFolder, "batch_result.csv")
			v.controller.OnGradeBatch(sol, dir, outPath)
		} else {
			v.ShowError(fmt.Errorf("Missing input files or output path"))
		}
	})

	v.ui.Bind("goExecEncrypt", func(inPath, outFolder string) {
		if inPath != "" && outFolder != "" && v.controller != nil {
			baseName := strings.TrimSuffix(filepath.Base(inPath), filepath.Ext(inPath))
			outPath := filepath.Join(outFolder, baseName+".solution")
			v.controller.OnEncrypt(inPath, outPath)
		} else {
			v.ShowError(fmt.Errorf("Input and Output are required"))
		}
	})
	
	v.ui.Bind("goExecExam", func(dir, outFolder string) {
		if dir != "" && outFolder != "" && v.controller != nil {
			outPath := filepath.Join(outFolder, "exam_student_uml.exe")
			v.controller.OnBuildExam(dir, outPath)
		} else {
			v.ShowError(fmt.Errorf("Directory and Output are required"))
		}
	})
}

// Helpers
func (v *instructorLorcaView) selectFile(elementID, patternStr string) {
	if v.dialogBusy { return }
	v.dialogBusy = true
	defer func() { v.dialogBusy = false }()

	patterns := strings.Split(patternStr, ";")
	v.ui.Eval(`window.focus()`)
	file, err := zenity.SelectFile(
		zenity.Title("Select File"),
		zenity.FileFilters{zenity.FileFilter{Name: "UML Diagram", Patterns: patterns}},
	)
	if err == nil && file != "" {
		v.ui.Eval(fmt.Sprintf(`setValue("%s", "%s")`, elementID, sanitizeStr(file)))
	}
}

func (v *instructorLorcaView) selectDir(elementID string) {
	if v.dialogBusy { return }
	v.dialogBusy = true
	defer func() { v.dialogBusy = false }()

	v.ui.Eval(`window.focus()`)
	dir, err := zenity.SelectFile(
		zenity.Title("Select Directory"),
		zenity.Directory(),
	)
	if err == nil && dir != "" {
		v.ui.Eval(fmt.Sprintf(`setValue("%s", "%s")`, elementID, sanitizeStr(dir)))
	}
}

func sanitizeStr(s string) string {
	s = filepath.ToSlash(s)
	// Escape backslashes for JS strings if any remain
	s = strings.ReplaceAll(s, "\\", "\\\\")
	return s
}


func (v *instructorLorcaView) ShowError(err error) {
	v.ui.Eval(`document.getElementById("loading").style.display = "none"`)
	escapedMsg := strings.ReplaceAll(err.Error(), "'", "\\'")
	escapedMsg = strings.ReplaceAll(escapedMsg, "\n", " ")
	v.ui.Eval(fmt.Sprintf(`showNotification('%s', 'error')`, escapedMsg))
}

func (v *instructorLorcaView) ShowSuccess(msg string) {
	v.ui.Eval(`document.getElementById("loading").style.display = "none"`)
	escapedMsg := strings.ReplaceAll(msg, "'", "\\'")
	escapedMsg = strings.ReplaceAll(escapedMsg, "\n", " ")
	v.ui.Eval(fmt.Sprintf(`showNotification('%s', 'success')`, escapedMsg))
}

func (v *instructorLorcaView) ShowLoading() {
	v.ui.Eval(`document.getElementById("loading").style.display = "flex";`)
}

func (v *instructorLorcaView) HideLoading() {
	v.ui.Eval(`document.getElementById("loading").style.display = "none";`)
}

func (v *instructorLorcaView) ShowLiveCompareResult(result *instructor.CompareResult) {
	tmpPath := filepath.Join(os.TempDir(), "uml_tmp_admin_report.html")
	vis := visualizer.NewHTMLVisualizer()
	
	err := vis.ExportHTML(result.GradeResult, tmpPath)
	if err != nil {
		v.ShowError(fmt.Errorf("Render Error: %w", err))
		return
	}

	b, err := os.ReadFile(tmpPath)
	if err != nil {
		v.ShowError(fmt.Errorf("Read output error: %w", err))
		return
	}

	b64 := base64.StdEncoding.EncodeToString(b)
	v.ui.Eval(fmt.Sprintf("renderLiveResult('%s')", b64))
}

func (v *instructorLorcaView) Wait() {
	<-v.ui.Done()
}

func (v *instructorLorcaView) Close() {
	v.ui.Close()
}
