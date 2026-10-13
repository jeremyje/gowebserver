var isAdvancedUpload = (function () {
  var div = document.createElement('div');
  return (('draggable' in div) || ('ondragstart' in div && 'ondrop' in div)) &&
    'FormData' in window && 'FileReader' in window;
})();

var form        = document.querySelector('form');
var dropZone    = document.getElementById('drop-zone');
var input       = document.getElementById('file-input');
var labelStrong = document.getElementById('drop-label-strong');
var labelRest   = document.getElementById('drop-label-rest');
var errorDetail = document.getElementById('error-detail');
var restartBtns = dropZone.querySelectorAll('.restart');
var droppedFiles = false;

function showFiles(files) {
  if (!files || files.length === 0) return;
  labelStrong.textContent = files.length > 1 ? files.length + ' files' : files[0].name;
  labelRest.textContent = ' selected';
}

function resetLabel() {
  labelStrong.textContent = 'Choose files';
  labelRest.textContent = ' or drag and drop here';
}

function triggerSubmit() {
  var ev = document.createEvent('HTMLEvents');
  // Cancelable: Firefox submits the form for this event as it would for
  // a real one, unless the submit handler below can prevent that.
  ev.initEvent('submit', true, true);
  form.dispatchEvent(ev);
}

/* Ripple: remove + reflow + add so animation restarts every drop */
function flashRipple() {
  dropZone.classList.remove('is-dropped');
  void dropZone.offsetWidth;
  dropZone.classList.add('is-dropped');
  setTimeout(function () { dropZone.classList.remove('is-dropped'); }, 600);
}

/* Ajax flag so the server responds with JSON */
var ajaxFlag = document.createElement('input');
ajaxFlag.type  = 'hidden';
ajaxFlag.name  = 'ajax';
ajaxFlag.value = '1';
form.appendChild(ajaxFlag);

/* Auto-submit when files are picked via dialog */
input.addEventListener('change', function (e) {
  showFiles(e.target.files);
  triggerSubmit();
});

/* Drag and drop */
if (isAdvancedUpload) {
  ['drag', 'dragstart', 'dragend', 'dragover', 'dragenter', 'dragleave', 'drop'].forEach(function (ev) {
    dropZone.addEventListener(ev, function (e) { e.preventDefault(); e.stopPropagation(); });
  });

  ['dragover', 'dragenter'].forEach(function (ev) {
    dropZone.addEventListener(ev, function () { dropZone.classList.add('is-dragover'); });
  });

  ['dragleave', 'dragend', 'drop'].forEach(function (ev) {
    dropZone.addEventListener(ev, function () { dropZone.classList.remove('is-dragover'); });
  });

  dropZone.addEventListener('drop', function (e) {
    droppedFiles = e.dataTransfer.files;
    showFiles(droppedFiles);
    flashRipple();
    triggerSubmit();
  });
}

/* Click anywhere on the zone (except label / status overlay) to open picker */
dropZone.addEventListener('click', function (e) {
  if (!e.target.closest('.status') && !e.target.closest('label')) {
    input.click();
  }
});

/* Form submit — AJAX upload */
form.addEventListener('submit', function (e) {
  if (dropZone.classList.contains('is-uploading')) return false;

  dropZone.classList.remove('is-error', 'is-success');
  dropZone.classList.add('is-uploading');

  if (isAdvancedUpload) {
    e.preventDefault();

    var data = new FormData(form);
    if (droppedFiles) {
      Array.prototype.forEach.call(droppedFiles, function (file) {
        data.append(input.name, file);
      });
    }

    var xhr = new XMLHttpRequest();
    xhr.open(form.method, form.action, true);

    xhr.onload = function () {
      dropZone.classList.remove('is-uploading');
      if (xhr.status >= 200 && xhr.status < 400) {
        var resp = JSON.parse(xhr.responseText);
        dropZone.classList.add(resp.success ? 'is-success' : 'is-error');
        if (!resp.success) errorDetail.textContent = resp.error || '';
      } else {
        dropZone.classList.add('is-error');
        errorDetail.textContent = 'Server error (' + xhr.status + ').';
      }
    };

    xhr.onerror = function () {
      dropZone.classList.remove('is-uploading');
      dropZone.classList.add('is-error');
      errorDetail.textContent = 'Network error — please try again.';
    };

    xhr.send(data);
  }
});

/* Restart: clear state and re-open picker */
restartBtns.forEach(function (btn) {
  btn.addEventListener('click', function (e) {
    e.preventDefault();
    dropZone.classList.remove('is-error', 'is-success');
    droppedFiles = false;
    input.value  = '';
    resetLabel();
    input.click();
  });
});

input.addEventListener('focus', function () { input.classList.add('has-focus'); });
input.addEventListener('blur',  function () { input.classList.remove('has-focus'); });
