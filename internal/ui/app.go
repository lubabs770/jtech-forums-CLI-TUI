package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sam/jtech-tui/internal/api"
	"github.com/sam/jtech-tui/internal/config"
)

type App struct {
	cfg           *config.Config
	client        *api.Client
	opts          AppOptions
	stack         []tea.Model
	width         int
	height        int
	restoreStatus string
}

func NewApp(cfg *config.Config, client *api.Client, opts AppOptions) *App {
	return &App{cfg: cfg, client: client, opts: opts}
}

func (a *App) Init() tea.Cmd {
	if a.cfg.SessionCookie == "" {
		login := newLoginView(a.client, a.cfg)
		a.stack = []tea.Model{login}
		return login.Init()
	}
	return a.initAuthedStack()
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width, a.height = msg.Width, msg.Height
		cmd := a.resizeAll(msg)
		a.persistUIState()
		return a, cmd

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return a, tea.Quit
		}

	case loggedInMsg:
		a.cfg.SessionCookie = msg.cookie
		a.cfg.Save()
		cmd := a.initAuthedStack()
		a.persistUIState()
		return a, tea.Batch(cmd, tea.WindowSize())

	case popViewMsg:
		if len(a.stack) > 1 {
			a.stack = a.stack[:len(a.stack)-1]
		}
		a.persistUIState()
		return a, nil

	case unauthorizedMsg:
		a.cfg.SessionCookie = ""
		a.cfg.Save()
		login := newLoginView(a.client, a.cfg)
		a.stack = []tea.Model{login}
		a.persistUIState()
		return a, login.Init()

	case openTopicMsg:
		thread := newThreadView(a.client, msg.topic, msg.category, msg.parent, msg.feedIndex)
		a.stack = append(a.stack, thread)
		cmds := []tea.Cmd{thread.Init()}
		if a.width > 0 && a.height > 0 {
			sized, cmd := thread.Update(tea.WindowSizeMsg{Width: a.width, Height: a.height})
			a.stack[len(a.stack)-1] = sized
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		a.persistUIState()
		return a, tea.Batch(cmds...)

	case openCategoryMsg:
		if msg.cat.ID == 0 {
			catView := newCategoriesView(a.client)
			a.stack = append(a.stack, catView)
			sized, cmd := catView.Update(tea.WindowSizeMsg{Width: a.width, Height: a.height})
			a.stack[len(a.stack)-1] = sized
			a.persistUIState()
			return a, tea.Batch(catView.Init(), cmd)
		}
		catTopics := newCategoriesTopicsView(a.client, msg.cat, msg.parent)
		a.stack = append(a.stack, catTopics)
		sized, cmd := catTopics.Update(tea.WindowSizeMsg{Width: a.width, Height: a.height})
		a.stack[len(a.stack)-1] = sized
		a.persistUIState()
		return a, tea.Batch(catTopics.Init(), cmd)
	}

	if len(a.stack) == 0 {
		return a, nil
	}
	top := a.stack[len(a.stack)-1]
	updated, cmd := top.Update(msg)
	a.stack[len(a.stack)-1] = updated
	a.persistUIState()
	return a, cmd
}

