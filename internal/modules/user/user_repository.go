package user

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type UserRepository interface {
	// ─── User CRUD ───────────────────────────────────────────────────────────
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByPersonID(ctx context.Context, personID uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)

	// ─── สถานะ & การ tracking ────────────────────────────────────────────────
	UpdateStatus(ctx context.Context, userID uuid.UUID, status string) error
	UpdateRole(ctx context.Context, userID uuid.UUID, role string) error
	UpdateLastLogin(ctx context.Context, userID uuid.UUID, t time.Time) error
	UpdatePersonID(ctx context.Context, userID, personID uuid.UUID) error // ใช้ตอน claim person

	// ─── Email Verification ──────────────────────────────────────────────────
	SetEmailVerificationToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error
	GetByEmailVerificationToken(ctx context.Context, token string) (*User, error)
	SetEmailVerified(ctx context.Context, userID uuid.UUID) error

	// ─── Password Reset ──────────────────────────────────────────────────────
	SetPasswordResetToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error
	GetByPasswordResetToken(ctx context.Context, token string) (*User, error)
	ClearPasswordResetToken(ctx context.Context, userID uuid.UUID) error

	// ─── UserAuth (multi-provider) ───────────────────────────────────────────
	CreateAuth(ctx context.Context, a *UserAuth) error
	GetAuthByUserAndProvider(ctx context.Context, userID uuid.UUID, provider string) (*UserAuth, error)
	GetByEmailWithLocalAuth(ctx context.Context, email string) (*User, *UserAuth, error)
	UpdateLocalAuthSecret(ctx context.Context, userID uuid.UUID, secret string) error
	UpdateAuthRefreshToken(ctx context.Context, authID uuid.UUID, token *string) error
	SetForcePasswordChange(ctx context.Context, userID uuid.UUID, required bool) error
	MarkPasswordChanged(ctx context.Context, userID uuid.UUID, changedAt time.Time) error

	// ─── Pagination ──────────────────────────────────────────────────────────
	// ListOffset returns a page of users for offset-based pagination.
	// Returns total count of all active users (for TotalPages calculation).
	ListOffset(ctx context.Context, limit, offset int) ([]*User, int64, error)

	// ListAfterCursor returns users whose ID > afterID, ordered by ID ASC.
	// Fetch limit+1 to detect hasMore — caller trims the extra item.
	// Pass uuid.Nil (zero value) to start from the beginning.
	ListAfterCursor(ctx context.Context, afterID uuid.UUID, limit int) ([]*User, error)
}
