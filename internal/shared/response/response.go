// Package response provides utilities for HTTP responses.
package response

// Envelope standardizes API responses
type Envelope struct {
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

// Success helper
func Success(data interface{}) Envelope {
	return Envelope{Data: data}
}

// Fail helper
func Fail(err string) Envelope {
	return Envelope{Error: err}
}