func (a *App) View() string {
	if len(a.stack) == 0 {
		return ""
	}
	view := a.stack[len(a.stack)-1].View()
	if a.opts.DevMode {
		view += "\n" + a.debugFooter()
	}
	return view
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

func (a *App) initAuthedStack() tea.Cmd {
	stack, cmds, status := a.buildInitialStack()
	a.stack = stack
	a.restoreStatus = status
	return tea.Batch(cmds...)
}

func (a *App) buildInitialStack() ([]tea.Model, []tea.Cmd, string) {
	state := a.opts.ResumeState
	feedName := a.cfg.DefaultFeed
	if state != nil && state.Feed != "" {
		feedName = state.Feed
	}

	feed := newFeedView(a.client, feedName)
	if state != nil {
		feed.restoreSelection(state.FeedIndex)
	}

	stack := []tea.Model{feed}
	cmds := []tea.Cmd{feed.Init()}
	status := "fresh"

	if state == nil || state.View == "" {
		return stack, cmds, status
	}

	status = "restored " + state.View
	switch state.View {
	case "categories":
		catView := newCategoriesView(a.client)
		catView.restoreSelection(state.CategoriesIndex)
		stack = append(stack, catView)
		cmds = append(cmds, catView.Init())
	case "category_topics":
		if state.CategoryID == 0 || state.CategorySlug == "" {
			return stack, cmds, "resume degraded to feed"
		}
		catView := newCategoriesView(a.client)
		catView.restoreSelection(state.CategoriesIndex)
		catTopics := newCategoriesTopicsView(a.client, restoredCategory(state), restoredParentCategory(state))
		catTopics.restoreSelection(state.CategoryTopicsIndex)
		stack = append(stack, catView, catTopics)
		cmds = append(cmds, catView.Init(), catTopics.Init())
	case "thread":
		if state.TopicID == 0 {
			return stack, cmds, "resume degraded to feed"
		}
		var category *api.Category
		if state.CategoryID != 0 && state.CategorySlug != "" {
			cat := restoredCategory(state)
			category = &cat
		}
		if state.ThreadSource == "category_topics" && category != nil {
			catView := newCategoriesView(a.client)
			catView.restoreSelection(state.CategoriesIndex)
			catTopics := newCategoriesTopicsView(a.client, *category, restoredParentCategory(state))
			catTopics.restoreSelection(state.CategoryTopicsIndex)
			stack = append(stack, catView, catTopics)
			cmds = append(cmds, catView.Init(), catTopics.Init())
		}
		thread := newThreadView(a.client, restoredTopic(state), category, restoredParentCategory(state), restoredFeedIndex(state))
		thread.restoreScroll(state.ThreadYOffset)
		stack = append(stack, thread)
		cmds = append(cmds, thread.Init())
	default:
		status = "resume ignored"
	}

	return stack, cmds, status
}

func restoredCategory(state *config.UIState) api.Category {
	return api.Category{
		ID:               state.CategoryID,
		Slug:             state.CategorySlug,
		Name:             state.CategoryName,
		Color:            state.CategoryColor,
		TextColor:        state.CategoryTextColor,
		ParentCategoryID: state.ParentCategoryID,
	}
}

func restoredParentCategory(state *config.UIState) *api.Category {
	if state.ParentCategoryID == 0 {
		return nil
	}
	parent := api.Category{
		ID:        state.ParentCategoryID,
		Slug:      state.ParentCategorySlug,
		Name:      state.ParentCategoryName,
		Color:     state.ParentCategoryColor,
		TextColor: state.ParentCategoryTextColor,
	}
	return &parent
}

func restoredFeedIndex(state *config.UIState) int {
	for i, feed := range feeds {
		if feed == state.Feed {
			return i
		}
	}
	return 0
}

func restoredTopic(state *config.UIState) api.Topic {
	return api.Topic{
		ID:         state.TopicID,
		Slug:       state.TopicSlug,
		Title:      state.TopicTitle,
		CategoryID: state.TopicCategoryID,
	}
}

func (a *App) persistUIState() {
	if !a.opts.DevMode {
		return
	}
	state := a.snapshotUIState()
	if state == nil {
		return
	}
	_ = config.SaveUIState(state)
}

func (a *App) snapshotUIState() *config.UIState {
	if len(a.stack) == 0 {
		return nil
	}

	state := &config.UIState{}
	if feed, ok := a.stack[0].(*feedView); ok {
		state.Feed = feed.currentFeed()
		state.FeedIndex = feed.list.Index()
	}

	switch top := a.stack[len(a.stack)-1].(type) {
	case *feedView:
		state.View = "feed"
	case *categoriesView:
		state.View = "categories"
		state.CategoriesIndex = top.list.Index()
	case *categoryTopicsView:
		state.View = "category_topics"
		state.CategoriesIndex = a.categoriesSelection()
		state.CategoryTopicsIndex = top.list.Index()
		fillCategoryState(state, top.cat)
		fillParentCategoryState(state, top.parent)
	case *threadView:
		state.View = "thread"
		state.CategoriesIndex = a.categoriesSelection()
		state.CategoryTopicsIndex = a.categoryTopicsSelection()
		state.ThreadYOffset = top.viewport.YOffset
		state.TopicID = top.topic.ID
		state.TopicSlug = top.topic.Slug
		state.TopicTitle = top.topic.Title
		state.TopicCategoryID = top.topic.CategoryID
		if catTopics := a.currentCategoryTopicsView(); catTopics != nil {
			state.ThreadSource = "category_topics"
			fillCategoryState(state, catTopics.cat)
			fillParentCategoryState(state, catTopics.parent)
		} else {
			state.ThreadSource = "feed"
			if top.category != nil {
				fillCategoryState(state, *top.category)
			}
			fillParentCategoryState(state, nil)
		}
	default:
		return nil
	}

	return state
}

func fillCategoryState(state *config.UIState, cat api.Category) {
	state.CategoryID = cat.ID
	state.CategorySlug = cat.Slug
	state.CategoryName = cat.Name
	state.CategoryColor = cat.Color
	state.CategoryTextColor = cat.TextColor
	state.ParentCategoryID = cat.ParentCategoryID
}

func fillParentCategoryState(state *config.UIState, cat *api.Category) {
	if cat == nil {
		state.ParentCategoryID = 0
		state.ParentCategorySlug = ""
		state.ParentCategoryName = ""
		state.ParentCategoryColor = ""
		state.ParentCategoryTextColor = ""
		return
	}
	state.ParentCategoryID = cat.ID
	state.ParentCategorySlug = cat.Slug
	state.ParentCategoryName = cat.Name
	state.ParentCategoryColor = cat.Color
	state.ParentCategoryTextColor = cat.TextColor
}

func (a *App) categoriesSelection() int {
	for i := len(a.stack) - 1; i >= 0; i-- {
		if v, ok := a.stack[i].(*categoriesView); ok {
			return v.list.Index()
		}
	}
	return 0
}

func (a *App) categoryTopicsSelection() int {
	for i := len(a.stack) - 1; i >= 0; i-- {
		if v, ok := a.stack[i].(*categoryTopicsView); ok {
			return v.list.Index()
		}
	}
	return 0
}

func (a *App) currentCategoryTopicsView() *categoryTopicsView {
	for i := len(a.stack) - 1; i >= 0; i-- {
		if v, ok := a.stack[i].(*categoryTopicsView); ok {
			return v
		}
	}
	return nil
}

func (a *App) debugFooter() string {
	state := a.snapshotUIState()
	if state == nil {
		return helpStyle.Render("dev")
	}

	parts := []string{
		"dev",
		"view=" + state.View,
	}
	if state.Feed != "" {
		parts = append(parts, "feed="+state.Feed)
	}
	if state.TopicID != 0 {
		parts = append(parts, fmt.Sprintf("topic=%d", state.TopicID))
	}
	if state.CategoryID != 0 {
		parts = append(parts, fmt.Sprintf("cat=%d", state.CategoryID))
	}
	parts = append(parts, "status="+a.currentStatus())
	if a.restoreStatus != "" {
		parts = append(parts, "resume="+a.restoreStatus)
	}

	return lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(strings.Join(parts, " • "))
}

func (a *App) currentStatus() string {
	if len(a.stack) == 0 {
		return "idle"
	}
	switch v := a.stack[len(a.stack)-1].(type) {
	case *feedView:
		return v.debugStatus()
	case *categoriesView:
		return v.debugStatus()
	case *categoryTopicsView:
		return v.debugStatus()
	case *threadView:
		return v.debugStatus()
	default:
		return "ready"
	}
}
