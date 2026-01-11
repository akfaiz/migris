package migris

import (
	"sync"

	"github.com/akfaiz/migris/schema"
)

var (
	// Global state for dry-run configuration
	// This is used by migration contexts to determine if they should use dry-run mode
	globalDryRunMu     sync.RWMutex
	globalDryRunState  bool
	globalDryRunConfig schema.DryRunConfig
)

// setGlobalDryRunState sets the global dry-run state
func setGlobalDryRunState(enabled bool, config schema.DryRunConfig) {
	globalDryRunMu.Lock()
	defer globalDryRunMu.Unlock()
	globalDryRunState = enabled
	globalDryRunConfig = config
}

// getGlobalDryRunState returns the current global dry-run state
func getGlobalDryRunState() (bool, schema.DryRunConfig) {
	globalDryRunMu.RLock()
	defer globalDryRunMu.RUnlock()
	return globalDryRunState, globalDryRunConfig
}
