# HTTP API Reference

The web server exposes a JSON REST API on `http://localhost:8080` (default). All request and response bodies use `application/json` unless noted.

## Table of Contents

- [`POST /api/speak`](#post-apispeak)
- [`POST /api/ocr`](#post-apiocr)
- [`POST /api/summarize`](#post-apisummarize)
- [`GET /api/summarizer-config`](#get-apisummarizer-config)
- [`GET /api/voices`](#get-apivoices)
- [`GET /api/models`](#get-apimodels)
- [Playing Opus audio in the browser](#playing-opus-audio-in-the-browser)

---

## `POST /api/speak`

Synthesize text to Ogg Opus audio. Requires `GEMINI_API_KEY`.

**Request body**

```json
{
  "text":  "Hello, world!",
  "voice": "Kore",
  "model": "gemini-2.5-flash-preview-tts"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `text` | string | yes | Text to synthesize |
| `voice` | string | no | Voice name. Defaults to `DEFAULT_VOICE` env var |
| `model` | string | no | TTS model. Defaults to `DEFAULT_MODEL` env var |

**Response `200 OK`**

```json
{
  "opus": "<base64-encoded Ogg Opus>"
}
```

Decode the `opus` field from base64 to get the raw Ogg Opus bytes (24 kHz · PCM-16 · mono).

**Errors**

| Status | Body | When |
|--------|------|------|
| `400` | `{"error": "text is required"}` | Empty text |
| `400` | `{"error": "invalid voice: ..."}` | Unknown voice name |
| `400` | `{"error": "invalid model: ..."}` | Unknown model name |
| `429` | `{"error": "rate limited — wait a moment and try again"}` | Gemini quota exceeded |
| `503` | `{"error": "TTS unavailable — GEMINI_API_KEY not configured"}` | Missing API key |
| `500` | `{"error": "..."}` | Unexpected server error |

**curl example**

```sh
curl -s -X POST http://localhost:8080/api/speak \
  -H 'Content-Type: application/json' \
  -d '{"text": "Hello, world!", "voice": "Kore", "model": "gemini-2.5-flash-preview-tts"}' \
  | jq -r '.opus' \
  | base64 -d > hello.opus

mpv hello.opus
```

---

## `POST /api/ocr`

Extract text from one or more uploaded images using Tesseract OCR.

**Request**

`Content-Type: multipart/form-data`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `files` | file[] | yes | One or more image files (PNG, JPEG, WebP, TIFF). Use `file` as a fallback for single-file requests |

**Response `200 OK`**

```json
{
  "text": "Extracted text from the image..."
}
```

When multiple files are uploaded, extracted text is joined with a blank line between each image. Returns an empty string in `text` if no text was detected.

**Errors**

| Status | Body | When |
|--------|------|------|
| `400` | `{"error": "at least one file is required"}` | No file field in form |
| `400` | `{"error": "invalid multipart form"}` | Malformed form data |
| `500` | `{"error": "..."}` | Tesseract not installed or other error |

**curl example**

```sh
# Single image
curl -s -X POST http://localhost:8080/api/ocr \
  -F "files=@screenshot.png" \
  | jq -r '.text'

# Multiple images
curl -s -X POST http://localhost:8080/api/ocr \
  -F "files=@page1.png" \
  -F "files=@page2.png" \
  | jq -r '.text'

# Extract text then pipe into /api/speak
TEXT=$(curl -s -X POST http://localhost:8080/api/ocr \
  -F "files=@screenshot.png" \
  | jq -r '.text')

curl -s -X POST http://localhost:8080/api/speak \
  -H 'Content-Type: application/json' \
  -d "{\"text\": \"$TEXT\", \"voice\": \"Kore\"}" \
  | jq -r '.opus' \
  | base64 -d > output.opus
```

---

## `POST /api/summarize`

Summarize text using a configured AI provider (Gemini, Groq, or OpenRouter). The server uses the provider set by `SUMMARIZER_PROVIDER`; the request can override this per-call via `provider` and `apiKey`.

**Request body**

```json
{
  "text":        "Long text to summarize...",
  "instruction": "Summarize in three bullet points.",
  "provider":    "groq",
  "apiKey":      "gsk_..."
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `text` | string | yes | Text to summarize |
| `instruction` | string | no | Custom summarization instruction. Defaults to a built-in prompt using headers and bullet lists |
| `provider` | string | no | Override the server provider: `gemini`, `groq`, or `openrouter` |
| `apiKey` | string | no | API key for the provider (used by the web UI Settings page; falls back to the server's env var) |

**Response `200 OK`**

```json
{
  "summary":  "## Key Points\n\n- ...",
  "provider": "groq",
  "model":    "llama-3.3-70b-versatile"
}
```

The `summary` field is Markdown-formatted text. `provider` and `model` reflect what was actually used.

**Errors**

| Status | Body | When |
|--------|------|------|
| `400` | `{"error": "text is required"}` | Empty text |
| `400` | `{"error": "..."}` | Unknown provider or missing API key for requested provider |
| `429` | `{"error": "rate limited — wait a moment and try again"}` | Provider quota exceeded |
| `503` | `{"error": "no summarizer configured — set a provider and API key"}` | No provider configured server-side and none supplied in the request |
| `500` | `{"error": "..."}` | Unexpected server error |

**curl examples**

```sh
# Using the server's configured provider
curl -s -X POST http://localhost:8080/api/summarize \
  -H 'Content-Type: application/json' \
  -d '{"text": "Go is a statically typed language..."}' \
  | jq -r '.summary'

# Override provider and key per-request
curl -s -X POST http://localhost:8080/api/summarize \
  -H 'Content-Type: application/json' \
  -d '{"text": "Go is a statically typed language...", "provider": "groq", "apiKey": "gsk_..."}' \
  | jq '{summary, provider, model}'

# Custom instruction
curl -s -X POST http://localhost:8080/api/summarize \
  -H 'Content-Type: application/json' \
  -d '{"text": "...", "instruction": "Summarize in one sentence."}' \
  | jq -r '.summary'
```

---

## `GET /api/summarizer-config`

Returns the server's current summarizer configuration. Does **not** expose the API key.

**Response `200 OK`**

```json
{
  "provider": "groq",
  "model":    "llama-3.3-70b-versatile"
}
```

Both fields are empty strings when no provider is configured server-side.

**curl example**

```sh
curl -s http://localhost:8080/api/summarizer-config | jq .
```

---

## `GET /api/voices`

List all available voices.

**Response `200 OK`**

```json
{
  "voices":  ["Achernar", "Achird", "Algenib", "..."],
  "default": "Kore"
}
```

**curl example**

```sh
curl -s http://localhost:8080/api/voices | jq '.voices[]'
```

---

## `GET /api/models`

List all available TTS models.

**Response `200 OK`**

```json
{
  "models":  ["gemini-2.5-flash-preview-tts", "gemini-2.5-pro-preview-tts", "..."],
  "default": "gemini-3.1-flash-tts-preview"
}
```

**curl example**

```sh
curl -s http://localhost:8080/api/models | jq '.models[]'
```

---

## Playing Opus audio in the browser

```js
const res  = await fetch('/api/speak', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ text: 'Hello', voice: 'Kore' }),
});
const { opus } = await res.json();
const bytes = Uint8Array.from(atob(opus), c => c.charCodeAt(0));
const blob  = new Blob([bytes], { type: 'audio/opus' });

const ctx = new AudioContext();
const buf = await ctx.decodeAudioData(await blob.arrayBuffer());
const src = ctx.createBufferSource();
src.buffer = buf;
src.connect(ctx.destination);
src.start();
```
