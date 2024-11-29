// Package timer provides structures and functions for managing timers.
package timer

import (
	"sync"
	"time"

	"github.com/tencentcloud/dbauth-sdk-go/dbauth/internal/constants"
)

// Manager represents a timer manager.
type Manager struct {
	timers map[string]*time.Timer
	mu     sync.Mutex
}

// NewManager creates a new timer manager.
func NewManager() *Manager {
	return &Manager{
		timers: make(map[string]*time.Timer),
	}
}

// SaveTimer saves a timer with the provided key, delay, and task.
func (tm *Manager) SaveTimer(key string, delay int64, task func()) {
	if key == "" || delay <= 0 || delay > constants.MaxDelay {
		return
	}

	tm.mu.Lock()
	defer tm.mu.Unlock()

	if timer, exists := tm.timers[key]; exists {
		timer.Stop()
		delete(tm.timers, key)
	}

	tm.timers[key] = time.AfterFunc(time.Duration(delay)*time.Millisecond, task)
}
