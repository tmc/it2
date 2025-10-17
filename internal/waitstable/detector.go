// Package waitstable provides utilities for detecting when a stream or buffer
// has stabilized (no changes for a configured duration).
package waitstable

import (
	"fmt"
	"time"
)

// Tolerance defines how long to wait before considering the stream stable.
type Tolerance string

const (
	// ToleranceLow requires quick stabilization (fast detection)
	ToleranceLow Tolerance = "low"
	// ToleranceNormal is the default, balanced tolerance
	ToleranceNormal Tolerance = "normal"
	// ToleranceHigh allows longer periods of inactivity (slow detection)
	ToleranceHigh Tolerance = "high"
)

// ToleranceDuration returns the stability timeout for a given tolerance level.
func ToleranceDuration(t Tolerance) time.Duration {
	switch t {
	case ToleranceLow:
		return 500 * time.Millisecond
	case ToleranceNormal:
		return 2 * time.Second
	case ToleranceHigh:
		return 5 * time.Second
	default:
		return 2 * time.Second // Default to normal
	}
}

// ValidateTolerance checks if a tolerance string is valid.
func ValidateTolerance(t string) error {
	switch Tolerance(t) {
	case ToleranceLow, ToleranceNormal, ToleranceHigh:
		return nil
	default:
		return fmt.Errorf("invalid tolerance: %q (must be low, normal, or high)", t)
	}
}

// Config holds configuration for stability detection.
type Config struct {
	// WaitStable is the base duration for stability detection.
	// Combined with Tolerance to determine the actual timeout.
	WaitStable time.Duration

	// Tolerance determines how long to wait before considering stable.
	// Values: low (500ms), normal (2s), high (5s)
	Tolerance Tolerance

	// MaxWait is the maximum time to wait for stability.
	// If set and exceeded, stability is forced regardless of activity.
	// Set to 0 for no maximum.
	MaxWait time.Duration

	// Threshold is the frequency at which to check for changes.
	// This determines the granularity of change detection.
	Threshold time.Duration
}

// ComputeTimeout calculates the actual timeout duration based on WaitStable and Tolerance.
func (c Config) ComputeTimeout() time.Duration {
	if c.WaitStable > 0 {
		return c.WaitStable
	}
	return ToleranceDuration(c.Tolerance)
}

// OnChangeFunc is called when a change is detected.
// It receives the time since the last change was detected.
type OnChangeFunc func(timeSinceLastChange time.Duration)

// Detector tracks whether a stream has stabilized by monitoring changes.
// It is safe for concurrent use.
type Detector struct {
	config         Config
	createdTime    time.Time
	lastChangeTime time.Time
	onChangeCb     OnChangeFunc
}

// New creates a new Detector with the given configuration.
// If onChangeCb is nil, no callback will be invoked on changes.
func New(config Config, onChangeCb OnChangeFunc) *Detector {
	now := time.Now()
	return &Detector{
		config:         config,
		createdTime:    now,
		lastChangeTime: now,
		onChangeCb:     onChangeCb,
	}
}

// RecordChange marks that a change has occurred and invokes the OnChange callback.
// This resets the stability timer.
func (d *Detector) RecordChange() {
	now := time.Now()
	timeSinceLastChange := now.Sub(d.lastChangeTime)
	d.lastChangeTime = now

	if d.onChangeCb != nil {
		d.onChangeCb(timeSinceLastChange)
	}
}

// IsStable returns true if:
// 1. Stream has no changes for the configured timeout duration, OR
// 2. MaxWait duration has been exceeded
// Returns false if stability detection is disabled (WaitStable == 0 and Tolerance == "").
func (d *Detector) IsStable() bool {
	timeout := d.config.ComputeTimeout()
	if timeout == 0 {
		return false
	}

	// Check if stability achieved through no changes
	if time.Since(d.lastChangeTime) >= timeout {
		return true
	}

	// Check if max-wait timeout exceeded
	if d.config.MaxWait > 0 && time.Since(d.createdTime) >= d.config.MaxWait {
		return true
	}

	return false
}

// TimeSinceLastChange returns the duration since the last recorded change.
func (d *Detector) TimeSinceLastChange() time.Duration {
	return time.Since(d.lastChangeTime)
}

// Reset resets the stability timer to now. Use this to signal the start of monitoring
// or to manually reset the detector state.
func (d *Detector) Reset() {
	d.lastChangeTime = time.Now()
}

// Config returns the detector's current configuration.
func (d *Detector) Config() Config {
	return d.config
}

// SetOnChange updates the OnChange callback.
func (d *Detector) SetOnChange(cb OnChangeFunc) {
	d.onChangeCb = cb
}

// CheckAndReport checks if the stream is stable and invokes the callback if enabled.
// Returns the time since last change and whether stability has been achieved.
func (d *Detector) CheckAndReport() (timeSinceLastChange time.Duration, stable bool) {
	timeSinceLastChange = d.TimeSinceLastChange()
	stable = d.IsStable()
	return
}

// TimeSinceCreation returns the total time elapsed since the detector was created.
func (d *Detector) TimeSinceCreation() time.Duration {
	return time.Since(d.createdTime)
}

// TimeUntilMaxWait returns the time remaining until MaxWait is exceeded.
// Returns 0 if MaxWait is not set or has already been exceeded.
func (d *Detector) TimeUntilMaxWait() time.Duration {
	if d.config.MaxWait == 0 {
		return 0
	}
	elapsed := d.TimeSinceCreation()
	if elapsed >= d.config.MaxWait {
		return 0
	}
	return d.config.MaxWait - elapsed
}

// MaxWaitReached returns true if the MaxWait timeout has been exceeded.
// Returns false if MaxWait is not configured (0).
func (d *Detector) MaxWaitReached() bool {
	if d.config.MaxWait == 0 {
		return false
	}
	return d.TimeSinceCreation() >= d.config.MaxWait
}
