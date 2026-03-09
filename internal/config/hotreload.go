package config

import (
	"fmt"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// HotReloader handles configuration hot-reloading
type HotReloader struct {
	configPath string
	loader     *Loader
	config     *Config
	mu         sync.RWMutex
	watcher    *fsnotify.Watcher
	onChange   func(*Config) error
	stop       chan struct{}
}

// NewHotReloader creates a new configuration hot-reloader
func NewHotReloader(configPath string, opts ...LoaderOption) (*HotReloader, error) {
	loaderOpts := append([]LoaderOption{WithConfigPath(configPath)}, opts...)
	loader := NewLoader(loaderOpts...)

	reloader := &HotReloader{
		configPath: configPath,
		loader:     loader,
		stop:       make(chan struct{}),
	}

	// Load initial configuration
	if err := reloader.reload(); err != nil {
		return nil, err
	}

	return reloader, nil
}

// SetOnChange sets the callback function to be called when config changes
func (h *HotReloader) SetOnChange(callback func(*Config) error) {
	h.onChange = callback
}

// Start starts watching for configuration changes
func (h *HotReloader) Start() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	h.watcher = watcher

	// Watch the config file directory
	dir := h.configPath
	if dir == "" {
		dir = "."
	}
	if err := watcher.Add(dir); err != nil {
		err := watcher.Close()
		if err != nil {
			return fmt.Errorf("failed to close watcher: %w", err)
		}
		return fmt.Errorf("failed to watch config path: %w", err)
	}

	go h.watchLoop()

	return nil
}

// watchLoop watches for configuration changes
func (h *HotReloader) watchLoop() {
	defer h.watcher.Close()

	for {
		select {
		case <-h.stop:
			return
		case event, ok := <-h.watcher.Events:
			if !ok {
				return
			}

			// Only handle write events
			if event.Op&fsnotify.Write == fsnotify.Write {
				// Debounce - wait a bit before reloading
				select {
				case <-h.stop:
					return
				case <-time.After(100 * time.Millisecond):
					if err := h.reload(); err != nil {
						fmt.Printf("Error reloading config: %v\n", err)
						continue
					}

					// Call the onChange callback if set
					if h.onChange != nil {
						h.mu.RLock()
						cfg := h.config
						h.mu.RUnlock()

						if err := h.onChange(cfg); err != nil {
							fmt.Printf("Error in config change callback: %v\n", err)
						}
					}
				}
			}
		case err, ok := <-h.watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("Watcher error: %v\n", err)
		}
	}
}

// reload reloads the configuration
func (h *HotReloader) reload() error {
	newConfig, err := h.loader.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	h.mu.Lock()
	h.config = newConfig
	h.mu.Unlock()

	return nil
}

// Get returns the current configuration
func (h *HotReloader) Get() *Config {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.config
}

// Stop stops the hot-reloader
func (h *HotReloader) Stop() error {
	close(h.stop)
	if h.watcher != nil {
		return h.watcher.Close()
	}
	return nil
}

// WatchFile creates a new hot-reloader for a specific config file
func WatchFile(path string) (*HotReloader, error) {
	return NewHotReloader(path)
}

// WatchDirectory watches all config files in a directory
func WatchDirectory(dir string) (*HotReloader, error) {
	configFile := FindConfigFile(dir)
	if configFile == "" {
		return nil, fmt.Errorf("no config file found in directory: %s", dir)
	}
	return NewHotReloader(configFile)
}

// ConfigWithHotReload creates a config with automatic reloading
type ConfigWithHotReload struct {
	*HotReloader
}

// NewConfigWithHotReload creates a new config with hot-reloading enabled
func NewConfigWithHotReload(configPath string) (*ConfigWithHotReload, error) {
	reloader, err := NewHotReloader(configPath)
	if err != nil {
		return nil, err
	}
	return &ConfigWithHotReload{HotReloader: reloader}, nil
}

// Get is a convenience method to get the current config
func (c *ConfigWithHotReload) Get() *Config {
	return c.HotReloader.Get()
}
