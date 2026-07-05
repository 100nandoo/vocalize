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
| `SPEECH_PROVIDER` | No | `gemini` | Active speech provider: `gemini` or `kokoro-heart` |
| `KOKORO_HEART_URL` | No | provider default | Override the `kokoro heart` upstream speech endpoint |
| `INTI_MAIN_KEY` | No | — | Main key that always authenticates all API requests. Set this when deploying publicly. See [API key management](#api-key-management). `INTI_MASTER_KEY` is still accepted as a fallback |
| `DEFAULT_VOICE` | No | provider-specific | Default TTS voice for the selected speech provider |
| `DEFAULT_MODEL` | No | `gemini-3.1-flash-tts-preview` | Default TTS model for Gemini. Ignored by `kokoro-heart` |
| `PORT` | No | `8282` | HTTP server port |
| `HOST` | No | `127.0.0.1` | HTTP server bind address. Set to `0.0.0.0` to listen on all interfaces |
| `INTI_CONFIG_DIR` | No | OS default | Override the directory where `inti.toml` is stored |
| `TELEGRAM_BOT_TOKEN` | No | — | When set, `./inti serve` also starts the Telegram bot in the same process |
| `SUMMARIZER_PROVIDER` | No | auto-detected | Active summarizer: `gemini`, `groq`, or `openrouter`. Auto-detected from available API keys if not set |
| `GROQ_API_KEY` | No | — | Groq API key. Enables Groq as a summarizer provider |
| `GROQ_MODEL` | No | `openai/gpt-oss-120b` | Groq model to use for summarization |
| `OPENROUTER_API_KEY` | No | — | OpenRouter API key. Enables OpenRouter as a summarizer provider |
| `OPENROUTER_MODEL` | No | `openrouter/free` | OpenRouter router to use for summarization; the app always uses the free router for OpenRouter |

### Example `.env`

```sh
GEMINI_API_KEY=AIza...
SPEECH_PROVIDER=gemini

# Protect the API when deploying publicly
# Generate with: openssl rand -hex 32
#           or:  python3 -c "import secrets; print(secrets.token_hex(32))"
INTI_MAIN_KEY=change_me_to_a_strong_secret

# Optional overrides
# DEFAULT_VOICE=Puck
# DEFAULT_MODEL=gemini-2.5-flash-preview-tts
# KOKORO_HEART_URL=https://koboldai-koboldcpp-tiefighter.hf.space/v1/audio/speech
# PORT=8282
# HOST=0.0.0.0
# TELEGRAM_BOT_TOKEN=123456:telegram_bot_token
```

To use `kokoro heart`, switch the speech provider and voice:

```sh
SPEECH_PROVIDER=kokoro-heart
DEFAULT_VOICE=cheery
```

---

## Config file location

Runtime settings changed via the web UI, including summarizer provider/model, speech provider selection, and appearance theme, plus API keys created via the API keys page and Telegram bot auth/session preferences, are persisted to `inti.toml` on disk.

The web frontend paints dark on first load. The light/dark toggle still stores an immediate local preference in browser `localStorage` under `inti-theme`, but the persisted `[appearance] theme` value overrides that local preference after the page loads.

```toml
[appearance]
theme = "dark" # "light" or "dark"; missing or legacy values fall back to dark
```

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

When deployed publicly (e.g. via Cloudflare Tunnel), set `INTI_MAIN_KEY` to protect all API endpoints.

**Authentication rules:**

- If `INTI_MAIN_KEY` is set — auth is enforced for HTML pages and every `/api/*` request.
- If `INTI_MAIN_KEY` is not set — `INTI_MASTER_KEY` is checked as a fallback for backward compatibility.
- Static assets such as CSS, JS, and the embedded SVG logo are publicly accessible so protected HTML pages can load correctly.

**Bootstrapping with a main key:**

1. Generate a strong secret:
   ```sh
   openssl rand -hex 32
   # or
   python3 -c "import secrets; print(secrets.token_hex(32))"
   ```
2. Add it to `.env`:
   ```sh
   INTI_MAIN_KEY=<generated secret>
   ```
3. Restart the server, then open `http://localhost:8282/api-keys.html?key=<your main key>`.
4. Create API keys to share with others. They use those keys; you keep the main key private.

**Using the API with a key:**

```sh
curl -s http://localhost:8282/api/voices \
  -H 'X-API-Key: inti_...'
```

The web UI reads the key from the page URL `?key=...` and includes it with every request.

## Experimental upstreams

`kokoro heart` is documented and intentionally supported, but it depends on a public upstream KoboldCpp + Kokoro deployment that may break or change without notice. Inti still treats that as an accepted product dependency and keeps its own `/api/speak` response normalized to Opus.

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
