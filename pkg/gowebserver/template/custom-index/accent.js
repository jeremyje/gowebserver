// ===== Accent from the backdrop =====
// The most vivid hue across some media elements, as OKLCH hue
// (degrees) and chroma, or null when they are all gray or muted.
// Neutral pixels (grays, beiges, skin, sky haze) are ignored outright,
// and the rest are weighted by chroma cubed, so any vibrant color on
// screen beats large muted areas. Each source is {el, w} where w
// scales its influence. Lightness is left to the theme (--accent-l).
var vividCanvas = null, vividLin = null;
var VIVID_BINS = 24;
// One element's colors, by hue bin: four runs of VIVID_BINS numbers
// (weight, weighted cos and sin of the hue, weighted chroma), or null
// if it can't be read. A loaded photo never changes, so its tally is
// kept on the element and the photo is drawn and read back once, not
// every time the page stops scrolling with it still on screen.
function vividTally(el) {
  var still = el.tagName === 'IMG' && el.complete && el.currentSrc;
  if (still && el._vivid && el._vividSrc === el.currentSrc) return el._vivid;
  if (!vividCanvas) {
    vividCanvas = doc.createElement('canvas');
    vividCanvas.width = 16;
    vividCanvas.height = 12;
    // sRGB to linear, for each of the 256 values a channel can have.
    vividLin = new Float64Array(256);
    for (var v = 0; v < 256; v++) {
      var f = v / 255;
      vividLin[v] = f <= 0.04045 ? f / 12.92 : Math.pow((f + 0.055) / 1.055, 2.4);
    }
  }
  var ctx = vividCanvas.getContext('2d', { willReadFrequently: true });
  var BINS = VIVID_BINS, MIN_C = 0.08, px;
  try {
    ctx.drawImage(el, 0, 0, vividCanvas.width, vividCanvas.height);
    px = ctx.getImageData(0, 0, vividCanvas.width, vividCanvas.height).data;
  } catch (e) { return null; }
  var t = new Float64Array(BINS * 4);
  for (var i = 0; i < px.length; i += 4) {
    var r = vividLin[px[i]], g = vividLin[px[i + 1]], b = vividLin[px[i + 2]];
    // Linear sRGB to OKLab.
    var l = Math.cbrt(0.4122214708 * r + 0.5363325363 * g + 0.0514459929 * b);
    var m = Math.cbrt(0.2119034982 * r + 0.6806995451 * g + 0.1073969566 * b);
    var s = Math.cbrt(0.0883024619 * r + 0.2817188376 * g + 0.6299787005 * b);
    var L = 0.2104542553 * l + 0.793617785 * m - 0.0040720468 * s;
    var A = 1.9779984951 * l - 2.428592205 * m + 0.4505937099 * s;
    var B = 0.0259040371 * l + 0.7827717662 * m - 0.808675766 * s;
    var c = Math.sqrt(A * A + B * B);
    if (c < MIN_C || L < 0.3 || L > 0.95) continue;
    var hue = Math.atan2(B, A);
    var bin = Math.floor(((hue / (2 * Math.PI) + 1) % 1) * BINS) % BINS;
    var wt = c * c * c;
    t[bin] += wt; t[BINS + bin] += wt * Math.cos(hue); t[2 * BINS + bin] += wt * Math.sin(hue); t[3 * BINS + bin] += wt * c;
  }
  if (still) { el._vivid = t; el._vividSrc = el.currentSrc; }
  return t;
}

function vividColor(sources) {
  var BINS = VIVID_BINS;
  var w = new Float64Array(BINS), sx = new Float64Array(BINS),
    sy = new Float64Array(BINS), sc = new Float64Array(BINS);
  sources.forEach(function (src) {
    var t = vividTally(src.el);
    if (!t) return;
    for (var k = 0; k < BINS; k++) {
      w[k] += t[k] * src.w; sx[k] += t[BINS + k] * src.w;
      sy[k] += t[2 * BINS + k] * src.w; sc[k] += t[3 * BINS + k] * src.w;
    }
  });
  var best = -1, bestW = 0;
  for (var k = 0; k < BINS; k++) {
    var score = w[k] + 0.5 * (w[(k + 1) % BINS] + w[(k + BINS - 1) % BINS]);
    if (score > bestW) { bestW = score; best = k; }
  }
  // A few clearly colorful pixels; a lone speck doesn't count.
  if (best < 0 || bestW < 0.006) return null;
  var h = Math.atan2(sy[best], sx[best]) * 180 / Math.PI;
  // Never hand back a muted color: chroma is floored well above gray.
  return { h: (h + 360) % 360, c: Math.min(0.22, Math.max(0.14, 1.1 * sc[best] / w[best])) };
}

// Accent sources: the photos on the page, and the current slide while
// the slideshow is open. A slide with no vivid color falls back to the
// page's color, and a page with none to the default accent.
var accent = (function () {
  var src = {}, lastH = null, shownH = '', shownC = '', on = false;
  // Every element inherits these, so each write restyles the whole
  // page: only what actually changed is written.
  function apply() {
    var col = src.ss || src.page;
    if (!col) {
      if (on) { on = false; root.classList.remove('amb-accent'); }
      return;
    }
    // Take the short way around the hue circle when transitioning.
    var h = col.h;
    if (lastH !== null) h = lastH + (((h - lastH) % 360) + 540) % 360 - 180;
    lastH = h;
    var hs = h.toFixed(1), cs = col.c.toFixed(3);
    if (hs !== shownH) { shownH = hs; root.style.setProperty('--amb-h', hs); }
    if (cs !== shownC) { shownC = cs; root.style.setProperty('--amb-c', cs); }
    if (!on) { on = true; root.classList.add('amb-accent'); }
  }
  return {
    set: function (key, col) { src[key] = col; apply(); },
    clear: function (key) { delete src[key]; apply(); }
  };
})();

// ===== Ambient backdrop =====
// Draws a media element into a tiny canvas; CSS scales and blurs it.
// onColor is called with the element each time the backdrop changes.
function Ambient(el, onColor) {
  var canvases = el ? el.querySelectorAll('canvas') : [];
  var cur = 0, last = null;
  return {
    show: function (src) {
      if (!el || !src || src === last) return;
      var next = canvases[1 - cur];
      try {
        next.getContext('2d').drawImage(src, 0, 0, next.width, next.height);
      } catch (e) { return; }
      last = src;
      if (onColor) onColor(src);
      next.classList.add('on');
      canvases[cur].classList.remove('on');
      cur = 1 - cur;
      el.classList.add('on');
    },
    reset: function () { last = null; }
  };
}
var pageAmbient = Ambient($('ambient'));
