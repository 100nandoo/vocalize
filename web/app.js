const VOICES = [
  { name: "Zephyr",        gender: "Female", characteristic: "Bright" },
  { name: "Puck",          gender: "Male",   characteristic: "Upbeat" },
  { name: "Charon",        gender: "Male",   characteristic: "Informative" },
  { name: "Kore",          gender: "Female", characteristic: "Firm" },
  { name: "Fenrir",        gender: "Male",   characteristic: "Excitable" },
  { name: "Leda",          gender: "Female", characteristic: "Youthful" },
  { name: "Orus",          gender: "Male",   characteristic: "Firm" },
  { name: "Aoede",         gender: "Female", characteristic: "Breezy" },
  { name: "Callirrhoe",    gender: "Female", characteristic: "Easy-going" },
  { name: "Autonoe",       gender: "Female", characteristic: "Bright" },
  { name: "Enceladus",     gender: "Male",   characteristic: "Breathy" },
  { name: "Iapetus",       gender: "Male",   characteristic: "Clear" },
  { name: "Umbriel",       gender: "Male",   characteristic: "Easy-going" },
  { name: "Algieba",       gender: "Male",   characteristic: "Smooth" },
  { name: "Despina",       gender: "Female", characteristic: "Smooth" },
  { name: "Erinome",       gender: "Female", characteristic: "Clear" },
  { name: "Algenib",       gender: "Male",   characteristic: "Gravelly" },
  { name: "Rasalgethi",    gender: "Male",   characteristic: "Informative" },
  { name: "Laomedeia",     gender: "Female", characteristic: "Upbeat" },
  { name: "Achernar",      gender: "Female", characteristic: "Soft" },
  { name: "Alnilam",       gender: "Male",   characteristic: "Firm" },
  { name: "Schedar",       gender: "Male",   characteristic: "Even" },
  { name: "Gacrux",        gender: "Female", characteristic: "Mature" },
  { name: "Pulcherrima",   gender: "Male",   characteristic: "Forward" },
  { name: "Achird",        gender: "Male",   characteristic: "Friendly" },
  { name: "Zubenelgenubi", gender: "Male",   characteristic: "Casual" },
  { name: "Vindemiatrix",  gender: "Female", characteristic: "Gentle" },
  { name: "Sadachbia",     gender: "Male",   characteristic: "Lively" },
  { name: "Sadaltager",    gender: "Male",   characteristic: "Knowledgeable" },
  { name: "Sulafat",       gender: "Female", characteristic: "Warm" },
];

const DEFAULT_VOICE = "Kore";

const imgModal       = document.getElementById('img-preview-modal');
const imgModalImg    = document.getElementById('img-preview-img');
const imgModalClose  = document.getElementById('img-preview-close');
const imgModalBack   = document.getElementById('img-preview-backdrop');

function openImgPreview(src) {
  imgModalImg.src = src;
  imgModal.hidden = false;
  document.addEventListener('keydown', onModalKey);
}
function closeImgPreview() {
  imgModal.hidden = true;
  imgModalImg.src = '';
  document.removeEventListener('keydown', onModalKey);
}
function onModalKey(e) { if (e.key === 'Escape') closeImgPreview(); }

imgModalClose.addEventListener('click', closeImgPreview);
imgModalBack.addEventListener('click', closeImgPreview);

const dropZone       = document.getElementById('drop-zone');
const fileInput      = document.getElementById('file-input');
const fileStaging    = document.getElementById('file-staging');
const fileList       = document.getElementById('file-list');
const clearFilesBtn  = document.getElementById('clear-files-btn');
const runOcrBtn      = document.getElementById('run-ocr-btn');
const ocrResult      = document.getElementById('ocr-result');
const ocrText        = document.getElementById('ocr-text');
const ocrCopyBtn     = document.getElementById('ocr-copy-btn');
const ocrCopyLabel   = document.getElementById('ocr-copy-label');
const ocrSynthBtn    = document.getElementById('ocr-synthesize-btn');

