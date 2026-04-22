const STORAGE_KEY = 'vocalize:summarizer';

const sumProviderSelect = document.getElementById('sum-provider-select');
const keyGemini         = document.getElementById('key-gemini');
const keyGroq           = document.getElementById('key-groq');
const keyOpenRouter     = document.getElementById('key-openrouter');
const sumSaveBtn        = document.getElementById('sum-save-btn');
const sumClearBtn       = document.getElementById('sum-clear-btn');
const sumSaveStatus     = document.getElementById('sum-save-status');

function load() {
  try {
    const saved = JSON.parse(localStorage.getItem(STORAGE_KEY) || '{}');
    if (saved.provider)         sumProviderSelect.value = saved.provider;
    if (saved.keys?.gemini)     keyGemini.value     = saved.keys.gemini;
    if (saved.keys?.groq)       keyGroq.value       = saved.keys.groq;
    if (saved.keys?.openrouter) keyOpenRouter.value = saved.keys.openrouter;
  } catch {}
}

function save() {
  localStorage.setItem(STORAGE_KEY, JSON.stringify({
    provider: sumProviderSelect.value,
    keys: {
      gemini:      keyGemini.value.trim(),
      groq:        keyGroq.value.trim(),
      openrouter:  keyOpenRouter.value.trim(),
    },
  }));
  sumSaveStatus.textContent = 'Saved';
  sumSaveStatus.className = 'status-text';
  setTimeout(() => { sumSaveStatus.textContent = ''; }, 2000);
}

function clearAll() {
  localStorage.removeItem(STORAGE_KEY);
  document.getElementById('groq-usage-section').hidden = true;
  sumProviderSelect.value = '';
  keyGemini.value = '';
  keyGroq.value = '';
  keyOpenRouter.value = '';
  sumSaveStatus.textContent = 'Cleared';
  sumSaveStatus.className = 'status-text';
  setTimeout(() => { sumSaveStatus.textContent = ''; }, 2000);
}

sumSaveBtn.addEventListener('click', save);
sumClearBtn.addEventListener('click', clearAll);

load();
renderGroqUsage();

function fmtDuration(ms) {
  if (ms <= 0) return 'now';
  const s = Math.round(ms / 1000);
  if (s < 60) return `${s}s`;
  const m = Math.floor(s / 60), rs = s % 60;
  if (m < 60) return rs > 0 ? `${m}m ${rs}s` : `${m}m`;
  const h = Math.floor(m / 60), rm = m % 60;
  return rm > 0 ? `${h}h ${rm}m` : `${h}h`;
}

function fmtAgo(ms) {
  const s = Math.round(ms / 1000);
  if (s < 60) return `${s}s ago`;
  if (s < 3600) return `${Math.round(s / 60)}m ago`;
  return `${Math.round(s / 3600)}h ago`;
}

function renderGroqUsage() {
  try {
    const saved = JSON.parse(localStorage.getItem(STORAGE_KEY) || '{}');
    const limits = saved.groqLimits;
    if (!limits) return;

    document.getElementById('groq-usage-section').hidden = false;

    const now = Date.now();

    // Timestamp
    if (limits.capturedAt) {
      document.getElementById('groq-usage-ts').textContent = `Updated ${fmtAgo(now - limits.capturedAt)}`;
    }

    // Requests (RPD) — static counts
    const limReq = parseInt(limits.limitRequests) || 0;
    const remReq = parseInt(limits.remainingRequests) || 0;
    if (limReq > 0) {
      const pct = Math.round((remReq / limReq) * 100);
      const bar = document.getElementById('groq-req-bar');
      bar.style.width = pct + '%';
      bar.classList.toggle('low', pct < 20);
      document.getElementById('groq-req-numbers').textContent =
        `${remReq.toLocaleString()} / ${limReq.toLocaleString()}`;
    }

    // Tokens (TPM) — static counts
    const limTok = parseInt(limits.limitTokens) || 0;
    const remTok = parseInt(limits.remainingTokens) || 0;
    if (limTok > 0) {
      const pct = Math.round((remTok / limTok) * 100);
      const bar = document.getElementById('groq-tok-bar');
      bar.style.width = pct + '%';
      bar.classList.toggle('low', pct < 20);
      document.getElementById('groq-tok-numbers').textContent =
        `${remTok.toLocaleString()} / ${limTok.toLocaleString()}`;
    }

    // Live countdown via absolute timestamps
    function tickResets() {
      const n = Date.now();
      const reqEl = document.getElementById('groq-req-reset');
      const tokEl = document.getElementById('groq-tok-reset');
      if (limits.resetRequestsAt) {
        reqEl.textContent = `Resets in ${fmtDuration(limits.resetRequestsAt - n)}`;
      }
      if (limits.resetTokensAt) {
        tokEl.textContent = `Resets in ${fmtDuration(limits.resetTokensAt - n)}`;
      }
    }
    tickResets();
    const timer = setInterval(tickResets, 1000);
    // Stop ticking when both resets have passed
    setTimeout(() => clearInterval(timer),
      Math.max(limits.resetRequestsAt || 0, limits.resetTokensAt || 0) - Date.now() + 2000);
  } catch {}
}
