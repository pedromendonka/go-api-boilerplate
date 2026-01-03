// Package timeutil provides utility functions for working with time.
package timeutil

import "time"

// Now returns current time (wrapper to ease testing)
func Now() time.Time { return time.Now() }
