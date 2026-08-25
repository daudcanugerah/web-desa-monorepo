package clock

import "time"

// Clock is an interface for time operations to enable testing
// Requirement 29.3: Use clock.Clock interface for all time operations to enable testing
type Clock interface {
	// Now returns the current time
	Now() time.Time
}

// RealClock implements Clock using the actual system time
type RealClock struct{}

// Now returns the current system time
func (RealClock) Now() time.Time {
	return time.Now()
}

// FixedClock implements Clock with a fixed time for testing
type FixedClock struct {
	time time.Time
}

// NewFixedClock creates a new FixedClock with the specified time
func NewFixedClock(t time.Time) *FixedClock {
	return &FixedClock{time: t}
}

// Now returns the fixed time
func (c *FixedClock) Now() time.Time {
	return c.time
}

// SetTime updates the fixed time
func (c *FixedClock) SetTime(t time.Time) {
	c.time = t
}
