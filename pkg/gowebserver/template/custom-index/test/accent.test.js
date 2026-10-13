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

// The page with canvases that hand back, for whatever was drawn into
// them last, that picture's color in every pixel. drawn lists what was
// drawn straight from an element, each of which decodes a picture in a
// browser, right then. opts.bitmaps gives the page fetch, which has
// every picture cached but those in opts.uncached (fetched lists what
// was asked for), and createImageBitmap, whose bitmaps (listed in
// bitmaps) are made off to the side from what was fetched.
// opts.filters: false makes the canvases ones that can't filter.
function accent(opts) {
  opts = opts || {};
  const drawn = [], bitmaps = [], fetched = [];
  const globals = {};
  if (opts.bitmaps) {
    globals.fetch = (url, o) => {
      fetched.push([url, o.cache]);
      if ((opts.uncached || []).includes(url)) return Promise.reject(new TypeError('Failed to fetch'));
      return Promise.resolve({ ok: true, blob: () => Promise.resolve({ blob: true, rgb: COLORS[url] }) });
    };
    globals.createImageBitmap = (blob, o) => {
      if (!blob.blob || !blob.rgb) return Promise.reject(new Error('cannot decode'));
      const bm = { bitmap: true, rgb: blob.rgb, w: o.resizeWidth, h: o.resizeHeight, closed: false, close() { this.closed = true; } };
      bitmaps.push(bm);
      return Promise.resolve(bm);
    };
  }
  const pg = page('<div id="ambient"><canvas width="192" height="128"></canvas><canvas width="192" height="128"></canvas></div>', { globals });
  const canvas = (node) => {
    let last = null;
    const ctx = {
      clearRect() { last = null; },
      drawImage(pic) {
        if (pic.broken) throw new Error('tainted');
        last = pic;
        if (pic.tagName) drawn.push(pic);
        node.shows = pic;
        node.filtered = ctx.filter;
      },
      getImageData() {
        const data = new Uint8ClampedArray(node.width * node.height * 4);
        if (last) for (let i = 0; i < data.length; i += 4) data.set(last.rgb.concat(255), i);
        return { data };
      },
    };
    if (opts.filters !== false) ctx.filter = 'none';
    node.getContext = () => ctx;
    // What a canvas drawn from another one shows is that one's picture.
    Object.defineProperty(node, 'rgb', { get: () => last && last.rgb });
    return node;
  };
  const create = pg.document.createElement.bind(pg.document);
  pg.document.createElement = (tag) => (tag === 'canvas' ? canvas({ width: 0, height: 0 }) : create(tag));
  const el = pg.document.getElementById('ambient');
  el.querySelectorAll('canvas').forEach((c) => { c.width = 192; c.height = 128; canvas(c); });
  el.style.setProperty('--backdrop-blur', '36px');
  el.style.setProperty('--backdrop-tone', 'saturate(1.5)');
  el.rect = { width: 1536, height: 1024 };
  const a = load(pg, ['custom-index/util.js', 'custom-index/accent.js'], ['vividColor', 'accent', 'pageAmbient', 'Ambient', 'root']);
  return Object.assign(a, { pg, el, drawn, bitmaps, fetched, canvases: el.querySelectorAll('canvas') });
}

const RED = [220, 30, 30], BLUE = [30, 60, 220], GRAY = [128, 128, 128];
const COLORS = { 'red.jpg': RED, 'blue.jpg': BLUE, 'slow.jpg': RED, 'gray.jpg': GRAY };

// A thumbnail, or with big a photo straight from a camera.
function img(src, rgb, big) {
  return { tagName: 'IMG', complete: true, currentSrc: src, rgb, naturalWidth: big ? 6000 : 512, naturalHeight: big ? 4000 : 341 };
}
const big = (src, rgb) => img(src, rgb, true);
const isRed = (col) => col.h > 15 && col.h < 40, isBlue = (col) => col.h > 250 && col.h < 280;

test('the color is the vivid one, whatever else is in view', async () => {
  const a = accent();
  assert.ok(isRed(await a.vividColor([{ el: img('gray.jpg', GRAY), w: 2 }, { el: img('red.jpg', RED), w: 1 }])));
  assert.ok(isBlue(await a.vividColor([{ el: img('blue.jpg', BLUE), w: 1 }])));
  assert.equal(await a.vividColor([{ el: img('gray.jpg', GRAY), w: 1 }]), null);
});

