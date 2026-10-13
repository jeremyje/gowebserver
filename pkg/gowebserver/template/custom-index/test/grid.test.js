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

const DAY = 86400;

// A folder of photos then videos, a day apart from 2024-01-01, newest
// first when sorted by date.
function folder(photos, videos) {
  const media = [];
  for (let i = 0; i < photos; i++) media.push([0, 'p' + i + '.jpg', 'p' + i + '.jpg', 1000 + i, 1704110400 + i * DAY]);
  for (let i = 0; i < (videos || 0); i++) media.push([1, 'v' + i + '.mp4', 'v' + i + '.mp4', 5000 + i, 1704110400 + i * DAY]);
  return media;
}

// The grid over media in a window 1280 wide and 800 tall, with the grid
// starting 100px down the page. The parts of the page it talks to but
// that aren't under test are stand-ins.
function setup(media, opts) {
  opts = Object.assign({ sort: 'name', thumbs: true, videoThumbs: true, width: 1280 }, opts);
  const opened = [];
  const pg = page('<p id="summary"></p><p id="no-match"></p><div id="media"></div>', {
    globals: {
      GOWS: { sort: opts.sort, thumbs: opts.thumbs, videoThumbs: opts.videoThumbs, media },
      pageAmbient: { show() { } },
      accent: { set() { } },
      vividColor: () => null,
      scrubber: { measure() { }, place() { }, refresh() { } },
      slideshow: { open: (list, pos) => opened.push([list.length, pos]) },
    },
  });
  const el = pg.document.getElementById('media');
  el.clientWidth = opts.width;
  el.getBoundingClientRect = () => ({ top: 100 - pg.scrollY });
  const g = load(pg, ['custom-index/util.js', 'custom-index/thumbs.js', 'custom-index/data.js', 'custom-index/grid.js'],
    ['grid', 'applyFilter', 'all', 'fmtDuration']);
  g.applyFilter('');
  const tiles = () => el.querySelectorAll('.tile');
  return Object.assign(g, {
    pg, el, opened, tiles,
    rows: () => el.children,
    names: () => tiles().map((a) => a.getAttribute('aria-label')),
    // Answers every thumbnail request so far, then shows the batch.
    arrive(type) {
      pg.requests.splice(0).forEach((r) => { if (r.el.tagName === 'IMG') r.el.fire(type || 'load'); });
      pg.frame();
    },
  });
}

test('fmtDuration', () => {
  const g = setup([]);
  assert.equal(g.fmtDuration(0), '0:00');
  assert.equal(g.fmtDuration(9.9), '0:09');
  assert.equal(g.fmtDuration(75), '1:15');
  assert.equal(g.fmtDuration(3600), '1:00:00');
  assert.equal(g.fmtDuration(3725), '1:02:05');
});

test('an empty folder has no grid', () => {
  const g = setup([]);
  assert.equal(g.el.hidden, true);
  assert.equal(g.rows().length, 0);
});

test('photos are laid out in square tiles that fill the width', () => {
  const g = setup(folder(20));
  // 1280 wide: six columns of 208px tiles with 4px gaps.
  const first = g.el.querySelectorAll('.m-row')[1];
  assert.equal(first.children.length, 6);
  const w = (1280 - 4 * 5) / 6;
  assert.equal(first.children[0].style.width, w + 'px');
  assert.equal(first.children[0].style.height, w + 'px');
  assert.equal(first.children[1].style.left, (w + 4) + 'px');
  // A heading, then four rows for 20 photos.
  assert.equal(g.el.style.height, (56 + 4 * (w + 4)) + 'px');
  assert.equal(g.rows()[0].querySelector('h2').textContent, 'Photos');
  assert.equal(g.rows()[0].querySelector('.m-count').textContent, '20');
});

test('a phone gets at least three columns', () => {
  const g = setup(folder(20), { width: 320 });
  assert.equal(g.el.querySelectorAll('.m-row')[1].children.length, 3);
});

test('videos come first, in 16:9 tiles', () => {
  const g = setup(folder(4, 5));
  const heads = g.el.querySelectorAll('.m-head h2').map((h) => h.textContent);
  assert.deepEqual(heads, ['Videos', 'Photos']);
  const v = g.el.querySelector('.tile.vid');
  // Three columns with 12px gaps.
  const w = (1280 - 12 * 2) / 3;
  assert.equal(v.style.width, w + 'px');
  assert.equal(v.style.height, Math.round(w * 9 / 16) + 'px');
});

test('sorted by date, photos are grouped by month', () => {
  const g = setup(folder(40), { sort: 'date' });
  const groups = g.el.querySelectorAll('.m-group');
  const fmt = new Intl.DateTimeFormat(undefined, { month: 'long', year: 'numeric' });
  assert.equal(groups[0].firstChild.textContent, fmt.format(new Date(1704110400 * 1000)));
  assert.equal(groups[0].querySelector('.m-count').textContent, '31');
  assert.equal(groups[1].querySelector('.m-count').textContent, '9');
});

