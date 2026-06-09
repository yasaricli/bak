// Package render builds the self-contained HTML page for a diff.
package render

import (
	"encoding/base64"
	"html"
	"strconv"
	"strings"

	"github.com/yasaricli/bak/internal/diff"
)

const githubRepo = "https://github.com/yasaricli/bak"

func HTML(files []diff.File, title, branch string, logoData []byte) string {
	totalAdd, totalDel := 0, 0
	for _, f := range files {
		totalAdd += f.Added
		totalDel += f.Removed
	}

	var b strings.Builder
	b.WriteString(htmlHead(title))
	b.WriteString(`<body>`)
	b.WriteString(sidebar(files, branch))
	b.WriteString(`<div id="right">`)
	b.WriteString(toolbar(len(files), totalAdd, totalDel))
	b.WriteString(mainContent(files))
	b.WriteString(`</div>`)
	b.WriteString(footer())
	return b.String()
}

func htmlHead(title string) string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>` + html.EscapeString(title) + `</title>
<link id="hljs-css" rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/github-dark.min.css">
<style>
` + css() + `
</style>
</head>
`
}

func sidebarFooter() string {
	return `<div id="sidebar-footer">` +
		`<div id="live-indicator" title="Watching for changes">` +
		`<span id="live-dot"></span>` +
		`<span id="live-label">Live</span>` +
		`</div>` +
		`<a href="` + githubRepo + `" target="_blank" rel="noopener noreferrer" id="github-btn">` +
		`<svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor">` +
		`<path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"/>` +
		`</svg>` +
		`<svg width="11" height="11" viewBox="0 0 16 16" fill="currentColor" class="star-svg">` +
		`<path d="M8 .25a.75.75 0 0 1 .673.418l1.882 3.815 4.21.612a.75.75 0 0 1 .416 1.279l-3.046 2.97.719 4.192a.751.751 0 0 1-1.088.791L8 12.347l-3.766 1.98a.75.75 0 0 1-1.088-.79l.72-4.194L.818 6.374a.75.75 0 0 1 .416-1.28l4.21-.611L7.327.668A.75.75 0 0 1 8 .25Z"/>` +
		`</svg>` +
		`<span>Star</span>` +
		`</a>` +
		`</div>`
}

func toolbar(nFiles, totalAdd, totalDel int) string {
	var b strings.Builder
	b.WriteString(`<div id="toolbar">`)
	b.WriteString(`<div id="toolbar-stats">`)
	if totalAdd > 0 {
		b.WriteString(`<span class="ts-add">+` + strconv.Itoa(totalAdd) + `</span>`)
	}
	if totalDel > 0 {
		b.WriteString(`<span class="ts-del">−` + strconv.Itoa(totalDel) + `</span>`)
	}
	b.WriteString(`<span class="ts-files">` + strconv.Itoa(nFiles) + ` files</span>`)
	b.WriteString(`</div>`)
	b.WriteString(`<div id="toolbar-actions">`)
	b.WriteString(`<button class="tb-btn" onclick="openSearch()" title="Search (/)">Search</button>`)
	b.WriteString(`<button class="tb-btn" id="btn-view" onclick="toggleView()" title="Toggle split view (s)">Split</button>`)
	b.WriteString(`<button class="tb-btn" id="btn-theme" onclick="toggleTheme()" title="Toggle theme (t)">Light</button>`)
	b.WriteString(`<button class="tb-btn" onclick="changeFontSize(-1)" title="Decrease font size">A−</button>`)
	b.WriteString(`<button class="tb-btn" onclick="changeFontSize(1)" title="Increase font size">A+</button>`)
	b.WriteString(`<button class="tb-btn" onclick="exportHTML()" title="Export as HTML (e)">Export</button>`)
	b.WriteString(`<button class="tb-btn" onclick="showHelp()" title="Keyboard shortcuts (?)">?</button>`)
	b.WriteString(`</div></div>`)
	return b.String()
}

