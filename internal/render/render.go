// Package render builds the self-contained HTML page for a diff.
package render

import (
	"encoding/base64"
	"html"
	"strconv"
	"strings"

	"github.com/yasaricli/bak/internal/diff"
)

func HTML(files []diff.File, title, branch string) string {
	totalAdd, totalDel := 0, 0
	for _, f := range files {
		totalAdd += f.Added
		totalDel += f.Removed
	}

	var b strings.Builder
	b.WriteString(header(title))
	b.WriteString(sidebar(files, branch))
	b.WriteString(`<div id="right">`)
	b.WriteString(toolbar(len(files), totalAdd, totalDel))
	b.WriteString(mainContent(files))
	b.WriteString(`</div>`)
	b.WriteString(footer())
	return b.String()
}

func header(title string) string {
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
<body>
`
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
	b.WriteString(`</div></nav>`)
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

		// File header
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

		// Body
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
  var main  = document.getElementById('main');
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
	icons := map[string]string{
		"go": "🔵", "js": "🟡", "ts": "🔷", "tsx": "🔷", "jsx": "🟡",
		"py": "🐍", "rs": "🦀", "rb": "💎", "java": "☕",
		"html": "🌐", "css": "🎨", "scss": "🎨", "json": "📋",
		"md": "📝", "yaml": "⚙️", "yml": "⚙️", "toml": "⚙️",
		"sh": "💻", "sql": "🗄️", "proto": "📡",
		"png": "🖼️", "jpg": "🖼️", "jpeg": "🖼️", "gif": "🖼️",
		"webp": "🖼️", "svg": "🖼️", "ico": "🖼️", "bmp": "🖼️",
	}
	if icon, ok := icons[ext]; ok {
		return icon
	}
	return "📄"
}

func css() string {
	return `
*, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

:root {
  --bg:        #0d1117;
  --bg-alt:    #161b22;
  --bg-hover:  #1c2128;
  --border:    #30363d;
  --muted:     #484f58;
  --subtle:    #8b949e;
  --text:      #c9d1d9;
  --blue:      #58a6ff;
  --green:     #3fb950;
  --red:       #f85149;
  --yellow:    #e3b341;
  --green-bg:  #0d4429;
  --green-ln:  #0a3320;
  --green-txt: #aff5b4;
  --red-bg:    #3d1a1a;
  --red-ln:    #2f1313;
  --red-txt:   #ffa198;
  --hunk-bg:   #1a2433;
  --font:      'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
  --fs:        12px;
}

:root.light {
  --bg:        #ffffff;
  --bg-alt:    #f6f8fa;
  --bg-hover:  #eaeef2;
  --border:    #d0d7de;
  --muted:     #8c959f;
  --subtle:    #57606a;
  --text:      #24292f;
  --blue:      #0969da;
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

/* ── Sidebar ───────────────────────────────────── */
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
  gap: 6px;
  padding: 5px 14px;
  cursor: pointer;
  font-size: 12px;
  border-left: 2px solid transparent;
  transition: background 0.12s;
}
.file-item:hover  { background: var(--bg-hover); }
.file-item.active { background: var(--bg-hover); border-left-color: var(--blue); }

.file-icon  { font-size: 13px; flex-shrink: 0; line-height: 1; }
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

.diff-file-icon  { font-size: 14px; line-height: 1; }
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
