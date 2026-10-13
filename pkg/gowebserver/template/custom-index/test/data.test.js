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
const { page, load, plain } = require('../../harness.js');

const ROWS = [['docs', true], ['Photos 2024', true], ['notes.txt', false], ['Report.PDF', false], ['photo-list.csv', false]];
const MEDIA = [[0, 'a%20b.jpg', 'a b.jpg', 10, 100], [1, 'clip.mp4', 'clip.mp4', 20, 200], [0, 'Photo.png', 'Photo.png', 30, 300]];

function setup(rows, media) {
  const html = '<p id="summary"></p><p id="no-match" hidden></p>' + rows.map((r) =>
    '<div class="row"' + (r[1] ? ' data-dir' : '') + '><a class="row-main"><span class="name">' + r[0] + '</span></a></div>').join('');
  const relayouts = [];
  const pg = page(html, {
    globals: {
      GOWS: { sort: 'name', thumbs: true, videoThumbs: true, media },
      grid: { relayout: (keep) => relayouts.push(keep) },
    },
  });
  const d = load(pg, ['custom-index/util.js', 'custom-index/data.js'], ['all', 'applyFilter', 'escapeHTML']);
  return {
    pg, relayouts, applyFilter: d.applyFilter, all: d.all, escapeHTML: d.escapeHTML,
    summary: () => pg.document.getElementById('summary').textContent,
    noMatch: () => !pg.document.getElementById('no-match').hidden,
    hidden: () => pg.document.querySelectorAll('.row').map((r) => r.hidden),
  };
}

test('escapeHTML', () => {
  const t = setup([], []);
  assert.equal(t.escapeHTML('<a href="x">\'&\'</a>'), '&lt;a href=&quot;x&quot;&gt;&#39;&amp;&#39;&lt;/a&gt;');
  assert.equal(t.escapeHTML(12), '12');
});

test('media entries are unpacked from the compact array', () => {
  const t = setup(ROWS, MEDIA);
  assert.deepEqual(plain(t.all[1]), { video: true, url: 'clip.mp4', name: 'clip.mp4', size: 20, t: 200, lower: 'clip.mp4', pos: 0, row: -1 });
  assert.equal(t.all[0].url, 'a%20b.jpg');
  assert.equal(t.all[2].lower, 'photo.png');
});

test('with no filter everything shows and is counted', () => {
  const t = setup(ROWS, MEDIA);
  t.applyFilter('');
  assert.equal(t.summary(), '2 folders, 3 files, 2 photos, 1 video');
  assert.deepEqual(t.hidden(), [false, false, false, false, false]);
  assert.equal(t.noMatch(), false);
  assert.deepEqual(t.all.map((it) => it.pos), [0, 0, 1], 'each item knows its place among its kind');
  assert.deepEqual(t.relayouts, [false]);
});

test('a filter matches names anywhere, whatever the case', () => {
  const t = setup(ROWS, MEDIA);
  t.applyFilter('  PHOTO ');
  assert.deepEqual(t.hidden(), [true, false, true, true, false]);
  assert.equal(t.summary(), '1 folder, 1 file, 1 photo');
  assert.equal(t.all[2].pos, 0, 'positions are among what is showing');
});

test('clearing the filter shows the rows again', () => {
  const t = setup(ROWS, MEDIA);
  t.applyFilter('photo');
  t.applyFilter('');
  assert.deepEqual(t.hidden(), [false, false, false, false, false]);
  assert.equal(t.summary(), '2 folders, 3 files, 2 photos, 1 video');
});

test('a filter that matches nothing says so', () => {
  const t = setup(ROWS, MEDIA);
  t.applyFilter('zzz');
  assert.equal(t.summary(), '');
  assert.equal(t.noMatch(), true);
});

test('an empty folder is not a failed search', () => {
  const t = setup([], []);
  t.applyFilter('');
  assert.equal(t.noMatch(), false);
});

test('the same filter again does no work', () => {
  const t = setup(ROWS, MEDIA);
  t.applyFilter('photo');
  t.applyFilter('photo ');
  assert.equal(t.relayouts.length, 1);
});

test('a filter pass writes only the rows that change', () => {
  const t = setup(ROWS, MEDIA);
  t.applyFilter('');
  const writes = [];
  t.pg.document.querySelectorAll('.row').forEach((r, i) => {
    let hidden = r.hidden;
    Object.defineProperty(r, 'hidden', { get: () => hidden, set: (v) => { hidden = v; writes.push(i); } });
  });
  t.applyFilter('photo');
  assert.deepEqual(writes, [0, 2, 3]);
  t.applyFilter('photos');
  assert.deepEqual(writes, [0, 2, 3, 4]);
});
