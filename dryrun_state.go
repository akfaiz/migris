package migris

import (
	"sync"
)

var (
	// Global state for dry-run configuration
	// This is used by migration contexts to determine if they should use dry-run mode
	globalDryRunMu    sync.RWMutex
	globalDryRunState bool
)

// setGlobalDryRunState sets the global dry-run state
func setGlobalDryRunState(enabled bool) {
	globalDryRunMu.Lock()
	defer globalDryRunMu.Unlock()
	globalDryRunState = enabled
}

// getGlobalDryRunState returns the current global dry-run state
func getGlobalDryRunState() bool {
	globalDryRunMu.RLock()
	defer globalDryRunMu.RUnlock()
	return globalDryRunState
}