const textInput      = document.getElementById('text-input');
const modelSelect    = document.getElementById('model-select');
const genderFilter   = document.getElementById('gender-filter');
const voiceSelect    = document.getElementById('voice-select');
const synthesizeBtn  = document.getElementById('synthesize-btn');
const synthesizeLabel = document.getElementById('synthesize-label');
const speakBtn       = document.getElementById('speak-btn');
const downloadBtn    = document.getElementById('download-btn');
const playingBar     = document.getElementById('playing-bar');
const statusText     = document.getElementById('status-text');
const providerSelect   = document.getElementById('provider-select');
const sumModelWrap     = document.getElementById('sum-model-wrap');
const sumModelSelect   = document.getElementById('sum-model-select');
const summarizeBtn     = document.getElementById('summarize-btn');
const summarizeSpeakBtn = document.getElementById('summarize-speak-btn');
const summaryResult    = document.getElementById('summary-result');
const summaryText      = document.getElementById('summary-text');
const summaryCopyBtn   = document.getElementById('summary-copy-btn');
const summaryCopyLabel = document.getElementById('summary-copy-label');
const summarySpeakBtn  = document.getElementById('summary-speak-btn');
const feed           = document.getElementById('feed');
const feedEmpty      = document.getElementById('feed-empty');


let lastWavBlob  = null;
let processing   = false;
let stagedFiles  = [];
let dragSrcIndex = null;

// --- OCR drop zone ---

dropZone.addEventListener('dragover', (e) => {
  e.preventDefault();
  dropZone.classList.add('drag-active');
});

dropZone.addEventListener('dragleave', () => {
  dropZone.classList.remove('drag-active');
});

dropZone.addEventListener('drop', (e) => {
  e.preventDefault();
  dropZone.classList.remove('drag-active');
  const files = Array.from(e.dataTransfer.files).filter(f => f.type.startsWith('image/'));
  if (files.length) addStagedFiles(files);
});

fileInput.addEventListener('change', () => {
  const files = Array.from(fileInput.files);
  if (files.length) addStagedFiles(files);
  fileInput.value = '';
});

function addStagedFiles(files) {
  stagedFiles.push(...files);
  renderFileList();
}

function renderFileList() {
  fileList.innerHTML = '';

  stagedFiles.forEach((file, i) => {
    const item = document.createElement('li');
    item.className = 'file-item';
    item.draggable = true;
    item.dataset.index = i;

    const url = URL.createObjectURL(file);
    item.innerHTML = `
      <span class="drag-handle" title="Drag to reorder">⠿</span>
      <img class="file-thumb" src="${url}" alt="" />
      <span class="file-name" title="${escHtml(file.name)}">${escHtml(file.name)}</span>
      <button class="file-remove" data-index="${i}" title="Remove">×</button>`;

    const thumb = item.querySelector('img');
    thumb.addEventListener('load', () => URL.revokeObjectURL(url));
    thumb.addEventListener('click', () => {
      const previewUrl = URL.createObjectURL(file);
      openImgPreview(previewUrl);
      imgModalImg.addEventListener('load', () => URL.revokeObjectURL(previewUrl), { once: true });
    });

    item.addEventListener('dragstart', (e) => {
      dragSrcIndex = i;
      e.dataTransfer.effectAllowed = 'move';
      requestAnimationFrame(() => item.classList.add('dragging'));
    });
    item.addEventListener('dragend', () => {
      item.classList.remove('dragging');
      fileList.querySelectorAll('.file-item').forEach(el => el.classList.remove('drag-over'));
    });
    item.addEventListener('dragover', (e) => {
      e.preventDefault();
      e.dataTransfer.dropEffect = 'move';
      fileList.querySelectorAll('.file-item').forEach(el => el.classList.remove('drag-over'));
      item.classList.add('drag-over');
    });
    item.addEventListener('drop', (e) => {
      e.preventDefault();
      item.classList.remove('drag-over');
      if (dragSrcIndex !== null && dragSrcIndex !== i) {
        const [moved] = stagedFiles.splice(dragSrcIndex, 1);
        stagedFiles.splice(i, 0, moved);
        renderFileList();
      }
    });

    fileList.appendChild(item);
  });

  fileStaging.hidden = stagedFiles.length === 0;
}