func sidebar(files []diff.File, branch string) string {
	var b strings.Builder
	b.WriteString(`<nav id="sidebar">`)
	if branch != "" {
		b.WriteString(`<div id="branch-bar">`)
		b.WriteString(`<span class="branch-icon">⎇</span>`)
		b.WriteString(`<span class="branch-name">` + html.EscapeString(branch) + `</span>`)
		b.WriteString(`</div>`)
	}
	b.WriteString(`<div id="sidebar-header">`)
	b.WriteString(`<span>Changed Files</span><span class="file-count">` + strconv.Itoa(len(files)) + `</span>`)
	b.WriteString(`</div>`)
	b.WriteString(`<div id="sidebar-filter"><input id="filter-input" placeholder="Filter files…" oninput="filterFiles(this.value)" autocomplete="off" spellcheck="false"></div>`)
	b.WriteString(`<div id="file-list">`)
	for i, f := range files {
		name := f.DisplayName()
		b.WriteString(`<div class="file-item" onclick="jumpTo('file-` + strconv.Itoa(i) + `',this)" id="nav-` + strconv.Itoa(i) + `">`)
		b.WriteString(`<span class="file-icon">` + fileIcon(name) + `</span>`)
		b.WriteString(`<span class="file-name" title="` + html.EscapeString(name) + `">` + html.EscapeString(name) + `</span>`)
		b.WriteString(`<span class="file-stats">`)
		if f.Added > 0 {
			b.WriteString(`<span class="stat-add">+` + strconv.Itoa(f.Added) + `</span>`)
		}
		if f.Removed > 0 {
			b.WriteString(`<span class="stat-del">−` + strconv.Itoa(f.Removed) + `</span>`)
		}
		b.WriteString(`</span></div>`)
	}
	b.WriteString(`</div>`)
	b.WriteString(sidebarFooter())
	b.WriteString(`</nav>`)
	return b.String()
}

func mainContent(files []diff.File) string {
	var b strings.Builder
	b.WriteString(`<main id="main">`)

	if len(files) == 0 {
		b.WriteString(`<div class="empty-state">`)
		b.WriteString(`<div class="empty-icon">✓</div>`)
		b.WriteString(`<div class="empty-title">No changes</div>`)
		b.WriteString(`<div class="empty-sub">Working tree is clean.</div>`)
		b.WriteString(`</div>`)
	}

	for i, f := range files {
		name := f.DisplayName()
		lang := fileLang(name)
		b.WriteString(`<section class="diff-file" id="file-` + strconv.Itoa(i) + `" data-lang="` + lang + `">`)

		b.WriteString(`<div class="diff-file-header" onclick="toggleFile(this)">`)
		b.WriteString(`<span class="diff-chevron">▾</span>`)
		b.WriteString(`<span class="diff-file-icon">` + fileIcon(name) + `</span>`)
		b.WriteString(`<span class="diff-file-name">` + html.EscapeString(name) + `</span>`)
		b.WriteString(`<span class="diff-file-stats">`)
		if f.Added > 0 {
			b.WriteString(`<span class="sa">+` + strconv.Itoa(f.Added) + `</span>`)
		}
		if f.Removed > 0 {
			b.WriteString(`<span class="sd">−` + strconv.Itoa(f.Removed) + `</span>`)
		}
		b.WriteString(`</span></div>`)

		switch {
		case f.OldImage != nil || f.NewImage != nil:
			b.WriteString(imageSection(f))
		case f.IsBinary:
			b.WriteString(`<div class="binary-notice">⊘ Binary file — no preview available</div>`)
		default:
			b.WriteString(`<div class="diff-body"><table class="diff-table">`)
			for _, line := range f.Lines {
				switch line.Type {
				case diff.LineHunk:
					b.WriteString(`<tr class="line-hunk" onclick="toggleHunk(this)"><td colspan="3"><span class="hunk-chevron">▾</span>` + html.EscapeString(line.Content) + `</td></tr>`)
				case diff.LineAdd:
					ln := strconv.Itoa(line.NewNum)
					b.WriteString(`<tr class="line-add">`)
					b.WriteString(`<td class="ln"></td>`)
					b.WriteString(`<td class="ln" onclick="permalink(this,` + strconv.Itoa(i) + `,` + ln + `)">` + ln + `</td>`)
					b.WriteString(`<td class="lc"><span class="sign">+</span>` + html.EscapeString(line.Content) + `</td>`)
					b.WriteString(`</tr>`)
				case diff.LineDel:
					ln := strconv.Itoa(line.OldNum)
					b.WriteString(`<tr class="line-del">`)
					b.WriteString(`<td class="ln" onclick="permalink(this,` + strconv.Itoa(i) + `,` + ln + `)">` + ln + `</td>`)
					b.WriteString(`<td class="ln"></td>`)
					b.WriteString(`<td class="lc"><span class="sign">−</span>` + html.EscapeString(line.Content) + `</td>`)
					b.WriteString(`</tr>`)
				case diff.LineCtx:
					ln := strconv.Itoa(line.NewNum)
					b.WriteString(`<tr class="line-ctx">`)
					b.WriteString(`<td class="ln">` + strconv.Itoa(line.OldNum) + `</td>`)
					b.WriteString(`<td class="ln" onclick="permalink(this,` + strconv.Itoa(i) + `,` + ln + `)">` + ln + `</td>`)
					b.WriteString(`<td class="lc"><span class="sign"> </span>` + html.EscapeString(line.Content) + `</td>`)
					b.WriteString(`</tr>`)
				}
			}
			b.WriteString(`</table></div>`)
		}
		b.WriteString(`</section>`)
	}

	b.WriteString(`</main>`)
	return b.String()
}

