package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sam/jtech-tui/internal/api"
)

type newTopicForm struct {
	title      textinput.Model
	categories []api.Category
	catIndex   int
	focused    int // 0=title, 1=category
}

func newNewTopicForm(cats []api.Category) *newTopicForm {
	t := textinput.New()
	t.Placeholder = "Topic title"
	t.Focus()
	return &newTopicForm{title: t, categories: cats}
}

type newTopicFormDoneMsg struct {
	title      string
	categoryID int
}

func (f *newTopicForm) Update(msg tea.Msg) (*newTopicForm, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab":
			// Two fields — either direction just toggles.
			f.focused = (f.focused + 1) % 2
			if f.focused == 0 {
				f.title.Focus()
			} else {
				f.title.Blur()
			}
		case "up", "k":
			if f.focused == 1 && f.catIndex > 0 {
				f.catIndex--
			}
		case "down", "j":
			if f.focused == 1 && f.catIndex < len(f.categories)-1 {
				f.catIndex++
			}
		case "enter":
			if f.focused == 0 && f.title.Value() != "" {
				f.focused = 1
				f.title.Blur()
			} else if f.focused == 1 && f.title.Value() != "" && len(f.categories) > 0 {
				catID := f.categories[f.catIndex].ID
				title := f.title.Value()
				return f, func() tea.Msg {
					return newTopicFormDoneMsg{title: title, categoryID: catID}
				}
			}
		}
	}
	var cmd tea.Cmd
	f.title, cmd = f.title.Update(msg)
	return f, cmd
}

func (f *newTopicForm) View() string {
	var sb strings.Builder
	sb.WriteString(titleStyle.Render("New Topic") + "\n\n")
	sb.WriteString(f.title.View() + "\n\n")
	sb.WriteString("Category:\n")
	for i, c := range f.categories {
		if i == f.catIndex && f.focused == 1 {
			sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("69")).Render("▶ "+c.Name) + "\n")
		} else {
			sb.WriteString(fmt.Sprintf("  %s\n", c.Name))
		}
	}
	sb.WriteString("\n" + helpStyle.Render("tab switch • j/k navigate category • enter confirm • esc cancel"))
	return overlayStyle.Render(sb.String())
}
