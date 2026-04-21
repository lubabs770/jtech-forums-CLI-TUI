```

   _ _            _        __                                
  (_) |_ ___  ___| |__    / _| ___  _ __ _   _ _ __ ___  ___ 
  | | __/ _ \/ __| '_ \  | |_ / _ \| '__| | | | '_ ` _ \/ __|
  | | ||  __/ (__| | | | |  _| (_) | |  | |_| | | | | | \__ \
 _/ |\__\___|\___|_| |_| |_|  \___/|_|   \__,_|_| |_| |_|___/
|__/                                                         

```

# jtech-forums

A terminal UI for browsing and posting on a Discourse-powered forum, written
in Go with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## What it does

- Log in with your forum username and password (stores a session cookie under
  `~/.config/jtech/`)
- Browse the `latest`, `new`, `top`, `unseen`, and `categories` feeds
- Open a thread, read posts rendered with [Glamour](https://github.com/charmbracelet/glamour)
  Markdown, and reply using your `$EDITOR`
- Create new topics in any category
- Developer mode with restart-friendly UI resume via `--dev`

## Install
...

## Configuration

On first run you'll be prompted to log in. Config lives at
`~/.config/jtech-tui/config.json` and stores the base URL, default feed, and
session cookie. Developer mode stores resumable UI state at
`~/.config/jtech-tui/ui-state.json`.

## Runtime Flags

- `--feed <name>`: start on `latest`, `new`, `top`, `unseen`, or `categories`
- `--dev`: enable developer mode, UI-state resume, and a debug status footer
- `--no-alt-screen`: run without the terminal alternate screen

## Keybindings

### Feed list
| Key | Action |
| --- | --- |
| `enter` | Open selected topic |
| `n` | New topic |
| `tab` / `shift+tab` | Next / previous feed |
| `j` / `k` | Move down / up |
| `h` | Back |

### Thread view
| Key | Action |
| --- | --- |
| `j` / `k` | Scroll |
| `gg` | Jump to top |
| `r` | Reply (opens `$EDITOR`) |
| `h` | Back |

### Login / new topic form
| Key | Action |
| --- | --- |
| `tab` / `shift+tab` | Switch fields |
| `enter` | Submit / advance |
| `esc` | Cancel (new topic) |
| `ctrl+c` | Quit |

## Development

```sh
go test ./...
go build ./...
go run ./cmd/jtech --dev
```

Project layout:

- `cmd/jtech/` — entry point
- `internal/api/` — Discourse HTTP client
- `internal/config/` — config load/save
- `internal/ui/` — Bubble Tea views (login, feed, thread, categories, new
  topic)
