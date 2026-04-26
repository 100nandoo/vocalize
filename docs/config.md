# Configuration Reference

## Table of Contents

- [Environment variables](#environment-variables)
- [Config file location](#config-file-location)
- [API key management](#api-key-management)
- [Open the config folder](#open-the-config-folder)

---

## Environment variables

All configuration is done via environment variables, loaded from a `.env` file in the project root (or set directly in the shell).

Copy the example file to get started:

```sh
cp .env.example .env
```

### Full variable reference

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `GEMINI_API_KEY` | Yes (for TTS) | — | Google Gemini API key. Required for TTS and Gemini summarization |
| `INTI_MASTER_KEY` | No | — | Master key that always authenticates all API requests. Set this when deploying publicly. See [API key management](#api-key-management) |
| `DEFAULT_VOICE` | No | `Kore` | Default TTS voice. Must be one of the 30 valid voice names |
| `DEFAULT_MODEL` | No | `gemini-3.1-flash-tts-preview` | Default TTS model |
| `PORT` | No | `8282` | HTTP server port |
| `HOST` | No | `127.0.0.1` | HTTP server bind address. Set to `0.0.0.0` to listen on all interfaces |
| `INTI_CONFIG_DIR` | No | OS default | Override the directory where `inti.toml` is stored |
| `SUMMARIZER_PROVIDER` | No | auto-detected | Active summarizer: `gemini`, `groq`, or `openrouter`. Auto-detected from available API keys if not set |
| `GROQ_API_KEY` | No | — | Groq API key. Enables Groq as a summarizer provider |
| `GROQ_MODEL` | No | `llama-3.3-70b-versatile` | Groq model to use for summarization |
| `OPENROUTER_API_KEY` | No | — | OpenRouter API key. Enables OpenRouter as a summarizer provider |
| `OPENROUTER_MODEL` | No | `google/gemma-3-27b-it:free` | OpenRouter model to use for summarization |

### Example `.env`

```sh
GEMINI_API_KEY=AIza...

# Protect the API when deploying publicly
# Generate with: openssl rand -hex 32
#           or:  python3 -c "import secrets; print(secrets.token_hex(32))"
INTI_MASTER_KEY=change_me_to_a_strong_secret

# Optional overrides
# DEFAULT_VOICE=Puck
# DEFAULT_MODEL=gemini-2.5-flash-preview-tts
# PORT=8282
# HOST=0.0.0.0
```

---

## Config file location

Runtime settings changed via the web UI (summarizer provider, model) and API keys created via the API keys page are persisted to `inti.toml` on disk.

| OS | Default path |
|----|-------------|
| macOS | `~/Library/Application Support/inti/inti.toml` |
| Linux / Debian | `~/.config/inti/inti.toml` |
| Windows | `%APPDATA%\inti\inti.toml` |

Override the location by setting `INTI_CONFIG_DIR` in your `.env`:

```sh
INTI_CONFIG_DIR=/data/inti
```

---

## API key management

When deployed publicly (e.g. via Cloudflare Tunnel), set `INTI_MASTER_KEY` to protect all API endpoints.

**Authentication rules:**

- If `INTI_MASTER_KEY` is set — auth is always enforced. Every `/api/*` request must include a valid `X-API-Key` header.
- If `INTI_MASTER_KEY` is not set — the server runs in setup mode until the first API key is created via the web UI, at which point auth is enforced automatically.
- Static pages (`/`, `/api-keys.html`, etc.) are always publicly accessible regardless of auth.

**Bootstrapping with a master key:**

1. Generate a strong secret:
   ```sh
   openssl rand -hex 32
   # or
   python3 -c "import secrets; print(secrets.token_hex(32))"
   ```
2. Add it to `.env`:
   ```sh
   INTI_MASTER_KEY=<generated secret>
   ```
3. Restart the server, then open `http://localhost:8282/api-keys.html`.
4. Paste the master key into **Your Access Key** and save — this stores it in the browser for admin calls.
5. Create API keys to share with others. They use those keys; you keep the master key private.

**Using the API with a key:**

```sh
curl -s http://localhost:8282/api/voices \
  -H 'X-API-Key: inti_...'
```

The web UI reads the stored key from `localStorage` automatically and includes it with every request.

---

## Open the config folder

A helper script opens the config directory in your file manager:

```sh
./open-config.sh
```

- **macOS** — opens Finder
- **Linux** — opens via `xdg-open`
- Falls back to printing the path if no GUI is available
- Respects `INTI_CONFIG_DIR` if set
