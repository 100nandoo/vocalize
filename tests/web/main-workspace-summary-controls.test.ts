import test from 'node:test';
import assert from 'node:assert/strict';

import {
  buildMainWorkspaceProviderOptions,
  loadMainWorkspaceModelOptions,
  loadMainWorkspaceSummarizerConfig,
  saveMainWorkspaceSummarizerConfig,
} from '../../web-src/src/lib/main-workspace-summary-controls.js';

test('buildMainWorkspaceProviderOptions keeps server default and filters providers by saved keys', () => {
  const options = buildMainWorkspaceProviderOptions({
    provider: 'groq',
    model: 'openai/gpt-oss-120b',
    keys: {
      gemini: '',
      groq: 'gsk_live',
      openrouter: 'sk-or-live',
    },
    groqLimits: null,
  });

  assert.deepEqual(options, [
    { value: '', label: 'Server default' },
    { value: 'groq', label: 'Groq' },
    { value: 'openrouter', label: 'OpenRouter' },
    { value: 'mock', label: 'Mock' },
  ]);
});

test('loadMainWorkspaceSummarizerConfig preserves the backend response contract', async () => {
  const result = await loadMainWorkspaceSummarizerConfig({
    apiURL: (path: string) => `http://localhost:8282${path}?key=secret`,
    fetchImpl: async (url) => {
      assert.equal(String(url), 'http://localhost:8282/api/summarizer-config?key=secret');
      return Response.json({
        provider: 'groq',
        model: 'openai/gpt-oss-120b',
        keys: {
          gemini: '',
          groq: 'gsk_saved',
          openrouter: '',
        },
      });
    },
  });

  assert.equal(result.provider, 'groq');
  assert.equal(result.model, 'openai/gpt-oss-120b');
  assert.deepEqual(result.keys, {
    gemini: '',
    groq: 'gsk_saved',
    openrouter: '',
  });
});

test('saveMainWorkspaceSummarizerConfig posts the current provider/model payload', async () => {
  const result = await saveMainWorkspaceSummarizerConfig({
    apiURL: (path: string) => path,
    provider: 'openrouter',
    model: '',
    keys: {
      gemini: '',
      groq: 'gsk_saved',
      openrouter: 'sk-or-saved',
    },
    fetchImpl: async (url, options = {}) => {
      assert.equal(String(url), '/api/summarizer-config');
      assert.deepEqual(JSON.parse(options.body as string), {
        provider: 'openrouter',
        model: '',
        keys: {
          gemini: '',
          groq: 'gsk_saved',
          openrouter: 'sk-or-saved',
        },
      });
      return Response.json(JSON.parse(options.body as string));
    },
  });

  assert.equal(result.provider, 'openrouter');
  assert.equal(result.model, '');
  assert.equal(result.keys.openrouter, 'sk-or-saved');
});

test('loadMainWorkspaceModelOptions keeps model selection on supported providers and hides unsupported ones', async () => {
  const loaded = await loadMainWorkspaceModelOptions({
    provider: 'groq',
    selectedModel: 'mixtral-scout',
    fetchImpl: async (url) => {
      assert.equal(String(url), '/summarizer-models/groq.json');
      return Response.json([
        { value: 'openai/gpt-oss-120b', label: 'gpt-oss-120b' },
        { value: 'mixtral-scout', label: 'Mixtral Scout' },
      ]);
    },
  });

  assert.equal(loaded.hidden, false);
  assert.equal(loaded.selectedModel, 'mixtral-scout');
  assert.deepEqual(loaded.options.map((option) => option.value), [
    'openai/gpt-oss-120b',
    'mixtral-scout',
  ]);

  const openRouter = await loadMainWorkspaceModelOptions({
    provider: 'openrouter',
    selectedModel: 'ignored',
  });

  assert.deepEqual(openRouter, {
    hidden: true,
    options: [],
    selectedModel: '',
  });
});
