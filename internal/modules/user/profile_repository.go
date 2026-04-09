package user

import (
	"context"

	"github.com/google/uuid"
)

// ProfileRepository handles all persistence operations for portfolio-related entities.
type ProfileRepository interface {
	// ── Profile (1:1 with users) ──────────────────────────────────────────
	GetProfileByUserID(ctx context.Context, userID uuid.UUID) (*Profile, error)
	UpsertProfile(ctx context.Context, p *Profile) error

	// ── Social Links (1:many via user_id) ────────────────────────────────
	ListSocialLinks(ctx context.Context, userID uuid.UUID) ([]*SocialLink, error)
	UpsertSocialLink(ctx context.Context, l *SocialLink) error
	DeleteSocialLink(ctx context.Context, linkID, userID uuid.UUID) error

	// ── Skills ────────────────────────────────────────────────────────────
	ListSkillsByUserID(ctx context.Context, userID uuid.UUID) ([]*Skill, error)
	CreateSkill(ctx context.Context, s *Skill) error
	DeleteSkill(ctx context.Context, skillID, userID uuid.UUID) error

	// ── Experiences ───────────────────────────────────────────────────────
	ListExperiencesByUserID(ctx context.Context, userID uuid.UUID) ([]*Experience, error)
	CreateExperience(ctx context.Context, e *Experience) error
	UpdateExperience(ctx context.Context, e *Experience) error
	DeleteExperience(ctx context.Context, expID, userID uuid.UUID) error

	// ── Portfolio Items ───────────────────────────────────────────────────
	// publishedOnly=true filters to status="published" (for public viewers).
	ListPortfolioByUserID(ctx context.Context, userID uuid.UUID, publishedOnly bool) ([]*PortfolioItem, error)
	GetPortfolioItemByID(ctx context.Context, itemID uuid.UUID) (*PortfolioItem, error)
	CreatePortfolioItem(ctx context.Context, item *PortfolioItem) error
	UpdatePortfolioItem(ctx context.Context, item *PortfolioItem) error
	DeletePortfolioItem(ctx context.Context, itemID, userID uuid.UUID) error

	// ── Tags (global pool) ────────────────────────────────────────────────
	GetOrCreateTags(ctx context.Context, names []string) ([]*PortfolioTag, error)
	// SetPortfolioItemTags replaces all tags on an item with the given tag IDs.
	SetPortfolioItemTags(ctx context.Context, itemID uuid.UUID, tagIDs []uuid.UUID) error
}
