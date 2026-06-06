// Package render builds the self-contained HTML page for a diff.
package render

import (
	"html"
	"strconv"
	"strings"

	"github.com/yasaricli/bak/internal/diff"
)

func HTML(files []diff.File, title string) string {
	var b strings.Builder
	b.WriteString(header(title))
	b.WriteString(sidebar(files))
	b.WriteString(mainContent(files))
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
<style>
` + css() + `
</style>
</head>
<body>
`
}

func sidebar(files []diff.File) string {
	var b strings.Builder
	b.WriteString(`<nav id="sidebar">`)
	b.WriteString(`<div id="sidebar-header">`)
	b.WriteString(`<span>Changed Files</span><span class="file-count">` + strconv.Itoa(len(files)) + `</span>`)
	b.WriteString(`</div>`)
	b.WriteString(`<div id="file-list">`)
	for i, f := range files {
		name := f.DisplayName()
		b.WriteString(`<div class="file-item" onclick="jumpTo('file-` + strconv.Itoa(i) + `',this)" id="nav-` + strconv.Itoa(i) + `">`)
		b.WriteString(`<span class="file-icon">`)
		b.WriteString(fileIcon(name))
		b.WriteString(`</span>`)
		b.WriteString(`<span class="file-name" title="` + html.EscapeString(name) + `">` + html.EscapeString(shortName(name)) + `</span>`)
		b.WriteString(`<span class="file-stats">`)
		if f.Added > 0 {
			b.WriteString(`<span class="stat-add">+` + strconv.Itoa(f.Added) + `</span>`)
		}
		if f.Removed > 0 {
			b.WriteString(`<span class="stat-del">-` + strconv.Itoa(f.Removed) + `</span>`)
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
		b.WriteString(`<section class="diff-file" id="file-` + strconv.Itoa(i) + `">`)

		// File header
		b.WriteString(`<div class="diff-file-header">`)
		b.WriteString(`<span class="diff-file-icon">` + fileIcon(name) + `</span>`)
		b.WriteString(`<span class="diff-file-name">` + html.EscapeString(name) + `</span>`)
		b.WriteString(`<span class="diff-file-stats">`)
		if f.Added > 0 {
			b.WriteString(`<span class="sa">+` + strconv.Itoa(f.Added) + `</span>`)
		}
		if f.Removed > 0 {
			b.WriteString(`<span class="sd">-` + strconv.Itoa(f.Removed) + `</span>`)
		}
		b.WriteString(`</span>`)
		b.WriteString(`</div>`)

		// Diff table
		b.WriteString(`<div class="diff-body"><table class="diff-table">`)
		for _, line := range f.Lines {
			switch line.Type {
			case diff.LineHunk:
				b.WriteString(`<tr class="line-hunk"><td colspan="3">` + html.EscapeString(line.Content) + `</td></tr>`)
			case diff.LineAdd:
				b.WriteString(`<tr class="line-add">`)
				b.WriteString(`<td class="ln"></td>`)
				b.WriteString(`<td class="ln">` + strconv.Itoa(line.NewNum) + `</td>`)
				b.WriteString(`<td class="lc"><span class="sign">+</span>` + html.EscapeString(line.Content) + `</td>`)
				b.WriteString(`</tr>`)
			case diff.LineDel:
				b.WriteString(`<tr class="line-del">`)
				b.WriteString(`<td class="ln">` + strconv.Itoa(line.OldNum) + `</td>`)
				b.WriteString(`<td class="ln"></td>`)
				b.WriteString(`<td class="lc"><span class="sign">-</span>` + html.EscapeString(line.Content) + `</td>`)
				b.WriteString(`</tr>`)
			case diff.LineCtx:
				b.WriteString(`<tr class="line-ctx">`)
				b.WriteString(`<td class="ln">` + strconv.Itoa(line.OldNum) + `</td>`)
				b.WriteString(`<td class="ln">` + strconv.Itoa(line.NewNum) + `</td>`)
				b.WriteString(`<td class="lc"><span class="sign"> </span>` + html.EscapeString(line.Content) + `</td>`)
				b.WriteString(`</tr>`)
			}
		}
		b.WriteString(`</table></div>`)
		b.WriteString(`</section>`)
	}

	b.WriteString(`</main>`)
	return b.String()
}

func footer() string {
	return `<script>
function jumpTo(id, el) {
  document.querySelectorAll('.file-item').forEach(e => e.classList.remove('active'));
  if (el) el.classList.add('active');
  var t = document.getElementById(id);
  if (t) t.scrollIntoView({ behavior: 'smooth', block: 'start' });
}

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

</script>
</body>
</html>
`
}

func shortName(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) <= 3 {
		return path
	}
	return ".../" + strings.Join(parts[len(parts)-2:], "/")
}

func fileIcon(name string) string {
	ext := ""
	if idx := strings.LastIndexByte(name, '.'); idx >= 0 {
		ext = strings.ToLower(name[idx+1:])
	}
	icons := map[string]string{
		"go":   "🔵", "js": "🟡", "ts": "🔷", "tsx": "🔷", "jsx": "🟡",
		"py":   "🐍", "rs": "🦀", "rb": "💎", "java": "☕",
		"html": "🌐", "css": "🎨", "scss": "🎨", "json": "📋",
		"md":   "📝", "yaml": "⚙️", "yml": "⚙️", "toml": "⚙️",
		"sh":   "💻", "sql": "🗄️", "proto": "📡",
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
  --green-bg:  #0d4429;
  --green-ln:  #0a3320;
  --green-txt: #aff5b4;
  --red-bg:    #3d1a1a;
  --red-ln:    #2f1313;
  --red-txt:   #ffa198;
  --hunk-bg:   #1a2433;
  --font:      'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
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

#sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
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

#file-list { overflow-y: auto; flex: 1; padding: 6px 0; }

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
}

.diff-file-icon { font-size: 14px; line-height: 1; }
.diff-file-name { flex: 1; font-size: 13px; font-weight: 600; color: var(--blue); word-break: break-all; }
.diff-file-stats { display: flex; gap: 8px; font-size: 12px; font-weight: 600; flex-shrink: 0; }
.diff-file-stats .sa { color: var(--green); }
.diff-file-stats .sd { color: var(--red); }

/* ── Diff table ────────────────────────────────── */
.diff-body { overflow-x: auto; }

.diff-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
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
}

.lc {
  padding: 0 14px;
  white-space: pre;
  tab-size: 4;
  width: 100%;
}

.sign {
  display: inline-block;
  width: 14px;
  user-select: none;
}

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
.line-ctx .sign      { color: var(--muted); }

/* Hunk header */
.line-hunk td {
  background: var(--hunk-bg);
  color: var(--blue);
  font-style: italic;
  padding: 4px 14px;
  opacity: .85;
}

/* ── Empty state ───────────────────────────────── */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 60vh;
  gap: 12px;
  color: var(--subtle);
}
.empty-icon  { font-size: 48px; }
.empty-title { font-size: 18px; font-weight: 600; color: var(--text); }
.empty-sub   { font-size: 13px; }

/* ── Scrollbar ─────────────────────────────────── */
::-webkit-scrollbar              { width: 8px; height: 8px; }
::-webkit-scrollbar-track        { background: var(--bg); }
::-webkit-scrollbar-thumb        { background: var(--border); border-radius: 4px; }
::-webkit-scrollbar-thumb:hover  { background: var(--muted); }
`
}
