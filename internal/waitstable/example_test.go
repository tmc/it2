package waitstable

import (
	"fmt"
	"time"
)

// ExampleDetector_Basic demonstrates basic usage of the Detector.
func ExampleDetector_Basic() {
	config := Config{
		Timeout:   2 * time.Second,
		Threshold: 500 * time.Millisecond,
	}

	d := New(config, nil)

	// Simulate changes
	d.RecordChange()
	fmt.Println("Change recorded")

	time.Sleep(500 * time.Millisecond)
	d.RecordChange()
	fmt.Println("Another change recorded")

	// Check stability
	fmt.Printf("Is stable: %v\n", d.IsStable())
	fmt.Printf("Time since last change: %v\n", d.TimeSinceLastChange())

	// Wait for stability
	time.Sleep(2500 * time.Millisecond)
	fmt.Printf("After wait - Is stable: %v\n", d.IsStable())

	// Output:
	// Change recorded
	// Another change recorded
	// Is stable: false
	// Time since last change: ~500ms
	// After wait - Is stable: true
}

// ExampleDetector_WithCallback demonstrates using a change callback.
func ExampleDetector_WithCallback() {
	changeCount := 0

	onChange := func(timeSinceLastChange time.Duration) {
		changeCount++
		fmt.Printf("Change #%d detected (%.0fs since last change)\n", changeCount, timeSinceLastChange.Seconds())
	}

	config := Config{
		Timeout:   1 * time.Second,
		Threshold: 100 * time.Millisecond,
	}

	d := New(config, onChange)

	// Record changes
	d.RecordChange()
	time.Sleep(200 * time.Millisecond)
	d.RecordChange()

	fmt.Printf("Total changes recorded: %d\n", changeCount)

	// Output:
	// Change #1 detected (0s since last change)
	// Change #2 detected (~0.2s since last change)
	// Total changes recorded: 2
}

// ExampleDetector_StreamMonitoring demonstrates monitoring a stream for stability.
func ExampleDetector_StreamMonitoring() {
	config := Config{
		Timeout:   1500 * time.Millisecond,
		Threshold: 500 * time.Millisecond,
	}

	d := New(config, nil)

	// Simulate stream activity
	fmt.Println("Monitoring stream...")

	for i := 0; i < 5; i++ {
		time.Sleep(300 * time.Millisecond)
		d.RecordChange()
		fmt.Printf("Event %d - Stable: %v\n", i+1, d.IsStable())
	}

	// Wait for final stability
	fmt.Println("Waiting for stability...")
	for !d.IsStable() {
		time.Sleep(100 * time.Millisecond)
	}
	fmt.Println("Stream stable!")

	// Output:
	// Monitoring stream...
	// Event 1 - Stable: false
	// Event 2 - Stable: false
	// Event 3 - Stable: false
	// Event 4 - Stable: false
	// Event 5 - Stable: false
	// Waiting for stability...
	// Stream stable!
}
