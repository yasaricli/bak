# bak

A local git diff viewer that opens a GitHub-style UI in the browser.

![dark theme diff viewer with sidebar](screen.png)

## Install

```sh
go install github.com/yasaricli/bak@latest
```

Requires Go 1.21+ and `git` in your PATH.

## Usage

```sh
bak              # working tree (unstaged changes)
bak --staged     # staged changes
bak HEAD         # last commit
bak HEAD~2       # two commits ago
bak HEAD~3..HEAD # range between commits
```

## Development

**Prerequisites:** Go 1.21+ and `git` in your PATH.

Clone and build locally:

```sh
git clone https://github.com/yasaricli/bak.git
cd bak
go build -o bak .
./bak              # run from the project directory
```

Run without building:

```sh
go run . --staged
```

Run tests:

```sh
go test ./...
```

To install your local build into `$GOPATH/bin`:

```sh
go install .
```

## How it works

1. Runs `git diff` with the given arguments
2. Parses the unified diff output
3. Starts a local HTTP server on a random available port
4. Opens the browser automatically (`open` on macOS)
5. In live mode, watches the repo and pushes updates over SSE
6. Runs until Ctrl-C

## Live mode

When you run `bak` with no arguments or `bak --staged`, the page auto-refreshes
as you edit files — no need to re-run the command.

- fsnotify watcher reacts to writes with ~250 ms debounce
- 1.5 s `git status` / `git diff` polling fallback catches anything missed
- Updates are pushed to the browser over Server-Sent Events
- Toolbar shows a ● Live indicator; set `BAK_DEBUG=1` for watcher logs
- State preserved across refreshes: scroll, theme, split view, search, file filter, permalink
- Off for fixed refs like `bak HEAD~1` (the diff can't change)

## UI

- Dark theme (GitHub dark style)
- Left sidebar listing changed files with `+/-` counts
- Clicking a file jumps to that section; active file tracks scroll
- Unified diff with old/new line numbers
- Added lines in green, removed lines in red
- File type icons
- Auto-refreshes on file changes in live mode
- Zero external dependencies — pure HTML/CSS/JS embedded in the binary

## Project structure

```
bak/
├── main.go               # CLI entry, watcher dispatcher
└── internal/
    ├── diff/
    │   └── diff.go       # unified diff parser
    ├── render/
    │   └── render.go     # HTML/CSS/JS page builder (SSE client embedded)
    ├── server/
    │   ├── server.go     # long-lived HTTP server
    │   └── sse.go        # SSE broker
    └── watcher/
        └── watcher.go    # fsnotify + polling fallback
```

## License

MIT
