# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```sh
# Dependencies (macOS)
brew install opus opusfile mpv

# Build
go build -o inti .

# Run web server
./inti serve                # http://localhost:8282
./inti serve --port 3000

# Run interactive TUI
./inti

# One-shot CLI
./inti speak "Hello, world!"
./inti speak --voice Puck --model gemini-2.5-pro-preview-tts "Hello"
./inti speak --export out.opus --play "Hello"
```

## Environment

```sh
cp .env.example .env
# Set GEMINI_API_KEY (required)
# Optional: DEFAULT_VOICE, DEFAULT_MODEL, PORT, HOST
```

## Architecture

Single Go binary with web assets embedded via `go:embed` (see `embed.go`).

**Entry flow:** `main.go` → `cmd/` (Cobra CLI) → three subcommands:
- `serve` — starts HTTP server (`internal/server/`)
- `speak` — one-shot TTS to stdout/file/player (`cmd/speak.go`)
- default (no subcommand) — launches Bubble Tea TUI (`internal/tui/`)

**Key packages:**
- `internal/gemini/` — wraps `google.golang.org/genai`, calls the Gemini TTS API, detects rate-limit errors
- `internal/audio/` — encodes PCM → Ogg Opus via `github.com/hraban/opus` (CGo, requires libopus), and invokes `mpv`/`ffplay`/`vlc` for playback
- `internal/tui/` — Bubble Tea MVC: `model.go` (state), `update.go` (Elm-style messages), `view.go` (render), `messages.go` (custom Msg types)
- `internal/server/` — `net/http` server; `handlers.go` exposes `POST /api/speak`, `GET /api/voices`, `GET /api/models`
- `internal/config/` — loads `.env` + env vars, exposes typed config struct

**Audio pipeline (TUI/CLI):** Gemini API returns raw PCM → `internal/audio/opus.go` wraps it in Ogg Opus container → written to temp file or exported path → player invoked via `exec.Command`.

**Web frontend** (`web/`): vanilla HTML/CSS/JS, embedded into the binary. Communicates with the HTTP API only.

## API

```
POST /api/speak   { "text": "...", "voice": "Kore", "model": "gemini-2.5-flash-preview-tts" }
                  → { "opus": "<base64 Ogg Opus>" }  |  429 on rate limit
GET  /api/voices  → { "voices": [...], "default": "Kore" }
GET  /api/models  → { "models": [...], "default": "gemini-2.5-flash-preview-tts" }
```
