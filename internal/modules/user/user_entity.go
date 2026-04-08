package user

import (
	"time"

	"github.com/google/uuid"
)

// User is a pure domain entity. It has no JSON tags by design —
// serialization is handled exclusively by the transport layer (UserResponse).
type User struct {
	ID          uuid.UUID
	Username    string
	Email       string
	Password    string // hashed; never exposed to transport layer
	DisplayName string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time // nil = active; non-nil = soft deleted
}