fileList.addEventListener('click', (e) => {
  const btn = e.target.closest('.file-remove');
  if (!btn) return;
  stagedFiles.splice(parseInt(btn.dataset.index), 1);
  renderFileList();
});

clearFilesBtn.addEventListener('click', () => {
  stagedFiles = [];
  renderFileList();
});

runOcrBtn.addEventListener('click', () => {
  if (stagedFiles.length && !processing) uploadImagesForOCR([...stagedFiles]);
});

async function uploadImagesForOCR(files) {
  const label = files.length === 1 ? files[0].name : `${files.length} images`;
  dropZone.classList.add('ocr-loading');
  runOcrBtn.disabled = true;
  const item = addFeed('info', `OCR: ${label}`, 'extracting text…');

  try {
    const formData = new FormData();
    files.forEach(f => formData.append('files', f));

    const res = await fetch('/api/ocr', { method: 'POST', body: formData });
    if (!res.ok) {
      const body = await res.json().catch(() => ({ error: res.statusText }));
      throw new Error(body.error || res.statusText);
    }

    const { text } = await res.json();
    ocrText.value = text || '';
    ocrResult.hidden = false;
    summaryResult.hidden = true;
    summaryText.value = '';
    stagedFiles = [];
    renderFileList();

    const wordCount = text ? text.trim().split(/\s+/).filter(Boolean).length : 0;
    updateFeedItem(item, 'ok', `OCR: ${label}`, `${wordCount} word${wordCount !== 1 ? 's' : ''} extracted`);
  } catch (err) {
    updateFeedItem(item, 'fail', `OCR: ${label}`, err.message);
  } finally {
    dropZone.classList.remove('ocr-loading');
    runOcrBtn.disabled = false;
  }
}

ocrCopyBtn.addEventListener('click', async () => {
  if (!ocrText.value) return;
  try {
    await navigator.clipboard.writeText(ocrText.value);
    ocrCopyLabel.textContent = 'Copied!';
    setTimeout(() => { ocrCopyLabel.textContent = 'Copy'; }, 1500);
  } catch {
    ocrText.select();
    document.execCommand('copy');
    ocrCopyLabel.textContent = 'Copied!';
    setTimeout(() => { ocrCopyLabel.textContent = 'Copy'; }, 1500);
  }
});

ocrSynthBtn.addEventListener('click', () => {
  const text = ocrText.value.trim();
  if (!text || processing) return;
  textInput.value = text;
  lastWavBlob = null;
  speakBtn.disabled = true;
  downloadBtn.disabled = true;
  synthesizeText(text);
});

// --- Model dropdown ---

async function loadModels() {
  try {
    const res = await fetch('/api/models');
    if (!res.ok) return;
    const { models, default: defaultModel } = await res.json();
    modelSelect.innerHTML = '';
    models.forEach(m => {
      const opt = document.createElement('option');
      opt.value = m;
      opt.textContent = m;
      if (m === defaultModel) opt.selected = true;
      modelSelect.appendChild(opt);
    });
  } catch {
    // server unreachable — leave dropdown empty, speak will fail with a clear error
  }
}

loadModels();

// --- Summarizer settings (stored in localStorage by settings.html) ---

const STORAGE_KEY = 'vocalize:summarizer';

function getSummarizerOverride() {
  try {
    const saved = JSON.parse(localStorage.getItem(STORAGE_KEY) || '{}');
    const provider = saved.provider || '';
    const apiKey = provider && saved.keys ? (saved.keys[provider] || '') : '';
    return { provider, apiKey };
  } catch {
    return { provider: '', apiKey: '' };
  }
}

