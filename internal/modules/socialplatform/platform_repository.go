package socialplatform

import (
	"context"

	"github.com/google/uuid"
)

// SocialPlatformRepository handles persistence for master social platform records.
type SocialPlatformRepository interface {
	List(ctx context.Context) ([]*SocialPlatform, error)
	ListOffset(ctx context.Context, limit, offset int) ([]*SocialPlatform, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*SocialPlatform, error)
	GetByKey(ctx context.Context, key string) (*SocialPlatform, error)
	Create(ctx context.Context, p *SocialPlatform) error
	Update(ctx context.Context, p *SocialPlatform) error
	Delete(ctx context.Context, id uuid.UUID) error
}
