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

func TestSaveAndLoadUIState(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	original := &config.UIState{
		View:              "thread",
		ThreadSource:      "feed",
		Feed:              "latest",
		FeedIndex:         2,
		TopicID:           42,
		TopicTitle:        "Restored topic",
		ThreadYOffset:     7,
		CategoryID:        5,
		CategorySlug:      "general",
		CategoryName:      "General",
		CategoryColor:     "00B3FF",
		CategoryTextColor: "000000",
	}

	if err := config.SaveUIState(original); err != nil {
		t.Fatalf("SaveUIState() returned error: %v", err)
	}

	loaded, err := config.LoadUIState()
	if err != nil {
		t.Fatalf("LoadUIState() returned error: %v", err)
	}
	if loaded == nil {
		t.Fatal("LoadUIState() returned nil state")
	}
	if loaded.View != original.View || loaded.TopicID != original.TopicID || loaded.ThreadYOffset != original.ThreadYOffset {
		t.Errorf("loaded UI state mismatch: got %+v want %+v", loaded, original)
	}
}

func TestClearUIState(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	if err := config.SaveUIState(&config.UIState{View: "feed"}); err != nil {
		t.Fatalf("SaveUIState() returned error: %v", err)
	}
	if err := config.ClearUIState(); err != nil {
		t.Fatalf("ClearUIState() returned error: %v", err)
	}

	loaded, err := config.LoadUIState()
	if err != nil {
		t.Fatalf("LoadUIState() returned error: %v", err)
	}
	if loaded != nil {
		t.Errorf("expected nil UI state after clear, got %+v", loaded)
	}
}
