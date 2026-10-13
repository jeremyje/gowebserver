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
const { serve, open } = require('../../browser.js');

const SOURCE = 'package main\n\nfunc main() {\n\tprintln("' + 'a long line '.repeat(60) + '")\n}\n';
let srv, b;

test.before(async () => {
  srv = await serve({ files: { 'src/main.go': SOURCE } });
  b = await open();
  await b.page.goto(srv.url + 'src/main.go?view=rich');
});

test.after(async () => {
  if (b) await b.close();
  if (srv) srv.close();
});

test('the file is shown, highlighted, under its name', async () => {
  assert.equal(await b.page.locator('#title').textContent(), 'main.go');
  assert.equal(await b.page.title(), 'main.go');
  const lines = await b.page.locator('#code .cl').allTextContents();
  assert.equal(lines.join(''), SOURCE);
  assert.ok(await b.page.locator('#code span[class]').count() > 10, 'the code is marked up for color');
});

test('breadcrumbs lead back up', async () => {
  const hrefs = await b.page.locator('#crumbs a').evaluateAll((as) => as.map((a) => a.getAttribute('href')));
  assert.deepEqual(hrefs, ['/', '/src/']);
});

test('wrap folds long lines and is remembered', async () => {
  const wide = () => b.page.locator('#code .chroma').first().evaluate((el) => el.scrollWidth > el.clientWidth);
  assert.equal(await wide(), true, 'a long line scrolls sideways');
  await b.page.click('#wrap');
  assert.equal(await b.page.locator('#wrap').getAttribute('aria-pressed'), 'true');
  assert.equal(await wide(), false, 'wrapped, it fits');
  await b.page.reload();
  assert.equal(await b.page.locator('#wrap').getAttribute('aria-pressed'), 'true');
  assert.equal(await wide(), false);
});

test('nothing went wrong along the way', () => {
  assert.deepEqual(b.problems, []);
});
