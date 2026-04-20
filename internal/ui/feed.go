package ui

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sam/jtech-tui/internal/api"
)

var feeds = []string{"latest", "new", "top", "unseen", "categories"}

type topicItem struct{ topic api.Topic }

func (t topicItem) FilterValue() string { return t.topic.Title }
func (t topicItem) Title() string       { return t.topic.Title }
func (t topicItem) Description() string {
	return fmt.Sprintf("%d replies • %s", t.topic.ReplyCount, formatTime(t.topic.LastPostedAt))
}

type feedLoadedMsg struct {
	topics []api.Topic
	err    error
}

type openCategoryListMsg struct{}

type feedView struct {
	client            *api.Client
	feedIndex         int
	list              list.Model
	spinner           spinner.Model
	loading           bool
	err               string
	width             int
	height            int
	form              *newTopicForm
	categories        []api.Category
	pendingTitle      string
	pendingCategoryID int
}

func newFeedView(client *api.Client, defaultFeed string) *feedView {
	idx := 0
	for i, f := range feeds {
		if f == defaultFeed {
			idx = i
			break
		}
	}
	l := list.New(nil, list.NewDefaultDelegate(), 0, 0)
	l.SetShowHelp(false)
	l.SetShowTitle(false)
	l.DisableQuitKeybindings()

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	return &feedView{client: client, feedIndex: idx, list: l, spinner: sp}
}

func (f *feedView) currentFeed() string { return feeds[f.feedIndex] }

func (f *feedView) loadFeed() tea.Cmd {
	feed := f.currentFeed()
	if feed == "categories" {
		return func() tea.Msg { return openCategoryListMsg{} }
	}
	client := f.client
	return func() tea.Msg {
		topics, err := client.GetFeed(feed)
		return feedLoadedMsg{topics: topics, err: err}
	}
}

func (f *feedView) Init() tea.Cmd {
	f.loading = true
	return tea.Batch(f.spinner.Tick, f.loadFeed())
}

func (f *feedView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Route keyboard to form if active
	if f.form != nil {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				f.form = nil
				return f, nil
			}
			updated, cmd := f.form.Update(msg)
			f.form = updated
			return f, cmd
		case newTopicFormDoneMsg:
			f.pendingTitle = msg.title
			f.pendingCategoryID = msg.categoryID
			f.form = nil
			return f, openEditor(fmt.Sprintf("# %s\n\nWrite your post here...\n", msg.title))
		}
		var cmd tea.Cmd
		updated, cmd := f.form.Update(msg)
		f.form = updated
		return f, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		f.width, f.height = msg.Width, msg.Height
		f.list.SetSize(msg.Width, msg.Height-3)

	case tea.KeyMsg:
		switch msg.String() {
		case "h":
			return f, func() tea.Msg { return popViewMsg{} }
		case "tab", "shift+tab":
			if msg.String() == "shift+tab" {
				f.feedIndex = (f.feedIndex - 1 + len(feeds)) % len(feeds)
			} else {
				f.feedIndex = (f.feedIndex + 1) % len(feeds)
			}
			// If landing on "categories", push the sub-view without wiping
			// this view's state so returning with `h` restores the prior feed.
			if f.currentFeed() == "categories" {
				// revert index so the underlying feed stays intact
				if msg.String() == "shift+tab" {
					f.feedIndex = (f.feedIndex + 1) % len(feeds)
				} else {
					f.feedIndex = (f.feedIndex - 1 + len(feeds)) % len(feeds)
				}
				return f, func() tea.Msg { return openCategoryMsg{} }
			}
			f.loading = true
			f.err = ""
			f.list.SetItems(nil)
			return f, tea.Batch(f.spinner.Tick, f.loadFeed())
		case "enter":
			if item, ok := f.list.SelectedItem().(topicItem); ok {
				return f, func() tea.Msg { return openTopicMsg{topic: item.topic} }
			}
		case "n":
			if len(f.categories) == 0 {
				client := f.client
				return f, func() tea.Msg {
					cats, err := client.GetCategories()
					return catsForFormMsg{cats: cats, err: err}
				}
			}
			f.form = newNewTopicForm(f.categories)
			return f, textinput.Blink
		}

	case feedLoadedMsg:
		f.loading = false
		if msg.err != nil {
			if isUnauthorized(msg.err) {
				return f, func() tea.Msg { return unauthorizedMsg{} }
			}
			f.err = msg.err.Error()
			return f, nil
		}
		items := make([]list.Item, len(msg.topics))
		for i, t := range msg.topics {
			items[i] = topicItem{topic: t}
		}
		f.list.SetItems(items)

	case openCategoryListMsg:
		return f, func() tea.Msg { return openCategoryMsg{} }

	case catsForFormMsg:
		if msg.err == nil {
			f.categories = msg.cats
			f.form = newNewTopicForm(f.categories)
		}
		return f, textinput.Blink

	case newTopicFormDoneMsg:
		f.pendingTitle = msg.title
		f.pendingCategoryID = msg.categoryID
		f.form = nil
		return f, openEditor(fmt.Sprintf("# %s\n\nWrite your post here...\n", msg.title))

	case editorFinishedMsg:
		if msg.err != nil || strings.TrimSpace(msg.content) == "" {
			return f, nil
		}
		body := strings.TrimSpace(msg.content)
		// Strip leading "# Title\n\n" template line if present
		if strings.HasPrefix(body, "# ") {
			if idx := strings.Index(body, "\n"); idx != -1 {
				body = strings.TrimSpace(body[idx:])
			}
		}
		title := f.pendingTitle
		catID := f.pendingCategoryID
		client := f.client
		feed := f.currentFeed()
		return f, func() tea.Msg {
			err := client.CreateTopic(title, body, catID)
			if err != nil {
				if isUnauthorized(err) {
					return unauthorizedMsg{}
				}
				return newTopicErrMsg{err: err}
			}
			topics, err := client.GetFeed(feed)
			return feedLoadedMsg{topics: topics, err: err}
		}

	case newTopicErrMsg:
		f.err = msg.err.Error()

	case spinner.TickMsg:
		var cmd tea.Cmd
		f.spinner, cmd = f.spinner.Update(msg)
		return f, cmd
	}

	var cmd tea.Cmd
	f.list, cmd = f.list.Update(msg)
	return f, cmd
}

func (f *feedView) baseView() string {
	var tabs []string
	for i, feed := range feeds {
		if i == f.feedIndex {
			tabs = append(tabs, activeTabStyle.Render(feed))
		} else {
			tabs = append(tabs, tabStyle.Render(feed))
		}
	}
	header := strings.Join(tabs, "")

	if f.loading {
		return header + "\n\n" + f.spinner.View() + " Loading..."
	}
	if f.err != "" {
		return header + "\n\n" + errStyle.Render(f.err)
	}
	return header + "\n" + f.list.View() + "\n" + helpStyle.Render("enter open • n new topic • tab/shift+tab switch feed • h back")
}

func (f *feedView) View() string {
	if f.form != nil {
		return lipgloss.Place(f.width, f.height, lipgloss.Center, lipgloss.Center, f.form.View())
	}
	return f.baseView()
}

func isUnauthorized(err error) bool {
	var e *api.ErrUnauthorized
	return errors.As(err, &e)
}

func formatTime(s string) string {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return s
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}
