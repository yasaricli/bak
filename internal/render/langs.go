package render

import "strings"

func fileExt(name string) string {
	if idx := strings.LastIndexByte(name, '.'); idx >= 0 {
		return strings.ToLower(name[idx+1:])
	}
	return ""
}

func fileLang(name string) string {
	langs := map[string]string{
		"go":    "go",
		"js":    "javascript",
		"ts":    "typescript",
		"tsx":   "typescript",
		"jsx":   "javascript",
		"py":    "python",
		"rs":    "rust",
		"rb":    "ruby",
		"java":  "java",
		"html":  "html",
		"css":   "css",
		"scss":  "scss",
		"json":  "json",
		"md":    "markdown",
		"yaml":  "yaml",
		"yml":   "yaml",
		"sh":    "bash",
		"bash":  "bash",
		"sql":   "sql",
		"c":     "c",
		"cpp":   "cpp",
		"cs":    "csharp",
		"php":   "php",
		"swift": "swift",
		"kt":    "kotlin",
		"toml":  "ini",
		"xml":   "xml",
		"vue":   "xml",
	}
	if l, ok := langs[fileExt(name)]; ok {
		return l
	}
	return ""
}
