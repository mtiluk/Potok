package sync

import (
	"fmt"
	"io/fs"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

var ignoreDirs = map[string]bool{
	".potok": true,
	".git":   true,
}

type EventType int

const (
	EventCreate EventType = iota
	EventUpdate
	EventDelete
	EventRename
)

func (e EventType) String() string {
	switch e {
	case EventCreate:
		return "CREATE"
	case EventUpdate:
		return "UPDATE"
	case EventDelete:
		return "DELETE"
	case EventRename:
		return "RENAME"
	default:
		return "UNKNOWN"
	}
}

type FileEvent struct {
	Type    EventType
	RelPath string
	AbsPath string
}

type Watcher struct {
	vaultRoot string
	fsw       *fsnotify.Watcher
	Events    chan []FileEvent
	Errors    chan error
	debounce  time.Duration
	done      chan struct{}
	logger    *slog.Logger
}

func NewWatcher(vaultRoot string, debounce time.Duration, logger *slog.Logger) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("create fsnotify watcher: %w", err)
	}

	w := &Watcher{
		vaultRoot: vaultRoot,
		fsw:       fsw,
		Events:    make(chan []FileEvent, 16),
		Errors:    make(chan error, 8),
		debounce:  debounce,
		done:      make(chan struct{}),
		logger:    logger,
	}

	if err := w.addRecursive(vaultRoot); err != nil {
		fsw.Close()
		return nil, fmt.Errorf("initial walk: %w", err)
	}

	return w, nil
}

func (w *Watcher) addRecursive(root string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		if ignoreDirs[d.Name()] {
			return filepath.SkipDir
		}
		w.logger.Debug("watching directory", "path", path)
		return w.fsw.Add(path)
	})
}

func (w *Watcher) relPath(absPath string) string {
	rel, err := filepath.Rel(w.vaultRoot, absPath)
	if err != nil {
		return absPath
	}
	return filepath.ToSlash(rel)
}

func (w *Watcher) shouldIgnore(path string) bool {
	rel := w.relPath(path)
	parts := strings.Split(rel, "/")
	for _, p := range parts {
		if ignoreDirs[p] {
			return true
		}
	}
	return false
}

func classify(e fsnotify.Event) EventType {
	switch {
	case e.Has(fsnotify.Remove):
		return EventDelete
	case e.Has(fsnotify.Rename):
		return EventRename
	case e.Has(fsnotify.Create):
		return EventCreate
	case e.Has(fsnotify.Write):
		return EventUpdate
	default:
		return EventUpdate
	}
}

func (w *Watcher) Start() {
	pending := make(map[string]FileEvent)
	var timer *time.Timer
	var timerC <-chan time.Time

	defer close(w.Events)
	defer close(w.Errors)

	for {
		select {
		case <-w.done:
			if len(pending) > 0 {
				w.flush(pending)
			}
			return

		case ev, ok := <-w.fsw.Events:
			if !ok {
				return
			}

			if w.shouldIgnore(ev.Name) {
				continue
			}

			if ev.Has(fsnotify.Create) {
				_ = filepath.WalkDir(ev.Name, func(path string, d fs.DirEntry, err error) error {
					if err != nil {
						return nil
					}
					if d.IsDir() && !ignoreDirs[d.Name()] {
						_ = w.fsw.Add(path)
					}
					return nil
				})
			}

			rel := w.relPath(ev.Name)
			fe := FileEvent{
				Type:    classify(ev),
				RelPath: rel,
				AbsPath: ev.Name,
			}

			pending[rel] = fe

			if timer == nil {
				timer = time.NewTimer(w.debounce)
				timerC = timer.C
			} else {
				timer.Reset(w.debounce)
			}

		case <-timerC:
			w.flush(pending)
			pending = make(map[string]FileEvent)
			timer = nil
			timerC = nil

		case err, ok := <-w.fsw.Errors:
			if !ok {
				return
			}
			w.Errors <- err
		}
	}
}

func (w *Watcher) flush(pending map[string]FileEvent) {
	batch := make([]FileEvent, 0, len(pending))
	for _, fe := range pending {
		batch = append(batch, fe)
	}
	if len(batch) > 0 {
		w.Events <- batch
	}
}

func (w *Watcher) Close() {
	close(w.done)
	w.fsw.Close()
}
