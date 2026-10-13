// Copyright 2026 Cloudfra
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

'use strict';

const test = require('node:test');
const assert = require('node:assert/strict');
const { page, load } = require('../../harness.js');

const HTML = '<select id="code-theme"></select><button id="wrap" aria-pressed="false"></button>' +
  '<button id="copy" title="Copy contents" hidden></button>' +
  '<div id="code"><span class="line"><span class="ln">1</span><span class="cl">package main\n</span></span>' +
  '<span class="line"><span class="ln">2</span><span class="cl">func main() {}\n</span></span></div>';

function setup(opts) {
  opts = Object.assign({ html: HTML }, opts);
  const copied = [];
  const clipboard = { writeText: (text) => { copied.push(text); return opts.denied ? Promise.reject(new Error('denied')) : Promise.resolve(); } };
  const pg = page(opts.html, Object.assign({ navigator: opts.noClipboard ? {} : { clipboard } }, opts));
  if (opts.wrapped) pg.document.documentElement.classList.add('wrap');
  if (opts.insecure) pg.isSecureContext = false;
  load(pg, ['rich-view/rich-view.js'], []);
  return { pg, copied, root: pg.document.documentElement, $: (id) => pg.document.getElementById(id) };
}

test('picking code colors loads the page with them, at the same line', () => {
  const t = setup({ location: { hash: '#L12' } });
  t.$('code-theme').value = 'main.go?view=rich&theme=nord';
  t.$('code-theme').fire('change');
  assert.equal(t.pg.location.href, 'main.go?view=rich&theme=nord#L12');
});

test('wrap toggles long lines and is remembered', () => {
  const t = setup();
  const wrap = t.$('wrap');
  assert.equal(wrap.getAttribute('aria-pressed'), 'false');
  wrap.click();
  assert.equal(t.root.classList.contains('wrap'), true);
  assert.equal(wrap.getAttribute('aria-pressed'), 'true');
  assert.equal(t.pg.stored['gows-wrap'], '1');
  wrap.click();
  assert.equal(t.root.classList.contains('wrap'), false);
  assert.equal(wrap.getAttribute('aria-pressed'), 'false');
  assert.equal('gows-wrap' in t.pg.stored, false);
});

test('the wrap button shows a choice restored before the page drew', () => {
  const t = setup({ wrapped: true });
  assert.equal(t.$('wrap').getAttribute('aria-pressed'), 'true');
});

test('copy takes the code without its line numbers, and says so for a moment', async () => {
  const t = setup();
  const copy = t.$('copy');
  assert.equal(copy.hidden, false);
  copy.click();
  assert.deepEqual(t.copied, ['package main\nfunc main() {}\n']);
  await t.pg.settle();
  assert.equal(copy.classList.contains('done'), true);
  assert.equal(copy.title, 'Copied');
  t.pg.tick(1500);
  assert.equal(copy.classList.contains('done'), false);
  assert.equal(copy.title, 'Copy contents');
});

test('a refused copy claims nothing', async () => {
  const t = setup({ denied: true });
  t.$('copy').click();
  await t.pg.settle();
  assert.equal(t.$('copy').classList.contains('done'), false);
});

test('copy stays hidden where the browser can\'t copy', () => {
  assert.equal(setup({ noClipboard: true }).$('copy').hidden, true);
  assert.equal(setup({ insecure: true }).$('copy').hidden, true);
});

test('a file too big to show has no code, colors or copy, and the page still works', () => {
  const t = setup({ html: '<button id="wrap"></button><button id="copy" hidden></button>' });
  assert.equal(t.$('copy').hidden, true);
  t.$('wrap').click();
  assert.equal(t.root.classList.contains('wrap'), true);
});
