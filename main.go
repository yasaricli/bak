package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/yasaricli/bak/internal/diff"
	"github.com/yasaricli/bak/internal/render"
	"github.com/yasaricli/bak/internal/server"
)

func main() {
	args := os.Args[1:]

	diffArgs := buildDiffArgs(args)
	title := buildTitle(args)

	raw, err := gitDiff(diffArgs)
	if err != nil {
		fmt.Fprintln(os.Stderr, "git diff:", err)
		os.Exit(1)
	}

	files := diff.Parse(raw)

	// Append untracked files only when no specific ref/path args are given.
	if len(args) == 0 || (len(args) == 1 && (args[0] == "--staged" || args[0] == "--cached")) {
		untracked, err := untrackedFiles()
		if err == nil {
			files = append(files, untracked...)
		}
	}

	page := render.HTML(files, title)

	port, err := server.FreePort()
	if err != nil {
		fmt.Fprintln(os.Stderr, "port:", err)
		os.Exit(1)
	}

	server.Open(page, port)
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
		content, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		f := diff.File{NewName: p}
		for i, line := range strings.Split(string(content), "\n") {
			f.Lines = append(f.Lines, diff.Line{
				Type:    diff.LineAdd,
				Content: line,
				NewNum:  i + 1,
			})
			f.Added++
		}
		files = append(files, f)
	}
	return files, nil
}
