package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sam/jtech-tui/internal/api"
	"github.com/sam/jtech-tui/internal/config"
	"github.com/sam/jtech-tui/internal/ui"
)

func main() {
	feedFlag := flag.String("feed", "", "starting feed: latest, new, top, unseen, categories")
	devFlag := flag.Bool("dev", false, "enable developer mode with resume and status output")
	noAltScreenFlag := flag.Bool("no-alt-screen", false, "run without the terminal alternate screen")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		os.Exit(1)
	}

	if *feedFlag != "" {
		cfg.DefaultFeed = *feedFlag
	}

	if *devFlag {
		*noAltScreenFlag = true
	}

	client, err := api.New(cfg.ForumURL, cfg.SessionCookie)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating client: %v\n", err)
		os.Exit(1)
	}

	var resumeState *config.UIState
	if *devFlag {
		resumeState, err = config.LoadUIState()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error loading UI state: %v\n", err)
			os.Exit(1)
		}
	}

	app := ui.NewApp(cfg, client, ui.AppOptions{
		DevMode:     *devFlag,
		ResumeState: resumeState,
	})
	opts := []tea.ProgramOption{}
	if !*noAltScreenFlag {
		opts = append(opts, tea.WithAltScreen())
	}
	p := tea.NewProgram(app, opts...)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
