package waitstable

import (
	"sync"
	"testing"
	"time"
)

func TestDetectorIsStable(t *testing.T) {
	tests := []struct {
		name              string
		config            Config
		setup             func(*Detector)
		expectedStable    bool
		expectedWaitTime  time.Duration
	}{
		{
			name: "stability disabled returns false",
			config: Config{
				WaitStable: 0,
				Threshold:  500 * time.Millisecond,
			},
			expectedStable: false,
		},
		{
			name: "immediately after creation is not stable",
			config: Config{
				WaitStable: 1 * time.Second,
				Threshold:  500 * time.Millisecond,
			},
			expectedStable: false,
		},
		{
			name: "after timeout duration passes is stable",
			config: Config{
				WaitStable: 100 * time.Millisecond,
				Threshold:  10 * time.Millisecond,
			},
			setup: func(d *Detector) {
				time.Sleep(150 * time.Millisecond)
			},
			expectedStable: true,
		},
		{
			name: "change resets stability timer",
			config: Config{
				WaitStable: 100 * time.Millisecond,
				Threshold:  10 * time.Millisecond,
			},
			setup: func(d *Detector) {
				time.Sleep(50 * time.Millisecond)
				d.RecordChange()
				time.Sleep(50 * time.Millisecond)
			},
			expectedStable: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d := New(test.config, nil)

			if test.setup != nil {
				test.setup(d)
			}

			if got := d.IsStable(); got != test.expectedStable {
				t.Errorf("IsStable() = %v, want %v", got, test.expectedStable)
			}
		})
	}
}

func TestDetectorRecordChange(t *testing.T) {
	callCount := 0
	var recordedTime time.Duration

	onChange := func(timeSinceLastChange time.Duration) {
		callCount++
		recordedTime = timeSinceLastChange
	}

	config := Config{
		WaitStable: 1 * time.Second,
		Threshold:  500 * time.Millisecond,
	}

	d := New(config, onChange)

	// First change should be recorded
	time.Sleep(100 * time.Millisecond)
	d.RecordChange()

	if callCount != 1 {
		t.Errorf("callback called %d times, want 1", callCount)
	}

	if recordedTime < 100*time.Millisecond || recordedTime > 150*time.Millisecond {
		t.Errorf("recorded time %v outside expected range", recordedTime)
	}

	// Second change should also be recorded
	time.Sleep(50 * time.Millisecond)
	d.RecordChange()

	if callCount != 2 {
		t.Errorf("callback called %d times, want 2", callCount)
	}
}

func TestDetectorReset(t *testing.T) {
	config := Config{
		WaitStable: 100 * time.Millisecond,
		Threshold:  10 * time.Millisecond,
	}

	d := New(config, nil)

	// Wait for stability would be achieved
	time.Sleep(150 * time.Millisecond)
	if !d.IsStable() {
		t.Error("should be stable before reset")
	}

	// Reset should clear stability
	d.Reset()
	if d.IsStable() {
		t.Error("should not be stable immediately after reset")
	}

	// Wait for stability again
	time.Sleep(150 * time.Millisecond)
	if !d.IsStable() {
		t.Error("should be stable again after timeout")
	}
}

func TestDetectorTimeSinceLastChange(t *testing.T) {
	config := Config{
		WaitStable: 1 * time.Second,
		Threshold:  50 * time.Millisecond,
	}

	d := New(config, nil)

	// Immediately should be close to 0
	if d.TimeSinceLastChange() > 10*time.Millisecond {
		t.Error("time since creation should be very small")
	}

	// After sleep, should reflect elapsed time
	time.Sleep(100 * time.Millisecond)
	elapsed := d.TimeSinceLastChange()
	if elapsed < 90*time.Millisecond || elapsed > 150*time.Millisecond {
		t.Errorf("TimeSinceLastChange() = %v, want ~100ms", elapsed)
	}

	// After RecordChange, should reset
	d.RecordChange()
	time.Sleep(50 * time.Millisecond)
	elapsed = d.TimeSinceLastChange()
	if elapsed > 100*time.Millisecond {
		t.Errorf("TimeSinceLastChange() after reset = %v, want ~50ms", elapsed)
	}
}

