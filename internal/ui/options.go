package ui

import "github.com/sam/jtech-tui/internal/config"

type AppOptions struct {
	DevMode     bool
	ResumeState *config.UIState
}
