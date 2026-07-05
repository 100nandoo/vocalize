import test from 'node:test';
import assert from 'node:assert/strict';

import {
  loadRuntimeConfig,
  readConfigJSON,
  saveRuntimeConfig,
} from '../../web-src/src/lib/runtime-settings-transport.js';

test('readConfigJSON surfaces backend error payloads and fallback messages', async () => {
  await assert.rejects(
    readConfigJSON(new Response(JSON.stringify({ error: 'backend exploded' }), { status: 500 }), 'fallback'),
    /backend exploded/,
  );

  await assert.rejects(
    readConfigJSON(new Response(null, { status: 500, statusText: 'Server Error' }), 'fallback'),
    /Server Error|fallback/,
  );
});

test('loadRuntimeConfig fetches a saved runtime config through the shared transport', async () => {
  const result = await loadRuntimeConfig<{ provider: string; voice: string; model: string }>({
    apiURL: (path: string) => `http://localhost:8282${path}?key=secret`,
    path: '/api/speech-config',
    fetchImpl: async (url) => {
      assert.equal(String(url), 'http://localhost:8282/api/speech-config?key=secret');
      return Response.json({
        provider: 'gemini',
        voice: 'Kore',
        model: 'gemini-2.5-flash-preview-tts',
      });
    },
    errorMessage: 'Could not load speech settings.',
  });

  assert.deepEqual(result, {
    provider: 'gemini',
    voice: 'Kore',
    model: 'gemini-2.5-flash-preview-tts',
  });
});

test('saveRuntimeConfig posts JSON through the shared runtime transport', async () => {
  const result = await saveRuntimeConfig<{ provider: string; model: string }>({
    apiURL: (path: string) => path,
    path: '/api/summarizer-config',
    payload: {
      provider: 'groq',
      model: 'openai/gpt-oss-120b',
      keys: { gemini: '', groq: 'gsk_saved', openrouter: '' },
    },
    fetchImpl: async (url, options = {}) => {
      assert.equal(String(url), '/api/summarizer-config');
      assert.equal(options.method, 'POST');
      assert.equal((options.headers as Record<string, string>)['Content-Type'], 'application/json');
      return Response.json(JSON.parse(options.body as string));
    },
    errorMessage: 'Could not save summarizer settings.',
  });

  assert.equal(result.provider, 'groq');
  assert.equal(result.model, 'openai/gpt-oss-120b');
});
