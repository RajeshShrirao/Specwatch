package watcher

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	fsWatcher *fsnotify.Watcher
	debouncer *Debouncer
	options   Options
}

type Options struct {
	Debounce   time.Duration
	Extensions []string
}

func NewWatcher(opt Options) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	if opt.Debounce == 0 {
		opt.Debounce = 800 * time.Millisecond
	}

	return &Watcher{
		fsWatcher: fsWatcher,
		debouncer: NewDebouncer(opt.Debounce),
		options:   opt,
	}, nil
}

func (w *Watcher) Watch(root string, onChange func(string)) error {
	// Add root and all subdirectories to watcher
	err := w.addTree(root)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case event, ok := <-w.fsWatcher.Events:
				if !ok {
					return
				}
				if err := w.handleEvent(event, onChange); err != nil {
					fmt.Printf("Watcher error: %v\n", err)
				}
			case err, ok := <-w.fsWatcher.Errors:
				if !ok {
					return
				}
				fmt.Printf("Watcher error: %v\n", err)
			}
		}
	}()

	return nil
}

func (w *Watcher) addTree(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// Skip hidden directories like .git
			if strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
				return filepath.SkipDir
			}
			return w.fsWatcher.Add(path)
		}
		return nil
	})
}

func (w *Watcher) handleEvent(event fsnotify.Event, onChange func(string)) error {
	if isTemporaryFile(event.Name) {
		return nil
	}

	info, err := os.Stat(event.Name)
	if err == nil && info.IsDir() {
		if event.Op&(fsnotify.Create|fsnotify.Rename) != 0 {
			return w.addTree(event.Name)
		}
		return nil
	}

	if event.Op&fsnotify.Remove == fsnotify.Remove {
		if w.matchesExtension(event.Name) {
			w.debouncer.Debounce(event.Name, func() {
				onChange(event.Name)
			})
		}
		return nil
	}

	if err != nil {
		return nil
	}

	if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename|fsnotify.Chmod) == 0 {
		return nil
	}

	if !w.matchesExtension(event.Name) {
		return nil
	}

	w.debouncer.Debounce(event.Name, func() {
		onChange(event.Name)
	})

	return nil
}

func (w *Watcher) matchesExtension(path string) bool {
	if len(w.options.Extensions) == 0 {
		return true
	}

	ext := filepath.Ext(path)
	if ext != "" {
		ext = ext[1:]
	}

	for _, candidate := range w.options.Extensions {
		if strings.EqualFold(candidate, ext) {
			return true
		}
	}

	return false
}

func isTemporaryFile(path string) bool {
	name := filepath.Base(path)
	return strings.HasPrefix(name, ".") ||
		strings.HasSuffix(name, "~") ||
		strings.HasSuffix(name, ".swp") ||
		strings.HasSuffix(name, ".swx") ||
		strings.HasSuffix(name, ".tmp")
}

func (w *Watcher) Close() error {
	return w.fsWatcher.Close()
}
