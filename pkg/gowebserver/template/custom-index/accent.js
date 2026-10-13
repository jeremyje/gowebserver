// ===== Accent from the backdrop =====
// The most vivid hue across some media elements, as OKLCH hue
// (degrees) and chroma, or null when they are all gray or muted.
// Neutral pixels (grays, beiges, skin, sky haze) are ignored outright,
// and the rest are weighted by chroma cubed, so any vibrant color on
// screen beats large muted areas. Each source is {el, w} where w
// scales its influence. Lightness is left to the theme (--accent-l).
var vividCanvas = null, vividLin = null;
var VIVID_BINS = 24;

// Something that draws el at w by h without holding the page up.
// Drawing a photo decodes it then and there, and a large one (the
// tiles show the files themselves when the server can't resize them)
// takes long enough to freeze the page, a screen of them for seconds.
// So a large photo is decoded off to the side instead, into a bitmap
// of that size, from the file the browser already has: the cache is
// the only place it's fetched from, never the network again. (A bitmap
// made from the element itself would be decoded right here too.) A
// few at a time, since each is a whole file in memory while it's
// decoded. Anything else, and a photo where that isn't possible, is
// drawn as it is.
var SMALL_PICTURE_OVER = 2e6, smallQueue = [], smallActive = 0, SMALL_CONCURRENCY = 2;
function smallPicture(el, w, h) {
  if (el.tagName !== 'IMG' || !window.createImageBitmap || !window.fetch ||
    !(el.naturalWidth * el.naturalHeight > SMALL_PICTURE_OVER)) return Promise.resolve(el);
  var src = el.currentSrc;
  return new Promise(function (resolve) {
    smallQueue.push(function () {
      return fetch(src, { cache: 'only-if-cached', mode: 'same-origin' }).then(function (res) {
        if (!res.ok) throw new Error('not cached');
        return res.blob();
      }).then(function (blob) {
        return createImageBitmap(blob, { resizeWidth: w, resizeHeight: h, resizeQuality: 'low' });
      }).catch(function () {
        // Not the picture it was (its tile scrolled away): nothing to draw.
        return el.currentSrc === src ? el : null;
      }).then(resolve);
    });
    pumpSmall();
  });
}

function pumpSmall() {
  while (smallActive < SMALL_CONCURRENCY && smallQueue.length) {
    smallActive++;
    smallQueue.shift()().then(function () { smallActive--; pumpSmall(); });
  }
}

// One element's colors, by hue bin: four runs of VIVID_BINS numbers
// (weight, weighted cos and sin of the hue, weighted chroma), or null
// if it can't be read. A loaded photo never changes, so its tally is
// kept by its URL and the photo is read once: not every time the page
// stops scrolling with it still on screen, and not again when its row
// scrolls away and comes back as new elements.
var vividSeen = new Map(), VIVID_SEEN_MAX = 4000;
function vividTally(el) {
  var still = el.tagName === 'IMG' && el.complete && el.currentSrc;
  if (still && vividSeen.has(still)) return Promise.resolve(vividSeen.get(still));
  return smallPicture(el, 16, 12).then(function (pic) {
    var t = pic && vividCount(pic);
    if (pic && pic.close) pic.close();
    if (t && still) {
      // The oldest goes first; a Map keeps the order they were added in.
      if (vividSeen.size >= VIVID_SEEN_MAX) vividSeen.delete(vividSeen.keys().next().value);
      vividSeen.set(still, t);
    }
    return t || null;
  });
}

function vividCount(pic) {
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
    ctx.clearRect(0, 0, vividCanvas.width, vividCanvas.height);
    ctx.drawImage(pic, 0, 0, vividCanvas.width, vividCanvas.height);
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
  return t;
}

// Resolves to that color once every source has been read.
function vividColor(sources) {
  return Promise.all(sources.map(function (src) { return vividTally(src.el); })).then(function (tallies) {
    return vividPick(sources, tallies);
  });
}

function vividPick(sources, tallies) {
  var BINS = VIVID_BINS;
  var w = new Float64Array(BINS), sx = new Float64Array(BINS),
    sy = new Float64Array(BINS), sc = new Float64Array(BINS);
  sources.forEach(function (src, i) {
    var t = tallies[i];
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
  var src = {}, asked = {}, lastH = null, shownH = '', shownC = '', on = false;
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
    // Takes key's color from sources ({el, w}, as for vividColor) once
    // they have been read; only the latest request for a key counts.
    from: function (key, sources) {
      var n = asked[key] = (asked[key] || 0) + 1;
      vividColor(sources).then(function (col) {
        if (asked[key] !== n) return;
        src[key] = col;
        apply();
      });
    },
    clear: function (key) {
      asked[key] = (asked[key] || 0) + 1;
      delete src[key];
      apply();
    }
  };
})();

// ===== Ambient backdrop =====
// A media element as a soft wash of its colors: a 24x16 copy of it,
// enlarged and blurred. The blur is done here, once, into a canvas
// still small enough that it costs almost nothing to keep or to draw;
// CSS only stretches that over the page. (As a CSS filter the blur was
// redone over the whole viewport for every frame of a scroll.) The
// amount of blur and the color adjustment are the backdrop's
// --backdrop-blur and --backdrop-tone. Where a canvas can't filter,
// .css-blur has CSS do it after all.
// onColor is called with the element each time the backdrop changes.
var ambientTiny = null;
function Ambient(el, onColor) {
  var canvases = el ? el.querySelectorAll('canvas') : [];
  var cur = 0, last = null, blur = 0, tone = '', filters = null;

  function setUp() {
    filters = 'filter' in canvases[0].getContext('2d');
    if (!filters) { el.classList.add('css-blur'); return; }
    var css = getComputedStyle(el);
    blur = parseFloat(css.getPropertyValue('--backdrop-blur')) || 0;
    tone = css.getPropertyValue('--backdrop-tone').trim();
  }

  function draw(pic, canvas) {
    if (!ambientTiny) {
      ambientTiny = doc.createElement('canvas');
      ambientTiny.width = 24;
      ambientTiny.height = 16;
    }
    var tiny = ambientTiny.getContext('2d'), ctx = canvas.getContext('2d');
    tiny.clearRect(0, 0, ambientTiny.width, ambientTiny.height);
    tiny.drawImage(pic, 0, 0, ambientTiny.width, ambientTiny.height);
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    if (filters) {
      // The blur is in CSS pixels of the backdrop as shown; here it's
      // in the canvas's own, which are that much smaller.
      var shown = el.offsetWidth || innerWidth;
      ctx.filter = ('blur(' + (blur * canvas.width / shown).toFixed(2) + 'px) ' + tone).trim();
    }
    ctx.drawImage(ambientTiny, 0, 0, canvas.width, canvas.height);
  }

  return {
    show: function (src) {
      if (!el || !src || src === last) return;
      last = src;
      smallPicture(src, 24, 16).then(function (pic) {
        // Another one was asked for meanwhile.
        if (last !== src) { if (pic && pic.close) pic.close(); return; }
        var next = canvases[1 - cur];
        try {
          if (!pic) throw new Error('nothing to draw');
          if (filters === null) setUp();
          draw(pic, next);
        } catch (e) {
          last = null;
          return;
        } finally {
          if (pic && pic.close) pic.close();
        }
        if (onColor) onColor(src);
        next.classList.add('on');
        canvases[cur].classList.remove('on');
        cur = 1 - cur;
        el.classList.add('on');
      });
    },
    reset: function () { last = null; }
  };
}
var pageAmbient = Ambient($('ambient'));
