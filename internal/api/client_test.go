package api_test

import (
	"encoding/json"
	"errors"
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

func TestGetFeed_Latest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/latest.json" {
			http.Error(w, "not found", 404)
			return
		}
		json.NewEncoder(w).Encode(topicListFixture())
	}))
	defer srv.Close()

	client, _ := api.New(srv.URL, "tok")
	topics, err := client.GetFeed("latest")
	if err != nil {
		t.Fatalf("GetFeed: %v", err)
	}
	if len(topics) != 1 || topics[0].Title != "Hello" {
		t.Errorf("unexpected topics: %+v", topics)
	}
}

func TestGetCategories(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"category_list": map[string]any{
				"categories": []map[string]any{
					{"id": 5, "name": "General", "slug": "general", "topic_count": 10},
				},
			},
		})
	}))
	defer srv.Close()

	client, _ := api.New(srv.URL, "tok")
	cats, err := client.GetCategories()
	if err != nil {
		t.Fatalf("GetCategories: %v", err)
	}
	if len(cats) != 1 || cats[0].Name != "General" {
		t.Errorf("unexpected categories: %+v", cats)
	}
}

func TestGetThread(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/t/42.json" {
			http.Error(w, "not found", 404)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{
			"id":    42,
			"title": "Test thread",
			"post_stream": map[string]any{
				"posts": []map[string]any{
					{"id": 1, "post_number": 1, "username": "alice", "raw": "Hello world", "created_at": "2026-01-01"},
				},
			},
		})
	}))
	defer srv.Close()

	client, _ := api.New(srv.URL, "tok")
	thread, err := client.GetThread(42)
	if err != nil {
		t.Fatalf("GetThread: %v", err)
	}
	if thread.Title != "Test thread" {
		t.Errorf("unexpected title: %q", thread.Title)
	}
	if len(thread.PostStream.Posts) != 1 || thread.PostStream.Posts[0].Username != "alice" {
		t.Errorf("unexpected posts: %+v", thread.PostStream.Posts)
	}
}

func TestGetFeed_Returns403(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	client, _ := api.New(srv.URL, "expired")
	_, err := client.GetFeed("latest")
	if err == nil {
		t.Error("expected error on 403")
	}
	var apiErr *api.ErrUnauthorized
	if !errors.As(err, &apiErr) {
		t.Errorf("expected ErrUnauthorized, got %T: %v", err, err)
	}
}
