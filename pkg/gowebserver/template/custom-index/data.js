function escapeHTML(s) {
  return String(s).replace(/[&<>"']/g, function (c) {
    return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c];
  });
}

// ===== Data =====
var all = CFG.media.map(function (m, i) {
  return { video: m[0] === 1, url: m[1], name: m[2], size: m[3], t: m[4], lower: m[2].toLowerCase(), pos: 0, row: -1 };
});
var vis = { photos: [], videos: [] };
// The list's rows, with what filtering needs to know about each kept
// here, so a filter pass reads nothing back from the page and only
// touches the rows that change. Names are read on the first filter,
// not on load.
var rowEls = doc.querySelectorAll('.row');
var rowDir = new Uint8Array(rowEls.length), rowHidden = new Uint8Array(rowEls.length);
var rowNames = null, lastQuery = null;
for (var ri = 0; ri < rowEls.length; ri++) rowDir[ri] = rowEls[ri].hasAttribute('data-dir') ? 1 : 0;

function summarize(folders, files) {
  var parts = [];
  if (folders) parts.push(plural(folders, 'folder', 'folders'));
  if (files) parts.push(plural(files, 'file', 'files'));
  if (vis.photos.length) parts.push(plural(vis.photos.length, 'photo', 'photos'));
  if (vis.videos.length) parts.push(plural(vis.videos.length, 'video', 'videos'));
  $('summary').textContent = parts.join(', ');
  $('no-match').hidden = !!(folders || files || vis.photos.length || vis.videos.length) || !rowEls.length && !all.length;
}

function applyFilter(q) {
  q = q.trim().toLowerCase();
  // The same filter again (a trailing space, say) changes nothing.
  if (q === lastQuery) return;
  lastQuery = q;
  var i, n = rowEls.length, folders = 0, files = 0;
  if (q && !rowNames) {
    rowNames = new Array(n);
    for (i = 0; i < n; i++) rowNames[i] = rowEls[i].querySelector('.name').textContent.toLowerCase();
  }
  for (i = 0; i < n; i++) {
    var hide = q && rowNames[i].indexOf(q) < 0 ? 1 : 0;
    if (hide !== rowHidden[i]) { rowHidden[i] = hide; rowEls[i].hidden = !!hide; }
    if (!hide) { if (rowDir[i]) folders++; else files++; }
  }
  vis.photos = []; vis.videos = [];
  for (i = 0; i < all.length; i++) {
    var it = all[i];
    if (q && it.lower.indexOf(q) < 0) continue;
    var list = it.video ? vis.videos : vis.photos;
    it.pos = list.length;
    list.push(it);
  }
  summarize(folders, files);
  if (grid) grid.relayout(false);
}
