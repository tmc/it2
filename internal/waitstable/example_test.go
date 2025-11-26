package waitstable

import (
	"fmt"
	"time"
)

// Example demonstrates basic usage of the Detector.
func Example() {
	config := Config{
		WaitStable: 2 * time.Second,
		Threshold:  500 * time.Millisecond,
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
	fmt.Println("Time since last change: some duration")

	// Wait for stability
	time.Sleep(2500 * time.Millisecond)
	fmt.Printf("After wait - Is stable: %v\n", d.IsStable())

	// Output:
	// Change recorded
	// Another change recorded
	// Is stable: false
	// Time since last change: some duration
	// After wait - Is stable: true
}

// ExampleNew demonstrates using a change callback.
func ExampleNew() {
	changeCount := 0

	onChange := func(timeSinceLastChange time.Duration) {
		changeCount++
		fmt.Printf("Change #%d detected\n", changeCount)
	}

	config := Config{
		WaitStable: 1 * time.Second,
		Threshold:  100 * time.Millisecond,
	}

	d := New(config, onChange)

	// Record changes
	d.RecordChange()
	time.Sleep(200 * time.Millisecond)
	d.RecordChange()

	fmt.Printf("Total changes recorded: %d\n", changeCount)

	// Output:
	// Change #1 detected
	// Change #2 detected
	// Total changes recorded: 2
}

// ExampleDetector_IsStable demonstrates monitoring a stream for stability.
func ExampleDetector_IsStable() {
	config := Config{
		WaitStable: 1500 * time.Millisecond,
		Threshold:  500 * time.Millisecond,
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
