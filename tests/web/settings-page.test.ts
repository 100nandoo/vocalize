import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

import {
  flushAsyncWork,
  installDom,
  requiredElement,
  setInputValue,
  setSelectValue,
  teardownPage,
} from './svelte-page-test-helpers.ts';

test('settings route uses the Svelte secondary-page entrypoint', () => {
  const html = readFileSync(new URL('../../web-src/settings.html', import.meta.url), 'utf8');

  assert.match(html, /<script type="module" src="\.\/src\/entries\/settings\.js"><\/script>/);
});

test('settings page loads and saves through the current backend APIs', async (t) => {
  const requests: Array<{ url: string; method: string; body: unknown }> = [];
  const themePreviewCalls: Array<[action: 'apply' | 'persist', theme: string]> = [];

  const dom = installDom('http://localhost:8282/settings.html?key=main-secret');
  t.after(() => teardownPage(dom));

  (
    window as typeof window & {
      IntiTheme: {
        apply: (theme: string) => void;
        persist: (theme: string) => void;
      };
    }
  ).IntiTheme = {
    apply: (theme) => themePreviewCalls.push(['apply', theme]),
    persist: (theme) => themePreviewCalls.push(['persist', theme]),
  };

  globalThis.fetch = async (url, options = {}) => {
    const method = options.method || 'GET';
    const body = options.body ? JSON.parse(options.body as string) : null;
    const urlText = String(url);
    requests.push({ url: urlText, method, body });

    if (method === 'GET' && urlText === 'http://localhost:8282/api/summarizer-config?key=main-secret') {
      return Response.json({
        provider: 'groq',
        model: 'openai/gpt-oss-120b',
        keys: {
          gemini: '',
          groq: 'gsk_existing',
          openrouter: '',
        },
      });
    }

    if (method === 'GET' && urlText === 'http://localhost:8282/api/theme-config?key=main-secret') {
      return Response.json({
        theme: 'dark',
        summaryDownloadFormat: 'md',
        ocrPromotionBehavior: 'append',
        summaryPromotionBehavior: 'replace',
      });
    }

    if (method === 'GET' && urlText === '/summarizer-models/groq.json') {
      return Response.json([
        { value: 'openai/gpt-oss-120b', label: 'gpt-oss-120b' },
        { value: 'mixtral-scout', label: 'Mixtral Scout' },
      ]);
    }

    if (method === 'POST' && urlText === 'http://localhost:8282/api/summarizer-config?key=main-secret') {
      return Response.json(body);
    }

    if (method === 'POST' && urlText === 'http://localhost:8282/api/theme-config?key=main-secret') {
      return Response.json(body);
    }

    throw new Error(`Unexpected fetch to ${method} ${urlText}`);
  };

  await import(`../../web/assets/settings.js?test=${Date.now()}`);
  await flushAsyncWork();

  assert.deepEqual(
    [...document.querySelectorAll('.header-settings-link')].map((link) => link.getAttribute('href')),
    ['/?key=main-secret', '/api-keys.html?key=main-secret', '/settings.html?key=main-secret'],
  );
  assert.match(document.body.textContent ?? '', /Runtime Settings/);
  assert.match(document.body.textContent ?? '', /Visual Theme/);
  assert.match(document.body.textContent ?? '', /Providers/);
  assert.equal(requiredElement<HTMLSelectElement>('sum-provider-select').value, 'groq');
  assert.equal(requiredElement<HTMLSelectElement>('sum-model-select').value, 'openai/gpt-oss-120b');
  assert.equal(requiredElement<HTMLSelectElement>('appearance-theme-select').value, 'dark');
  assert.deepEqual(
    [...requiredElement<HTMLSelectElement>('appearance-theme-select').options].map((option) => option.value),
    ['light', 'dark'],
  );
  assert.equal(requiredElement<HTMLSelectElement>('summary-download-format-select').value, 'md');
  assert.equal(document.getElementById('ocr-promotion-behavior-select'), null);
  assert.equal(document.getElementById('summary-promotion-behavior-select'), null);
  assert.equal(requiredElement<HTMLInputElement>('key-groq').value, 'gsk_existing');

  setSelectValue(requiredElement<HTMLSelectElement>('sum-model-select'), 'mixtral-scout');
  setSelectValue(requiredElement<HTMLSelectElement>('appearance-theme-select'), 'dark');
  setSelectValue(requiredElement<HTMLSelectElement>('summary-download-format-select'), 'txt');
  setInputValue(requiredElement<HTMLInputElement>('key-groq'), 'gsk_saved');
  setInputValue(requiredElement<HTMLInputElement>('key-openrouter'), 'sk-or-saved');

  requiredElement<HTMLButtonElement>('sum-save-btn').click();
  await flushAsyncWork();

  assert.deepEqual(themePreviewCalls, [
    ['apply', 'dark'],
    ['persist', 'dark'],
  ]);
  assert.deepEqual(requests.slice(-2), [
    {
      url: 'http://localhost:8282/api/summarizer-config?key=main-secret',
      method: 'POST',
      body: {
        provider: 'groq',
        model: 'mixtral-scout',
        keys: {
          gemini: '',
          groq: 'gsk_saved',
          openrouter: 'sk-or-saved',
        },
      },
    },
    {
      url: 'http://localhost:8282/api/theme-config?key=main-secret',
      method: 'POST',
      body: {
        theme: 'dark',
        summaryDownloadFormat: 'txt',
        ocrPromotionBehavior: 'replace',
        summaryPromotionBehavior: 'replace',
      },
    },
  ]);
  assert.match(document.body.textContent ?? '', /Saved/);
});

