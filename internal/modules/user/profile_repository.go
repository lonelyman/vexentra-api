package user

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProfileRepository handles all persistence operations for portfolio-related entities.
type ProfileRepository interface {
	WithTx(tx *gorm.DB) ProfileRepository

	// ── Profile (1:1 with persons) ────────────────────────────────────────
	GetProfileByPersonID(ctx context.Context, personID uuid.UUID) (*Profile, error)
	UpsertProfile(ctx context.Context, p *Profile) error
	SetProfileAvatarFileID(ctx context.Context, personID, fileID uuid.UUID) error

	// ── Social Links (1:many via person_id) ──────────────────────────────
	ListSocialLinks(ctx context.Context, personID uuid.UUID) ([]*SocialLink, error)
	UpsertSocialLink(ctx context.Context, l *SocialLink) error
	DeleteSocialLink(ctx context.Context, linkID, personID uuid.UUID) error

	// ── Skills ────────────────────────────────────────────────────────────
	ListSkillsByPersonID(ctx context.Context, personID uuid.UUID) ([]*Skill, error)
	CreateSkill(ctx context.Context, s *Skill) error
	UpdateSkill(ctx context.Context, s *Skill) error
	DeleteSkill(ctx context.Context, skillID, personID uuid.UUID) error

	// ── Experiences ───────────────────────────────────────────────────────
	ListExperiencesByPersonID(ctx context.Context, personID uuid.UUID) ([]*Experience, error)
	CreateExperience(ctx context.Context, e *Experience) error
	UpdateExperience(ctx context.Context, e *Experience) error
	DeleteExperience(ctx context.Context, expID, personID uuid.UUID) error

	// ── Portfolio Items ───────────────────────────────────────────────────
	// publishedOnly=true filters to status="published" (for public viewers).
	ListPortfolioByPersonID(ctx context.Context, personID uuid.UUID, publishedOnly bool) ([]*PortfolioItem, error)
	GetPortfolioItemByID(ctx context.Context, itemID uuid.UUID) (*PortfolioItem, error)
	CreatePortfolioItem(ctx context.Context, item *PortfolioItem) error
	UpdatePortfolioItem(ctx context.Context, item *PortfolioItem) error
	DeletePortfolioItem(ctx context.Context, itemID, personID uuid.UUID) error

	// ── Tags (global pool) ────────────────────────────────────────────────
	GetOrCreateTags(ctx context.Context, names []string) ([]*PortfolioTag, error)
	// SetPortfolioItemTags replaces all tags on an item with the given tag IDs.
	SetPortfolioItemTags(ctx context.Context, itemID uuid.UUID, tagIDs []uuid.UUID) error
}