func footer() string {
	return `<div id="search-bar">
  <input id="search-input" placeholder="Search…" autocomplete="off" spellcheck="false">
  <span id="search-count"></span>
  <button class="sb-btn" onclick="searchNav(-1)" title="Previous">↑</button>
  <button class="sb-btn" onclick="searchNav(1)" title="Next">↓</button>
  <button class="sb-btn" onclick="closeSearch()">✕</button>
</div>
<div id="help-modal" hidden onclick="if(event.target===this)hideHelp()">
  <div id="help-box">
    <div id="help-title">Keyboard Shortcuts</div>
    <table id="help-table">
      <tr><td><kbd>j</kbd> / <kbd>k</kbd></td><td>Next / prev hunk</td></tr>
      <tr><td><kbd>n</kbd> / <kbd>p</kbd></td><td>Next / prev file</td></tr>
      <tr><td><kbd>/</kbd></td><td>Open search</td></tr>
      <tr><td><kbd>Esc</kbd></td><td>Close search / modal</td></tr>
      <tr><td><kbd>t</kbd></td><td>Toggle theme</td></tr>
      <tr><td><kbd>s</kbd></td><td>Toggle split view</td></tr>
      <tr><td><kbd>e</kbd></td><td>Export HTML</td></tr>
      <tr><td><kbd>?</kbd></td><td>Show this help</td></tr>
    </table>
    <button class="tb-btn" onclick="hideHelp()">Close</button>
  </div>
</div>
<div id="toast"></div>
<script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/highlight.min.js"></script>
<script>
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
</script>
</body>
</html>
`
}

func imageSection(f diff.File) string {
	var b strings.Builder
	b.WriteString(`<div class="img-section">`)
	if f.OldImage != nil && f.NewImage != nil {
		b.WriteString(`<div class="img-compare">`)
		b.WriteString(`<div class="img-panel img-panel-del"><div class="img-label img-label-del">Before</div>`)
		b.WriteString(imageTag(f.OldImage, f.OldName))
		b.WriteString(`</div>`)
		b.WriteString(`<div class="img-panel img-panel-add"><div class="img-label img-label-add">After</div>`)
		b.WriteString(imageTag(f.NewImage, f.NewName))
		b.WriteString(`</div></div>`)
	} else if f.NewImage != nil {
		b.WriteString(imageTag(f.NewImage, f.NewName))
	} else {
		b.WriteString(imageTag(f.OldImage, f.OldName))
	}
	b.WriteString(`</div>`)
	return b.String()
}

func imageTag(data []byte, name string) string {
	ext := ""
	if idx := strings.LastIndexByte(name, '.'); idx >= 0 {
		ext = strings.ToLower(name[idx+1:])
	}
	mime := "image/png"
	switch ext {
	case "jpg", "jpeg":
		mime = "image/jpeg"
	case "gif":
		mime = "image/gif"
	case "webp":
		mime = "image/webp"
	case "svg":
		mime = "image/svg+xml"
	case "ico":
		mime = "image/x-icon"
	case "bmp":
		mime = "image/bmp"
	case "avif":
		mime = "image/avif"
	}
	return `<img src="data:` + mime + `;base64,` + base64.StdEncoding.EncodeToString(data) + `" class="img-preview" alt="">`
}