test('settings page labels and scopes clear provider settings narrowly', async (t) => {
  const requests: Array<{ url: string; method: string; body: unknown }> = [];

  const dom = installDom('http://localhost:8282/settings.html?key=main-secret');
  t.after(() => teardownPage(dom));

  (
    window as typeof window & {
      IntiTheme: {
        apply: () => void;
        persist: () => void;
      };
    }
  ).IntiTheme = {
    apply() {},
    persist() {},
  };

  globalThis.fetch = async (url, options = {}) => {
    const method = options.method || 'GET';
    const body = options.body ? JSON.parse(options.body as string) : null;
    const urlText = String(url);
    requests.push({ url: urlText, method, body });

    if (method === 'GET' && urlText === 'http://localhost:8282/api/summarizer-config?key=main-secret') {
      return Response.json({
        provider: 'groq',
        model: 'openai/gpt-oss-120b',
        keys: {
          gemini: 'AIza-existing',
          groq: 'gsk_existing',
          openrouter: 'sk-or-existing',
        },
      });
    }

    if (method === 'GET' && urlText === 'http://localhost:8282/api/theme-config?key=main-secret') {
      return Response.json({
        theme: 'dark',
        summaryDownloadFormat: 'txt',
        ocrPromotionBehavior: 'replace',
        summaryPromotionBehavior: 'append',
      });
    }

    if (method === 'GET' && urlText === '/summarizer-models/groq.json') {
      return Response.json([
        { value: 'openai/gpt-oss-120b', label: 'gpt-oss-120b' },
      ]);
    }

    if (method === 'POST' && urlText === 'http://localhost:8282/api/summarizer-config?key=main-secret') {
      return Response.json(body);
    }

    throw new Error(`Unexpected fetch to ${method} ${urlText}`);
  };

  await import(`../../web/assets/settings.js?test=${Date.now()}`);
  await flushAsyncWork();

  assert.equal(requiredElement<HTMLButtonElement>('sum-clear-btn').textContent.trim(), 'Clear Provider Settings');
  assert.equal(requiredElement<HTMLSelectElement>('appearance-theme-select').value, 'dark');
  assert.equal(requiredElement<HTMLSelectElement>('summary-download-format-select').value, 'txt');
  assert.equal(document.getElementById('ocr-promotion-behavior-select'), null);
  assert.equal(document.getElementById('summary-promotion-behavior-select'), null);

  requiredElement<HTMLButtonElement>('sum-clear-btn').click();
  await flushAsyncWork();

  assert.deepEqual(requests.at(-1), {
    url: 'http://localhost:8282/api/summarizer-config?key=main-secret',
    method: 'POST',
    body: {
      provider: '',
      model: '',
      keys: {
        gemini: '',
        groq: '',
        openrouter: '',
      },
    },
  });
  assert.equal(requiredElement<HTMLSelectElement>('sum-provider-select').value, '');
  assert.equal(requiredElement<HTMLInputElement>('key-gemini').value, '');
  assert.equal(requiredElement<HTMLInputElement>('key-groq').value, '');
  assert.equal(requiredElement<HTMLInputElement>('key-openrouter').value, '');
  assert.equal(requiredElement<HTMLSelectElement>('appearance-theme-select').value, 'dark');
  assert.equal(requiredElement<HTMLSelectElement>('summary-download-format-select').value, 'txt');
  assert.equal(document.getElementById('ocr-promotion-behavior-select'), null);
  assert.equal(document.getElementById('summary-promotion-behavior-select'), null);
});

