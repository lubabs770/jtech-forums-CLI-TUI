package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sam/jtech-tui/internal/api"
)

func TestLogin_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/session" || r.Method != http.MethodPost {
			http.Error(w, "not found", 404)
			return
		}
		http.SetCookie(w, &http.Cookie{Name: "_t", Value: "session-token-xyz"})
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"user": map[string]any{"username": "testuser"}})
	}))
	defer srv.Close()

	client, err := api.New(srv.URL, "")
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	cookie, err := client.Login("testuser", "password123")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if cookie != "session-token-xyz" {
		t.Errorf("expected cookie 'session-token-xyz', got %q", cookie)
	}
}

func TestLogin_BadCredentials(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{"error": "invalid credentials"})
	}))
	defer srv.Close()

	client, _ := api.New(srv.URL, "")
	_, err := client.Login("bad", "wrong")
	if err == nil {
		t.Error("expected error for bad credentials")
	}
}

func TestNew_WithCookie(t *testing.T) {
	var gotCookie string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCookie = r.Header.Get("Cookie")
		json.NewEncoder(w).Encode(topicListFixture())
	}))
	defer srv.Close()

	client, _ := api.New(srv.URL, "existing-cookie")
	client.GetFeed("latest")
	if gotCookie == "" {
		t.Error("expected cookie to be sent")
	}
}

func topicListFixture() map[string]any {
	return map[string]any{
		"topic_list": map[string]any{
			"topics": []map[string]any{
				{"id": 1, "title": "Hello", "slug": "hello", "posts_count": 3, "reply_count": 2},
			},
		},
	}
}