func TestDetectorSetOnChange(t *testing.T) {
	config := Config{
		WaitStable: 1 * time.Second,
		Threshold:  500 * time.Millisecond,
	}

	d := New(config, nil)

	// Initially no callback
	d.RecordChange() // Should not panic

	// Set callback
	callCount := 0
	d.SetOnChange(func(t time.Duration) {
		callCount++
	})

	d.RecordChange()
	if callCount != 1 {
		t.Errorf("callback not called after SetOnChange")
	}
}

func TestDetectorCheckAndReport(t *testing.T) {
	config := Config{
		WaitStable: 100 * time.Millisecond,
		Threshold:  10 * time.Millisecond,
	}

	d := New(config, nil)

	// Initially not stable
	elapsed, stable := d.CheckAndReport()
	if stable {
		t.Error("should not be stable immediately")
	}
	if elapsed > 10*time.Millisecond {
		t.Errorf("elapsed time should be near 0, got %v", elapsed)
	}

	// After timeout
	time.Sleep(150 * time.Millisecond)
	elapsed, stable = d.CheckAndReport()
	if !stable {
		t.Error("should be stable after timeout")
	}
	if elapsed < 140*time.Millisecond {
		t.Errorf("elapsed time should be ~150ms, got %v", elapsed)
	}
}

func TestDetectorConcurrency(t *testing.T) {
	config := Config{
		WaitStable: 100 * time.Millisecond,
		Threshold:  10 * time.Millisecond,
	}

	d := New(config, nil)
	var mu sync.Mutex
	callCount := 0

	d.SetOnChange(func(t time.Duration) {
		mu.Lock()
		callCount++
		mu.Unlock()
	})

	// Spawn multiple goroutines recording changes
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(time.Duration(i*10) * time.Millisecond)
			d.RecordChange()
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(time.Duration(i*15) * time.Millisecond)
			_ = d.IsStable()
		}()
	}

	wg.Wait()

	mu.Lock()
	if callCount < 10 {
		t.Errorf("callback called %d times, expected at least 10", callCount)
	}
	mu.Unlock()
}

func TestDetectorConfig(t *testing.T) {
	config := Config{
		WaitStable: 1 * time.Second,
		Tolerance:  ToleranceNormal,
		Threshold:  500 * time.Millisecond,
	}

	d := New(config, nil)
	retrieved := d.Config()

	if retrieved.WaitStable != config.WaitStable {
		t.Errorf("Config.WaitStable = %v, want %v", retrieved.WaitStable, config.WaitStable)
	}
	if retrieved.Tolerance != config.Tolerance {
		t.Errorf("Config.Tolerance = %v, want %v", retrieved.Tolerance, config.Tolerance)
	}
	if retrieved.Threshold != config.Threshold {
		t.Errorf("Config.Threshold = %v, want %v", retrieved.Threshold, config.Threshold)
	}
}

func TestToleranceDuration(t *testing.T) {
	tests := []struct {
		tolerance Tolerance
		expected  time.Duration
	}{
		{ToleranceLow, 500 * time.Millisecond},
		{ToleranceNormal, 2 * time.Second},
		{ToleranceHigh, 5 * time.Second},
		{Tolerance("unknown"), 2 * time.Second}, // Default to normal
	}

	for _, test := range tests {
		got := ToleranceDuration(test.tolerance)
		if got != test.expected {
			t.Errorf("ToleranceDuration(%q) = %v, want %v", test.tolerance, got, test.expected)
		}
	}
}

