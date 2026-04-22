# HTTP API Reference

The web server exposes a JSON REST API on `http://localhost:8080` (default). All request and response bodies use `application/json` unless noted.

## Table of Contents

- [`POST /api/speak`](#post-apispeak)
- [`POST /api/ocr`](#post-apiocr)
- [`GET /api/voices`](#get-apivoices)
- [`GET /api/models`](#get-apimodels)
- [Playing Opus audio in the browser](#playing-opus-audio-in-the-browser)

---

## `POST /api/speak`

Synthesize text to Ogg Opus audio.

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

Extract text from an uploaded image using Tesseract OCR.

**Request**

`Content-Type: multipart/form-data`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `file` | file | yes | Image file (PNG, JPEG, WebP, TIFF, or any Tesseract-supported format) |

**Response `200 OK`**

```json
{
  "text": "Extracted text from the image..."
}
```

Returns an empty string in `text` if no text was detected.

**Errors**

| Status | Body | When |
|--------|------|------|
| `400` | `{"error": "file is required"}` | No file field in form |
| `400` | `{"error": "invalid multipart form"}` | Malformed form data |
| `500` | `{"error": "..."}` | Tesseract not installed or other error |

**curl example**

```sh
# Extract text only
curl -s -X POST http://localhost:8080/api/ocr \
  -F "file=@screenshot.png" \
  | jq -r '.text'

# Extract text then pipe into /api/speak
TEXT=$(curl -s -X POST http://localhost:8080/api/ocr \
  -F "file=@screenshot.png" \
  | jq -r '.text')

curl -s -X POST http://localhost:8080/api/speak \
  -H 'Content-Type: application/json' \
  -d "{\"text\": \"$TEXT\", \"voice\": \"Kore\"}" \
  | jq -r '.opus' \
  | base64 -d > output.opus
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
