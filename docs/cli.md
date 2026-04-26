# CLI Reference

All subcommands share the same binary: `./inti`.

## Table of Contents

- [Global](#global)
- [`speak` — Synthesize text](#speak--synthesize-text)
- [`summarize` — Summarize text](#summarize--summarize-text)
- [`ocr` — Extract text from an image](#ocr--extract-text-from-an-image)
- [`serve` — Start the web server](#serve--start-the-web-server)
- [`pdf` — Convert PDF to images](#pdf--convert-pdf-to-images)
- [Interactive TUI](#interactive-tui)
- [Models](#models)
- [Voices](#voices)

## Global

```sh
./inti --help
./inti [command] --help
```

---

## `speak` — Synthesize text

```sh
./inti speak [flags] <text>
```

Synthesizes the given text and plays it. Exits when playback finishes. Requires `GEMINI_API_KEY`.

| Flag | Default | Description |
|------|---------|-------------|
| `--voice <name>` | `$DEFAULT_VOICE` | Voice name (see [Voices](#voices)) |
| `--model <name>` | `$DEFAULT_MODEL` | TTS model (see [Models](#models)) |
| `--export <path>` | — | Save audio to `.opus` file (skips playback unless `--play` is also set) |
| `--play` | `false` | Play audio even when `--export` is set |

**Examples**

```sh
# Basic
./inti speak "Hello, world!"

# Choose a voice
./inti speak --voice Puck "Hello, world!"

# Choose a model
./inti speak --model gemini-2.5-pro-preview-tts "Hello, world!"

# Save to file (no playback)
./inti speak --export hello.opus "Hello, world!"

# Save and play
./inti speak --export hello.opus --play "Hello, world!"
```

---

## `summarize` — Summarize text

```sh
./inti summarize [flags] <text>
```

Summarizes the given text using a configured AI provider and prints the result to stdout. Supports Gemini, Groq (free tier), and OpenRouter (free models). Does **not** require `GEMINI_API_KEY` unless the provider is set to `gemini`.

| Flag | Default | Description |
|------|---------|-------------|
| `--instruction <text>` | — | Custom summarization instruction. Defaults to a structured prompt with headers and bullet lists |
| `--provider <name>` | `$SUMMARIZER_PROVIDER` | Override provider: `gemini`, `groq`, or `openrouter` |
| `--api-key <key>` | env var | API key for the provider (overrides `GROQ_API_KEY` / `OPENROUTER_API_KEY`) |

The provider is auto-detected from env vars if `--provider` is not set: `GEMINI_API_KEY` → gemini, `GROQ_API_KEY` → groq, `OPENROUTER_API_KEY` → openrouter.

**Examples**

```sh
# Using the server-configured provider
./inti summarize "Go is a statically typed, compiled language..."

# Groq (free tier)
GROQ_API_KEY=gsk_... ./inti summarize "Go is a statically typed language..."

# OpenRouter (free models)
OPENROUTER_API_KEY=sk-or-... ./inti summarize "Go is a statically typed language..."

# Override provider and key inline
./inti summarize --provider groq --api-key gsk_... "Long article text..."

# Custom instruction
./inti summarize --instruction "Summarize in one sentence." "Long article text..."
```

---

## `ocr` — Extract text from an image

```sh
./inti ocr [flags] <image-path>
```

Runs Tesseract OCR on the image and prints the extracted text. Supports PNG, JPEG, WebP, TIFF, and any format Tesseract accepts.

Optionally synthesizes the extracted text with TTS using `--speak`.

| Flag | Default | Description |
|------|---------|-------------|
| `--speak` | `false` | Synthesize and play the extracted text |
| `--voice <name>` | `$DEFAULT_VOICE` | Voice name (only used with `--speak`) |
| `--model <name>` | `$DEFAULT_MODEL` | TTS model (only used with `--speak`) |
| `--export <path>` | — | Save audio to `.opus` file (only used with `--speak`) |
| `--play` | `false` | Play audio even when `--export` is set (only used with `--speak`) |

**Examples**

```sh
# Extract and print text
./inti ocr screenshot.png

# Extract then speak
./inti ocr --speak invoice.jpg

# Extract, speak, and save audio
./inti ocr --speak --export invoice.opus invoice.jpg

# Extract with a specific voice
./inti ocr --speak --voice Fenrir notes.png
```

---

## `serve` — Start the web server

```sh
./inti serve [flags]
```

Starts an HTTP server serving the web UI at `http://localhost:8282`. `GEMINI_API_KEY` is only required if you use TTS; summarization works with Groq or OpenRouter keys alone.

| Flag | Default | Description |
|------|---------|-------------|
| `--port <n>` | `8282` (or `$PORT`) | Port to listen on |
| `--host <addr>` | `127.0.0.1` (or `$HOST`) | Address to bind to |

**Examples**

```sh
./inti serve
./inti serve --port 3000
./inti serve --port 3000 --host 0.0.0.0

# Summarize-only mode (no TTS)
GROQ_API_KEY=gsk_... ./inti serve
```

---

## `pdf` — Convert PDF to images

```sh
./inti pdf [flags] <pdf-path>
```

Converts each page of the PDF to a numbered PNG image.

| Flag | Default | Description |
|------|---------|-------------|
| `--output <dir>` | `./` | Output directory for PNG files |

**Examples**

```sh
# Convert to current directory
./inti pdf report.pdf

# Specify output directory
./inti pdf report.pdf --output /tmp/pages
```

---

## Interactive TUI

Running `./inti` without any subcommand launches the Bubble Tea terminal UI. Requires `GEMINI_API_KEY`.

```sh
./inti
```

Press **Enter** on an empty prompt to open the command menu. Navigate with **↑ ↓**, select with **Enter**, dismiss with **Esc**. Use **↑ ↓** while typing to scroll the history.

| Command | Description |
|---------|-------------|
| `speak <text>` | Synthesize and play |
| `voice <name>` | Switch voice |
| `model <name>` | Switch TTS model |
| `export [path]` | Save last audio as Opus |
| `status` | Show current config |
| `clear` | Clear the history |
| `help` | List commands |
| `q` / `Ctrl+C` | Quit |

---

## Models

| Model | Notes |
|-------|-------|
| `gemini-2.5-flash-preview-tts` | Fast |
| `gemini-2.5-pro-preview-tts` | Higher quality |
| `gemini-3.1-flash-tts-preview` | Latest preview (default) |

## Voices

30 voices available:

| Voice | Gender | Style |
|-------|--------|-------|
| **Kore** _(default)_ | Female | Firm |
| Zephyr | Female | Bright |
| Puck | Male | Upbeat |
| Charon | Male | Informative |
| Fenrir | Male | Excitable |
| Leda | Female | Youthful |
| Orus | Male | Firm |
| Aoede | Female | Breezy |
| Callirrhoe | Female | Easy-going |
| Autonoe | Female | Bright |
| Enceladus | Male | Breathy |
| Iapetus | Male | Clear |
| Umbriel | Male | Easy-going |
| Algieba | Male | Smooth |
| Despina | Female | Smooth |
| Erinome | Female | Clear |
| Algenib | Male | Gravelly |
| Rasalgethi | Male | Informative |
| Laomedeia | Female | Upbeat |
| Achernar | Female | Soft |
| Alnilam | Male | Firm |
| Schedar | Male | Even |
| Gacrux | Female | Mature |
| Pulcherrima | Male | Forward |
| Achird | Male | Friendly |
| Zubenelgenubi | Male | Casual |
| Vindemiatrix | Female | Gentle |
| Sadachbia | Male | Lively |
| Sadaltager | Male | Knowledgeable |
| Sulafat | Female | Warm |
