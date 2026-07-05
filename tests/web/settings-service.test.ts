import test from 'node:test';
import assert from 'node:assert/strict';

import {
  clearSummarizerSettings,
  loadSettings,
  saveSettings,
} from '../../web-src/src/lib/settings-service.js';

test('loadSettings fetches summarizer and appearance settings together', async () => {
  const fetchCalls: string[] = [];
  const apiURL = (path: string) => `http://localhost:8282${path}?key=secret`;

  const result = await loadSettings({
    apiURL,
    fetchImpl: async (url) => {
      const urlText = String(url);
      fetchCalls.push(urlText);
      if (urlText.includes('/api/summarizer-config')) {
        return Response.json({ provider: 'groq', model: 'llama', keys: { groq: 'gsk_test' } });
      }
      if (urlText.includes('/api/theme-config')) {
        return Response.json({
          theme: 'dark',
          summaryDownloadFormat: 'md',
          ocrPromotionBehavior: 'append',
          summaryPromotionBehavior: 'replace',
        });
      }
      throw new Error(`Unexpected request: ${urlText}`);
    },
  });

  assert.deepEqual(fetchCalls, [
    'http://localhost:8282/api/summarizer-config?key=secret',
    'http://localhost:8282/api/theme-config?key=secret',
  ]);
  assert.equal(result.summarizerConfig.provider, 'groq');
  assert.equal(result.appearanceConfig.theme, 'dark');
  assert.equal(result.appearanceConfig.ocrPromotionBehavior, 'replace');
  assert.equal(result.appearanceConfig.summaryPromotionBehavior, 'replace');
});

test('saveSettings preserves the current backend contracts', async () => {
  const bodies: unknown[] = [];
  const apiURL = (path: string) => path;

  const result = await saveSettings({
    apiURL,
    provider: 'groq',
    model: 'openai/gpt-oss-120b',
    keys: { gemini: '', groq: 'gsk_saved', openrouter: '' },
    appearanceConfig: {
      theme: 'dark',
      summaryDownloadFormat: 'txt',
      ocrPromotionBehavior: 'replace',
      summaryPromotionBehavior: 'replace',
    },
    fetchImpl: async (_url, options = {}) => {
      const body = JSON.parse(options.body as string);
      bodies.push(body);
      return Response.json(body);
    },
  });

  assert.deepEqual(bodies, [
    {
      provider: 'groq',
      model: 'openai/gpt-oss-120b',
      keys: { gemini: '', groq: 'gsk_saved', openrouter: '' },
    },
    {
      theme: 'dark',
      summaryDownloadFormat: 'txt',
      ocrPromotionBehavior: 'replace',
      summaryPromotionBehavior: 'replace',
    },
  ]);
  assert.equal(result.summarizerConfig.model, 'openai/gpt-oss-120b');
  assert.equal(result.appearanceConfig.summaryDownloadFormat, 'txt');
});

test('clearSummarizerSettings resets only provider-side settings', async () => {
  const result = await clearSummarizerSettings({
    apiURL: (path: string) => path,
    fetchImpl: async (_url, options = {}) => {
      assert.deepEqual(JSON.parse(options.body as string), {
        provider: '',
        model: '',
        keys: { gemini: '', groq: '', openrouter: '' },
      });
      return Response.json({
        provider: '',
        model: '',
        keys: { gemini: '', groq: '', openrouter: '' },
      });
    },
  });

  assert.equal(result.provider, '');
  assert.deepEqual(result.keys, { gemini: '', groq: '', openrouter: '' });
});
