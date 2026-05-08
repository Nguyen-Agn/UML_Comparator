package view

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/ncruces/zenity"
	"github.com/zserge/lorca"

	"uml_compare/domain"
	_ "embed"
)

//go:embed mermaid_editor_view.html
var editorHTMLContent string

// editorLorcaView implements a simplified domain.IGeneratorView using Lorca.
type editorLorcaView struct {
	ui         lorca.UI
	controller domain.IGeneratorController
	dialogBusy bool
}

// NewMermaidEditorView creates and initialises the Mermaid Editor Lorca window.
func NewMermaidEditorView() (domain.IGeneratorView, error) {
	b64 := base64.StdEncoding.EncodeToString([]byte(editorHTMLContent))
	url := "data:text/html;base64," + b64

	ui, err := lorca.New("", "", 1000, 700, "--remote-allow-origins=*")
	if err != nil {
		return nil, fmt.Errorf("editor_view: lorca new: %w", err)
	}
	ui.Load(url)

	v := &editorLorcaView{ui: ui}
	v.bindFunctions()
	return v, nil
}

func (v *editorLorcaView) SetGeneratorController(c domain.IGeneratorController) {
	v.controller = c
}

func (v *editorLorcaView) bindFunctions() {
	v.ui.Bind("goSaveMermaid", func(mermaidCode string) string {
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
			if v.controller != nil {
				v.controller.OnSave(mermaidCode, path)
			}
			return `{"success": true}`
		}
		return `{"error": "Không có đường dẫn"}`
	})
}

// Stubs for IGeneratorView interface to reuse GeneratorController safely
func (v *editorLorcaView) ShowGeneratedUML(code string) {}
func (v *editorLorcaView) ShowError(err error) {
	b, _ := json.Marshal(err.Error())
	v.ui.Eval(fmt.Sprintf(`showNotification(%s, true)`, string(b)))
}
func (v *editorLorcaView) ShowSuccess(msg string) {
	b, _ := json.Marshal(msg)
	v.ui.Eval(fmt.Sprintf(`showNotification(%s, false)`, string(b)))
}
func (v *editorLorcaView) ShowLoading() {}
func (v *editorLorcaView) HideLoading() {}
func (v *editorLorcaView) Wait() {
	<-v.ui.Done()
}
func (v *editorLorcaView) Close() {
	v.ui.Close()
}
