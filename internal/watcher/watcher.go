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
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
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
				// We only care about write events (file saves)
				if event.Op&fsnotify.Write == fsnotify.Write {
					// Skip directories and temporary files
					if info, err := os.Stat(event.Name); err == nil && !info.IsDir() {
						// Filter by extensions if provided
						if len(w.options.Extensions) > 0 {
							ext := filepath.Ext(event.Name)
							if ext != "" {
								ext = ext[1:] // remove dot
							}
							found := false
							for _, e := range w.options.Extensions {
								if e == ext {
									found = true
									break
								}
							}
							if !found {
								continue
							}
						}

						w.debouncer.Debounce(event.Name, func() {
							onChange(event.Name)
						})
					}
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

func (w *Watcher) Close() error {
	return w.fsWatcher.Close()
}