test('only the rows near the window are in the page, however big the folder', () => {
  const g = setup(folder(50000));
  const w = (1280 - 4 * 5) / 6;
  assert.equal(g.el.style.height, (56 + Math.ceil(50000 / 6) * (w + 4)) + 'px', 'the page is as tall as the whole folder');
  assert.ok(g.tiles().length <= 80, g.tiles().length + ' tiles mounted');

  g.pg.tick(100000);
  g.pg.scroll(900000);
  assert.ok(g.tiles().length <= 120, g.tiles().length + ' tiles mounted after scrolling');
  const row = Math.floor((900000 - 100 - 56) / (w + 4));
  assert.ok(g.names().includes('p' + row * 6 + '.jpg'), 'the rows under the window are the mounted ones');
  assert.ok(!g.names().includes('p0.jpg'), 'and the first rows are gone');
});

test('thumbnails are fetched a batch at a time and tiles fill in together', async () => {
  const g = setup(folder(200));
  await g.pg.settle();
  assert.deepEqual(g.pg.requests.map((r) => r.url), [0, 1, 2, 3].map((i) => 'p' + i + '.jpg?width=256&height=256&fit=outside'));
  assert.equal(g.el.querySelectorAll('.tile.ready').length, 0);
  g.arrive();
  assert.deepEqual(g.el.querySelectorAll('.tile.ready').map((a) => a.getAttribute('aria-label')), ['p0.jpg', 'p1.jpg', 'p2.jpg', 'p3.jpg']);
  await g.pg.settle();
  assert.equal(g.pg.requests.length, 4, 'the next batch is on its way');
});

test('a thumbnail that fails leaves the placeholder', async () => {
  const g = setup(folder(4));
  await g.pg.settle();
  g.arrive('error');
  assert.equal(g.el.querySelectorAll('.tile.err').length, 4);
  assert.equal(g.el.querySelectorAll('.tile.ready').length, 0);
});

test('tiles scrolled past give up their place in the queue', async () => {
  const g = setup(folder(5000));
  await g.pg.settle();
  const first = g.pg.requests.map((r) => r.el);
  g.pg.requests.length = 0;
  g.pg.tick(100000);
  g.pg.scroll(60000);
  assert.ok(first.every((img) => img.src === ''), 'requests for unmounted tiles are dropped');
  await g.pg.settle();
  const w = (1280 - 4 * 5) / 6;
  const top = Math.floor((60000 - 100 - 56 - 400) / (w + 4)) * 6;
  const got = g.pg.requests.map((r) => Number(/p(\d+)/.exec(r.url)[1]));
  assert.equal(got.length, 4);
  assert.ok(got.every((n) => n >= top - 6), 'the next batch is from what is on screen now: ' + got);
});

test('nothing is fetched during a fling; it loads when the scroll settles', async () => {
  const g = setup(folder(5000));
  await g.pg.settle();
  g.pg.requests.length = 0;
  // 50,000px in 10ms.
  g.pg.tick(10);
  g.pg.scroll(50000);
  await g.pg.settle();
  assert.ok(g.tiles().length > 0, 'tiles are laid out');
  assert.equal(g.pg.requests.length, 0, 'but not loaded');
  g.pg.tick(110);
  g.pg.frame();
  await g.pg.settle();
  assert.equal(g.pg.requests.length, 4);
});

test('at most three videos load at once, each with its thumbnail', async () => {
  const g = setup(folder(0, 9));
  assert.equal(g.el.querySelectorAll('video').length, 3);
  await g.pg.settle();
  assert.deepEqual(g.pg.requests.filter((r) => r.el.tagName === 'IMG').map((r) => r.url),
    ['v0.mp4?width=512&height=288&fit=outside', 'v1.mp4?width=512&height=288&fit=outside', 'v2.mp4?width=512&height=288&fit=outside']);
  // Its length arrives with the metadata, and frees the slot.
  const v = g.el.querySelector('video');
  v.duration = 75;
  v.fire('loadedmetadata');
  assert.equal(v.parentNode.querySelector('.dur').textContent, '1:15');
  assert.equal(g.el.querySelectorAll('video').length, 4);
});

test('a video whose thumbnail fails shows the video alone', async () => {
  const g = setup(folder(0, 1));
  await g.pg.settle();
  g.arrive('error');
  const tile = g.el.querySelector('.tile.vid');
  assert.equal(tile.querySelector('.vthumb'), null);
  assert.equal(tile.classList.contains('err'), false);
  assert.ok(tile.querySelector('video'));
});

test('without server frames a video tile asks the browser for one', () => {
  const g = setup(folder(0, 1), { videoThumbs: false });
  assert.equal(g.el.querySelector('video').src, 'v0.mp4#t=0.5');
  assert.equal(g.el.querySelector('.vthumb'), null);
});

test('filtering lays out only what matches', () => {
  const g = setup(folder(200));
  g.applyFilter('p19');
  assert.deepEqual(g.names(), ['p19.jpg', 'p190.jpg', 'p191.jpg', 'p192.jpg', 'p193.jpg', 'p194.jpg', 'p195.jpg', 'p196.jpg', 'p197.jpg', 'p198.jpg', 'p199.jpg']);
  assert.equal(g.rows()[0].querySelector('.m-count').textContent, '11');
});

test('clicking a tile opens the slideshow there', () => {
  const g = setup(folder(20));
  const e = g.tiles()[7].fire('click');
  assert.equal(e.defaultPrevented, true);
  assert.deepEqual(g.opened, [[20, 7]]);
  // A modified click is left to the browser (open in a new tab).
  g.tiles()[8].fire('click', { ctrlKey: true });
  assert.equal(g.opened.length, 1);
});
