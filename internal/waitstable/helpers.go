package waitstable

import (
	"time"
)

// FlagOptions groups flag values commonly used when adding stability flags to commands
type FlagOptions struct {
	// Enabled indicates if stability monitoring is enabled
	Enabled bool
	// Tolerance is the stability tolerance level (low/normal/high)
	Tolerance string
	// MaxWait is the maximum total time to wait
	MaxWait time.Duration
	// Threshold is the check frequency
	Threshold time.Duration
}

// ComputeConfig creates a Config from FlagOptions, computing the appropriate timeout.
// When enabled is true, it uses the tolerance level to determine the wait duration.
// Returns a config with WaitStable=0 if not enabled (which makes IsStable() return false).
func (o FlagOptions) ComputeConfig() Config {
	if !o.Enabled {
		return Config{
			WaitStable: 0,
			Tolerance:  Tolerance(o.Tolerance),
			MaxWait:    o.MaxWait,
			Threshold:  o.Threshold,
		}
	}

	// When enabled, use the tolerance level to determine the wait duration
	duration := ToleranceDuration(Tolerance(o.Tolerance))

	return Config{
		WaitStable: duration,
		Tolerance:  Tolerance(o.Tolerance),
		MaxWait:    o.MaxWait,
		Threshold:  o.Threshold,
	}
}

// IsEnabled returns true if stability monitoring should be active
func (o FlagOptions) IsEnabled() bool {
	return o.Enabled
}