func fileLang(name string) string {
	ext := ""
	if idx := strings.LastIndexByte(name, '.'); idx >= 0 {
		ext = strings.ToLower(name[idx+1:])
	}
	langs := map[string]string{
		"go": "go", "js": "javascript", "ts": "typescript",
		"tsx": "typescript", "jsx": "javascript", "py": "python",
		"rs": "rust", "rb": "ruby", "java": "java",
		"html": "html", "css": "css", "scss": "scss",
		"json": "json", "md": "markdown", "yaml": "yaml",
		"yml": "yaml", "sh": "bash", "bash": "bash",
		"sql": "sql", "c": "c", "cpp": "cpp", "cs": "csharp",
		"php": "php", "swift": "swift", "kt": "kotlin",
		"toml": "ini", "xml": "xml", "vue": "xml",
	}
	if l, ok := langs[ext]; ok {
		return l
	}
	return ""
}

func fileIcon(name string) string {
	ext := ""
	if idx := strings.LastIndexByte(name, '.'); idx >= 0 {
		ext = strings.ToLower(name[idx+1:])
	}
	switch ext {
	case "go":
		return badge("#00ACD7", "#fff", "go")
	case "js":
		return badge("#F7DF1E", "#1a1a1a", "JS")
	case "ts":
		return badge("#3178C6", "#fff", "TS")
	case "tsx":
		return badge("#3178C6", "#fff", "TS")
	case "jsx":
		return badge("#61DAFB", "#1a1a1a", "JX")
	case "py":
		return badge("#3776AB", "#fff", "Py")
	case "rs":
		return badge("#CE422B", "#fff", "Rs")
	case "rb":
		return badge("#CC342D", "#fff", "Rb")
	case "java":
		return badge("#B07219", "#fff", "Jv")
	case "html":
		return badge("#E34F26", "#fff", "HT")
	case "css":
		return badge("#1572B6", "#fff", "CS")
	case "scss":
		return badge("#CC6699", "#fff", "SC")
	case "json":
		return badge("#4a4a4a", "#F7DF1E", "{}")
	case "md":
		return badge("#083FA1", "#fff", "MD")
	case "yaml", "yml":
		return badge("#CB171E", "#fff", "YL")
	case "toml":
		return badge("#9C4221", "#fff", "TL")
	case "sh", "bash":
		return badge("#4EAA25", "#fff", "SH")
	case "sql":
		return badge("#336791", "#fff", "SQ")
	case "c":
		return badge("#555555", "#fff", "C")
	case "cpp":
		return badge("#f34b7d", "#fff", "C+")
	case "cs":
		return badge("#239120", "#fff", "C#")
	case "php":
		return badge("#777BB4", "#fff", "PH")
	case "swift":
		return badge("#FA7343", "#fff", "Sw")
	case "kt":
		return badge("#7F52FF", "#fff", "Kt")
	case "vue":
		return badge("#4FC08D", "#fff", "Vu")
	case "xml":
		return badge("#0060AC", "#fff", "XM")
	case "proto":
		return badge("#4285F4", "#fff", "Pb")
	case "png", "jpg", "jpeg", "gif", "webp", "svg", "ico", "bmp", "avif":
		return iconImage()
	}
	return iconFile()
}

func badge(bg, fg, text string) string {
	fs := "7.5"
	if len(text) == 1 {
		fs = "9"
	}
	return `<svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">` +
		`<rect width="16" height="16" rx="3" fill="` + bg + `"/>` +
		`<text x="8" y="8" text-anchor="middle" dominant-baseline="central" font-size="` + fs + `" font-weight="700" fill="` + fg + `" font-family="ui-sans-serif,system-ui,sans-serif">` + html.EscapeString(text) + `</text>` +
		`</svg>`
}

func iconImage() string {
	return `<svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">` +
		`<rect x="0.5" y="2.5" width="15" height="11" rx="1.5" fill="#1a7f37" stroke="#2ea043" stroke-width="0.5"/>` +
		`<circle cx="4.5" cy="5.5" r="1.5" fill="#aff5b4"/>` +
		`<path d="M0.5 10.5l3.5-3.5 2.5 2.5 2-2 5 4.5H0.5z" fill="#aff5b4" opacity="0.75"/>` +
		`</svg>`
}

func iconFile() string {
	return `<svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">` +
		`<path d="M3.5 1h6.586L13.5 4.414V15H3.5V1z" fill="#30363d" stroke="#484f58" stroke-width="0.5"/>` +
		`<path d="M10 1v3.5H13.5" fill="none" stroke="#484f58" stroke-width="0.5"/>` +
		`<line x1="5.5" y1="7" x2="10.5" y2="7" stroke="#8b949e" stroke-width="0.75"/>` +
		`<line x1="5.5" y1="9.5" x2="10.5" y2="9.5" stroke="#8b949e" stroke-width="0.75"/>` +
		`<line x1="5.5" y1="12" x2="8.5" y2="12" stroke="#8b949e" stroke-width="0.75"/>` +
		`</svg>`
}