test('a photo is read once, even as a new element', async () => {
  const a = accent();
  const first = await a.vividColor([{ el: img('red.jpg', RED), w: 1 }, { el: img('blue.jpg', BLUE), w: 1 }]);
  assert.equal(a.drawn.length, 2);
  // Its row scrolled away and came back: new elements, same pictures.
  const again = await a.vividColor([{ el: img('red.jpg', RED), w: 1 }, { el: img('blue.jpg', BLUE), w: 1 }]);
  assert.equal(a.drawn.length, 2);
  assert.deepEqual(again, first);
});

test('an element showing another picture is read again', async () => {
  const a = accent();
  const el = img('red.jpg', RED);
  assert.ok(isRed(await a.vividColor([{ el, w: 1 }])));
  el.currentSrc = 'blue.jpg';
  el.rgb = BLUE;
  assert.ok(isBlue(await a.vividColor([{ el, w: 1 }])));
  assert.equal(a.drawn.length, 2);
});

test('what is still changing or loading is read every time', async () => {
  const a = accent();
  const video = { tagName: 'VIDEO', currentSrc: 'clip.mp4', rgb: RED };
  const loading = { tagName: 'IMG', complete: false, currentSrc: 'red.jpg', rgb: RED };
  for (let i = 0; i < 2; i++) await a.vividColor([{ el: video, w: 1 }, { el: loading, w: 1 }]);
  assert.equal(a.drawn.length, 4);
});

test('a picture that cannot be read is left out', async () => {
  const a = accent();
  const col = await a.vividColor([{ el: Object.assign(img('x.jpg', RED), { broken: true }), w: 1 }, { el: img('blue.jpg', BLUE), w: 1 }]);
  assert.ok(isBlue(col));
});

test('a large photo is read from a small bitmap, not decoded on the spot', async () => {
  const a = accent({ bitmaps: true });
  const video = { tagName: 'VIDEO', currentSrc: 'clip.mp4', rgb: GRAY };
  const thumb = img('gray.jpg', GRAY);
  assert.ok(isRed(await a.vividColor([{ el: big('red.jpg', RED), w: 1 }, { el: video, w: 1 }, { el: thumb, w: 1 }])));
  // The video and the thumbnail, quick to draw, were drawn directly.
  assert.deepEqual(a.drawn, [video, thumb]);
  // The photo came from the browser's cache and nowhere else.
  assert.deepEqual(a.fetched, [['red.jpg', 'only-if-cached']]);
  assert.deepEqual(a.bitmaps.map((b) => [b.w, b.h, b.closed]), [[16, 12, true]]);
});

test('a large photo the browser did not keep is left out, and asked for once', async () => {
  const a = accent({ bitmaps: true, uncached: ['gone.jpg'] });
  const gone = big('gone.jpg', RED);
  // Never decoded on the spot: the color comes from the others.
  assert.ok(isBlue(await a.vividColor([{ el: gone, w: 1 }, { el: big('blue.jpg', BLUE), w: 1 }])));
  assert.equal(await a.vividColor([{ el: gone, w: 1 }]), null);
  assert.equal(a.drawn.length, 0);
  assert.equal(a.fetched.filter((f) => f[0] === 'gone.jpg').length, 1);
});

test('a browser that never finds one in its cache draws them as before', async () => {
  const names = ['a.jpg', 'b.jpg', 'c.jpg', 'd.jpg', 'e.jpg'];
  const a = accent({ bitmaps: true, uncached: names });
  const photos = names.map((n) => big(n, RED));
  assert.equal(await a.vividColor(photos.slice(0, 3).map((el) => ({ el, w: 1 }))), null);
  // The fourth miss in a row with nothing ever found: it can't be asked.
  assert.ok(isRed(await a.vividColor([{ el: photos[3], w: 1 }])));
  assert.ok(isRed(await a.vividColor([{ el: photos[4], w: 1 }])));
  assert.deepEqual(a.drawn, [photos[3], photos[4]]);
  assert.equal(a.fetched.length, 4);
});

test('large photos are decoded a couple at a time', async () => {
  const a = accent({ bitmaps: true });
  const make = a.pg.createImageBitmap, waiting = [];
  let most = 0, now = 0;
  a.pg.createImageBitmap = (blob, o) => new Promise((resolve) => {
    most = Math.max(most, ++now);
    waiting.push(() => { now--; resolve(make(blob, o)); });
  });
  const sources = [];
  for (let i = 0; i < 6; i++) sources.push({ el: big(i % 2 ? 'red.jpg' : 'blue.jpg', i % 2 ? RED : BLUE), w: 1 });
  const pending = a.vividColor(sources);
  for (let i = 0; i < 6; i++) {
    await a.pg.settle();
    waiting.shift()();
  }
  assert.ok(await pending);
  assert.equal(most, 2);
});

