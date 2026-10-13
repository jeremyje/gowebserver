// ===== Filter =====
(function () {
  var input = $('filter'), timer = 0;
  input.addEventListener('input', function () {
    clearTimeout(timer);
    timer = setTimeout(function () { applyFilter(input.value); }, 120);
  });
  input.addEventListener('keydown', function (e) {
    if (e.key === 'Escape' && input.value) { input.value = ''; applyFilter(''); e.stopPropagation(); }
  });
  doc.addEventListener('keydown', function (e) {
    if (e.key === '/' && !e.metaKey && !e.ctrlKey && doc.activeElement.tagName !== 'INPUT' && !doc.querySelector('dialog[open]')) {
      e.preventDefault();
      input.focus();
    }
  });
})();

applyFilter($('filter').value || '');
