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
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		os.Exit(1)
	}

	if *feedFlag != "" {
		cfg.DefaultFeed = *feedFlag
	}

	client, err := api.New(cfg.ForumURL, cfg.SessionCookie)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating client: %v\n", err)
		os.Exit(1)
	}

	app := ui.NewApp(cfg, client)
	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
