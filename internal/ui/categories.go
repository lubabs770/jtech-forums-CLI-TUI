package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sam/jtech-tui/internal/api"
)

// --- Category list view ---

type catItem struct{ cat api.Category }

func (c catItem) FilterValue() string { return c.cat.Name }
func (c catItem) Title() string       { return categoryBadge(c.cat) }
func (c catItem) Description() string { return categoryDescription(c.cat) }

type catsLoadedMsg struct {
	cats []api.Category
	err  error
}

type categoriesView struct {
	client       *api.Client
	list         list.Model
	spinner      spinner.Model
	categoryByID map[int]api.Category
	loading      bool
	err          string
	width        int
	height       int
	restoreIndex int
}

func newCategoriesView(client *api.Client) *categoriesView {
	l := list.New(nil, list.NewDefaultDelegate(), 0, 0)
	l.SetShowHelp(false)
	l.SetShowTitle(false)
	l.DisableQuitKeybindings()
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	return &categoriesView{client: client, list: l, spinner: sp, restoreIndex: -1}
}

func (v *categoriesView) Init() tea.Cmd {
	v.loading = true
	client := v.client
	return tea.Batch(v.spinner.Tick, func() tea.Msg {
		cats, err := client.GetCategories()
		return catsLoadedMsg{cats: cats, err: err}
	})
}

func (v *categoriesView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width, v.height = msg.Width, msg.Height
		v.list.SetSize(msg.Width, msg.Height-5)

	case tea.KeyMsg:
		switch msg.String() {
		case "h":
			return v, func() tea.Msg { return popViewMsg{} }
		case "enter":
			if item, ok := v.list.SelectedItem().(catItem); ok {
				var parent *api.Category
				if found, ok := v.categoryByID[item.cat.ParentCategoryID]; ok {
					parentCopy := found
					parent = &parentCopy
				}
				return v, func() tea.Msg { return openCategoryMsg{cat: item.cat, parent: parent} }
			}
		}

	case catsLoadedMsg:
		v.loading = false
		if msg.err != nil {
			if isUnauthorized(msg.err) {
				return v, func() tea.Msg { return unauthorizedMsg{} }
			}
			v.err = msg.err.Error()
			return v, nil
		}
		items := make([]list.Item, len(msg.cats))
		v.categoryByID = categoryMap(msg.cats)
		for i, c := range msg.cats {
			items[i] = catItem{cat: c}
		}
		v.list.SetItems(items)
		if len(items) > 0 {
			index := v.restoreIndex
			if index < 0 {
				index = 0
			}
			if index >= len(items) {
				index = len(items) - 1
			}
			v.list.Select(index)
		}
		v.restoreIndex = -1

	case spinner.TickMsg:
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		return v, cmd
	}

	var cmd tea.Cmd
	v.list, cmd = v.list.Update(msg)
	return v, cmd
}

func (v *categoriesView) View() string {
	header := renderFeedTabs(len(feeds) - 1)
	if v.loading {
		return header + "\n\n" + v.spinner.View() + " Loading..."
	}
	if v.err != "" {
		return header + "\n\n" + errStyle.Render(v.err)
	}
	return header + "\n" + v.list.View() + "\n" + helpStyle.Render("enter open • h back")
}

func (v *categoriesView) restoreSelection(index int) {
	v.restoreIndex = index
}

func (v *categoriesView) debugStatus() string {
	if v.loading {
		return "loading"
	}
	if v.err != "" {
		return "error: " + v.err
	}
	return fmt.Sprintf("ready:%d", len(v.list.Items()))
}

// --- Category topics view ---

type catTopicsLoadedMsg struct {
	topics []api.Topic
	err    error
}

type categoryTopicsView struct {
	client       *api.Client
	cat          api.Category
	parent       *api.Category
	list         list.Model
	spinner      spinner.Model
	loading      bool
	err          string
	width        int
	height       int
	restoreIndex int
}

func newCategoriesTopicsView(client *api.Client, cat api.Category, parent *api.Category) *categoryTopicsView {
	l := list.New(nil, list.NewDefaultDelegate(), 0, 0)
	l.SetShowHelp(false)
	l.SetShowTitle(false)
	l.DisableQuitKeybindings()
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	return &categoryTopicsView{client: client, cat: cat, parent: parent, list: l, spinner: sp, restoreIndex: -1}
}

func (v *categoryTopicsView) Init() tea.Cmd {
	v.loading = true
	client, cat := v.client, v.cat
	return tea.Batch(v.spinner.Tick, func() tea.Msg {
		topics, err := client.GetCategoryTopics(cat.Slug, cat.ID)
		return catTopicsLoadedMsg{topics: topics, err: err}
	})
}

func (v *categoryTopicsView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width, v.height = msg.Width, msg.Height
		v.list.SetSize(msg.Width, msg.Height-5)

	case tea.KeyMsg:
		switch msg.String() {
		case "h":
			return v, func() tea.Msg { return popViewMsg{} }
		case "enter":
			if item, ok := v.list.SelectedItem().(topicItem); ok {
				catCopy := v.cat
				return v, func() tea.Msg {
					return openTopicMsg{topic: item.topic, category: &catCopy, parent: v.parent, feedIndex: len(feeds) - 1}
				}
			}
		}

	case catTopicsLoadedMsg:
		v.loading = false
		if msg.err != nil {
			if isUnauthorized(msg.err) {
				return v, func() tea.Msg { return unauthorizedMsg{} }
			}
			v.err = msg.err.Error()
			return v, nil
		}
		items := make([]list.Item, len(msg.topics))
		for i, t := range msg.topics {
			items[i] = topicItem{
				topic:      t,
				categories: map[int]api.Category{v.cat.ID: v.cat},
			}
		}
		v.list.SetItems(items)
		if len(items) > 0 {
			index := v.restoreIndex
			if index < 0 {
				index = 0
			}
			if index >= len(items) {
				index = len(items) - 1
			}
			v.list.Select(index)
		}
		v.restoreIndex = -1

	case spinner.TickMsg:
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		return v, cmd
	}

	var cmd tea.Cmd
	v.list, cmd = v.list.Update(msg)
	return v, cmd
}

func (v *categoryTopicsView) View() string {
	header := renderFeedTabs(len(feeds)-1) + "\n" + categorySectionHeaderWithParent(v.parent, v.cat)
	if v.loading {
		return header + "\n" + v.spinner.View() + " Loading..."
	}
	if v.err != "" {
		return header + "\n" + errStyle.Render(v.err)
	}
	return header + "\n" + v.list.View() + "\n" + helpStyle.Render("enter open • h back")
}

func (v *categoryTopicsView) restoreSelection(index int) {
	v.restoreIndex = index
}

func (v *categoryTopicsView) debugStatus() string {
	if v.loading {
		return "loading"
	}
	if v.err != "" {
		return "error: " + v.err
	}
	return fmt.Sprintf("ready:%d", len(v.list.Items()))
}
