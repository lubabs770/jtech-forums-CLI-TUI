package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// UIState stores resumable UI context for developer mode.
type UIState struct {
	View                    string `json:"view"`
	ThreadSource            string `json:"thread_source,omitempty"`
	Feed                    string `json:"feed,omitempty"`
	FeedIndex               int    `json:"feed_index,omitempty"`
	CategoriesIndex         int    `json:"categories_index,omitempty"`
	CategoryTopicsIndex     int    `json:"category_topics_index,omitempty"`
	CategoryID              int    `json:"category_id,omitempty"`
	CategorySlug            string `json:"category_slug,omitempty"`
	CategoryName            string `json:"category_name,omitempty"`
	CategoryColor           string `json:"category_color,omitempty"`
	CategoryTextColor       string `json:"category_text_color,omitempty"`
	ParentCategoryID        int    `json:"parent_category_id,omitempty"`
	ParentCategorySlug      string `json:"parent_category_slug,omitempty"`
	ParentCategoryName      string `json:"parent_category_name,omitempty"`
	ParentCategoryColor     string `json:"parent_category_color,omitempty"`
	ParentCategoryTextColor string `json:"parent_category_text_color,omitempty"`
	TopicID                 int    `json:"topic_id,omitempty"`
	TopicSlug               string `json:"topic_slug,omitempty"`
	TopicTitle              string `json:"topic_title,omitempty"`
	TopicCategoryID         int    `json:"topic_category_id,omitempty"`
	ThreadYOffset           int    `json:"thread_y_offset,omitempty"`
}

func uiStatePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "jtech-tui", "ui-state.json"), nil
}

// LoadUIState reads the persisted UI state. It returns nil if no state exists.
func LoadUIState() (*UIState, error) {
	path, err := uiStatePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var state UIState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

// SaveUIState writes the developer-mode UI state to disk.
func SaveUIState(state *UIState) error {
	if state == nil {
		return ClearUIState()
	}

	path, err := uiStatePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// ClearUIState removes any persisted UI state.
func ClearUIState() error {
	path, err := uiStatePath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
