package person

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type PersonRepository interface {
	Create(ctx context.Context, p *Person) error
	GetByID(ctx context.Context, id uuid.UUID) (*Person, error)
	GetByLinkedUserID(ctx context.Context, userID uuid.UUID) (*Person, error)
	GetByInviteEmail(ctx context.Context, email string) (*Person, error)
	GetByInviteToken(ctx context.Context, token string) (*Person, error) // สำหรับ invite link flow
	SetInviteToken(ctx context.Context, personID uuid.UUID, token string, expiresAt time.Time) error
	ClearInviteToken(ctx context.Context, personID uuid.UUID) error
	Update(ctx context.Context, p *Person) error
	LinkUser(ctx context.Context, personID, userID uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error // soft-delete
}
