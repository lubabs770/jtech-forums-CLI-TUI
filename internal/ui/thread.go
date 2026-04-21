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
	client         *api.Client
	topic          api.Topic
	category       *api.Category
	parent         *api.Category
	feedIndex      int
	thread         *api.Thread
	visiblePosts   []api.Post
	loadedStart    int
	viewport       viewport.Model
	spinner        spinner.Model
	loading        bool
	err            string
	width          int
	height         int
	lastKey        string
	restoreYOffset int
}

const threadPageSize = 30
const threadPreloadThreshold = 20

func newThreadView(client *api.Client, topic api.Topic, category *api.Category, parent *api.Category, feedIndex int) *threadView {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	return &threadView{client: client, topic: topic, category: category, parent: parent, feedIndex: feedIndex, spinner: sp, restoreYOffset: -1}
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
		if msg.Height > 3 {
			wasAtBottom := v.viewport.AtBottom()
			v.viewport.Width = msg.Width
			v.viewport.Height = msg.Height - 3
			if len(v.visiblePosts) > 0 {
				v.viewport.SetContent(renderPosts(v.visiblePosts, msg.Width))
				if wasAtBottom {
					v.viewport.GotoBottom()
				}
			}
		}

	case tea.KeyMsg:
		key := msg.String()
		switch key {
		case "h":
			v.lastKey = ""
			return v, func() tea.Msg { return popViewMsg{} }
		case "r":
			v.lastKey = ""
			if v.thread != nil {
				return v, openEditor("")
			}
		case "g":
			if v.lastKey == "g" {
				v.viewport.GotoTop()
				v.lastKey = ""
				return v, nil
			}
		}
		v.lastKey = key

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
		v.loadedStart = initialThreadStart(msg.thread.PostStream.Posts)
		v.visiblePosts = visibleThreadPosts(msg.thread.PostStream.Posts, v.loadedStart)
		if v.width > 0 && v.height > 3 {
			v.viewport = viewport.New(v.width, v.height-3)
			v.viewport.SetContent(renderPosts(v.visiblePosts, v.width))
			if v.restoreYOffset >= 0 {
				v.viewport.SetYOffset(v.restoreYOffset)
				v.restoreYOffset = -1
			} else {
				v.viewport.GotoBottom()
			}
		}

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
	v.maybeLoadOlderPosts()
	return v, cmd
}

func (v *threadView) View() string {
	headerParts := []string{}
	if v.feedIndex >= 0 && v.feedIndex < len(feeds) {
		headerParts = append(headerParts, renderFeedTabs(v.feedIndex))
	}
	if v.category != nil {
		headerParts = append(headerParts, topicContextBar(v.parent, *v.category, v.topic.Title, formatTime(v.topic.LastPostedAt)))
	} else {
		headerParts = append(headerParts, topicHeaderBar(v.topic.Title, formatTime(v.topic.LastPostedAt)))
	}
	header := strings.Join(headerParts, "\n")
	if v.loading {
		return header + "\n\n" + v.spinner.View() + " Loading..."
	}
	if v.err != "" {
		return header + "\n\n" + errStyle.Render(v.err) + "\n\n" + v.viewport.View()
	}
	footer := helpStyle.Render(fmt.Sprintf("j/k scroll • r reply • h back  %d%%", int(v.viewport.ScrollPercent()*100)))
	return header + "\n" + v.viewport.View() + "\n" + footer
}

func (v *threadView) restoreScroll(yOffset int) {
	v.restoreYOffset = yOffset
}

func (v *threadView) debugStatus() string {
	if v.loading {
		return "loading"
	}
	if v.err != "" {
		return "error: " + v.err
	}
	if v.thread == nil {
		return "ready:empty"
	}
	return fmt.Sprintf("ready:posts=%d", len(v.thread.PostStream.Posts))
}

func initialThreadStart(posts []api.Post) int {
	if len(posts) <= threadPageSize {
		return 0
	}
	return len(posts) - threadPageSize
}

func visibleThreadPosts(posts []api.Post, start int) []api.Post {
	if start < 0 {
		start = 0
	}
	if start > len(posts) {
		start = len(posts)
	}
	return posts[start:]
}

func (v *threadView) maybeLoadOlderPosts() {
	if v.thread == nil || v.width == 0 || v.loadedStart == 0 || v.viewport.YOffset > threadPreloadThreshold {
		return
	}

	nextStart := v.loadedStart - threadPageSize
	if nextStart < 0 {
		nextStart = 0
	}
	if nextStart == v.loadedStart {
		return
	}

	oldContent := renderPosts(v.visiblePosts, v.width)
	oldLines := renderedLineCount(oldContent)

	v.loadedStart = nextStart
	v.visiblePosts = visibleThreadPosts(v.thread.PostStream.Posts, v.loadedStart)
	newContent := renderPosts(v.visiblePosts, v.width)
	newLines := renderedLineCount(newContent)

	v.viewport.SetContent(newContent)
	v.viewport.SetYOffset(v.viewport.YOffset + (newLines - oldLines))
}

func renderedLineCount(content string) int {
	if content == "" {
		return 0
	}
	return strings.Count(content, "\n") + 1
}
