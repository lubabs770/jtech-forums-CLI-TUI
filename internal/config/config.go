package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// Config holds the application configuration.
type Config struct {
	ForumURL      string `json:"forum_url"`
	DefaultFeed   string `json:"default_feed"`
	SessionCookie string `json:"session_cookie"`
}

// configPath returns the path to the config file using the current HOME directory.
func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "jtech-tui", "config.json"), nil
}

// Load reads the config from disk, returning defaults if the file doesn't exist.
func Load() (*Config, error) {
	cfg := &Config{
		ForumURL:    "https://forums.jtechforums.org",
		DefaultFeed: "latest",
	}

	path, err := configPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Save writes the config to disk, creating directories as needed.
func (c *Config) Save() error {
	path, err := configPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}