const SUMMARIZER_MODELS = {
  gemini: [
    { value: '',                  label: 'Server default' },
    { value: 'gemini-2.0-flash',  label: 'gemini-2.0-flash' },
    { value: 'gemini-1.5-flash',  label: 'gemini-1.5-flash' },
    { value: 'gemini-1.5-pro',    label: 'gemini-1.5-pro' },
  ],
  groq: [
    { value: '',                                          label: 'Server default' },
    { value: 'llama-3.3-70b-versatile',                  label: 'llama-3.3-70b-versatile' },
    { value: 'llama-3.1-8b-instant',                     label: 'llama-3.1-8b-instant' },
    { value: 'meta-llama/llama-4-scout-17b-16e-instruct', label: 'llama-4-scout-17b' },
    { value: 'qwen/qwen3-32b',                           label: 'qwen3-32b' },
    { value: 'groq/compound',                            label: 'compound' },
    { value: 'groq/compound-mini',                       label: 'compound-mini' },
    { value: 'openai/gpt-oss-120b',                      label: 'gpt-oss-120b' },
    { value: 'openai/gpt-oss-20b',                       label: 'gpt-oss-20b' },
    { value: 'allam-2-7b',                               label: 'allam-2-7b' },
  ],
  openrouter: [
    { value: '',                                                    label: 'Server default' },
    { value: 'google/gemma-3-27b-it:free',                         label: 'Gemma 3 27B (free)' },
    { value: 'google/gemma-3-12b-it:free',                         label: 'Gemma 3 12B (free)' },
    { value: 'google/gemma-3-4b-it:free',                          label: 'Gemma 3 4B (free)' },
    { value: 'google/gemma-3n-e4b-it:free',                        label: 'Gemma 3n 4B (free)' },
    { value: 'google/gemma-3n-e2b-it:free',                        label: 'Gemma 3n 2B (free)' },
    { value: 'google/gemma-4-31b-it:free',                         label: 'Gemma 4 31B (free)' },
    { value: 'google/gemma-4-26b-a4b-it:free',                     label: 'Gemma 4 26B A4B (free)' },
    { value: 'meta-llama/llama-3.3-70b-instruct:free',             label: 'Llama 3.3 70B (free)' },
    { value: 'meta-llama/llama-3.2-3b-instruct:free',              label: 'Llama 3.2 3B (free)' },
    { value: 'qwen/qwen3-next-80b-a3b-instruct:free',              label: 'Qwen3 Next 80B (free)' },
    { value: 'qwen/qwen3-coder:free',                              label: 'Qwen3 Coder (free)' },
    { value: 'nvidia/nemotron-3-super-120b-a12b:free',             label: 'Nemotron 3 Super 120B (free)' },
    { value: 'nvidia/nemotron-3-nano-30b-a3b:free',                label: 'Nemotron 3 Nano 30B (free)' },
    { value: 'nvidia/nemotron-nano-9b-v2:free',                    label: 'Nemotron Nano 9B (free)' },
    { value: 'nvidia/nemotron-nano-12b-v2-vl:free',                label: 'Nemotron Nano 12B VL (free)' },
    { value: 'openai/gpt-oss-120b:free',                           label: 'gpt-oss-120b (free)' },
    { value: 'openai/gpt-oss-20b:free',                            label: 'gpt-oss-20b (free)' },
    { value: 'nousresearch/hermes-3-llama-3.1-405b:free',          label: 'Hermes 3 405B (free)' },
    { value: 'minimax/minimax-m2.5:free',                          label: 'MiniMax M2.5 (free)' },
    { value: 'arcee-ai/trinity-large-preview:free',                label: 'Trinity Large Preview (free)' },
    { value: 'liquid/lfm-2.5-1.2b-instruct:free',                  label: 'LFM2.5 1.2B Instruct (free)' },
    { value: 'liquid/lfm-2.5-1.2b-thinking:free',                  label: 'LFM2.5 1.2B Thinking (free)' },
    { value: 'inclusionai/ling-2.6-flash:free',                    label: 'Ling 2.6 Flash (free)' },
    { value: 'z-ai/glm-4.5-air:free',                              label: 'GLM 4.5 Air (free)' },
    { value: 'cognitivecomputations/dolphin-mistral-24b-venice-edition:free', label: 'Dolphin Mistral 24B Venice (free)' },
  ],
};

