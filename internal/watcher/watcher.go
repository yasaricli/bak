// Package watcher recursively watches a repository for filesystem changes
// and emits debounced events when `git status --porcelain` actually changes.
package watcher

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// skipDirs are directories we never need to watch: either git metadata, large
// dependency stores, or build output. Skipping them keeps the kqueue/inotify
// FD count low so we don't silently fail to register watches deep in the tree.
var skipDirs = map[string]bool{
	".git": true, "node_modules": true, ".next": true, "dist": true,
	"build": true, "vendor": true, "target": true, "coverage": true,
	"__pycache__": true, ".cache": true, ".venv": true, "venv": true,
	".idea": true, ".vscode": true, ".turbo": true, ".parcel-cache": true,
}

var debugWatcher = os.Getenv("BAK_DEBUG") != ""

func dbg(format string, a ...any) {
	if debugWatcher {
		fmt.Fprintf(os.Stderr, "[watcher] "+format+"\n", a...)
	}
}

type Event struct{}

type Watcher struct {
	root     string
	debounce time.Duration
	events   chan Event
	fsw      *fsnotify.Watcher
}

func New(root string, debounce time.Duration) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &Watcher{
		root:     filepath.Clean(root),
		debounce: debounce,
		events:   make(chan Event, 1),
		fsw:      fsw,
	}, nil
}

func (w *Watcher) Events() <-chan Event { return w.events }

func (w *Watcher) Close() error {
	return w.fsw.Close()
}

// Start blocks until ctx is cancelled. Walks the repo, adds every non-skipped
// directory to the underlying fsnotify watcher, then runs the event loop plus
// a polling fallback that catches anything fsnotify misses.
func (w *Watcher) Start(ctx context.Context) error {
	dbg("start root=%s debounce=%s", w.root, w.debounce)
	ok, fail, err := w.addTree(w.root)
	if err != nil {
		dbg("addTree error: %v", err)
		return err
	}
	if fail > 0 {
		fmt.Fprintf(os.Stderr, "bak: watching %d directories (%d skipped due to limit/error)\n", ok, fail)
	} else {
		fmt.Fprintf(os.Stderr, "bak: watching %d directories\n", ok)
	}

	dirty := make(chan struct{}, 1)
	go w.eventLoop(ctx, dirty)
	go w.debounceLoop(ctx, dirty)
	go w.pollLoop(ctx, dirty)

	<-ctx.Done()
	return w.fsw.Close()
}

func (w *Watcher) addTree(root string) (ok, fail int, err error) {
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, werr error) error {
		if werr != nil {
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		if skipDirs[d.Name()] {
			return filepath.SkipDir
		}
		if addErr := w.fsw.Add(path); addErr != nil {
			dbg("add %s: %v", path, addErr)
			fail++
		} else {
			dbg("watch %s", path)
			ok++
		}
		return nil
	})
	return ok, fail, err
}


func (w *Watcher) ignored(path string) bool {
	clean := filepath.Clean(path)
	gitDir := filepath.Join(w.root, ".git")
	if clean == gitDir || strings.HasPrefix(clean, gitDir+string(filepath.Separator)) {
		return true
	}
	return false
}

func (w *Watcher) eventLoop(ctx context.Context, dirty chan<- struct{}) {
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-w.fsw.Events:
			if !ok {
				return
			}
			dbg("event op=%s path=%s", ev.Op, ev.Name)
			if w.ignored(ev.Name) {
				continue
			}
			if ev.Op&fsnotify.Create != 0 {
				if info, err := os.Stat(ev.Name); err == nil && info.IsDir() {
					_, _, _ = w.addTree(ev.Name)
				}
			}
			select {
			case dirty <- struct{}{}:
			default:
			}
		case _, ok := <-w.fsw.Errors:
			if !ok {
				return
			}
		}
	}
}

func (w *Watcher) debounceLoop(ctx context.Context, dirty <-chan struct{}) {
	var timer *time.Timer
	var fire <-chan time.Time

	for {
		select {
		case <-ctx.Done():
			return
		case <-dirty:
			if timer == nil {
				timer = time.NewTimer(w.debounce)
				fire = timer.C
			} else {
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				timer.Reset(w.debounce)
			}
		case <-fire:
			timer = nil
			fire = nil
			dbg("debounce fire")
			select {
			case w.events <- Event{}:
				dbg("event dispatched")
			default:
				dbg("event channel full, dropped")
			}
		}
	}
}

// pollLoop is the safety net: even when fsnotify misses an event (FD limit,
// symlink, kqueue quirk), polling git's view of the repo every 1.5s detects
// the change and signals the debounce channel. main.go's HTML hash dedup
// suppresses redundant broadcasts, so this is cheap when nothing changed.
func (w *Watcher) pollLoop(ctx context.Context, dirty chan<- struct{}) {
	var last [32]byte
	tick := time.NewTicker(1500 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			cur := w.gitFingerprint()
			if cur != last {
				dbg("poll detected change")
				last = cur
				select {
				case dirty <- struct{}{}:
				default:
				}
			}
		}
	}
}

func (w *Watcher) gitFingerprint() [32]byte {
	var buf bytes.Buffer
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = w.root
	if out, err := cmd.Output(); err == nil {
		buf.Write(out)
	}
	cmd = exec.Command("git", "diff")
	cmd.Dir = w.root
	if out, err := cmd.Output(); err == nil {
		buf.Write(out)
	}
	return sha256.Sum256(buf.Bytes())
}

