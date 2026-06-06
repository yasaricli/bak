// Package diff parses unified diff output into structured file diffs.
package diff

import (
	"strconv"
	"strings"
)

type LineType string

const (
	LineAdd  LineType = "add"
	LineDel  LineType = "del"
	LineCtx  LineType = "ctx"
	LineHunk LineType = "hunk"
)

type Line struct {
	Type    LineType
	Content string
	OldNum  int
	NewNum  int
}

type File struct {
	OldName string
	NewName string
	Added   int
	Removed int
	Lines   []Line
}

func (f *File) DisplayName() string {
	if f.NewName != "" {
		return f.NewName
	}
	return f.OldName
}

func Parse(raw string) []File {
	var files []File
	var cur *File
	oldLine, newLine := 0, 0

	for _, l := range strings.Split(raw, "\n") {
		switch {
		case strings.HasPrefix(l, "diff --git "):
			if cur != nil {
				files = append(files, *cur)
			}
			cur = &File{}

		case cur == nil:
			continue

		case strings.HasPrefix(l, "--- "):
			name := strings.TrimPrefix(strings.TrimPrefix(l, "--- "), "a/")
			if name != "/dev/null" {
				cur.OldName = name
			}

		case strings.HasPrefix(l, "+++ "):
			name := strings.TrimPrefix(strings.TrimPrefix(l, "+++ "), "b/")
			if name != "/dev/null" {
				cur.NewName = name
			}

		case strings.HasPrefix(l, "@@ "):
			oldLine, newLine = parseHunkHeader(l)
			cur.Lines = append(cur.Lines, Line{Type: LineHunk, Content: l})

		case strings.HasPrefix(l, "+") && !strings.HasPrefix(l, "+++"):
			cur.Lines = append(cur.Lines, Line{Type: LineAdd, Content: l[1:], NewNum: newLine})
			cur.Added++
			newLine++

		case strings.HasPrefix(l, "-") && !strings.HasPrefix(l, "---"):
			cur.Lines = append(cur.Lines, Line{Type: LineDel, Content: l[1:], OldNum: oldLine})
			cur.Removed++
			oldLine++

		case strings.HasPrefix(l, " ") || l == "":
			content := ""
			if len(l) > 0 {
				content = l[1:]
			}
			cur.Lines = append(cur.Lines, Line{Type: LineCtx, Content: content, OldNum: oldLine, NewNum: newLine})
			oldLine++
			newLine++
		}
	}

	if cur != nil {
		files = append(files, *cur)
	}
	return files
}

func parseHunkHeader(s string) (oldStart, newStart int) {
	s = strings.TrimPrefix(s, "@@ ")
	parts := strings.Fields(s)
	if len(parts) < 2 {
		return 0, 0
	}
	return parseLineNum(parts[0]), parseLineNum(parts[1])
}

func parseLineNum(s string) int {
	s = strings.TrimLeft(s, "-+")
	if idx := strings.IndexByte(s, ','); idx >= 0 {
		s = s[:idx]
	}
	n, _ := strconv.Atoi(s)
	return n
}
