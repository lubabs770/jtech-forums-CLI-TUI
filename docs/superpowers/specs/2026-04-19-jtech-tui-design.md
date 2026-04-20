# jtech TUI — Design Spec
# Date: 2026-04-19

## Overview

A terminal UI client for [forums.jtechforums.org](https://forums.jtechforums.org/) built in Go using Bubble Tea. Supports browsing forum feeds and threads, replying to topics, and creating new topics. Auth via username/password session cookie. Compose via `$EDITOR`.

---

## Architecture

**Language:** Go + Bubble Tea  
**Project layout:**

```
jtech-tui/
├── cmd/jtech/
│   └── main.go              # entry point, reads config, initialises app
├── internal/
│   ├── api/
│   │   └── client.go        # Discourse HTTP client + all API methods
│   ├── config/
│   │   └── config.go        # load/save ~/.config/jtech-tui/config.json
│   └── ui/
│       ├── app.go            # root Bubble Tea model, view stack
│       ├── login.go          # username/password form
│       ├── feed.go           # latest / new / top / unseen / categories feed
│       ├── categories.go     # category browser (feed=categories only)
│       └── thread.go         # thread reading view
└── go.mod
```

**Navigation model:** a `[]tea.Model` stack in `app.go`. `enter` pushes a new view; `h` pops back. No global router.

---

## Configuration

Stored at `~/.config/jtech-tui/config.json`:

```json
{
  "forum_url": "https://forums.jtechforums.org",
  "default_feed": "latest",
  "session_cookie": "<_t cookie value>"
}
```

`default_feed` can be `latest`, `new`, `top`, `unseen`, or `categories`. Overridable via CLI flag: `jtech --feed new`.

---

## Auth

1. On startup, load config. If `session_cookie` is present, attach it to all requests via `http.CookieJar`.
2. If missing or expired (any 403 response), push `LoginView` onto the stack.
3. `LoginView`: two fields — username and password. Submit with `enter`.
4. POST `https://forums.jtechforums.org/session` with `{"login": "...", "password": "..."}`.
5. On success, extract the `_t` cookie and persist it to config.
6. On 403 mid-session, clear the saved cookie and push `LoginView` — transparent re-auth, no crash.

---

## API Layer (`internal/api/client.go`)

All requests to `https://forums.jtechforums.org`. Session cookie attached via shared `http.CookieJar`.

```go
Login(username, password string) (http.Cookie, error)
GetFeed(feed string) ([]Topic, error)           // feed: latest|new|top|unseen
GetCategories() ([]Category, error)
GetCategoryTopics(slug string, id int) ([]Topic, error)
GetThread(id int) (Thread, error)
PostReply(topicID int, raw string) error
CreateTopic(title, raw string, categoryID int) error
```

**Endpoints used:**

| Method | Endpoint |
|---|---|
| POST | `/session` |
| GET | `/latest.json` |
| GET | `/new.json` |
| GET | `/top.json` |
| GET | `/unseen.json` |
| GET | `/categories.json` |
| GET | `/c/:slug/:id.json` |
| GET | `/t/:id.json` |
| POST | `/posts` |

---

## UI & Navigation

**View stack flow:**

```
LoginView  (if no session cookie)
    ↓
FeedView   (default_feed determines starting feed; tab switches between feeds)
    ↓ enter on category (feed=categories)
CategoryTopicsView
    ↓ enter on topic (from FeedView or CategoryTopicsView)
ThreadView
```

**Composing (launched from ThreadView or FeedView, not a separate view):**
- `r` from `ThreadView` → write temp file, suspend TUI, open `$EDITOR`, resume, POST reply
- `n` from `FeedView` → show a small overlay form (title field + category selector) rendered inside `FeedView`; on submit open `$EDITOR` for body, resume, POST topic
- Falls back to `nano` if `$EDITOR` is unset

**Keybindings:**

| Key | Action |
|---|---|
| `j` / `↓` | move down |
| `k` / `↑` | move up |
| `enter` | open selected item |
| `h` | go back (pop view stack) |
| `r` | reply to current topic |
| `n` | new topic |
| `tab` | cycle feed (latest → new → top → unseen → categories) — loads immediately |
| `ctrl+c` | quit |

---

## Error Handling

- Network errors: show inline error message in current view, allow retry
- 403: clear session, push LoginView
- 422 (post validation): show Discourse's error message inline
- `$EDITOR` exits non-zero or file is empty: abort post silently, return to thread

---

## Out of Scope (v1)

- Search (`/search.json`)
- Reactions
- Notifications
- User profiles
- Moderation actions
