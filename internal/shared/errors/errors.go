// Package apperrors defines custom error types and helpers for the application.
package apperrors

import "fmt"

// Wrapf wraps an error with a formatted message, returns nil if err is nil.
func Wrapf(err error, fmtStr string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	prefix := fmt.Sprintf(fmtStr, args...)
	return fmt.Errorf("%s: %w", prefix, err)
}
