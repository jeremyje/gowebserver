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

// Times the file browser in a real browser, against a real server:
//   make benchmark-browser [BROWSER=firefox] [PHOTOS=5000]
//     [PHOTO_SIZE=3000x2000] [THUMBNAILS=off]
//
// It opens a folder of PHOTOS pictures the server has never seen (cold:
// every thumbnail has to be made), scrolls through it, then loads it
// again (warm: the server has the thumbnails it made). It scrolls twice:
// flung, too fast for the grid to load anything, and browsed, slowly
// with pauses, which is when pictures load, fade in and color the
// backdrop. THUMBNAILS=off has the tiles show the files themselves, as
// when the server can't resize them. Times are in
// milliseconds from the start of each page load, and include the server,
// so compare runs on one machine.

'use strict';

const { serve, open } = require('./browser.js');

const PHOTOS = Number(process.env.PHOTOS) || 2000;
const PHOTO_SIZE = (process.env.PHOTO_SIZE || '640x480').split('x').map(Number);
const SCROLL_FRAMES = 180, SCROLL_STEP = 160;
// Slow enough that the grid loads as it goes, with a stop every so
// often that's long enough for the backdrop to follow.
const BROWSE_FRAMES = 240, BROWSE_STEP = 30, BROWSE_PAUSE_EVERY = 40, BROWSE_PAUSE = 700;

function report(name, value, unit) {
  console.log(name.padEnd(34) + String(value).padStart(10) + ' ' + unit);
}

// In the page: resolves, with the time, once every tile on screen has
// its picture (or its placeholder).
function screenFilled() {
  return new Promise((resolve) => {
    (function check() {
      const tiles = Array.from(document.querySelectorAll('.tile')).filter((t) => {
        const r = t.getBoundingClientRect();
        return r.bottom > 0 && r.top < innerHeight;
      });
      if (tiles.length && tiles.every((t) => t.classList.contains('ready') || t.classList.contains('err'))) resolve(performance.now());
      else requestAnimationFrame(check);
    })();
  });
}

// In the page: scrolls a step a frame and resolves to the time each
// frame took and how long the page was busy for 50ms or more at a time
// (where the browser reports that). With pauseEvery, it stops for
// pause ms after that many frames; the stops aren't frames.
function scrollFrames(args) {
  return new Promise((resolve) => {
    const times = [];
    let last = performance.now(), busy = 0;
    try {
      new PerformanceObserver((list) => list.getEntries().forEach((e) => { busy += e.duration; })).observe({ entryTypes: ['longtask'] });
    } catch (e) { }
    (function step() {
      const now = performance.now();
      times.push(now - last);
      last = now;
      if (times.length > args.frames) { resolve({ times: times.slice(1), busy }); return; }
      if (args.pauseEvery && times.length % args.pauseEvery === 0) {
        setTimeout(() => {
          last = performance.now();
          scrollBy(0, args.step);
          requestAnimationFrame(step);
        }, args.pause);
        return;
      }
      scrollBy(0, args.step);
      requestAnimationFrame(step);
    })();
  });
}

async function scroll(b, label, args) {
  const r = await b.page.evaluate(scrollFrames, args);
  const frames = r.times.sort((x, y) => x - y);
  const at = (q) => frames[Math.min(frames.length - 1, Math.floor(frames.length * q))];
  report(label + ' median frame', at(0.5).toFixed(1), 'ms');
  report(label + ' 95th percentile frame', at(0.95).toFixed(1), 'ms');
  report(label + ' slowest frame', at(1).toFixed(1), 'ms');
  report(label + ' frames over 50ms', frames.filter((t) => t > 50).length, 'of ' + frames.length);
  report(label + ' page busy 50ms or more', r.busy.toFixed(0), 'ms');
}

async function load(b, url, label) {
  await b.page.addInitScript(() => {
    new MutationObserver((records, obs) => {
      if (document.querySelector('.tile.ready')) { window.firstTile = performance.now(); obs.disconnect(); }
    }).observe(document, { subtree: true, attributes: true, attributeFilter: ['class'] });
  });
  await b.page.goto(url, { waitUntil: 'domcontentloaded' });
  const filled = await b.page.evaluate(screenFilled);
  const m = await b.page.evaluate(() => {
    const nav = performance.getEntriesByType('navigation')[0];
    return {
      html: nav.responseEnd, ready: nav.domContentLoadedEventEnd, first: window.firstTile,
      bytes: nav.transferSize, elements: document.querySelectorAll('*').length, tiles: document.querySelectorAll('.tile').length,
    };
  });
  report(label + ' page received', m.html.toFixed(0), 'ms');
  report(label + ' page ready', m.ready.toFixed(0), 'ms');
  report(label + ' first thumbnail shown', m.first.toFixed(0), 'ms');
  report(label + ' screen filled', filled.toFixed(0), 'ms');
  return m;
}

(async () => {
  const args = process.env.THUMBNAILS === 'off' ? ['-thumbnails.enabled=false'] : [];
  const srv = await serve({ photos: PHOTOS, photoSize: PHOTO_SIZE, args });
  const b = await open();
  try {
    console.log(b.name + ' ' + b.browser.version() + ', ' + PHOTOS + ' photos of ' + PHOTO_SIZE.join('x') + (args.length ? ', no thumbnails' : ''));
    const m = await load(b, srv.url, 'cold:');
    report('page size', (m.bytes / 1024).toFixed(0), 'kB');
    report('elements in page', m.elements, '');
    report('tiles in page', m.tiles, 'of ' + PHOTOS);

    await scroll(b, 'fling:', { frames: SCROLL_FRAMES, step: SCROLL_STEP });
    // From the top again: the fling may have reached the end.
    await b.page.evaluate(() => scrollTo(0, 0));
    await b.page.waitForTimeout(500);
    await scroll(b, 'browse:', { frames: BROWSE_FRAMES, step: BROWSE_STEP, pauseEvery: BROWSE_PAUSE_EVERY, pause: BROWSE_PAUSE });
    const before = await b.page.evaluate(() => performance.now());
    report('screen filled after scrolling', (await b.page.evaluate(screenFilled) - before).toFixed(0), 'ms');
    report('elements in page after scroll', await b.page.evaluate(() => document.querySelectorAll('*').length), '');

    await load(b, srv.url, 'warm:');
    if (b.problems.length) console.log('problems:\n  ' + b.problems.join('\n  '));
  } finally {
    await b.close();
    srv.close();
  }
})().catch((e) => { console.error(e); process.exit(1); });