func css() string {
	return `
*, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

:root {
  --bg:        #1e1e1e;
  --bg-alt:    #252526;
  --bg-hover:  #2d2d2d;
  --border:    #3e3e42;
  --muted:     #6e6e6e;
  --subtle:    #888888;
  --text:      #d4d4d4;
  --blue:      #4fc1ff;
  --green:     #4ec94e;
  --red:       #f44747;
  --yellow:    #ddb348;
  --green-bg:  #1a3320;
  --green-ln:  #142918;
  --green-txt: #b5ffb5;
  --red-bg:    #3a1818;
  --red-ln:    #2d1212;
  --red-txt:   #ffb5b5;
  --hunk-bg:   #1a2133;
  --font:      'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
  --fs:        12px;
}

:root.light {
  --bg:        #ffffff;
  --bg-alt:    #f3f3f3;
  --bg-hover:  #e8e8e8;
  --border:    #d4d4d4;
  --muted:     #a0a0a0;
  --subtle:    #6e6e6e;
  --text:      #333333;
  --blue:      #0066cc;
  --green:     #1a7f37;
  --red:       #cf222e;
  --yellow:    #9a6700;
  --green-bg:  #e6ffec;
  --green-ln:  #ccffd8;
  --green-txt: #116329;
  --red-bg:    #ffebe9;
  --red-ln:    #ffd7d5;
  --red-txt:   #82071e;
  --hunk-bg:   #ddf4ff;
}

html, body { height: 100%; }
body {
  font-family: var(--font);
  background: var(--bg);
  color: var(--text);
  display: flex;
  height: 100vh;
  overflow: hidden;
}

/* ── Sidebar footer ─────────────────────────────── */
#sidebar-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 9px 14px;
  border-top: 1px solid var(--border);
  flex-shrink: 0;
  gap: 8px;
}

#live-indicator {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 11px;
  color: var(--subtle);
  user-select: none;
}

#live-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--muted);
  transition: background 0.4s;
}

#live-dot.connected {
  background: var(--green);
  animation: live-pulse 2.4s ease-in-out infinite;
}

#live-dot.disconnected {
  background: var(--red);
}

@keyframes live-pulse {
  0%, 100% { opacity: 1; }
  50%       { opacity: 0.35; }
}

#github-btn {
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 4px 9px;
  background: var(--bg-hover);
  border: 1px solid var(--border);
  border-radius: 6px;
  color: var(--subtle);
  text-decoration: none;
  font-size: 11px;
  font-family: var(--font);
  transition: background 0.1s, border-color 0.1s, color 0.1s;
}

#github-btn:hover {
  background: var(--border);
  border-color: var(--subtle);
  color: var(--text);
}

.star-svg { color: var(--yellow); }

/* ── Sidebar ────────────────────────────────────── */
#sidebar {
  width: 272px;
  min-width: 272px;
  background: var(--bg-alt);
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

#branch-bar {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 9px 16px;
  font-size: 12px;
  font-weight: 600;
  color: var(--blue);
  border-bottom: 1px solid var(--border);
  flex-shrink: 0;
}
.branch-icon { font-size: 14px; }
.branch-name { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

#sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: .06em;
  color: var(--subtle);
  border-bottom: 1px solid var(--border);
  flex-shrink: 0;
}
.file-count {
  background: var(--muted);
  color: var(--text);
  border-radius: 10px;
  padding: 1px 7px;
  font-size: 11px;
  font-weight: 600;
}

#sidebar-filter {
  padding: 8px 10px;
  border-bottom: 1px solid var(--border);
  flex-shrink: 0;
}
#filter-input {
  width: 100%;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 6px;
  color: var(--text);
  font-family: var(--font);
  font-size: 12px;
  padding: 5px 10px;
  outline: none;
}
#filter-input:focus { border-color: var(--blue); }

#file-list { overflow-y: auto; flex: 1; padding: 4px 0; }

.file-item {
  display: flex;
  align-items: center;
  gap: 7px;
  padding: 5px 14px;
  cursor: pointer;
  font-size: 12px;
  border-left: 2px solid transparent;
  transition: background 0.12s;
}
.file-item:hover  { background: var(--bg-hover); }
.file-item.active { background: var(--bg-hover); border-left-color: var(--blue); }

.file-icon  { flex-shrink: 0; display: flex; align-items: center; }
.file-name  { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: var(--text); }
.file-stats { display: flex; gap: 5px; flex-shrink: 0; font-weight: 600; }
.stat-add   { color: var(--green); }
.stat-del   { color: var(--red); }

/* ── Right panel ────────────────────────────────── */
#right { flex: 1; display: flex; flex-direction: column; overflow: hidden; min-width: 0; }

/* ── Toolbar ────────────────────────────────────── */
#toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  height: 40px;
  flex-shrink: 0;
  border-bottom: 1px solid var(--border);
  background: var(--bg-alt);
  gap: 12px;
}
#toolbar-stats {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 12px;
  font-weight: 600;
}
.ts-add    { color: var(--green); }
.ts-del    { color: var(--red); }
.ts-files  { color: var(--subtle); font-weight: 400; }
#toolbar-actions { display: flex; gap: 4px; }

.tb-btn {
  background: var(--bg-hover);
  border: 1px solid var(--border);
  border-radius: 6px;
  color: var(--text);
  cursor: pointer;
  font-family: var(--font);
  font-size: 11px;
  padding: 3px 9px;
  transition: background 0.1s, border-color 0.1s;
}
.tb-btn:hover { background: var(--border); border-color: var(--subtle); }

/* ── Main ──────────────────────────────────────── */
#main { flex: 1; overflow-y: auto; padding: 20px 24px; }

/* ── Diff file block ───────────────────────────── */
.diff-file {
  background: var(--bg-alt);
  border: 1px solid var(--border);
  border-radius: 8px;
  margin-bottom: 20px;
  overflow: hidden;
}
.diff-file-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  background: var(--bg-hover);
  border-bottom: 1px solid var(--border);
  cursor: pointer;
  user-select: none;
}
.diff-file-header:hover { filter: brightness(1.08); }
.diff-chevron {
  font-size: 13px;
  color: var(--subtle);
  flex-shrink: 0;
  transition: transform 0.15s ease;
}
.diff-file.collapsed .diff-chevron    { transform: rotate(-90deg); }
.diff-file.collapsed .diff-body,
.diff-file.collapsed .img-section,
.diff-file.collapsed .binary-notice   { display: none; }
.diff-file.collapsed                  { border-bottom: none; }

.diff-file-icon  { display: flex; align-items: center; flex-shrink: 0; }
.diff-file-name  { flex: 1; font-size: 13px; font-weight: 600; color: var(--blue); word-break: break-all; }
.diff-file-stats { display: flex; gap: 8px; font-size: 12px; font-weight: 600; flex-shrink: 0; }
.diff-file-stats .sa { color: var(--green); }
.diff-file-stats .sd { color: var(--red); }

/* ── Diff table ────────────────────────────────── */
.diff-body { overflow-x: auto; }

.diff-table {
  width: 100%;
  border-collapse: collapse;
  font-size: var(--fs);
  line-height: 1.65;
}
.diff-table td { padding: 0; vertical-align: top; }

.ln {
  width: 1%;
  min-width: 44px;
  padding: 0 10px;
  text-align: right;
  color: var(--muted);
  user-select: none;
  white-space: nowrap;
  cursor: pointer;
}
.ln:hover { color: var(--blue); }

.lc {
  padding: 0 14px;
  white-space: pre;
  tab-size: 4;
  width: 100%;
}
.split-lc { width: 50%; border-right: 1px solid var(--border); }

.sign { display: inline-block; width: 14px; user-select: none; }

/* Added */
.line-add            { background: var(--green-bg); }
.line-add .ln        { background: var(--green-ln); color: var(--green); }
.line-add .lc        { color: var(--green-txt); }
.line-add .sign      { color: var(--green); }

/* Removed */
.line-del            { background: var(--red-bg); }
.line-del .ln        { background: var(--red-ln); color: var(--red); }
.line-del .lc        { color: var(--red-txt); }
.line-del .sign      { color: var(--red); }

/* Context */
.line-ctx .sign { color: var(--muted); }

/* Hunk */
.line-hunk td {
  background: var(--hunk-bg);
  color: var(--blue);
  font-style: italic;
  padding: 4px 14px;
  cursor: pointer;
  opacity: .85;
}
.line-hunk:hover td { opacity: 1; }
.hunk-chevron {
  display: inline-block;
  margin-right: 6px;
  transition: transform 0.15s;
  font-style: normal;
}

/* Split view — td-level colors */
.split-body td.line-add    { background: var(--green-bg); color: var(--green-txt); }
.split-body td.line-add.ln { background: var(--green-ln); color: var(--green); }
.split-body td.line-del    { background: var(--red-bg);   color: var(--red-txt); }
.split-body td.line-del.ln { background: var(--red-ln);   color: var(--red); }

/* Permalink highlight */
.permalink-hl { outline: 2px solid var(--yellow); }

/* Search */
.lc.search-match { background-color: rgba(227, 179, 65, 0.18); outline: 1px solid var(--yellow); }
.lc.search-cur   { background-color: rgba(227, 179, 65, 0.40); outline: 2px solid var(--yellow); }

#search-bar {
  display: none;
  position: fixed;
  top: 48px;
  right: 20px;
  background: var(--bg-alt);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 6px 10px;
  gap: 6px;
  align-items: center;
  z-index: 200;
  box-shadow: 0 4px 16px rgba(0,0,0,.4);
}
#search-bar.open { display: flex; }
#search-input {
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 6px;
  color: var(--text);
  font-family: var(--font);
  font-size: 13px;
  padding: 4px 10px;
  outline: none;
  width: 220px;
}
#search-input:focus { border-color: var(--blue); }
#search-count { font-size: 11px; color: var(--subtle); white-space: nowrap; min-width: 60px; }
.sb-btn {
  background: none;
  border: none;
  color: var(--subtle);
  cursor: pointer;
  font-size: 14px;
  padding: 2px 6px;
  border-radius: 4px;
}
.sb-btn:hover { background: var(--bg-hover); color: var(--text); }

/* Help modal */
#help-modal {
  position: fixed; inset: 0;
  background: rgba(0,0,0,.6);
  display: flex; align-items: center; justify-content: center;
  z-index: 300;
}
#help-modal[hidden] { display: none; }
#help-box {
  background: var(--bg-alt);
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 24px 28px;
  min-width: 340px;
}
#help-title { font-size: 15px; font-weight: 700; margin-bottom: 16px; color: var(--text); }
#help-table { width: 100%; border-collapse: collapse; margin-bottom: 20px; }
#help-table td { padding: 5px 8px; font-size: 13px; color: var(--text); }
#help-table td:first-child { white-space: nowrap; }
kbd {
  background: var(--bg-hover);
  border: 1px solid var(--border);
  border-radius: 4px;
  padding: 1px 6px;
  font-family: var(--font);
  font-size: 11px;
}

/* Toast */
#toast {
  position: fixed;
  bottom: 20px; right: 20px;
  background: var(--text); color: var(--bg);
  padding: 8px 16px;
  border-radius: 6px;
  font-size: 13px;
  opacity: 0;
  transition: opacity 0.2s;
  pointer-events: none;
  z-index: 400;
}
#toast.show { opacity: 1; }

/* Image preview */
.img-section  { padding: 20px; }
.img-compare  { display: flex; gap: 16px; }
.img-panel    { flex: 1; min-width: 0; border-radius: 6px; overflow: hidden; }
.img-panel-del { border: 1px solid var(--red); }
.img-panel-add { border: 1px solid var(--green); }
.img-label    { padding: 5px 12px; font-size: 11px; font-weight: 600; text-transform: uppercase; letter-spacing: .05em; }
.img-label-del { background: var(--red-bg);   color: var(--red); }
.img-label-add { background: var(--green-bg); color: var(--green); }
.img-preview  { display: block; max-width: 100%; height: auto; }
.binary-notice { padding: 20px 24px; color: var(--subtle); font-size: 13px; }

/* Empty state */
.empty-state {
  display: flex; flex-direction: column;
  align-items: center; justify-content: center;
  height: 60vh; gap: 12px; color: var(--subtle);
}
.empty-icon  { font-size: 48px; }
.empty-title { font-size: 18px; font-weight: 600; color: var(--text); }
.empty-sub   { font-size: 13px; }

/* Scrollbar */
::-webkit-scrollbar              { width: 8px; height: 8px; }
::-webkit-scrollbar-track        { background: var(--bg); }
::-webkit-scrollbar-thumb        { background: var(--border); border-radius: 4px; }
::-webkit-scrollbar-thumb:hover  { background: var(--muted); }
`
}
