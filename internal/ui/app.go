package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sam/jtech-tui/internal/api"
	"github.com/sam/jtech-tui/internal/config"
)

type App struct {
	cfg    *config.Config
	client *api.Client
	stack  []tea.Model
	width  int
	height int
}

func NewApp(cfg *config.Config, client *api.Client) *App {
	return &App{cfg: cfg, client: client}
}

func (a *App) Init() tea.Cmd {
	if a.cfg.SessionCookie == "" {
		login := newLoginView(a.client, a.cfg)
		a.stack = []tea.Model{login}
		return login.Init()
	}
	feed := newFeedView(a.client, a.cfg.DefaultFeed)
	a.stack = []tea.Model{feed}
	return feed.Init()
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width, a.height = msg.Width, msg.Height
		return a, a.resizeAll(msg)

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return a, tea.Quit
		}

	case loggedInMsg:
		a.cfg.SessionCookie = msg.cookie
		a.cfg.Save()
		feed := newFeedView(a.client, a.cfg.DefaultFeed)
		a.stack = []tea.Model{feed}
		return a, tea.Batch(feed.Init(), tea.WindowSize())

	case popViewMsg:
		if len(a.stack) > 1 {
			a.stack = a.stack[:len(a.stack)-1]
		}
		return a, nil

	case unauthorizedMsg:
		a.cfg.SessionCookie = ""
		a.cfg.Save()
		login := newLoginView(a.client, a.cfg)
		a.stack = []tea.Model{login}
		return a, login.Init()

	case openTopicMsg:
		thread := newThreadView(a.client, msg.topic)
		a.stack = append(a.stack, thread)
		cmds := []tea.Cmd{thread.Init()}
		if a.width > 0 && a.height > 0 {
			sized, cmd := thread.Update(tea.WindowSizeMsg{Width: a.width, Height: a.height})
			a.stack[len(a.stack)-1] = sized
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		return a, tea.Batch(cmds...)

	case openCategoryMsg:
		if msg.cat.ID == 0 {
			catView := newCategoriesView(a.client)
			a.stack = append(a.stack, catView)
			sized, cmd := catView.Update(tea.WindowSizeMsg{Width: a.width, Height: a.height})
			a.stack[len(a.stack)-1] = sized
			return a, tea.Batch(catView.Init(), cmd)
		}
		catTopics := newCategoriesTopicsView(a.client, msg.cat)
		a.stack = append(a.stack, catTopics)
		sized, cmd := catTopics.Update(tea.WindowSizeMsg{Width: a.width, Height: a.height})
		a.stack[len(a.stack)-1] = sized
		return a, tea.Batch(catTopics.Init(), cmd)
	}

	if len(a.stack) == 0 {
		return a, nil
	}
	top := a.stack[len(a.stack)-1]
	updated, cmd := top.Update(msg)
	a.stack[len(a.stack)-1] = updated
	return a, cmd
}

func (a *App) View() string {
	if len(a.stack) == 0 {
		return ""
	}
	return a.stack[len(a.stack)-1].View()
}

func (a *App) resizeAll(msg tea.WindowSizeMsg) tea.Cmd {
	var cmds []tea.Cmd
	for i, v := range a.stack {
		updated, cmd := v.Update(msg)
		a.stack[i] = updated
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return tea.Batch(cmds...)
}
