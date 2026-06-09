// Package render builds the self-contained HTML page for a diff.
package render

import (
	_ "embed"
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"

	"github.com/yasaricli/bak/internal/diff"
)

//go:embed templates/page.html
var pageTemplate string

//go:embed templates/style.css
var styleCSS string

//go:embed templates/app.js
var appJS string

type pageData struct {
	Title    string
	Branch   string
	Files    []fileData
	TotalAdd int
	TotalDel int
	Style    template.HTML
	Script   template.HTML
}

type fileData struct {
	Index       int
	Name        string
	Lang        string
	Icon        template.HTML
	Added       int
	Removed     int
	IsBinary    bool
	HasOldImage bool
	HasNewImage bool
	OldImageSrc template.URL
	NewImageSrc template.URL
	Lines       []lineData
}

type lineData struct {
	Type    string
	OldNum  int
	NewNum  int
	Content string
	FileIdx int
}

func HTML(files []diff.File, title, branch string, _ []byte) string {
	tmpl, err := template.New("page").Parse(pageTemplate)
	if err != nil {
		return fmt.Sprintf("template parse error: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, buildPageData(files, title, branch)); err != nil {
		return fmt.Sprintf("template execute error: %v", err)
	}
	return buf.String()
}

func buildPageData(files []diff.File, title, branch string) pageData {
	data := pageData{
		Title:  title,
		Branch: branch,
		Style:  template.HTML("<style>\n" + styleCSS + "\n</style>"),
		Script: template.HTML("<script>\n" + appJS + "\n</script>"),
	}

	for i, f := range files {
		name := f.DisplayName()
		fd := fileData{
			Index:    i,
			Name:     name,
			Lang:     fileLang(name),
			Icon:     fileIcon(name),
			Added:    f.Added,
			Removed:  f.Removed,
			IsBinary: f.IsBinary,
		}

		if f.OldImage != nil {
			fd.HasOldImage = true
			fd.OldImageSrc = template.URL(imageDataURL(f.OldImage, f.OldName))
		}
		if f.NewImage != nil {
			fd.HasNewImage = true
			fd.NewImageSrc = template.URL(imageDataURL(f.NewImage, f.NewName))
		}

		for _, l := range f.Lines {
			fd.Lines = append(fd.Lines, lineData{
				Type:    string(l.Type),
				OldNum:  l.OldNum,
				NewNum:  l.NewNum,
				Content: l.Content,
				FileIdx: i,
			})
		}

		data.TotalAdd += f.Added
		data.TotalDel += f.Removed
		data.Files = append(data.Files, fd)
	}

	return data
}

func imageDataURL(data []byte, name string) string {
	return "data:" + imageMIME(fileExt(name)) + ";base64," + base64.StdEncoding.EncodeToString(data)
}

func imageMIME(ext string) string {
	switch ext {
	case "jpg", "jpeg":
		return "image/jpeg"
	case "gif":
		return "image/gif"
	case "webp":
		return "image/webp"
	case "svg":
		return "image/svg+xml"
	case "ico":
		return "image/x-icon"
	case "bmp":
		return "image/bmp"
	case "avif":
		return "image/avif"
	default:
		return "image/png"
	}
}
