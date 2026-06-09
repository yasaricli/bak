/* ── Live reload (SSE) ──────────────────────────── */
(function() {
  var dot = document.getElementById('live-dot');
  function setConnected(on) {
    if (!dot) return;
    dot.className = on ? 'connected' : 'disconnected';
  }
  try {
    var es = new EventSource('/events');
    es.addEventListener('open', function() { setConnected(true); });
    es.onmessage = function(e) {
      if (e.data === 'update') location.reload();
    };
    es.onerror = function() { setConnected(false); };
  } catch(e) {}
})();

/* ── Collapse file ──────────────────────────────── */
function toggleFile(header) {
  header.closest('.diff-file').classList.toggle('collapsed');
}

/* ── Collapse hunk ──────────────────────────────── */
function toggleHunk(hunkRow) {
  var rows = Array.from(hunkRow.closest('table').querySelectorAll('tr'));
  var idx = rows.indexOf(hunkRow);
  var col = hunkRow.classList.toggle('hunk-collapsed');
  hunkRow.querySelector('.hunk-chevron').style.transform = col ? 'rotate(-90deg)' : '';
  for (var i = idx + 1; i < rows.length; i++) {
    if (rows[i].classList.contains('line-hunk')) break;
    rows[i].style.display = col ? 'none' : '';
  }
}

/* ── Jump to file ───────────────────────────────── */
function jumpTo(id, el) {
  document.querySelectorAll('.file-item').forEach(function(e) { e.classList.remove('active'); });
  if (el) el.classList.add('active');
  var t = document.getElementById(id);
  if (!t) return;
  if (t.classList.contains('collapsed')) t.classList.remove('collapsed');
  t.scrollIntoView({ behavior: 'smooth', block: 'start' });
}

/* ── Scroll spy ─────────────────────────────────── */
(function () {
  var main = document.getElementById('main');
  var items = document.querySelectorAll('.diff-file');
  main.addEventListener('scroll', function () {
    var top = main.scrollTop + 100;
    var active = 0;
    items.forEach(function (el, i) { if (el.offsetTop <= top) active = i; });
    document.querySelectorAll('.file-item').forEach(function (el, i) {
      el.classList.toggle('active', i === active);
    });
    var nav = document.getElementById('nav-' + active);
    if (nav) nav.scrollIntoView({ block: 'nearest' });
  });
})();

/* ── Syntax highlight ───────────────────────────── */
(function () {
  document.querySelectorAll('.diff-file[data-lang]').forEach(function (section) {
    var lang = section.dataset.lang;
    if (!lang) return;
    var rows = Array.from(section.querySelectorAll('tr.line-add, tr.line-del, tr.line-ctx'));
    if (!rows.length) return;
    var code = rows.map(function (r) { return r.querySelector('.lc').textContent.slice(1); }).join('\n');
    var result;
    try { result = hljs.highlight(code, { language: lang, ignoreIllegals: true }); } catch (e) { return; }
    var hLines = result.value.split('\n');
    rows.forEach(function (row, i) {
      var lc = row.querySelector('.lc');
      var sign = lc.querySelector('.sign').outerHTML;
      lc.innerHTML = sign + (hLines[i] !== undefined ? hLines[i] : '');
    });
  });
})();

/* ── Theme ──────────────────────────────────────── */
function applyTheme(t) {
  document.documentElement.classList.toggle('light', t === 'light');
  document.getElementById('btn-theme').textContent = t === 'light' ? 'Dark' : 'Light';
  var link = document.getElementById('hljs-css');
  link.href = t === 'light'
    ? 'https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/github.min.css'
    : 'https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/github-dark.min.css';
}
function toggleTheme() {
  var isLight = document.documentElement.classList.contains('light');
  var t = isLight ? 'dark' : 'light';
  localStorage.setItem('bak-theme', t);
  applyTheme(t);
}
applyTheme(localStorage.getItem('bak-theme') || 'dark');

/* ── Font size ──────────────────────────────────── */
var _fs = parseInt(localStorage.getItem('bak-fs') || '12');
function applyFS() { document.documentElement.style.setProperty('--fs', _fs + 'px'); }
function changeFontSize(d) { _fs = Math.max(9, Math.min(22, _fs + d)); localStorage.setItem('bak-fs', _fs); applyFS(); }
applyFS();

/* ── Split view ─────────────────────────────────── */
var _split = false;
var _splitBuilt = {};

