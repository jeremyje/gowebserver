// ===== Scrubber =====
var scrubber = (function () {
  var el = $('scrub'), thumb = $('scrub-thumb'), label = $('scrub-label');
  var hideTimer = 0, dragging = false, grab = 0, enabled = false;
  var THUMB_H = 48;

  function metrics() {
    var track = el.getBoundingClientRect();
    return { top: track.top, h: track.height - THUMB_H, max: root.scrollHeight - innerHeight };
  }

  function refresh() {
    enabled = !!grid && root.scrollHeight > innerHeight * 4;
    root.classList.toggle('has-scrub', enabled);
    el.classList.toggle('on', enabled && finePointer.matches);
    update();
  }

  // Where the thumb goes and what the label says for the current
  // scroll position: measure() only reads the layout and place() only
  // writes, and only what changed. The grid calls them each frame it
  // draws (see its tick), around its own reads and writes.
  var shownY = null, shownLabel = null;
  function measure() {
    if (!enabled) return null;
    var m = metrics();
    var ratio = m.max > 0 ? scrollY / m.max : 0;
    return { y: ratio * m.h, label: grid.labelAt(scrollY + innerHeight / 2) };
  }

  function place(s) {
    if (!s) return;
    if (s.y !== shownY) {
      shownY = s.y;
      thumb.style.transform = 'translateY(' + s.y + 'px)';
      label.style.top = (s.y + THUMB_H / 2) + 'px';
    }
    if (s.label !== shownLabel) {
      shownLabel = s.label;
      label.textContent = s.label;
    }
  }

  function update() {
    place(measure());
  }

  function show() {
    if (!enabled) return;
    el.classList.add('on');
    clearTimeout(hideTimer);
    // Mouse users keep the scrubber; on touch it appears while scrolling.
    if (!dragging && !finePointer.matches) hideTimer = setTimeout(function () { el.classList.remove('on'); }, 1400);
  }

  addEventListener('scroll', show, { passive: true });
  addEventListener('resize', refresh, { passive: true });
  el.addEventListener('pointerleave', show);

  thumb.addEventListener('pointerdown', function (e) {
    e.preventDefault();
    dragging = true;
    thumb.setPointerCapture(e.pointerId);
    grab = e.clientY - thumb.getBoundingClientRect().top;
    el.classList.add('drag');
    if (grid) grid.setFast(true);
    show();
  });
  thumb.addEventListener('pointermove', function (e) {
    if (!dragging) return;
    var m = metrics();
    var ratio = Math.min(1, Math.max(0, (e.clientY - grab - m.top) / m.h));
    scrollTo(0, ratio * m.max);
  });
  var end = function () {
    if (!dragging) return;
    dragging = false;
    el.classList.remove('drag');
    if (grid) grid.setFast(false);
    show();
  };
  thumb.addEventListener('pointerup', end);
  thumb.addEventListener('pointercancel', end);

  return { refresh: refresh, measure: measure, place: place };
})();
