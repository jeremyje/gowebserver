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
const fs = require('node:fs');
const path = require('node:path');
const { serve, open } = require('../../browser.js');

let srv, b;

test.before(async () => {
  srv = await serve({ files: { 'a.txt': 'a' } });
  b = await open();
  await b.page.goto(srv.url + 'upload.asp');
});

test.after(async () => {
  if (b) await b.close();
  if (srv) srv.close();
});

// Every file under dir, by its name.
function uploaded(dir) {
  const out = {};
  for (const e of fs.readdirSync(dir, { recursive: true, withFileTypes: true })) {
    if (e.isFile()) out[e.name] = fs.readFileSync(path.join(e.parentPath || e.path, e.name), 'utf8');
  }
  return out;
}

test('chosen files upload at once and arrive whole', async () => {
  await b.page.setInputFiles('#file-input', [
    { name: 'one.txt', mimeType: 'text/plain', buffer: Buffer.from('first file\n') },
    { name: 'two.txt', mimeType: 'text/plain', buffer: Buffer.from('second file\n') },
  ]);
  await b.page.waitForSelector('#drop-zone.is-success');
  assert.equal(await b.page.locator('#drop-label-strong').textContent(), '2 files');
  assert.deepEqual(uploaded(srv.uploads), { 'one.txt': 'first file\n', 'two.txt': 'second file\n' });
});

test('upload more starts over', async () => {
  await b.page.locator('.status-success .restart').click();
  assert.equal(await b.page.locator('#drop-zone.is-success').count(), 0);
  assert.equal(await b.page.locator('#drop-label-strong').textContent(), 'Choose files');
});

test('nothing went wrong along the way', () => {
  assert.deepEqual(b.problems, []);
});
