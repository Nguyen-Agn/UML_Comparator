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
	_ "embed"
)

//go:embed instructor_view.html
var instructorHTMLContent string

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
	v.ui.Bind("goUpdateConfig", func(th float64, ai bool) {
		if v.controller != nil {
			v.controller.OnUpdateConfig(th, ai)
		}
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
