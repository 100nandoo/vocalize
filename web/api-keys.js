const API_KEY_STORAGE = 'inti:apiKey';

function getStoredKey() {
  return localStorage.getItem(API_KEY_STORAGE) || '';
}

function apiHeaders(extra = {}) {
  const k = getStoredKey();
  return k ? { ...extra, 'X-API-Key': k } : extra;
}

function formatDate(iso) {
  if (!iso) return '—';
  const d = new Date(iso);
  return d.toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' });
}

// ── DOM refs ──────────────────────────────────────────────────
const storedKeyInput = document.getElementById('stored-key-input');
const saveKeyBtn     = document.getElementById('save-key-btn');
const clearKeyBtn    = document.getElementById('clear-key-btn');
const keySaveStatus  = document.getElementById('key-save-status');

const newKeyName     = document.getElementById('new-key-name');
const createKeyBtn   = document.getElementById('create-key-btn');
const createStatus   = document.getElementById('create-status');
const setupBanner    = document.getElementById('setup-banner');
const keysTable      = document.getElementById('keys-table');
const keysBody       = document.getElementById('keys-body');
const keysEmpty      = document.getElementById('keys-empty');
const keysError      = document.getElementById('keys-error');

const keyModal       = document.getElementById('key-modal');
const keyModalValue  = document.getElementById('key-modal-value');
const keyModalCopy   = document.getElementById('key-modal-copy');
const keyModalSave   = document.getElementById('key-modal-save');
const keyModalBackdrop = document.getElementById('key-modal-backdrop');

// ── Stored key card ───────────────────────────────────────────
storedKeyInput.value = getStoredKey();

saveKeyBtn.addEventListener('click', () => {
  const val = storedKeyInput.value.trim();
  if (val) {
    localStorage.setItem(API_KEY_STORAGE, val);
  } else {
    localStorage.removeItem(API_KEY_STORAGE);
  }
  keySaveStatus.textContent = 'Saved';
  setTimeout(() => { keySaveStatus.textContent = ''; }, 2000);
});

clearKeyBtn.addEventListener('click', () => {
  localStorage.removeItem(API_KEY_STORAGE);
  storedKeyInput.value = '';
  keySaveStatus.textContent = 'Cleared';
  setTimeout(() => { keySaveStatus.textContent = ''; }, 2000);
});

// ── Key list ──────────────────────────────────────────────────
async function loadKeys() {
  keysError.style.display = 'none';
  try {
    const res = await fetch('/api/admin/keys', { headers: apiHeaders() });
    if (res.status === 401) {
      showError('Unauthorized — save your API key in the card above first.');
      return;
    }
    if (!res.ok) throw new Error(res.statusText);
    const { keys } = await res.json();
    renderKeys(keys || []);
  } catch (e) {
    showError('Could not load keys: ' + e.message);
  }
}

function renderKeys(keys) {
  if (keys.length === 0) {
    keysTable.style.display  = 'none';
    keysEmpty.style.display  = 'block';
    setupBanner.style.display = 'block';
  } else {
    keysEmpty.style.display   = 'none';
    setupBanner.style.display = 'none';
    keysTable.style.display   = 'table';
    keysBody.innerHTML = keys.map(k => `
      <tr>
        <td>${escHtml(k.name)}</td>
        <td><span class="key-prefix">${escHtml(k.prefix)}…</span></td>
        <td>${formatDate(k.createdAt)}</td>
        <td>${formatDate(k.lastUsedAt)}</td>
        <td>
          <button class="btn-secondary" style="font-size:12px;padding:4px 10px;"
                  onclick="deleteKey('${escHtml(k.id)}', '${escHtml(k.name)}')">
            Delete
          </button>
        </td>
      </tr>
    `).join('');
  }
}

function showError(msg) {
  keysError.textContent = msg;
  keysError.style.display = 'block';
}

function escHtml(s) {
  return String(s)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}

// ── Create key ────────────────────────────────────────────────
createKeyBtn.addEventListener('click', async () => {
  createKeyBtn.disabled = true;
  createStatus.textContent = 'Creating…';
  const name = newKeyName.value.trim();
  try {
    const res = await fetch('/api/admin/keys', {
      method: 'POST',
      headers: apiHeaders({ 'Content-Type': 'application/json' }),
      body: JSON.stringify({ name }),
    });
    if (res.status === 401) {
      createStatus.textContent = 'Unauthorized — save your API key above first.';
      return;
    }
    if (!res.ok) throw new Error(res.statusText);
    const { raw } = await res.json();
    newKeyName.value = '';
    createStatus.textContent = '';
    showKeyModal(raw);
    await loadKeys();
  } catch (e) {
    createStatus.textContent = 'Error: ' + e.message;
  } finally {
    createKeyBtn.disabled = false;
  }
});

// ── Delete key ────────────────────────────────────────────────
async function deleteKey(id, name) {
  if (!confirm(`Delete key "${name}"?\n\nAny requests using it will immediately return 401.`)) return;
  try {
    const res = await fetch(`/api/admin/keys/${id}`, {
      method: 'DELETE',
      headers: apiHeaders(),
    });
    if (!res.ok && res.status !== 204) throw new Error(res.statusText);
    await loadKeys();
  } catch (e) {
    showError('Delete failed: ' + e.message);
  }
}

// ── One-time key modal ────────────────────────────────────────
function showKeyModal(raw) {
  keyModalValue.textContent = raw;
  keyModal.style.display = 'flex';
}

function closeKeyModal() {
  keyModal.style.display = 'none';
  keyModalValue.textContent = '';
}

keyModalCopy.addEventListener('click', () => {
  navigator.clipboard.writeText(keyModalValue.textContent).then(() => {
    keyModalCopy.textContent = 'Copied!';
    setTimeout(() => { keyModalCopy.textContent = 'Copy'; }, 2000);
  });
});

keyModalSave.addEventListener('click', () => {
  const raw = keyModalValue.textContent;
  localStorage.setItem(API_KEY_STORAGE, raw);
  storedKeyInput.value = raw;
  closeKeyModal();
  keySaveStatus.textContent = 'Key saved to browser';
  setTimeout(() => { keySaveStatus.textContent = ''; }, 3000);
});

keyModalBackdrop.addEventListener('click', closeKeyModal);

// ── Init ──────────────────────────────────────────────────────
loadKeys();
