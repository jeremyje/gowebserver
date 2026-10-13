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
const { page, load, markup } = require('../../harness.js');

// The upload page's own form, in a browser that can (or, with basic,
// cannot) upload from script. sent collects the requests made.
function setup(opts) {
  opts = opts || {};
  const sent = [];
  class FormData {
    constructor(form) { this.form = form; this.files = []; }
    append(name, file) { this.files.push([name, file.name]); }
  }
  class XMLHttpRequest {
    open(method, url) { this.method = method; this.url = url; }
    send(data) { this.data = data; sent.push(this); }
    respond(status, body) { this.status = status; this.responseText = JSON.stringify(body); this.onload(); }
  }
  const globals = opts.basic ? {} : { FormData, XMLHttpRequest, FileReader: class { } };
  const pg = page(markup('upload/upload.html', '<form', '</form>'), { globals });
  const $ = (id) => pg.document.getElementById(id);
  // Elements in a real browser can be dragged; the script checks.
  if (!opts.basic) pg.document.createElement = ((make) => (tag) => Object.assign(make(tag), { draggable: false }))(pg.document.createElement.bind(pg.document));
  else pg.document.createElement = ((make) => (tag) => { const el = make(tag); delete el.draggable; return el; })(pg.document.createElement.bind(pg.document));
  let picks = 0;
  $('file-input').click = () => { picks++; };
  load(pg, ['upload/upload.js'], []);
  const zone = $('drop-zone');
  return {
    pg, $, sent, zone,
    form: pg.document.querySelector('form'),
    picks: () => picks,
    state: () => ['is-uploading', 'is-success', 'is-error', 'is-dragover', 'is-dropped'].filter((c) => zone.classList.contains(c)),
    label: () => $('drop-label-strong').textContent + $('drop-label-rest').textContent,
    choose(...names) { return $('file-input').fire('change', { target: { files: names.map((name) => ({ name })) } }); },
  };
}

test('the form is marked so the server answers in JSON', () => {
  const t = setup();
  const flag = t.form.querySelector('input[name=ajax]');
  assert.equal(flag.type, 'hidden');
  assert.equal(flag.value, '1');
});

test('choosing files names them and uploads at once', () => {
  const t = setup();
  t.choose('a.txt');
  assert.equal(t.label(), 'a.txt selected');
  assert.deepEqual(t.state(), ['is-uploading']);
  assert.equal(t.sent.length, 1);
  assert.equal(t.sent[0].method, 'post');
  assert.equal(t.sent[0].data.form, t.form);

  const many = setup();
  many.choose('a.txt', 'b.txt', 'c.txt');
  assert.equal(many.label(), '3 files selected');
});

test('the browser is kept from submitting the form as well', () => {
  // Firefox submits a form for a scripted submit event unless it's
  // prevented, which uploads twice and leaves the page for the JSON.
  const t = setup();
  let submit;
  t.form.addEventListener('submit', (e) => { submit = e; });
  t.choose('a.txt');
  assert.equal(submit.defaultPrevented, true);
});

test('a successful upload shows success', () => {
  const t = setup();
  t.choose('a.txt');
  t.sent[0].respond(200, { success: true });
  assert.deepEqual(t.state(), ['is-success']);
});

test('the server\'s refusal is shown with its reason', () => {
  const t = setup();
  t.choose('a.txt');
  t.sent[0].respond(200, { success: false, error: 'bad token' });
  assert.deepEqual(t.state(), ['is-error']);
  assert.equal(t.$('error-detail').textContent, 'bad token');
});

test('a server error and a network error each say what happened', () => {
  const t = setup();
  t.choose('a.txt');
  t.sent[0].respond(500, {});
  assert.deepEqual(t.state(), ['is-error']);
  assert.equal(t.$('error-detail').textContent, 'Server error (500).');

  const n = setup();
  n.choose('a.txt');
  n.sent[0].onerror();
  assert.deepEqual(n.state(), ['is-error']);
  assert.match(n.$('error-detail').textContent, /^Network error/);
});

test('a second submit while uploading is ignored', () => {
  const t = setup();
  t.choose('a.txt');
  t.choose('b.txt');
  assert.equal(t.sent.length, 1);
});

test('dropped files are uploaded under the file input\'s name', () => {
  const t = setup();
  t.zone.fire('dragover');
  assert.deepEqual(t.state(), ['is-dragover']);
  const e = t.zone.fire('drop', { dataTransfer: { files: [{ name: 'x.png' }, { name: 'y.png' }] } });
  assert.equal(e.defaultPrevented, true, 'the browser must not open the file');
  assert.equal(t.label(), '2 files selected');
  assert.deepEqual(t.state(), ['is-uploading', 'is-dropped']);
  const name = t.$('file-input').name;
  assert.deepEqual(t.sent[0].data.files, [[name, 'x.png'], [name, 'y.png']]);
  t.pg.tick(600);
  assert.deepEqual(t.state(), ['is-uploading'], 'the ripple ends');
});

test('leaving the zone ends the drag highlight', () => {
  const t = setup();
  t.zone.fire('dragenter');
  t.zone.fire('dragleave');
  assert.deepEqual(t.state(), []);
});

test('clicking the zone opens the file picker, but not from the label or a result', () => {
  const t = setup();
  t.zone.fire('click');
  assert.equal(t.picks(), 1);
  t.$('drop-label-strong').fire('click');
  assert.equal(t.picks(), 1, 'the label opens it by itself');
  t.zone.querySelector('.status').fire('click');
  assert.equal(t.picks(), 1);
});

test('upload more starts over and opens the picker', () => {
  const t = setup();
  t.zone.fire('drop', { dataTransfer: { files: [{ name: 'x.png' }] } });
  t.sent[0].respond(200, { success: true });
  const e = t.zone.querySelector('.restart').fire('click');
  assert.equal(e.defaultPrevented, true, 'it does not follow the link');
  assert.deepEqual(t.state().filter((c) => c !== 'is-dropped'), []);
  assert.equal(t.label(), 'Choose files or drag and drop here');
  assert.equal(t.picks(), 1);
  // The dropped files are forgotten.
  t.choose('b.txt');
  assert.deepEqual(t.sent[1].data.files, []);
});

test('a browser that cannot upload from script submits the form itself', () => {
  const t = setup({ basic: true });
  const e = t.choose('a.txt');
  assert.deepEqual(t.state(), ['is-uploading']);
  assert.equal(t.sent.length, 0);
  assert.equal(e.defaultPrevented, false);
  t.zone.fire('dragover');
  assert.deepEqual(t.state(), ['is-uploading'], 'and has no drag and drop');
});

test('focus on the hidden input shows on the zone\'s label', () => {
  const t = setup();
  const input = t.$('file-input');
  input.fire('focus');
  assert.equal(input.classList.contains('has-focus'), true);
  input.fire('blur');
  assert.equal(input.classList.contains('has-focus'), false);
});
