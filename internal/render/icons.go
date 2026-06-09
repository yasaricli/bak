package render

import "html/template"

func fileIcon(name string) template.HTML {
	switch fileExt(name) {
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

func badge(bg, fg, text string) template.HTML {
	fs := "7.5"
	if len(text) == 1 {
		fs = "9"
	}
	escaped := template.HTMLEscapeString(text)
	return template.HTML(
		`<svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">` +
			`<rect width="16" height="16" rx="3" fill="` + bg + `"/>` +
			`<text x="8" y="8" text-anchor="middle" dominant-baseline="central" font-size="` + fs + `" font-weight="700" fill="` + fg + `" font-family="ui-sans-serif,system-ui,sans-serif">` + escaped + `</text>` +
			`</svg>`,
	)
}

func iconImage() template.HTML {
	return template.HTML(
		`<svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">` +
			`<rect x="0.5" y="2.5" width="15" height="11" rx="1.5" fill="#1a7f37" stroke="#2ea043" stroke-width="0.5"/>` +
			`<circle cx="4.5" cy="5.5" r="1.5" fill="#aff5b4"/>` +
			`<path d="M0.5 10.5l3.5-3.5 2.5 2.5 2-2 5 4.5H0.5z" fill="#aff5b4" opacity="0.75"/>` +
			`</svg>`,
	)
}

func iconFile() template.HTML {
	return template.HTML(
		`<svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">` +
			`<path d="M3.5 1h6.586L13.5 4.414V15H3.5V1z" fill="#30363d" stroke="#484f58" stroke-width="0.5"/>` +
			`<path d="M10 1v3.5H13.5" fill="none" stroke="#484f58" stroke-width="0.5"/>` +
			`<line x1="5.5" y1="7" x2="10.5" y2="7" stroke="#8b949e" stroke-width="0.75"/>` +
			`<line x1="5.5" y1="9.5" x2="10.5" y2="9.5" stroke="#8b949e" stroke-width="0.75"/>` +
			`<line x1="5.5" y1="12" x2="8.5" y2="12" stroke="#8b949e" stroke-width="0.75"/>` +
			`</svg>`,
	)
}