test('settings page falls back removed theme values to dark before saving', async (t) => {
  const requests: Array<{ url: string; method: string; body: unknown }> = [];

  const dom = installDom('http://localhost:8282/settings.html?key=main-secret');
  t.after(() => teardownPage(dom));

  (
    window as typeof window & {
      IntiTheme: {
        apply: (theme: string) => void;
        persist: (theme: string) => void;
      };
    }
  ).IntiTheme = {
    apply() {},
    persist() {},
  };

  globalThis.fetch = async (url, options = {}) => {
    const method = options.method || 'GET';
    const body = options.body ? JSON.parse(options.body as string) : null;
    const urlText = String(url);
    requests.push({ url: urlText, method, body });

    if (method === 'GET' && urlText === 'http://localhost:8282/api/summarizer-config?key=main-secret') {
      return Response.json({
        provider: '',
        model: '',
        keys: {
          gemini: '',
          groq: '',
          openrouter: '',
        },
      });
    }

    if (method === 'GET' && urlText === 'http://localhost:8282/api/theme-config?key=main-secret') {
      return Response.json({
        theme: 'minimal',
        summaryDownloadFormat: 'md',
        ocrPromotionBehavior: 'append',
        summaryPromotionBehavior: 'append',
      });
    }

    if (method === 'POST' && urlText === 'http://localhost:8282/api/summarizer-config?key=main-secret') {
      return Response.json(body);
    }

    if (method === 'POST' && urlText === 'http://localhost:8282/api/theme-config?key=main-secret') {
      return Response.json(body);
    }

    throw new Error(`Unexpected fetch to ${method} ${urlText}`);
  };

  await import(`../../web/assets/settings.js?test=${Date.now()}`);
  await flushAsyncWork();

  assert.equal(requiredElement<HTMLSelectElement>('appearance-theme-select').value, 'dark');

  requiredElement<HTMLButtonElement>('sum-save-btn').click();
  await flushAsyncWork();

  assert.deepEqual(requests.slice(-2), [
    {
      url: 'http://localhost:8282/api/summarizer-config?key=main-secret',
      method: 'POST',
      body: {
        provider: '',
        model: '',
        keys: {
          gemini: '',
          groq: '',
          openrouter: '',
        },
      },
    },
    {
      url: 'http://localhost:8282/api/theme-config?key=main-secret',
      method: 'POST',
      body: {
        theme: 'dark',
        summaryDownloadFormat: 'md',
        ocrPromotionBehavior: 'replace',
        summaryPromotionBehavior: 'replace',
      },
    },
  ]);
});
