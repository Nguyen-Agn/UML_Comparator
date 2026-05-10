package view

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ncruces/zenity"
	"github.com/zserge/lorca"

	_ "embed"
	"uml_compare/domain"
	"uml_compare/src/visualizer"
)

//go:embed instructor_view.html
var instructorHTMLContent string

//go:embed instructor_style.css
var instructorCSSContent string

//go:embed instructor_script.js
var instructorJSContent string

type instructorLorcaView struct {
	ui         lorca.UI
	controller domain.InstructorController
	genCtrl    domain.IGeneratorController
	dialogBusy bool
}

func NewInstructorView() (domain.InstructorView, error) {
	fmt.Println("Initializing Lorca UI for Instructor...")

	html := strings.Replace(instructorHTMLContent, "<!-- INJECT_CSS -->", "<style>"+instructorCSSContent+"</style>", 1)
	html = strings.Replace(html, "<!-- INJECT_JS -->", "<script>"+instructorJSContent+"</script>", 1)

	b64Content := base64.StdEncoding.EncodeToString([]byte(html))
	url := "data:text/html;base64," + b64Content

	ui, err := lorca.New("", "", 1150, 800, "--remote-allow-origins=*")
	if err != nil {
		fmt.Printf("Lorca New Error: %v\n", err)
		return nil, err
	}

	v := &instructorLorcaView{
		ui: ui,
	}

	v.bindFunctions()
	ui.Load(url)

	return v, nil
}

func (v *instructorLorcaView) SetController(c domain.InstructorController) {
	v.controller = c
}

func (v *instructorLorcaView) SetGeneratorController(c domain.IGeneratorController) {
	v.genCtrl = c
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
	v.ui.Bind("goUpdateConfig", func(th float64, ai bool) {
		if v.controller != nil {
			v.controller.OnUpdateConfig(th, ai)
		}
	})

	// ── Generator Bindings ──────────────────────────
	v.ui.Bind("goExecGenerate", func(problemText string) {
		fmt.Println("--- Go: goExecGenerate called with prompt length:", len(problemText))
		if v.genCtrl == nil {
			fmt.Println("--- Go: Error - genCtrl is NIL!")
			return
		}
		v.genCtrl.OnGenerate(problemText)
	})

	v.ui.Bind("goExecSave", func(mermaidCode string) string {
		if v.dialogBusy {
			return `{"error": "Hộp thoại đang mở"}`
		}
		v.dialogBusy = true
		defer func() { v.dialogBusy = false }()

		path, err := zenity.SelectFileSave(
			zenity.Title("Lưu sơ đồ Mermaid"),
			zenity.Filename("diagram.mmd"),
			zenity.FileFilter{Name: "Mermaid Files (*.mmd)", Patterns: []string{"*.mmd"}},
		)
		if err != nil {
			if err == zenity.ErrCanceled {
				return `{"error": "Đã hủy lưu"}`
			}
			return `{"error": "` + err.Error() + `"}`
		}
		if path != "" {
			if v.genCtrl != nil {
				v.genCtrl.OnSave(mermaidCode, path)
			}
			return `{"success": true}`
		}
		return `{"error": "Không có đường dẫn"}`
	})

	v.ui.Bind("goSaveGenConfig", func(endpoint, model, key string) string {
		if v.genCtrl == nil {
			return `{"error": "No controller"}`
		}
		cfg := domain.GeneratorConfig{
			APIEndpoint: endpoint,
			Model:       model,
			APIKey:      key,
		}
		err := v.genCtrl.OnSaveConfig(cfg)
		if err != nil {
			return `{"error": "` + err.Error() + `"}`
		}
		return `{"success": true}`
	})

	v.ui.Bind("goLoadGenConfig", func() string {
		if v.genCtrl == nil {
			return `{}`
		}
		cfg := v.genCtrl.OnGetConfig()
		b, _ := json.Marshal(cfg)
		return string(b)
	})
}

// Helpers
func (v *instructorLorcaView) selectFile(elementID, patternStr string) {
	if v.dialogBusy {
		return
	}
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
	if v.dialogBusy {
		return
	}
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
	msgJSON, _ := json.Marshal(err.Error())
	v.ui.Eval(fmt.Sprintf(`showError(%s)`, string(msgJSON)))
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

func (v *instructorLorcaView) ShowLiveCompareResult(result *domain.CompareResult) {
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

func (v *instructorLorcaView) UpdateConfigUI(threshold float64, useAI bool, aiAvailable bool) {
	v.ui.Eval(fmt.Sprintf("updateConfigUI(%.2f, %v, %v)", threshold, useAI, aiAvailable))
}

func (v *instructorLorcaView) Wait() {
	<-v.ui.Done()
}

func (v *instructorLorcaView) Close() {
	v.ui.Close()
}

// ── Stubs for IGeneratorView ───────────────────────
func (v *instructorLorcaView) ShowGeneratedUML(code string) {
	b, _ := json.Marshal(code)
	v.ui.Eval(fmt.Sprintf(`showGeneratedUML(%s)`, string(b)))
}