function buildSplitView(section) {
  if (_splitBuilt[section.id]) return;
  _splitBuilt[section.id] = true;
  var unifiedBody = section.querySelector('.diff-body:not(.split-body)');
  if (!unifiedBody) return;
  var rows = Array.from(unifiedBody.querySelectorAll('tr'));
  var html = '<table class="diff-table"><colgroup><col style="width:1%"><col style="width:1%"><col style="width:50%"><col style="width:1%"><col style="width:1%"><col></colgroup>';
  var i = 0;
  while (i < rows.length) {
    var row = rows[i];
    if (row.classList.contains('line-hunk')) {
      html += '<tr class="line-hunk" onclick="toggleHunk(this)"><td colspan="6">' + row.querySelector('td').innerHTML + '</td></tr>';
      i++;
    } else if (row.classList.contains('line-ctx')) {
      var tds = row.querySelectorAll('td');
      var code = tds[2].innerHTML;
      html += '<tr class="line-ctx"><td class="ln">' + tds[0].textContent + '</td><td class="lc split-lc">' + code + '</td><td class="ln">' + tds[1].textContent + '</td><td class="lc split-lc">' + code + '</td></tr>';
      i++;
    } else {
      var dels = [], adds = [];
      while (i < rows.length && (rows[i].classList.contains('line-del') || rows[i].classList.contains('line-add'))) {
        if (rows[i].classList.contains('line-del')) dels.push(rows[i]);
        else adds.push(rows[i]);
        i++;
      }
      var max = Math.max(dels.length, adds.length);
      for (var j = 0; j < max; j++) {
        html += '<tr>';
        if (dels[j]) {
          var dt = dels[j].querySelectorAll('td');
          html += '<td class="ln line-del">' + dt[0].textContent + '</td><td class="lc split-lc line-del">' + dt[2].innerHTML + '</td>';
        } else {
          html += '<td class="ln"></td><td class="lc split-lc"></td>';
        }
        if (adds[j]) {
          var at = adds[j].querySelectorAll('td');
          html += '<td class="ln line-add">' + at[1].textContent + '</td><td class="lc split-lc line-add">' + at[2].innerHTML + '</td>';
        } else {
          html += '<td class="ln"></td><td class="lc split-lc"></td>';
        }
        html += '</tr>';
      }
    }
  }
  html += '</table>';
  var div = document.createElement('div');
  div.className = 'diff-body split-body';
  div.innerHTML = html;
  div.style.display = 'none';
  unifiedBody.after(div);
}

function toggleView() {
  _split = !_split;
  document.getElementById('btn-view').textContent = _split ? 'Unified' : 'Split';
  document.querySelectorAll('.diff-file').forEach(function (section) {
    if (_split) buildSplitView(section);
    var u = section.querySelector('.diff-body:not(.split-body)');
    var s = section.querySelector('.split-body');
    if (u) u.style.display = _split ? 'none' : '';
    if (s) s.style.display = _split ? '' : 'none';
  });
}

/* ── Search ─────────────────────────────────────── */
var _matches = [], _matchIdx = 0;

function openSearch() {
  var bar = document.getElementById('search-bar');
  bar.classList.add('open');
  document.getElementById('search-input').focus();
}
function closeSearch() {
  document.getElementById('search-bar').classList.remove('open');
  document.querySelectorAll('.lc.search-match,.lc.search-cur').forEach(function(el) {
    el.classList.remove('search-match','search-cur');
  });
  _matches = [];
  document.getElementById('search-count').textContent = '';
}
function doSearch(q) {
  document.querySelectorAll('.lc.search-match,.lc.search-cur').forEach(function(el) {
    el.classList.remove('search-match','search-cur');
  });
  _matches = [];
  if (!q) { document.getElementById('search-count').textContent = ''; return; }
  var re = new RegExp(q.replace(/[.*+?^${}()|[\]\\]/g,'\\$&'), 'i');
  document.querySelectorAll('.diff-body:not(.split-body) .lc, .split-body .lc').forEach(function(lc) {
    var body = lc.closest('.diff-body');
    if (body && body.style.display === 'none') return;
    if (re.test(lc.textContent)) { lc.classList.add('search-match'); _matches.push(lc); }
  });
  _matchIdx = 0;
  updateSearchCur();
}
function updateSearchCur() {
  document.querySelectorAll('.lc.search-cur').forEach(function(el) { el.classList.remove('search-cur'); });
  if (!_matches.length) { document.getElementById('search-count').textContent = 'No results'; return; }
  var cur = _matches[_matchIdx];
  cur.classList.add('search-cur');
  var section = cur.closest('.diff-file');
  if (section && section.classList.contains('collapsed')) section.classList.remove('collapsed');
  cur.scrollIntoView({ behavior: 'smooth', block: 'center' });
  document.getElementById('search-count').textContent = (_matchIdx+1) + ' / ' + _matches.length;
}
function searchNav(d) { if (!_matches.length) return; _matchIdx = (_matchIdx+d+_matches.length)%_matches.length; updateSearchCur(); }

