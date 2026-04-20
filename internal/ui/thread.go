package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/sam/jtech-tui/internal/api"
)

type threadLoadedMsg struct {
	thread *api.Thread
	err    error
}

type threadView struct {
	client   *api.Client
	topic    api.Topic
	thread   *api.Thread
	viewport viewport.Model
	spinner  spinner.Model
	loading  bool
	err      string
	width    int
	height   int
}

func newThreadView(client *api.Client, topic api.Topic) *threadView {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	return &threadView{client: client, topic: topic, spinner: sp}
}

func (v *threadView) Init() tea.Cmd {
	v.loading = true
	client, id := v.client, v.topic.ID
	return tea.Batch(v.spinner.Tick, func() tea.Msg {
		thread, err := client.GetThread(id)
		return threadLoadedMsg{thread: thread, err: err}
	})
}

func renderPosts(posts []api.Post, width int) string {
	wrapWidth := width - 4
	if wrapWidth < 20 {
		wrapWidth = 80
	}
	r, _ := glamour.NewTermRenderer(glamour.WithAutoStyle(), glamour.WithWordWrap(wrapWidth))
	var sb strings.Builder
	for _, p := range posts {
		sb.WriteString(usernameStyle.Render(p.Username))
		sb.WriteString("  ")
		sb.WriteString(sepStyle.Render(formatTime(p.CreatedAt)))
		sb.WriteString("\n")
		rendered, err := r.Render(p.Raw)
		if err != nil || strings.TrimSpace(rendered) == "" {
			sb.WriteString(stripHTML(p.Cooked))
		} else {
			sb.WriteString(rendered)
		}
		if width > 2 {
			sb.WriteString(sepStyle.Render(strings.Repeat("─", width-2)))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func stripHTML(s string) string {
	var result strings.Builder
	inTag := false
	for _, r := range s {
		if r == '<' {
			inTag = true
		} else if r == '>' {
			inTag = false
		} else if !inTag {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func (v *threadView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width, v.height = msg.Width, msg.Height
		v.viewport.Width = msg.Width
		v.viewport.Height = msg.Height - 3
		if v.thread != nil {
			v.viewport.SetContent(renderPosts(v.thread.PostStream.Posts, msg.Width))
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "h":
			return v, func() tea.Msg { return popViewMsg{} }
		case "r":
			if v.thread != nil {
				return v, openEditor("")
			}
		}

	case threadLoadedMsg:
		v.loading = false
		if msg.err != nil {
			if isUnauthorized(msg.err) {
				return v, func() tea.Msg { return unauthorizedMsg{} }
			}
			v.err = msg.err.Error()
			return v, nil
		}
		v.thread = msg.thread
		v.viewport = viewport.New(v.width, v.height-3)
		v.viewport.SetContent(renderPosts(msg.thread.PostStream.Posts, v.width))

	case editorFinishedMsg:
		if msg.err != nil || strings.TrimSpace(msg.content) == "" {
			return v, nil
		}
		topicID := v.topic.ID
		client := v.client
		raw := strings.TrimSpace(msg.content)
		return v, func() tea.Msg {
			err := client.PostReply(topicID, raw)
			if err != nil {
				if isUnauthorized(err) {
					return unauthorizedMsg{}
				}
				return replyErrMsg{err: err}
			}
			thread, err := client.GetThread(topicID)
			return threadLoadedMsg{thread: thread, err: err}
		}

	case replyErrMsg:
		v.err = msg.err.Error()

	case spinner.TickMsg:
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		return v, cmd
	}

	var cmd tea.Cmd
	v.viewport, cmd = v.viewport.Update(msg)
	return v, cmd
}

func (v *threadView) View() string {
	header := titleStyle.Render(v.topic.Title)
	if v.loading {
		return header + "\n\n" + v.spinner.View() + " Loading..."
	}
	if v.err != "" {
		return header + "\n\n" + errStyle.Render(v.err) + "\n\n" + v.viewport.View()
	}
	footer := helpStyle.Render(fmt.Sprintf("j/k scroll • r reply • h back  %d%%", int(v.viewport.ScrollPercent()*100)))
	return header + "\n" + v.viewport.View() + "\n" + footer
}
