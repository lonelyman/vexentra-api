package socialplatform

import (
	"time"

	"github.com/google/uuid"
)

// SocialPlatform defines an available social/contact platform.
// Managed by admins; drives UI rendering and user social-link validation.
type SocialPlatform struct {
	ID        uuid.UUID
	Key       string // unique slug e.g. "github" — referenced by social_links.platform
	Name      string // display name e.g. "GitHub"
	IconURL   string // optional icon asset URL
	SortOrder int
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