function populateModelSelect(provider) {
  const models = SUMMARIZER_MODELS[provider] || [];
  sumModelWrap.hidden = models.length === 0;
  sumModelSelect.innerHTML = '';
  models.forEach(({ value, label }) => {
    const opt = document.createElement('option');
    opt.value = value;
    opt.textContent = label;
    sumModelSelect.appendChild(opt);
  });
}

function populateProviderSelect() {
  const saved = (() => { try { return JSON.parse(localStorage.getItem(STORAGE_KEY) || '{}'); } catch { return {}; } })();
  const keys = saved.keys || {};
  const current = providerSelect.value || saved.provider || '';

  const PROVIDERS = [
    { value: 'gemini',      label: 'Gemini' },
    { value: 'groq',        label: 'Groq' },
    { value: 'openrouter',  label: 'OpenRouter' },
  ];

  providerSelect.innerHTML = '<option value="">Server default</option>';
  PROVIDERS.forEach(({ value, label }) => {
    if (keys[value]) {
      const opt = document.createElement('option');
      opt.value = value;
      opt.textContent = label;
      providerSelect.appendChild(opt);
    }
  });

  if (current && [...providerSelect.options].some(o => o.value === current)) {
    providerSelect.value = current;
  }
  populateModelSelect(providerSelect.value);
}
populateProviderSelect();

providerSelect.addEventListener('change', () => {
  populateModelSelect(providerSelect.value);
});

// --- Voice dropdown ---

function populateVoices(genderFilter = 'All', keepSelection = false) {
  const prev = keepSelection ? voiceSelect.value : DEFAULT_VOICE;
  const filtered = genderFilter === 'All' ? VOICES : VOICES.filter(v => v.gender === genderFilter);

  voiceSelect.innerHTML = '';
  filtered.forEach(v => {
    const opt = document.createElement('option');
    opt.value = v.name;
    opt.textContent = `${v.name} — ${v.characteristic}`;
    if (v.name === prev) opt.selected = true;
    voiceSelect.appendChild(opt);
  });

  // If previous selection no longer in list, default to first
  if (!filtered.find(v => v.name === prev) && filtered.length > 0) {
    voiceSelect.selectedIndex = 0;
  }
}

populateVoices();

genderFilter.addEventListener('change', () => {
  populateVoices(genderFilter.value, true);
});

// --- Events ---

synthesizeBtn.addEventListener('click', async () => {
  const text = textInput.value.trim();
  if (!text || processing) return;
  await synthesizeText(text);
});

speakBtn.addEventListener('click', async () => {
  if (!lastWavBlob || processing) return;
  await playAudio();
});

textInput.addEventListener('keydown', async (e) => {
  if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
    e.preventDefault();
    const text = textInput.value.trim();
    if (text && !processing) await synthesizeText(text);
  }
});

textInput.addEventListener('input', () => {
  lastWavBlob = null;
  speakBtn.disabled = true;
  downloadBtn.disabled = true;
  synthesizeBtn.disabled = !textInput.value.trim() || processing;
});

summarizeBtn.addEventListener('click', async () => {
  const text = ocrText.value.trim();
  if (!text || processing) return;
  await summarizeText(text, false);
});

summarizeSpeakBtn.addEventListener('click', async () => {
  const text = ocrText.value.trim();
  if (!text || processing) return;
  await summarizeText(text, true);
});

summaryCopyBtn.addEventListener('click', async () => {
  const text = summaryText.innerText;
  if (!text) return;
  try {
    await navigator.clipboard.writeText(text);
    summaryCopyLabel.textContent = 'Copied!';
    setTimeout(() => { summaryCopyLabel.textContent = 'Copy'; }, 1500);
  } catch {
    summaryCopyLabel.textContent = 'Copied!';
    setTimeout(() => { summaryCopyLabel.textContent = 'Copy'; }, 1500);
  }
});

summarySpeakBtn.addEventListener('click', async () => {
  const text = summaryText.innerText.trim();
  if (!text || processing) return;
  textInput.value = text;
  lastWavBlob = null;
  speakBtn.disabled = true;
  downloadBtn.disabled = true;
  await synthesizeText(text);
});

