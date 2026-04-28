package domain

import "sync"

// IAppConfig manages global comparison settings.
// Interface must be commented clear for each method.
type IAppConfig interface {
	// GetThreshold returns the universal similarity threshold (0.0 to 1.0).
	// Assume that Interface provide data exactly as comment.
	GetThreshold() float64

	// UseAI returns true if AI-based matching should be used, false for fuzzy string matching.
	// Assume that Interface provide data exactly as comment.
	UseAI() bool

	// SetThreshold updates the universal similarity threshold.
	SetThreshold(val float64)

	// SetUseAI updates the AI matching toggle.
	SetUseAI(val bool)
}

// GlobalConfig is the concrete implementation of IAppConfig.
type GlobalConfig struct {
	mu        sync.RWMutex
	threshold float64
	useAI     bool
}

// DefaultConfig is the global configuration instance for portability.
var DefaultConfig = &GlobalConfig{
	threshold: 0.4,
	useAI:     true, // Default to fuzzy as requested or safe default
}

func (c *GlobalConfig) GetThreshold() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.threshold
}

func (c *GlobalConfig) UseAI() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.useAI
}

func (c *GlobalConfig) SetThreshold(val float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.threshold = val
}

func (c *GlobalConfig) SetUseAI(val bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.useAI = val
}

var _ IAppConfig = (*GlobalConfig)(nil)
