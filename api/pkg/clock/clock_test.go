package clock

import (
	"testing"
	"time"
)

func TestRealClock_Now(t *testing.T) {
	clock := RealClock{}

	before := time.Now()
	now := clock.Now()
	after := time.Now()

	// Verify that the returned time is between before and after
	if now.Before(before) || now.After(after) {
		t.Errorf("RealClock.Now() returned time outside expected range: got %v, expected between %v and %v", now, before, after)
	}
}

func TestFixedClock_Now(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	clock := NewFixedClock(fixedTime)

	// Verify that Now() returns the fixed time
	now := clock.Now()
	if !now.Equal(fixedTime) {
		t.Errorf("FixedClock.Now() = %v, want %v", now, fixedTime)
	}

	// Verify that multiple calls return the same time
	now2 := clock.Now()
	if !now2.Equal(fixedTime) {
		t.Errorf("FixedClock.Now() second call = %v, want %v", now2, fixedTime)
	}
}

func TestFixedClock_SetTime(t *testing.T) {
	initialTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	clock := NewFixedClock(initialTime)

	// Verify initial time
	if !clock.Now().Equal(initialTime) {
		t.Errorf("Initial time = %v, want %v", clock.Now(), initialTime)
	}

	// Update the time
	newTime := time.Date(2024, 2, 20, 14, 45, 0, 0, time.UTC)
	clock.SetTime(newTime)

	// Verify updated time
	if !clock.Now().Equal(newTime) {
		t.Errorf("After SetTime, Now() = %v, want %v", clock.Now(), newTime)
	}
}

func TestClockInterface(t *testing.T) {
	// Verify that both implementations satisfy the Clock interface
	var _ Clock = RealClock{}
	var _ Clock = &FixedClock{}

	// Test that we can use them interchangeably
	clocks := []Clock{
		RealClock{},
		NewFixedClock(time.Now()),
	}

	for i, clock := range clocks {
		now := clock.Now()
		if now.IsZero() {
			t.Errorf("Clock %d returned zero time", i)
		}
	}
}

func TestFixedClock_UseCaseExample(t *testing.T) {
	// Example use case: testing time-dependent logic
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	clock := NewFixedClock(fixedTime)

	// Simulate advancing time
	clock.SetTime(fixedTime.Add(24 * time.Hour))

	expected := time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC)
	if !clock.Now().Equal(expected) {
		t.Errorf("After advancing 24 hours, Now() = %v, want %v", clock.Now(), expected)
	}
}
