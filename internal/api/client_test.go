package api_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/sam/jtech-tui/internal/api"
)

func TestLogin_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/session/csrf.json":
			json.NewEncoder(w).Encode(map[string]any{"csrf": "test-csrf-token"})
		case r.URL.Path == "/session" && r.Method == http.MethodPost:
			if r.Header.Get("X-CSRF-Token") != "test-csrf-token" {
				http.Error(w, "missing csrf", 403)
				return
			}
			http.SetCookie(w, &http.Cookie{Name: "_t", Value: "session-token-xyz"})
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"user": map[string]any{"username": "testuser"}})
		default:
			http.Error(w, "not found", 404)
		}
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
					{
						"id":                 5,
						"name":               "General",
						"slug":               "general",
						"topic_count":        10,
						"color":              "00b3ff",
						"text_color":         "000000",
						"style_type":         "icon",
						"icon":               "star-half-stroke",
						"emoji":              "telephone_receiver",
						"parent_category_id": 4,
					},
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
	if cats[0].Color != "00b3ff" || cats[0].TextColor != "000000" || cats[0].Icon != "star-half-stroke" || cats[0].Emoji != "telephone_receiver" || cats[0].ParentCategoryID != 4 {
		t.Errorf("expected category presentation fields to decode, got %+v", cats[0])
	}
}

func TestGetThread(t *testing.T) {
	var requestedPostIDs []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/t/42/last.json":
			stream := make([]int, 35)
			for i := range stream {
				stream[i] = i + 1
			}
			lastPosts := make([]map[string]any, 0, 20)
			for i := 16; i <= 35; i++ {
				lastPosts = append(lastPosts, map[string]any{
					"id":          i,
					"post_number": i,
					"username":    "alice",
					"raw":         "Hello world",
					"created_at":  "2026-01-01",
				})
			}
			json.NewEncoder(w).Encode(map[string]any{
				"id":    42,
				"title": "Test thread",
				"post_stream": map[string]any{
					"stream": stream,
					"posts":  lastPosts,
				},
			})
		case "/t/42/posts.json":
			requestedPostIDs = append([]string(nil), r.URL.Query()["post_ids[]"]...)
			posts := make([]map[string]any, 0, len(requestedPostIDs))
			for _, id := range requestedPostIDs {
				n, _ := strconv.Atoi(id)
				posts = append(posts, map[string]any{
					"id":          n,
					"post_number": n,
					"username":    "alice",
					"raw":         "Hello world",
					"created_at":  "2026-01-01",
				})
			}
			json.NewEncoder(w).Encode(map[string]any{
				"post_stream": map[string]any{"posts": posts},
			})
		default:
			http.Error(w, "not found", 404)
		}
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
	if len(thread.PostStream.Posts) != 30 || thread.PostStream.Posts[0].PostNumber != 6 || thread.PostStream.Posts[29].PostNumber != 35 {
		t.Errorf("unexpected posts: %+v", thread.PostStream.Posts)
	}
	if len(requestedPostIDs) != 10 || requestedPostIDs[0] != "6" || requestedPostIDs[9] != "15" {
		t.Errorf("expected previous 10 post ids, got %v", requestedPostIDs)
	}
}

func TestGetThreadPosts(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/t/42/posts.json" {
			http.Error(w, "not found", 404)
			return
		}
		got := r.URL.Query()["post_ids[]"]
		want := []string{"91", "92", "93"}
		if len(got) != len(want) {
			t.Fatalf("expected %d ids, got %v", len(want), got)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("expected post_ids %v, got %v", want, got)
			}
		}
		json.NewEncoder(w).Encode(map[string]any{
			"post_stream": map[string]any{
				"posts": []map[string]any{
					{"id": 93, "post_number": 93, "username": "carol", "raw": "third", "created_at": "2026-01-03"},
					{"id": 91, "post_number": 91, "username": "alice", "raw": "first", "created_at": "2026-01-01"},
					{"id": 92, "post_number": 92, "username": "bob", "raw": "second", "created_at": "2026-01-02"},
				},
			},
		})
	}))
	defer srv.Close()

	client, _ := api.New(srv.URL, "tok")
	posts, err := client.GetThreadPosts(42, []int{91, 92, 93})
	if err != nil {
		t.Fatalf("GetThreadPosts: %v", err)
	}
	if len(posts) != 3 {
		t.Fatalf("expected 3 posts, got %d", len(posts))
	}
	for i, want := range []int{91, 92, 93} {
		if posts[i].PostNumber != want {
			t.Fatalf("expected posts sorted by post number, got %+v", posts)
		}
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

func TestPostReply(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/session/csrf.json" {
			json.NewEncoder(w).Encode(map[string]any{"csrf": "test-csrf"})
			return
		}
		if r.URL.Path != "/posts" || r.Method != http.MethodPost {
			http.Error(w, "not found", 404)
			return
		}
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{"id": 99})
	}))
	defer srv.Close()

	client, _ := api.New(srv.URL, "tok")
	err := client.PostReply(42, "My reply text")
	if err != nil {
		t.Fatalf("PostReply: %v", err)
	}
	if gotBody["topic_id"] != float64(42) {
		t.Errorf("expected topic_id 42, got %v", gotBody["topic_id"])
	}
	if gotBody["raw"] != "My reply text" {
		t.Errorf("expected raw 'My reply text', got %v", gotBody["raw"])
	}
}

func TestCreateTopic(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/session/csrf.json" {
			json.NewEncoder(w).Encode(map[string]any{"csrf": "test-csrf"})
			return
		}
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{"id": 100, "topic_id": 55})
	}))
	defer srv.Close()

	client, _ := api.New(srv.URL, "tok")
	err := client.CreateTopic("My new topic", "Body text here", 5)
	if err != nil {
		t.Fatalf("CreateTopic: %v", err)
	}
	if gotBody["title"] != "My new topic" {
		t.Errorf("expected title, got %v", gotBody["title"])
	}
	if gotBody["category"] != float64(5) {
		t.Errorf("expected category 5, got %v", gotBody["category"])
	}
}
