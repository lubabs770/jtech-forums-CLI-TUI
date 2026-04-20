package ui

import (
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

func openEditor(initial string) tea.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nano"
	}
	f, err := os.CreateTemp("", "jtech-*.md")
	if err != nil {
		return func() tea.Msg { return editorFinishedMsg{err: err} }
	}
	if initial != "" {
		f.WriteString(initial)
	}
	f.Close()

	c := exec.Command(editor, f.Name())
	return tea.ExecProcess(c, func(err error) tea.Msg {
		defer os.Remove(f.Name())
		if err != nil {
			return editorFinishedMsg{err: err}
		}
		content, readErr := os.ReadFile(f.Name())
		if readErr != nil {
			return editorFinishedMsg{err: readErr}
		}
		return editorFinishedMsg{content: string(content)}
	})
}
