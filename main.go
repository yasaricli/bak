package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/yasaricli/bak/internal/diff"
	"github.com/yasaricli/bak/internal/render"
	"github.com/yasaricli/bak/internal/server"
)

//go:embed logo.png
var logoData []byte

func main() {
	args := os.Args[1:]
	diffArgs := buildDiffArgs(args)
	title := buildTitle(args)
	watchWorking := len(args) == 0 || (len(args) == 1 && (args[0] == "--staged" || args[0] == "--cached"))

	// Validate git repo on startup.
	if _, err := gitDiff(diffArgs); err != nil {
		fmt.Fprintln(os.Stderr, "git diff:", err)
		os.Exit(1)
	}

	buildPage := func() string {
		raw, _ := gitDiff(diffArgs)
		files := diff.Parse(raw)
		files = loadImages(files)
		if watchWorking {
			untracked, _ := untrackedFiles()
			files = append(files, untracked...)
		}
		return render.HTML(files, title, currentBranch(), logoData)
	}

	port, err := server.FreePort()
	if err != nil {
		fmt.Fprintln(os.Stderr, "port:", err)
		os.Exit(1)
	}

	server.Serve(buildPage, port)
}

func buildDiffArgs(args []string) []string {
	out := make([]string, 0, len(args))
	for _, a := range args {
		if a == "--staged" || a == "--cached" {
			out = append(out, "--staged")
		} else {
			out = append(out, a)
		}
	}
	return out
}

func buildTitle(args []string) string {
	if len(args) == 0 {
		return "bak: working tree"
	}
	return "bak: " + strings.Join(args, " ")
}

func gitDiff(args []string) (string, error) {
	cmd := exec.Command("git", append([]string{"diff"}, args...)...)
	out, err := cmd.Output()
	if err != nil {
		if e, ok := err.(*exec.ExitError); ok && len(e.Stderr) > 0 {
			return "", fmt.Errorf("%s", strings.TrimSpace(string(e.Stderr)))
		}
		return "", err
	}
	return string(out), nil
}

func untrackedFiles() ([]diff.File, error) {
	cmd := exec.Command("git", "ls-files", "--others", "--exclude-standard")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	paths := strings.Split(strings.TrimRight(string(out), "\n"), "\n")

	var files []diff.File
	for _, p := range paths {
		if p == "" {
			continue
		}
		f := diff.File{NewName: p}
		if isImageExt(p) {
			data, err := os.ReadFile(p)
			if err != nil {
				continue
			}
			f.NewImage = data
		} else {
			content, err := os.ReadFile(p)
			if err != nil {
				continue
			}
			for i, line := range strings.Split(string(content), "\n") {
				f.Lines = append(f.Lines, diff.Line{
					Type:    diff.LineAdd,
					Content: line,
					NewNum:  i + 1,
				})
				f.Added++
			}
		}
		files = append(files, f)
	}
	return files, nil
}

func loadImages(files []diff.File) []diff.File {
	for i := range files {
		f := &files[i]
		if !f.IsBinary || !isImageExt(f.DisplayName()) {
			continue
		}
		if f.NewName != "" {
			if data, err := os.ReadFile(f.NewName); err == nil {
				f.NewImage = data
			}
		}
		if f.OldName != "" {
			if data, err := exec.Command("git", "show", "HEAD:"+f.OldName).Output(); err == nil {
				f.OldImage = data
			}
		}
	}
	return files
}

func isImageExt(name string) bool {
	dot := strings.LastIndex(name, ".")
	if dot < 0 {
		return false
	}
	switch strings.ToLower(name[dot:]) {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".svg", ".ico", ".bmp", ".avif":
		return true
	}
	return false
}

func currentBranch() string {
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
