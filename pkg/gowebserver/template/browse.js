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

// A look at a page in a headless browser, for when something needs
// seeing rather than testing:
//   make browse TARGET=<url or folder> [SCREENSHOT=file.png] [BROWSER=firefox]
//
// A folder is served by the server first and its file browser opened.
// It prints what the page logged, what failed, how long loading took and
// what's in the file browser's grid, and saves a screenshot.

'use strict';

const fs = require('node:fs');
const path = require('node:path');
const { serve, open } = require('./browser.js');

(async () => {
  const target = process.argv[2], shot = process.argv[3];
  if (!target) throw new Error('usage: browse.js <url or folder> [screenshot.png]');
  const srv = /^https?:/.test(target) ? null : await serve({ dir: target });
  const b = await open();
  try {
    const logged = [];
    b.page.on('console', (m) => logged.push(m.type() + ': ' + m.text()));
    const requests = { count: 0, bytes: 0, slowest: [] };
    b.page.on('requestfinished', async (r) => {
      requests.count++;
      const t = r.timing();
      requests.slowest.push([Math.round(t.responseEnd), r.url()]);
      const sizes = await r.sizes().catch(() => null);
      if (sizes) requests.bytes += sizes.responseBodySize;
    });

    await b.page.goto(srv ? srv.url : target, { waitUntil: 'load' });
    // Long enough for thumbnails; a page that never goes quiet still gets looked at.
    await b.page.waitForLoadState('networkidle', { timeout: 15000 }).catch(() => console.log('still loading after 15s'));

    const info = await b.page.evaluate(() => {
      const nav = performance.getEntriesByType('navigation')[0];
      const n = (sel) => document.querySelectorAll(sel).length;
      return {
        title: document.title, url: location.href,
        received: Math.round(nav.responseEnd), ready: Math.round(nav.domContentLoadedEventEnd), loaded: Math.round(nav.loadEventEnd),
        elements: n('*'), rows: n('.row'), tiles: n('.tile'), shown: n('.tile.ready'), failed: n('.tile.err'),
        height: document.documentElement.scrollHeight,
      };
    });
    console.log(b.name + ' ' + b.browser.version());
    console.log('page      ' + info.title + '  ' + info.url);
    console.log('timing    received ' + info.received + 'ms, ready ' + info.ready + 'ms, loaded ' + info.loaded + 'ms');
    console.log('requests  ' + requests.count + ', ' + Math.round(requests.bytes / 1024) + ' kB');
    requests.slowest.sort((x, y) => y[0] - x[0]).slice(0, 5).forEach((r) => console.log('  ' + String(r[0]).padStart(6) + 'ms  ' + r[1]));
    console.log('page      ' + info.elements + ' elements, ' + info.height + 'px tall');
    if (info.rows || info.tiles) {
      console.log('browser   ' + info.rows + ' rows; ' + info.tiles + ' tiles in the page, ' + info.shown + ' shown, ' + info.failed + ' failed');
    }
    console.log('console   ' + (logged.length ? '\n  ' + logged.join('\n  ') : 'nothing logged'));
    console.log('problems  ' + (b.problems.length ? '\n  ' + b.problems.join('\n  ') : 'none'));
    if (shot) {
      fs.mkdirSync(path.dirname(path.resolve(shot)), { recursive: true });
      await b.page.screenshot({ path: shot });
      console.log('screenshot ' + path.resolve(shot));
    }
  } finally {
    await b.close();
    if (srv) srv.close();
  }
})().catch((e) => { console.error(e); process.exit(1); });