downloadBtn.addEventListener('click', () => {
  if (!lastWavBlob) return;
  const url = URL.createObjectURL(lastWavBlob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `vocalize-${Date.now()}.opus`;
  a.click();
  URL.revokeObjectURL(url);
  addFeed('ok', 'Downloaded', 'Opus file saved to your downloads folder');
});

// --- Core ---

async function synthesizeText(text) {
  const voice = voiceSelect.value;
  const model = modelSelect.value;
  setProcessing(true);
  setStatus('Synthesizing…', '');

  const startTime = performance.now();
  const wordCount = text.trim().split(/\s+/).length;
  const item = addFeed('info', `"${truncate(text, 60)}"`, `${model} · ${voice} · synthesizing…`);

  try {
    const res = await fetch('/api/speak', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ text, voice, model }),
    });

    if (!res.ok) {
      const body = await res.json().catch(() => ({ error: res.statusText }));
      throw new Error(body.error || res.statusText);
    }

    const { opus } = await res.json();
    const opusBytes = Uint8Array.from(atob(opus), c => c.charCodeAt(0));
    lastWavBlob = new Blob([opusBytes], { type: 'audio/opus' });
    downloadBtn.disabled = false;
    speakBtn.disabled = false;

    const duration = ((performance.now() - startTime) / 1000).toFixed(1);
    setStatus('', '');
    updateFeedItem(item, 'ok', `"${truncate(text, 60)}"`, `${wordCount} words · ${duration}s · ${model} · ${voice} · ${(opusBytes.length / 1024).toFixed(1)} KB`);
  } catch (err) {
    setStatus(err.message, 'error');
    updateFeedItem(item, 'fail', `"${truncate(text, 60)}"`, err.message);
  } finally {
    setProcessing(false);
  }
}

async function playAudio() {
  if (!lastWavBlob) return;

  setStatus('Playing…', '');
  setPlaying(true);

  try {
    const base64Wav = await new Promise((resolve) => {
      const reader = new FileReader();
      reader.onload = () => {
        const arr = reader.result.split(',')[1] || reader.result;
        resolve(arr);
      };
      reader.readAsDataURL(lastWavBlob);
    });

    await playWAV(base64Wav);

    setStatus('', '');
  } catch (err) {
    setStatus(err.message, 'error');
  } finally {
    setPlaying(false);
  }
}

async function playWAV(base64Wav) {
  const bytes = Uint8Array.from(atob(base64Wav), c => c.charCodeAt(0));
  const ctx   = new AudioContext();
  const buf   = await ctx.decodeAudioData(bytes.buffer);
  const src   = ctx.createBufferSource();
  src.buffer  = buf;
  src.connect(ctx.destination);
  src.start();
  return new Promise(resolve => { src.onended = () => { ctx.close(); resolve(); }; });
}

async function summarizeText(text, shouldSpeak) {
  setProcessing(true);
  setStatus('Summarizing…', '');

  const startTime = performance.now();
  const wordCount = text.trim().split(/\s+/).length;
  const item = addFeed('info', `"${truncate(text, 60)}"`, 'summarizing…');

  try {
    const reqProvider = providerSelect.value;
    const saved = (() => { try { return JSON.parse(localStorage.getItem(STORAGE_KEY) || '{}'); } catch { return {}; } })();
    const apiKey = reqProvider && saved.keys ? (saved.keys[reqProvider] || '') : '';
    const res = await fetch('/api/summarize', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ text, instruction: '', provider: reqProvider, apiKey, model: sumModelSelect.value }),
    });

    if (!res.ok) {
      const body = await res.json().catch(() => ({ error: res.statusText }));
      throw new Error(body.error || res.statusText);
    }

    const { summary, provider, model, rateLimits } = await res.json();
    if (rateLimits && provider === 'groq') {
      try {
        const now = Date.now();
        const stored = JSON.parse(localStorage.getItem(STORAGE_KEY) || '{}');
        stored.groqLimits = {
          ...rateLimits,
          capturedAt: now,
          resetRequestsAt: now + parseGroqDuration(rateLimits.resetRequests),
          resetTokensAt:   now + parseGroqDuration(rateLimits.resetTokens),
        };
        localStorage.setItem(STORAGE_KEY, JSON.stringify(stored));
      } catch {}
    }
    summaryText.innerHTML = renderMarkdown(summary || '');
    summaryResult.hidden = false;

    const duration = ((performance.now() - startTime) / 1000).toFixed(1);
    const modelTag = model ? ` · ${model}` : (provider ? ` · ${provider}` : '');
    setStatus('', '');
    updateFeedItem(item, 'ok', `"${truncate(text, 60)}"`, `${wordCount} words → summary · ${duration}s${modelTag}`);

    if (shouldSpeak) {
      textInput.value = summary;
      await synthesizeText(summary);
    }
  } catch (err) {
    setStatus(err.message, 'error');
    updateFeedItem(item, 'fail', `"${truncate(text, 60)}"`, err.message);
  } finally {
    setProcessing(false);
  }
}

