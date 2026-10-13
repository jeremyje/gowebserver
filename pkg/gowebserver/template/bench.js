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

// Runs the templates' JavaScript benchmarks: node bench.js [filter].
// Each is a */bench/*.bench.js file that calls harness.bench().
//
// They run on the harness's stand-in browser, which does no layout or
// painting, so they time the scripts' own work (what they compute and
// how many elements they touch) and nothing a real browser adds. Use
// them to compare a change against what it replaces, not as page timings.

'use strict';

const fs = require('node:fs');
const path = require('node:path');

process.env.BENCH_FILTER = process.argv[2] || '';
for (const dir of fs.readdirSync(__dirname).sort()) {
  const benchDir = path.join(__dirname, dir, 'bench');
  if (!fs.existsSync(benchDir)) continue;
  for (const f of fs.readdirSync(benchDir).sort()) {
    if (f.endsWith('.bench.js')) require(path.join(benchDir, f));
  }
}
