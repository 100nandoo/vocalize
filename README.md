# Inti

Text-to-speech powered by Gemini and `kokoro heart`, with an OCR and summarization workspace, a browser extension, and scriptable CLI/server workflows in a single Go binary.

## Table of Contents

- [Features](#features)
- [Documentation](#documentation)
- [Setup](#setup)
- [Usage](#usage)
  - [Web server](#web-server)
  - [Browser extension usage](#browser-extension-usage)
  - [One-shot CLI](#one-shot-cli)
- [API](#api)
- [Models](#models)
- [Voices](#voices)
- [Configuration](#configuration)
- [Deploying publicly](#deploying-publicly)
- [Project structure](#project-structure)
- [Requirements](#requirements)
- [Browser Extension](#browser-extension)
  - [Extension Install](#extension-install)
  - [Extension Setup](#extension-setup)
  - [Extension Usage](#extension-usage)
  - [Extension Keyboard Shortcuts](#extension-keyboard-shortcuts)
- [Star History](#star-history)

## Features

- **Web UI** — dark-first text workspace with an explicit light/dark Visual Theme, organized into Import/OCR, Text Workspace, Text to Speech, and Activity panels
- **Image OCR** — drag-and-drop or browse to upload images (multi-file supported); extracted text lands in an editable OCR output and can be sent into the workspace or TTS flow
- **Summarizer** — summarize text with Gemini, Groq (free tier), or OpenRouter (free models); results rendered as Markdown with copy, speak, and split-button download actions for `.txt` or `.md`; provider and API keys configurable in the Settings page without restarting the server
- **Speech providers** — generate speech with Gemini or `kokoro heart`, while keeping Inti's exported/downloaded speech contract normalized to Opus
- **Browser extension** — summarize article pages directly in Chrome desktop, Firefox desktop, and Firefox Android via the bundled `extension/` app
- **Synthesis metadata** — activity feed shows word count, duration, voice, model, and summarizer model used
- **API key authentication** — protect the server with a main key and issue per-user API keys via the built-in `/api-keys.html` management page
- **One-shot CLI** — pipe-friendly `speak`, `summarize`, and `ocr` subcommands for scripts and automation
- **PDF converter** — convert PDF pages to numbered PNG images with the `pdf` subcommand
- **Single binary** — web assets embedded via `go:embed`, no separate file serving
- **Rate limit handling** — quota errors surface as a friendly message instead of a raw API error

## Documentation

- [CLI reference](docs/cli.md) — all subcommands, flags, and examples
- [API reference](docs/api.md) — HTTP endpoints, request/response schemas, and curl examples
- [Configuration reference](docs/config.md) — all env vars, config file locations, and API key setup
- [Dev loop contract](docs/dev-loop.md) — the single-command embedded web development workflow and validation matrix
- [Extension contributing guide](docs/contributing.md) — development, build, packaging, and architecture notes for the browser extension in `extension/`

## Setup

```sh
cp .env.example .env
# Edit .env — GEMINI_API_KEY is required for Gemini TTS.
# Set SPEECH_PROVIDER=kokoro-heart to use kokoro heart instead.
# GROQ_API_KEY or OPENROUTER_API_KEY is enough for summarization-only use.
```

```sh
go build -o inti .
```

```sh
npm install
```

For local development with auto-rebuild and restart on Go or embedded web asset changes:

```sh
go install github.com/air-verse/air@latest
npm install
make dev
```

## Usage

### Web server

```sh
./inti serve
# Open http://localhost:8282
```

If `TELEGRAM_BOT_TOKEN` is set, the same `serve` process also starts the Telegram bot automatically. Without that env var, `serve` keeps the current web-only behavior.

During development, `make dev` is the single local entrypoint. It first verifies that `air` and the frontend dependencies are installed, refreshes the embedded `web/` assets once, then runs the persistent `npm run watch:web` loop beside [Air](https://github.com/air-verse/air). Air owns Go rebuilds and `serve` restarts, while raw `web-src/` edits belong to the web watcher and only trigger a restart after the generated embedded assets in `web/` are updated successfully.

`web-src/` is the handwritten source tree for the web app, including `web-src/public/` for static assets copied into the build. `web/` is the embedded output directory and should be treated as generated output only. The web test suite lives under `tests/web/`. Do not add handwritten product logic under `web/` or `web/assets/`.

For the Svelte-based web source in `web-src/`, run `npm run typecheck:web` to validate the current `.svelte` surface and any typed web modules before building. This is the enforced baseline for the ongoing web TypeScript migration work: `#31` wires Svelte-aware checking into normal validation, and later migration slices can tighten types file-by-file without blocking on the older untyped JavaScript helpers.

To rebuild the embedded web assets continuously without starting the full dev loop, run `npm run watch:web`. It watches the embedded web build plus the generated unauthorized page and stays running across rebuild failures so the next successful save recovers automatically.

The web UI is split into four panels:

- **Import / OCR** — stage one or more images, extract text, and edit the OCR output.
- **Text Workspace** — paste or import OCR text, choose a summarizer provider/model, and generate a Markdown summary.
- **Text to Speech** — choose a speech provider, provider-specific voice and model options, then generate speech with optional auto-play or download.
- **Activity** — review recent OCR, summarization, synthesis, and download events.

To use OCR, drop or browse images in **Import / OCR**. The extracted text appears in **OCR Output**, is copied into **Text Workspace**, and is also available in **Text to Speech** for direct synthesis.

When a summary is shown, the summary action row lets you **Copy**, **Use Summary for TTS**, or **Download** it. The **Download** control is a split button: the main action downloads plain text (`.txt`), and the menu lets you choose Markdown (`.md`). Summary files use human-readable filenames such as `inti-summary-2026-04-29.txt`.

To configure the summarizer provider and API key, click **Settings** in the top-right corner. To manage API keys for access control, click **API Keys**.

The web UI paints **dark** on first load. Use the **Light** / **Dark** toggle beside **Settings** to switch themes locally. To make a theme global for the web UI, open **Settings**, choose an Appearance theme, and save it to the server config.

Flags: `--port 3000`, `--host 0.0.0.0`

The same summarization backend can also be used by the browser extension in [`extension/`](extension/) for article-page summaries inside Chrome and Firefox.

### Browser extension usage

The browser extension uses your Inti summarization API to summarize the current article page:

- Toolbar: click the Inti toolbar button to open the extension UI and start a summary for the current page.
- Context menu: right-click the page and choose **Summarize Page with Inti** to start the same summary flow.
- Side UI: use the Chrome side panel or Firefox sidebar to summarize the active page from the persistent extension UI.

### One-shot CLI

```sh
# Synthesize text
./inti speak "Hello, world!"
./inti speak --voice Puck --export hello.opus "Hello, world!"
./inti speak --provider kokoro-heart --voice cheery --export hello.opus "Hello, world!"

# Summarize text
./inti summarize "Long article text..."
./inti summarize --provider groq --api-key gsk_... "Long article text..."

# OCR — extract text from an image
./inti ocr screenshot.png

# OCR then synthesize
./inti ocr --speak invoice.jpg
./inti ocr --speak --export invoice.opus invoice.jpg
```

See [docs/cli.md](docs/cli.md) for the full flag reference.

## API

```
POST /api/speak              { "text": "...", "provider"?: "gemini" | "kokoro-heart", "voice": "...", "model"?: "..." }
                             → { "opus": "<base64 Ogg Opus>" }

POST /api/ocr                multipart/form-data  files=<image(s)>
                             → { "text": "..." }

POST /api/summarize          { "text": "...", "instruction"?, "provider"?, "apiKey"? }
                             → { "summary": "...", "provider": "...", "model": "..." }

GET  /api/summarizer-config  → { "provider": "...", "model": "..." }
GET  /api/theme-config       → { "theme": "light" | "dark", ... }
POST /api/theme-config       { "theme": "light" | "dark", ... }
GET  /api/voices             → { "provider": "...", "voices": [...], "default": "..." }
GET  /api/models             → { "provider": "...", "models": [...], "default": "..." }

GET    /api/admin/keys       → { "keys": [...] }
POST   /api/admin/keys       { "name": "..." } → { "key": {...}, "raw": "inti_..." }
DELETE /api/admin/keys/{id}  → 204
```

All `/api/*` endpoints require an `X-API-Key` header once at least one key exists. See [Deploying publicly](#deploying-publicly).

See [docs/api.md](docs/api.md) for the full reference with curl examples.

## Models

Gemini speech models:

| Model                          | Notes            |
| ------------------------------ | ---------------- |
| `gemini-2.5-flash-preview-tts` | Fast             |
| `gemini-2.5-pro-preview-tts`   | Higher quality   |
| `gemini-3.1-flash-tts-preview` | Latest preview   |

`kokoro heart` does not expose model selection in Inti.

## Voices

Gemini voices:

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

`kokoro heart` currently exposes:

| Voice | Notes |
|-------|-------|
| `cheery` | Provider-specific upstream voice |

## Configuration

| Variable              | Default                        | Description                                                    |
| --------------------- | ------------------------------ | -------------------------------------------------------------- |
| `GEMINI_API_KEY`      | —                              | Required for TTS and Gemini summarization                      |
| `SPEECH_PROVIDER`     | `gemini`                       | Active speech provider: `gemini` or `kokoro-heart`             |
| `KOKORO_HEART_URL`    | upstream default               | Override the `kokoro heart` speech endpoint URL                |
| `INTI_MAIN_KEY`       | —                              | Main key for API authentication (recommended for public deployment). `INTI_MASTER_KEY` is still accepted as a fallback |
| `DEFAULT_VOICE`       | provider-specific              | Default voice name for the selected speech provider            |
| `DEFAULT_MODEL`       | `gemini-3.1-flash-tts-preview` | Default speech model; used by Gemini only                      |
| `PORT`                | `8282`                         | Web server port                                                |
| `HOST`                | `127.0.0.1`                    | Web server bind address                                        |
| `INTI_CONFIG_DIR`     | OS default                     | Override config/key storage directory                          |
| `SUMMARIZER_PROVIDER` | auto-detected                  | Summarizer provider: `gemini`, `groq`, or `openrouter`         |
| `GROQ_API_KEY`        | —                              | Required when provider is `groq`                               |
| `GROQ_MODEL`          | `openai/gpt-oss-120b`          | Groq model to use                                              |
| `OPENROUTER_API_KEY`  | —                              | Required when provider is `openrouter`                         |
| `OPENROUTER_MODEL`    | `openrouter/free`              | OpenRouter model/router to use; `openrouter/free` tracks currently available free models |

`SUMMARIZER_PROVIDER` is auto-detected if not set: uses `gemini` if `GEMINI_API_KEY` is present, then `groq` if `GROQ_API_KEY` is present, then `openrouter` if `OPENROUTER_API_KEY` is present.

`kokoro heart` is an experimental speech provider backed by a public upstream service. Inti documents and supports it intentionally, but accepts that the upstream may break or change without notice.

## Credits

`kokoro heart` support relies on [KoboldCpp](https://github.com/LostRuins/koboldcpp) and [Kokoro](https://github.com/hexgrad/kokoro). Credit to the KoboldCpp and Kokoro authors.

See [docs/config.md](docs/config.md) for the full reference including config file locations per OS.

## Deploying publicly

When exposing Inti via Cloudflare Tunnel or any public URL, set a main key to lock down the API:

1. Generate a secret:
   ```sh
   openssl rand -hex 32
   # or
   python3 -c "import secrets; print(secrets.token_hex(32))"
   ```
2. Add it to `.env`:
   ```sh
   INTI_MAIN_KEY=<generated secret>
   HOST=0.0.0.0
   ```
3. Open `http://your-host/api-keys.html?key=<your main key>` and create per-user API keys to share with others.

All API requests must then include the key in the header:
```sh
curl -s http://your-host/api/voices -H 'X-API-Key: inti_...'
```

The web UI expects the key in the page URL as `?key=...`.

## Project structure

```
├── main.go                    # Entry point
├── embed.go                   # Embeds web/ into binary
├── cmd/                       # CLI commands (root, speak, summarize, serve, ocr, pdf)
├── docs/                      # Main app docs plus browser extension contributor documentation
├── extension/                 # Browser extension source, manifests, and build tooling
├── internal/
│   ├── config/                # Env/config loading and validation
│   ├── gemini/                # Gemini TTS + summarization client
│   ├── summarizer/            # Summarizer interface + Groq and OpenRouter clients
│   ├── audio/                 # Opus encoder (Ogg container), platform audio player
│   ├── ocr/                   # Tesseract OCR wrapper
│   ├── pdf/                   # PDF-to-image converter (go-fitz/MuPDF)
│   └── server/                # HTTP server, REST handlers, API key auth middleware
└── web/                       # Embedded frontend (HTML/CSS/JS, shared theme script, logo SVG, settings, API keys page)
```

## Requirements

- Go 1.22+
- `libopus` and `libopusfile` (for building): `brew install opus opusfile` / `apt install libopus-dev libopusfile-dev`
- `mupdf` (for PDF conversion): `brew install mupdf` / `apt install libmupdf-dev`
- `tesseract` (for OCR): `brew install tesseract` / `apt install tesseract-ocr`
- An Opus-capable audio player for the CLI `speak` and `ocr --speak` commands: `mpv`, `ffplay`, or `vlc`
  - macOS: `brew install mpv`
  - Linux: `apt install mpv` or `apt install ffmpeg`

## Browser Extension

This repository also includes a browser extension under [`extension/`](extension/) that summarizes article pages with AI and targets Chrome desktop, Firefox desktop, and Firefox Android.

- Extension contributor guide: [`docs/contributing.md`](docs/contributing.md)

### Extension Install

Download the latest extension release from the repository [Releases](../../releases) page, then load the unpacked build for your browser.

#### Chrome

1. Open `chrome://extensions`
2. Enable **Developer mode**
3. Extract `chrome.zip` and select the extracted extension folder containing `manifest.json`

#### Firefox Desktop

1. Open `about:debugging#/runtime/this-firefox`
2. Click **Load Temporary Add-on**
3. Extract `firefox.zip` and select its `manifest.json`

#### Firefox Android

1. Install [Firefox for Android](https://www.mozilla.org/firefox/android/)
2. Open **Settings -> About Firefox** and tap the version number five times to enable debug mode
3. On desktop Firefox, open `about:debugging#/setup` and connect the Android device over USB
4. Select the device and click **Load Temporary Add-on**
5. Extract `firefox.zip` and select its `manifest.json`

### Extension Setup

1. Deploy your summarization API and note its base URL.
2. Open Inti extension settings and save the API URL.

The extension posts article data to your configured endpoint and stores its state in extension storage (`chrome.storage.local` / `browser.storage.local`), not page `localStorage`.

Extension settings surfaces:

- The full Options page manages `apiUrl`, optional summarization `instruction`, and `theme`.
- The popup/sidebar settings panel manages `apiUrl`, optional `apiKey`, and `theme`.
- When `apiKey` is set, the extension sends it as the `X-API-Key` header on summarization requests.

### Extension Usage

Click the Inti extension action on an article page and run **Summarize Article**. The extension extracts the page content, sends it to your API, and renders the summary in the platform-specific UI:

- Chrome desktop: side panel
- Firefox desktop: sidebar
- Firefox Android: page overlay

The last summary and saved settings are restored the next time you open the extension.

### Extension Keyboard Shortcuts

- `Alt + Shift + S` (`Cmd + Shift + S` on Mac): summarize the current page
- `Ctrl + Shift + Y` (`Cmd + Shift + Y` on Mac): toggle the Firefox desktop sidebar

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=100nandoo/inti&type=Date)](https://star-history.com/#100nandoo/inti&Date)
