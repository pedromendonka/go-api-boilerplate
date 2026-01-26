package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Context keys (unexported to prevent collisions).
const (
	userIDKey = "user_id"
)

// SetUserID stores the authenticated user's ID in the Gin context.
func SetUserID(c *gin.Context, id uuid.UUID) {
	c.Set(userIDKey, id)
}

// GetUserID retrieves the authenticated user's ID from the Gin context.
// Returns the ID and true if found, or zero UUID and false if not set.
func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	v, ok := c.Get(userIDKey)
	if !ok {
		return uuid.UUID{}, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}