document.getElementById('search-input').addEventListener('input', function() { doSearch(this.value); });
document.getElementById('search-input').addEventListener('keydown', function(e) {
  if (e.key==='Enter') { e.shiftKey ? searchNav(-1) : searchNav(1); }
  if (e.key==='Escape') closeSearch();
});

/* ── File filter ────────────────────────────────── */
function filterFiles(q) {
  var lower = q.toLowerCase();
  document.querySelectorAll('.file-item').forEach(function(item) {
    var name = item.querySelector('.file-name').textContent.toLowerCase();
    item.style.display = (!q || name.includes(lower)) ? '' : 'none';
  });
}

/* ── Permalink ──────────────────────────────────── */
function permalink(td, fileIdx, lineNum) {
  var hash = '#file-' + fileIdx + '-L' + lineNum;
  history.pushState(null, '', hash);
  navigator.clipboard.writeText(location.href).catch(function(){});
  showToast('Link copied!');
}
(function() {
  var m = location.hash.match(/#file-(\d+)-L(\d+)/);
  if (!m) return;
  var section = document.getElementById('file-' + m[1]);
  if (!section) return;
  if (section.classList.contains('collapsed')) section.classList.remove('collapsed');
  setTimeout(function() {
    var lns = section.querySelectorAll('.ln');
    for (var i = 0; i < lns.length; i++) {
      if (lns[i].textContent.trim() === m[2]) {
        lns[i].scrollIntoView({ behavior: 'smooth', block: 'center' });
        lns[i].closest('tr').classList.add('permalink-hl');
        break;
      }
    }
  }, 300);
})();

/* ── Export ─────────────────────────────────────── */
function exportHTML() {
  var content = '<!DOCTYPE html>' + document.documentElement.outerHTML;
  var blob = new Blob([content], { type: 'text/html; charset=utf-8' });
  var url = URL.createObjectURL(blob);
  var a = document.createElement('a');
  a.href = url; a.download = 'bak-' + new Date().toISOString().slice(0,10) + '.html';
  document.body.appendChild(a); a.click(); document.body.removeChild(a);
  URL.revokeObjectURL(url);
}

/* ── Help modal ─────────────────────────────────── */
function showHelp() { document.getElementById('help-modal').hidden = false; }
function hideHelp() { document.getElementById('help-modal').hidden = true; }

/* ── Toast ──────────────────────────────────────── */
function showToast(msg) {
  var t = document.getElementById('toast');
  t.textContent = msg; t.classList.add('show');
  setTimeout(function() { t.classList.remove('show'); }, 2000);
}

/* ── Keyboard shortcuts ─────────────────────────── */
document.addEventListener('keydown', function(e) {
  if (e.target.tagName === 'INPUT') return;
  var hunks = Array.from(document.querySelectorAll('.line-hunk:not([style*="display: none"])'));
  var files = Array.from(document.querySelectorAll('.diff-file'));
  switch(e.key) {
    case 'j': {
      var cur = hunks.findIndex(function(h) { return h.getBoundingClientRect().top > 10; });
      if (cur >= 0) { hunks[cur].scrollIntoView({ behavior:'smooth', block:'start' }); } break;
    }
    case 'k': {
      var above = hunks.filter(function(h) { return h.getBoundingClientRect().top < -5; });
      if (above.length) { above[above.length-1].scrollIntoView({ behavior:'smooth', block:'start' }); } break;
    }
    case 'n': {
      var cf = files.findIndex(function(f) { return f.getBoundingClientRect().top > 10; });
      if (cf >= 0) files[cf].scrollIntoView({ behavior:'smooth', block:'start' }); break;
    }
    case 'p': {
      var pf = files.filter(function(f) { return f.getBoundingClientRect().top < -5; });
      if (pf.length) pf[pf.length-1].scrollIntoView({ behavior:'smooth', block:'start' }); break;
    }
    case '/': e.preventDefault(); openSearch(); break;
    case 't': toggleTheme(); break;
    case 's': toggleView(); break;
    case 'e': exportHTML(); break;
    case '?': showHelp(); break;
    case 'Escape':
      if (!document.getElementById('help-modal').hidden) { hideHelp(); break; }
      closeSearch(); break;
  }
});
