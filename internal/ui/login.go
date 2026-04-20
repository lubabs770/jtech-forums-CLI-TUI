package ui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sam/jtech-tui/internal/api"
	"github.com/sam/jtech-tui/internal/config"
)

type loginView struct {
	client   *api.Client
	cfg      *config.Config
	username textinput.Model
	password textinput.Model
	focused  int
	err      string
	loading  bool
}

func newLoginView(client *api.Client, cfg *config.Config) *loginView {
	u := textinput.New()
	u.Placeholder = "Username"
	u.Focus()

	p := textinput.New()
	p.Placeholder = "Password"
	p.EchoMode = textinput.EchoPassword
	p.EchoCharacter = '•'

	return &loginView{client: client, cfg: cfg, username: u, password: p}
}

func (l *loginView) Init() tea.Cmd { return textinput.Blink }

type loginResultMsg struct {
	cookie string
	err    error
}

func (l *loginView) doLogin() tea.Cmd {
	u, p := l.username.Value(), l.password.Value()
	client := l.client
	return func() tea.Msg {
		cookie, err := client.Login(u, p)
		return loginResultMsg{cookie: cookie, err: err}
	}
}

func (l *loginView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if l.loading {
			return l, nil
		}
		switch msg.String() {
		case "tab", "down":
			l.focused = (l.focused + 1) % 2
			if l.focused == 0 {
				l.username.Focus()
				l.password.Blur()
			} else {
				l.password.Focus()
				l.username.Blur()
			}
		case "enter":
			if l.focused == 0 && l.username.Value() != "" {
				l.focused = 1
				l.password.Focus()
				l.username.Blur()
			} else if l.username.Value() != "" && l.password.Value() != "" {
				l.loading = true
				l.err = ""
				return l, l.doLogin()
			}
		}

	case loginResultMsg:
		l.loading = false
		if msg.err != nil {
			l.err = msg.err.Error()
			l.password.SetValue("")
			l.password.Focus()
			return l, textinput.Blink
		}
		return l, func() tea.Msg { return loggedInMsg{cookie: msg.cookie} }
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd
	l.username, cmd = l.username.Update(msg)
	cmds = append(cmds, cmd)
	l.password, cmd = l.password.Update(msg)
	cmds = append(cmds, cmd)
	return l, tea.Batch(cmds...)
}

func (l *loginView) View() string {
	s := titleStyle.Render("jtech forums") + "\n\n"
	s += l.username.View() + "\n"
	s += l.password.View() + "\n\n"
	if l.loading {
		s += "Logging in...\n"
	} else if l.err != "" {
		s += errStyle.Render(l.err) + "\n"
	}
	s += "\n" + helpStyle.Render("tab switch fields • enter submit • ctrl+c quit")
	return s
}
