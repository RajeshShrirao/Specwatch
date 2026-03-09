package watcher

import (
	"sync"
	"time"
)

// Debouncer manages debouncing for multiple keys (file paths)
type Debouncer struct {
	mu       sync.Mutex
	timers   map[string]*time.Timer
	interval time.Duration
}

// NewDebouncer creates a new Debouncer with the specified interval
func NewDebouncer(interval time.Duration) *Debouncer {
	return &Debouncer{
		timers:   make(map[string]*time.Timer),
		interval: interval,
	}
}

// Debounce schedule a function to be called after the interval.
// If a function is already scheduled for the same key, it is rescheduled.
func (d *Debouncer) Debounce(key string, f func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if timer, ok := d.timers[key]; ok {
		timer.Stop()
	}

	d.timers[key] = time.AfterFunc(d.interval, func() {
		f()
		d.mu.Lock()
		delete(d.timers, key)
		d.mu.Unlock()
	})
}