// --- UI helpers ---

function addFeed(kind, label, meta) {
  feedEmpty?.remove();
  const item = document.createElement('div');
  item.className = `feed-item ${kind}`;
  item.innerHTML = `
    <div class="feed-dot"></div>
    <div class="feed-content">
      <div class="feed-label">${escHtml(label)}</div>
      <div class="feed-meta">${escHtml(meta)}</div>
    </div>`;
  feed.prepend(item);
  return item;
}

function updateFeedItem(item, kind, label, meta) {
  item.className = `feed-item ${kind}`;
  item.querySelector('.feed-label').textContent = label;
  item.querySelector('.feed-meta').textContent  = meta;
}

function setProcessing(val) {
  processing = val;
  textInput.disabled      = val;
  modelSelect.disabled    = val;
  genderFilter.disabled   = val;
  voiceSelect.disabled    = val;
  providerSelect.disabled = val;
  sumModelSelect.disabled = val;
  synthesizeBtn.disabled  = val || !textInput.value.trim();
  speakBtn.disabled       = val || !lastWavBlob;
  downloadBtn.disabled    = val || !lastWavBlob;
  if (!val) textInput.focus();
}

function setPlaying(val) {
  playingBar.classList.toggle('active', val);
}

function setStatus(msg, kind) {
  statusText.textContent = msg;
  statusText.className   = 'status-text' + (kind ? ` ${kind}` : '');
}

function truncate(str, n) {
  return str.length > n ? str.slice(0, n) + '…' : str;
}

function escHtml(str) {
  return str.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');
}

function parseGroqDuration(s) {
  if (!s) return 0;
  let ms = 0;
  const h   = s.match(/(\d+(?:\.\d+)?)h/);
  const m   = s.match(/(\d+(?:\.\d+)?)m(?!s)/);
  const sec = s.match(/(\d+(?:\.\d+)?)s/);
  if (h)   ms += parseFloat(h[1])   * 3600000;
  if (m)   ms += parseFloat(m[1])   * 60000;
  if (sec) ms += parseFloat(sec[1]) * 1000;
  return ms;
}

function renderMarkdown(md) {
  const lines = md.split('\n');
  let html = '';
  let inList = false;

  for (const raw of lines) {
    const line = raw.trimEnd();

    const heading = line.match(/^(#{1,3})\s+(.+)/);
    if (heading) {
      if (inList) { html += '</ul>'; inList = false; }
      const lvl = heading[1].length;
      html += `<h${lvl}>${inlineMarkdown(heading[2])}</h${lvl}>`;
      continue;
    }

    const listItem = line.match(/^[-*]\s+(.+)/);
    if (listItem) {
      if (!inList) { html += '<ul>'; inList = true; }
      html += `<li>${inlineMarkdown(listItem[1])}</li>`;
      continue;
    }

    if (inList) { html += '</ul>'; inList = false; }
    if (line === '') continue;
    html += `<p>${inlineMarkdown(line)}</p>`;
  }

  if (inList) html += '</ul>';
  return html;
}

function inlineMarkdown(text) {
  return escHtml(text)
    .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
    .replace(/\*(.+?)\*/g,     '<em>$1</em>')
    .replace(/`(.+?)`/g,       '<code>$1</code>');
}