test('a photo that went away while its bitmap was made is not read', async () => {
  const names = ['a.jpg', 'b.jpg', 'c.jpg', 'red.jpg'];
  const a = accent({ bitmaps: true, uncached: names });
  await a.vividColor(names.slice(0, 3).map((n) => ({ el: big(n, RED), w: 1 })));
  // This one would be drawn directly, were it still there.
  const gone = big('red.jpg', RED);
  const pending = a.vividColor([{ el: gone, w: 1 }]);
  gone.currentSrc = '';
  assert.equal(await pending, null);
  assert.equal(a.drawn.length, 0);
  // It wasn't remembered as colorless either.
  assert.ok(isRed(await a.vividColor([{ el: img('red.jpg', RED), w: 1 }])));
});

test('the accent follows the latest request, however late an older one is', async () => {
  const a = accent({ bitmaps: true });
  const make = a.pg.createImageBitmap;
  let release;
  a.pg.createImageBitmap = (blob, o) => {
    if (blob.rgb !== RED) return make(blob, o);
    return new Promise((r) => { release = () => r(make(blob, o)); });
  };
  a.accent.from('page', [{ el: big('slow.jpg', RED), w: 1 }]);
  a.accent.from('page', [{ el: big('blue.jpg', BLUE), w: 1 }]);
  await a.pg.settle();
  const hue = () => Number(a.root.style.getPropertyValue('--amb-h'));
  assert.ok(isBlue({ h: hue() }), String(hue()));
  release();
  await a.pg.settle();
  assert.ok(isBlue({ h: hue() }), String(hue()));
  // Cleared, a request still out doesn't bring the color back.
  a.accent.from('page', [{ el: big('slow.jpg', RED), w: 1 }]);
  await a.pg.settle();
  a.accent.clear('page');
  release();
  await a.pg.settle();
  assert.equal(a.root.classList.contains('amb-accent'), false);
});

test('the backdrop is blurred once, into its canvas', async () => {
  const a = accent({ bitmaps: true });
  const photo = big('red.jpg', RED);
  a.pageAmbient.show(photo);
  await a.pg.settle();
  const [back, front] = a.canvases;
  assert.equal(front.classList.contains('on'), true);
  assert.equal(back.classList.contains('on'), false);
  assert.equal(a.el.classList.contains('on'), true);
  assert.equal(a.el.classList.contains('css-blur'), false);
  // 36px of a backdrop 1536 wide, in a canvas 192 wide.
  assert.equal(front.filtered, 'blur(4.50px) saturate(1.5)');
  assert.deepEqual(front.rgb, RED);
  assert.deepEqual(a.bitmaps.map((b) => [b.w, b.h, b.closed]), [[24, 16, true]]);
  assert.equal(a.drawn.length, 0);

  // The same picture again changes nothing; another one fades over it.
  a.pageAmbient.show(photo);
  a.pageAmbient.show(big('blue.jpg', BLUE));
  await a.pg.settle();
  assert.equal(a.bitmaps.length, 2);
  assert.equal(back.classList.contains('on'), true);
  assert.equal(front.classList.contains('on'), false);
  assert.deepEqual(back.rgb, BLUE);
});

test('only the last picture asked for becomes the backdrop', async () => {
  const a = accent({ bitmaps: true });
  const colored = [];
  const amb = a.Ambient(a.el, (src) => colored.push(src.currentSrc));
  amb.show(big('red.jpg', RED));
  amb.show(big('blue.jpg', BLUE));
  await a.pg.settle();
  assert.deepEqual(colored, ['blue.jpg']);
  assert.deepEqual(a.canvases[1].rgb, BLUE);
  assert.equal(a.canvases[0].classList.contains('on'), false);
  assert.ok(a.bitmaps.every((b) => b.closed));
});

test('a picture that cannot be drawn leaves the backdrop as it was', async () => {
  const a = accent();
  a.pageAmbient.show(img('red.jpg', RED));
  await a.pg.settle();
  const bad = Object.assign(img('x.jpg', BLUE), { broken: true });
  a.pageAmbient.show(bad);
  await a.pg.settle();
  assert.equal(a.canvases[1].classList.contains('on'), true);
  assert.equal(a.canvases[0].classList.contains('on'), false);
});

test('where a canvas cannot filter, CSS blurs the backdrop', async () => {
  const a = accent({ filters: false });
  a.pageAmbient.show(img('red.jpg', RED));
  await a.pg.settle();
  assert.equal(a.el.classList.contains('css-blur'), true);
  assert.equal(a.canvases[1].filtered, undefined);
  assert.deepEqual(a.canvases[1].rgb, RED);
});
