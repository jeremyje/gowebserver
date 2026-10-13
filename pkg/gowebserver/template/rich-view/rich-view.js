var doc = document, root = doc.documentElement;
function $(id) { return doc.getElementById(id); }
function store(key, value) {
  try {
    if (value === null) localStorage.removeItem(key); else localStorage.setItem(key, value);
  } catch (e) { }
}
var select = $('code-theme');
if (select) select.addEventListener('change', function () { location.href = select.value + location.hash; });

var wrap = $('wrap');
if (wrap) {
  var syncWrap = function () { wrap.setAttribute('aria-pressed', String(root.classList.contains('wrap'))); };
  syncWrap();
  wrap.addEventListener('click', function () {
    var on = root.classList.toggle('wrap');
    store('gows-wrap', on ? '1' : null);
    syncWrap();
  });
}

var copy = $('copy'), code = $('code');
if (copy && code && navigator.clipboard && window.isSecureContext) {
  copy.hidden = false;
  copy.addEventListener('click', function () {
    var text = Array.prototype.map.call(code.querySelectorAll('.cl'), function (l) { return l.textContent; }).join('');
    navigator.clipboard.writeText(text).then(function () {
      copy.classList.add('done');
      copy.title = 'Copied';
      setTimeout(function () { copy.classList.remove('done'); copy.title = 'Copy contents'; }, 1500);
    }, function () { });
  });
}
