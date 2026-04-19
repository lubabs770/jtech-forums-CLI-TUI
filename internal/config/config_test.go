package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sam/jtech-tui/internal/config"
)

func TestLoadDefaults(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.ForumURL != "https://forums.jtechforums.org" {
		t.Errorf("ForumURL = %q, want %q", cfg.ForumURL, "https://forums.jtechforums.org")
	}
	if cfg.DefaultFeed != "latest" {
		t.Errorf("DefaultFeed = %q, want %q", cfg.DefaultFeed, "latest")
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	original := &config.Config{
		ForumURL:      "https://forums.jtechforums.org",
		DefaultFeed:   "top",
		SessionCookie: "abc123",
	}

	if err := original.Save(); err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	loaded, err := config.Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if loaded.SessionCookie != original.SessionCookie {
		t.Errorf("SessionCookie = %q, want %q", loaded.SessionCookie, original.SessionCookie)
	}
	if loaded.DefaultFeed != original.DefaultFeed {
		t.Errorf("DefaultFeed = %q, want %q", loaded.DefaultFeed, original.DefaultFeed)
	}
}

func TestConfigPath(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	cfg := &config.Config{
		ForumURL:    "https://forums.jtechforums.org",
		DefaultFeed: "latest",
	}

	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	expectedPath := filepath.Join(tmpHome, ".config", "jtech-tui", "config.json")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("config file not found at expected path: %s", expectedPath)
	}
}