func TestValidateTolerance(t *testing.T) {
	tests := []struct {
		tolerance string
		valid     bool
	}{
		{"low", true},
		{"normal", true},
		{"high", true},
		{"invalid", false},
		{"", false},
	}

	for _, test := range tests {
		err := ValidateTolerance(test.tolerance)
		if (err == nil) != test.valid {
			t.Errorf("ValidateTolerance(%q) error = %v, valid = %v", test.tolerance, err, test.valid)
		}
	}
}

func TestComputeTimeout(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		expected time.Duration
	}{
		{
			name: "explicit wait-stable takes precedence",
			config: Config{
				WaitStable: 3 * time.Second,
				Tolerance:  ToleranceLow,
			},
			expected: 3 * time.Second,
		},
		{
			name: "tolerance used when wait-stable is 0",
			config: Config{
				WaitStable: 0,
				Tolerance:  ToleranceLow,
			},
			expected: 500 * time.Millisecond,
		},
		{
			name: "normal tolerance default",
			config: Config{
				WaitStable: 0,
				Tolerance:  ToleranceNormal,
			},
			expected: 2 * time.Second,
		},
	}

	for _, test := range tests {
		got := test.config.ComputeTimeout()
		if got != test.expected {
			t.Errorf("%s: ComputeTimeout() = %v, want %v", test.name, got, test.expected)
		}
	}
}

func TestDetectorMaxWait(t *testing.T) {
	config := Config{
		WaitStable: 1 * time.Second,
		Tolerance:  ToleranceNormal,
		MaxWait:    200 * time.Millisecond,
		Threshold:  50 * time.Millisecond,
	}

	d := New(config, nil)

	// Should not be stable immediately
	if d.IsStable() {
		t.Error("should not be stable immediately")
	}

	// After MaxWait exceeded, should be stable regardless of changes
	time.Sleep(250 * time.Millisecond)
	if !d.IsStable() {
		t.Error("should be stable after MaxWait exceeded")
	}
}

func TestDetectorTimeSinceCreation(t *testing.T) {
	config := Config{
		WaitStable: 1 * time.Second,
		Tolerance:  ToleranceNormal,
	}

	d := New(config, nil)

	// Should be very small initially
	if d.TimeSinceCreation() > 10*time.Millisecond {
		t.Error("TimeSinceCreation should be very small initially")
	}

	// After sleep should reflect elapsed time
	time.Sleep(100 * time.Millisecond)
	elapsed := d.TimeSinceCreation()
	if elapsed < 90*time.Millisecond || elapsed > 150*time.Millisecond {
		t.Errorf("TimeSinceCreation() = %v, want ~100ms", elapsed)
	}
}

func TestDetectorTimeUntilMaxWait(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		setup     func(*Detector)
		checkFunc func(*Detector) bool
	}{
		{
			name: "no max-wait returns 0",
			config: Config{
				WaitStable: 1 * time.Second,
				Tolerance:  ToleranceNormal,
				MaxWait:    0,
			},
			checkFunc: func(d *Detector) bool {
				return d.TimeUntilMaxWait() == 0
			},
		},
		{
			name: "time remaining decreases",
			config: Config{
				WaitStable: 1 * time.Second,
				Tolerance:  ToleranceNormal,
				MaxWait:    200 * time.Millisecond,
			},
			setup: func(d *Detector) {
				time.Sleep(50 * time.Millisecond)
			},
			checkFunc: func(d *Detector) bool {
				remaining := d.TimeUntilMaxWait()
				return remaining > 100*time.Millisecond && remaining < 200*time.Millisecond
			},
		},
		{
			name: "returns 0 after max-wait exceeded",
			config: Config{
				WaitStable: 1 * time.Second,
				Tolerance:  ToleranceNormal,
				MaxWait:    100 * time.Millisecond,
			},
			setup: func(d *Detector) {
				time.Sleep(150 * time.Millisecond)
			},
			checkFunc: func(d *Detector) bool {
				return d.TimeUntilMaxWait() == 0
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d := New(test.config, nil)
			if test.setup != nil {
				test.setup(d)
			}
			if !test.checkFunc(d) {
				t.Error("check function failed")
			}
		})
	}
}
