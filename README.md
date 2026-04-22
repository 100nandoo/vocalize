# Vocalize

Text-to-speech powered by Google Gemini, with a modern web UI and an interactive terminal — all in a single Go binary.

## Table of Contents

- [Features](#features)
- [Documentation](#documentation)
- [Setup](#setup)
- [Usage](#usage)
  - [Web server](#web-server)
  - [One-shot CLI](#one-shot-cli)
  - [Interactive TUI](#interactive-tui)
- [API](#api)
- [Models](#models)
- [Voices](#voices)
- [Configuration](#configuration)
- [Project structure](#project-structure)
- [Requirements](#requirements)

## Features

- **Web UI** — dark interface with model & voice dropdowns, gender filter, waveform indicator, and Opus download
- **Image OCR** — drag-and-drop or browse to upload an image; extracted text appears in a copyable box and can be synthesized in one click
- **Synthesis metadata** — activity feed shows word count and how long each synthesis took
- **Interactive TUI** — Bubble Tea terminal UI with scrollable history and a command menu
- **One-shot CLI** — pipe-friendly `speak` and `ocr` subcommands for scripts and automation
- **PDF converter** — convert PDF pages to numbered PNG images with the `pdf` subcommand
- **Single binary** — web assets embedded via `go:embed`, no separate file serving
- **Rate limit handling** — quota errors surface as a friendly message instead of a raw API error

## Documentation

- [CLI reference](docs/cli.md) — all subcommands, flags, and examples
- [API reference](docs/api.md) — HTTP endpoints, request/response schemas, and curl examples

## Setup

```sh
cp .env.example .env
# add your GEMINI_API_KEY to .env
```

```sh
go build -o vocalize .
```

## Usage

### Web server

```sh
./vocalize serve
# Open http://localhost:8080
```

Choose a **model** and **voice** from the dropdowns, type your text, and hit **Synthesize**. Download the result with the **Download** button.

To use OCR, drop or browse an image in the **Image OCR** card at the top. The extracted text appears in a copyable box — click **Synthesize** to convert it to speech.

Flags: `--port 3000`, `--host 0.0.0.0`

### One-shot CLI

```sh
# Synthesize text
./vocalize speak "Hello, world!"
./vocalize speak --voice Puck --export hello.opus "Hello, world!"

# OCR — extract text from an image
./vocalize ocr screenshot.png

# OCR then synthesize
./vocalize ocr --speak invoice.jpg
./vocalize ocr --speak --export invoice.opus invoice.jpg
```

See [docs/cli.md](docs/cli.md) for the full flag reference.

### Interactive TUI

```sh
./vocalize
```

Press **Enter** on an empty prompt to open the command menu. Navigate with **↑ ↓**, select with **Enter**, dismiss with **Esc**.

| Command          | Description              |
| ---------------- | ------------------------ |
| `speak <text>`   | Synthesize and play      |
| `voice <name>`   | Switch voice             |
| `model <name>`   | Switch TTS model         |
| `export [path]`  | Save last audio as Opus  |
| `status`         | Show current config      |
| `clear`          | Clear the history        |
| `help`           | List commands            |
| `q` / `Ctrl+C`   | Quit                     |

## API

```
POST /api/speak   { "text": "...", "voice": "Kore", "model": "..." }
                  → { "opus": "<base64 Ogg Opus>" }

POST /api/ocr     multipart/form-data  file=<image>
                  → { "text": "..." }

GET  /api/voices  → { "voices": [...], "default": "Kore" }
GET  /api/models  → { "models": [...], "default": "..." }
```

See [docs/api.md](docs/api.md) for the full reference with curl examples.

## Models

| Model                          | Notes            |
| ------------------------------ | ---------------- |
| `gemini-2.5-flash-preview-tts` | Fast             |
| `gemini-2.5-pro-preview-tts`   | Higher quality   |
| `gemini-3.1-flash-tts-preview` | Latest preview   |

## Voices

30 voices available, filterable by gender in the web UI:

| Voice                | Style         | Voice         | Style       |
| -------------------- | ------------- | ------------- | ----------- |
| **Kore** _(default)_ | Firm          | Zephyr        | Bright      |
| Puck                 | Upbeat        | Charon        | Informative |
| Fenrir               | Excitable     | Leda          | Youthful    |
| Orus                 | Firm          | Aoede         | Breezy      |
| Callirrhoe           | Easy-going    | Autonoe       | Bright      |
| Enceladus            | Breathy       | Iapetus       | Clear       |
| Umbriel              | Easy-going    | Algieba       | Smooth      |
| Despina              | Smooth        | Erinome       | Clear       |
| Algenib              | Gravelly      | Rasalgethi    | Informative |
| Laomedeia            | Upbeat        | Achernar      | Soft        |
| Alnilam              | Firm          | Schedar       | Even        |
| Gacrux               | Mature        | Pulcherrima   | Forward     |
| Achird               | Friendly      | Zubenelgenubi | Casual      |
| Vindemiatrix         | Gentle        | Sadachbia     | Lively      |
| Sadaltager           | Knowledgeable | Sulafat       | Warm        |

## Configuration

| Variable         | Default                        | Description                        |
| ---------------- | ------------------------------ | ---------------------------------- |
| `GEMINI_API_KEY` | —                              | Required. Your Gemini API key      |
| `DEFAULT_VOICE`  | `Kore`                         | Default voice name                 |
| `DEFAULT_MODEL`  | `gemini-2.5-flash-preview-tts` | Default TTS model                  |
| `PORT`           | `8080`                         | Web server port                    |
| `HOST`           | `127.0.0.1`                    | Web server bind address            |

## Project structure

```
├── main.go                    # Entry point
├── embed.go                   # Embeds web/ into binary
├── cmd/                       # CLI commands (root, speak, serve, ocr, pdf)
├── docs/                      # API and CLI documentation
├── internal/
│   ├── config/                # Env/config loading and validation
│   ├── gemini/                # Gemini TTS client + rate-limit detection
│   ├── audio/                 # Opus encoder (Ogg container), platform audio player
│   ├── tui/                   # Bubble Tea TUI (model, view, update)
│   ├── ocr/                   # Tesseract OCR wrapper
│   ├── pdf/                   # PDF-to-image converter (go-fitz/MuPDF)
│   └── server/                # HTTP server + REST handlers
└── web/                       # Embedded frontend (HTML/CSS/JS)
```

## Requirements

- Go 1.22+
- `libopus` and `libopusfile` (for building): `brew install opus opusfile` / `apt install libopus-dev libopusfile-dev`
- `mupdf` (for PDF conversion): `brew install mupdf` / `apt install libmupdf-dev`
- `tesseract` (for OCR): `brew install tesseract` / `apt install tesseract-ocr`
- An Opus-capable audio player for the CLI/TUI `speak` and `export` commands: `mpv`, `ffplay`, or `vlc`
  - macOS: `brew install mpv`
  - Linux: `apt install mpv` or `apt install ffmpeg`
