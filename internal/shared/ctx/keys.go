// Package ctx defines shared context keys for Gin context values to avoid stringly-typed keys.
package ctx

// contextKey is a private type for context keys to avoid collisions.
type contextKey string

const (
	// User holds the raw JWT claims map (jwt.MapClaims)
	User contextKey = "user"

	// UserID holds the typed user id (github.com/google/uuid.UUID)
	UserID contextKey = "user_id"

	// CurrentUser holds the application-level current user object (e.g., UserResponse)
	CurrentUser contextKey = "current_user"
)
