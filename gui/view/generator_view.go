package view

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ncruces/zenity"
	"github.com/zserge/lorca"

	"uml_compare/domain"
	_ "embed"
)

//go:embed generator_view.html
var generatorHTMLContent string

// generatorLorcaView implements domain.IGeneratorView using Lorca.
type generatorLorcaView struct {
	ui         lorca.UI
	controller domain.IGeneratorController
	dialogBusy bool
}

// NewGeneratorView creates and initialises the UML Generator Lorca window.
func NewGeneratorView() (domain.IGeneratorView, error) {
	b64 := base64.StdEncoding.EncodeToString([]byte(generatorHTMLContent))
	url := "data:text/html;base64," + b64

	ui, err := lorca.New("", "", 1200, 800, "--remote-allow-origins=*")
	if err != nil {
		return nil, fmt.Errorf("generator_view: lorca new: %w", err)
	}
	ui.Load(url)

	v := &generatorLorcaView{ui: ui}
	v.bindFunctions()
	return v, nil
}

func (v *generatorLorcaView) SetGeneratorController(c domain.IGeneratorController) {
	v.controller = c
}

// bindFunctions registers Go callbacks accessible from JavaScript.
func (v *generatorLorcaView) bindFunctions() {
	// Trigger AI generation
	v.ui.Bind("goExecGenerate", func(problemText string) {
		if v.controller == nil {
			return
		}
		v.controller.OnGenerate(problemText)
	})

	// Save to file — opens a Zenity save dialog then calls OnSave
	v.ui.Bind("goExecSave", func(mermaidCode string) {
		if v.dialogBusy {
			return
		}
		v.dialogBusy = true
		defer func() { v.dialogBusy = false }()

		v.ui.Eval(`window.focus()`)
		path, err := zenity.SelectFileSave(
			zenity.Title("Lưu file UML"),
			zenity.FileFilter{Name: "Mermaid UML", Patterns: []string{"*.mmd"}},
			zenity.ConfirmOverwrite(),
		)
		if err != nil || path == "" {
			return // user cancelled
		}
		if !strings.HasSuffix(path, ".mmd") {
			path += ".mmd"
		}
		if v.controller != nil {
			v.controller.OnSave(mermaidCode, path)
		}
	})

	// Return current config as a JS object
	v.ui.Bind("goGetConfig", func() domain.GeneratorConfig {
		if v.controller == nil {
			return domain.GeneratorConfig{}
		}
		return v.controller.OnGetConfig()
	})

	// Save config from settings modal; returns error string or ""
	v.ui.Bind("goSaveConfig", func(cfg domain.GeneratorConfig) string {
		if v.controller == nil {
			return "controller not initialised"
		}
		if err := v.controller.OnSaveConfig(cfg); err != nil {
			return err.Error()
		}
		return ""
	})
}

// ── IGeneratorView implementation ─────────────────────────────────────────────

func (v *generatorLorcaView) ShowLoading() {
	v.ui.Eval(`document.getElementById("loading").classList.add("active")`)
}

func (v *generatorLorcaView) HideLoading() {
	v.ui.Eval(`document.getElementById("loading").classList.remove("active")`)
}

func (v *generatorLorcaView) ShowError(err error) {
	v.HideLoading()
	v.ui.Eval(`enableGenerate()`)
	// Use JSON encoding to safely embed the string in JS (handles all special chars)
	msgJSON, _ := json.Marshal(err.Error())
	v.ui.Eval(fmt.Sprintf(`showError(%s)`, string(msgJSON)))
}

func (v *generatorLorcaView) ShowSuccess(msg string) {
	msgJSON, _ := json.Marshal(msg)
	v.ui.Eval(fmt.Sprintf(`showNotification(%s, 'success')`, string(msgJSON)))
}

func (v *generatorLorcaView) ShowGeneratedUML(mermaidCode string) {
	v.ui.Eval(`enableGenerate()`)
	// JSON-encode the mermaid code to safely pass multiline string with any special chars
	codeJSON, _ := json.Marshal(mermaidCode)
	v.ui.Eval(fmt.Sprintf(`showGeneratedUML(%s)`, string(codeJSON)))
}

func (v *generatorLorcaView) Wait() {
	<-v.ui.Done()
}

func (v *generatorLorcaView) Close() {
	v.ui.Close()
}
