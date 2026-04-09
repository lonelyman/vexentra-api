package usersvc

import (
	"context"
	"strings"

	"vexentra-api/internal/modules/user"
	"vexentra-api/pkg/custom_errors"
	"vexentra-api/pkg/logger"

	"github.com/google/uuid"
)

// GetFullProfileResult bundles all profile data for GET /api/v1/users/:id/profile.
type GetFullProfileResult struct {
	User        *user.User
	Profile     *user.Profile // nil if the user has not set up a profile yet
	Skills      []*user.Skill
	Experiences []*user.Experience
	Portfolio   []*user.PortfolioItem
}

type ProfileService interface {
	// GetFullProfile returns the complete profile for any user.
	// viewerIsOwner=true shows draft portfolio items; false shows published only.
	GetFullProfile(ctx context.Context, userID uuid.UUID, viewerIsOwner bool) (*GetFullProfileResult, error)

	UpsertProfile(ctx context.Context, userID uuid.UUID, p *user.Profile) error

	AddSkill(ctx context.Context, userID uuid.UUID, s *user.Skill) error
	RemoveSkill(ctx context.Context, skillID, userID uuid.UUID) error

	AddExperience(ctx context.Context, userID uuid.UUID, e *user.Experience) error
	UpdateExperience(ctx context.Context, expID, userID uuid.UUID, e *user.Experience) error
	RemoveExperience(ctx context.Context, expID, userID uuid.UUID) error

	AddPortfolioItem(ctx context.Context, userID uuid.UUID, item *user.PortfolioItem, tagNames []string) error
	UpdatePortfolioItem(ctx context.Context, itemID, userID uuid.UUID, item *user.PortfolioItem, tagNames []string) error
	RemovePortfolioItem(ctx context.Context, itemID, userID uuid.UUID) error
}

type profileService struct {
	userRepo    user.UserRepository
	profileRepo user.ProfileRepository
	logger      logger.Logger
}

func NewProfileService(userRepo user.UserRepository, profileRepo user.ProfileRepository, l logger.Logger) ProfileService {
	if l == nil {
		l = logger.Get()
	}
	return &profileService{
		userRepo:    userRepo,
		profileRepo: profileRepo,
		logger:      l,
	}
}

func (s *profileService) GetFullProfile(ctx context.Context, userID uuid.UUID, viewerIsOwner bool) (*GetFullProfileResult, error) {
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	profile, err := s.profileRepo.GetProfileByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	skills, err := s.profileRepo.ListSkillsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	exps, err := s.profileRepo.ListExperiencesByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Non-owners see only published portfolio items
	portfolio, err := s.profileRepo.ListPortfolioByUserID(ctx, userID, !viewerIsOwner)
	if err != nil {
		return nil, err
	}

	return &GetFullProfileResult{
		User:        u,
		Profile:     profile,
		Skills:      skills,
		Experiences: exps,
		Portfolio:   portfolio,
	}, nil
}

func (s *profileService) UpsertProfile(ctx context.Context, userID uuid.UUID, p *user.Profile) error {
	p.UserID = userID
	return s.profileRepo.UpsertProfile(ctx, p)
}

func (s *profileService) AddSkill(ctx context.Context, userID uuid.UUID, skill *user.Skill) error {
	skill.UserID = userID
	return s.profileRepo.CreateSkill(ctx, skill)
}

func (s *profileService) RemoveSkill(ctx context.Context, skillID, userID uuid.UUID) error {
	return s.profileRepo.DeleteSkill(ctx, skillID, userID)
}

func (s *profileService) AddExperience(ctx context.Context, userID uuid.UUID, e *user.Experience) error {
	e.UserID = userID
	return s.profileRepo.CreateExperience(ctx, e)
}

func (s *profileService) UpdateExperience(ctx context.Context, expID, userID uuid.UUID, e *user.Experience) error {
	e.ID = expID
	e.UserID = userID
	return s.profileRepo.UpdateExperience(ctx, e)
}

func (s *profileService) RemoveExperience(ctx context.Context, expID, userID uuid.UUID) error {
	return s.profileRepo.DeleteExperience(ctx, expID, userID)
}

func (s *profileService) AddPortfolioItem(ctx context.Context, userID uuid.UUID, item *user.PortfolioItem, tagNames []string) error {
	item.UserID = userID
	if item.Slug == "" {
		item.Slug = slugify(item.Title)
	}
	if item.Status == "" {
		item.Status = user.PortfolioStatusDraft
	}

	if err := s.profileRepo.CreatePortfolioItem(ctx, item); err != nil {
		return err
	}

	return s.syncTags(ctx, item, tagNames)
}

func (s *profileService) UpdatePortfolioItem(ctx context.Context, itemID, userID uuid.UUID, item *user.PortfolioItem, tagNames []string) error {
	// Verify ownership
	existing, err := s.profileRepo.GetPortfolioItemByID(ctx, itemID)
	if err != nil {
		return err
	}
	if existing.UserID != userID {
		return custom_errors.New(403, custom_errors.ErrForbidden, "ไม่มีสิทธิ์แก้ไข portfolio item นี้")
	}

	item.ID = itemID
	item.UserID = userID
	if item.Slug == "" {
		item.Slug = slugify(item.Title)
	}

	if err := s.profileRepo.UpdatePortfolioItem(ctx, item); err != nil {
		return err
	}

	if tagNames != nil {
		return s.syncTags(ctx, item, tagNames)
	}
	return nil
}

func (s *profileService) RemovePortfolioItem(ctx context.Context, itemID, userID uuid.UUID) error {
	return s.profileRepo.DeletePortfolioItem(ctx, itemID, userID)
}

// syncTags resolves tag names to IDs (creating new tags as needed) and
// replaces all tag associations on the given portfolio item.
func (s *profileService) syncTags(ctx context.Context, item *user.PortfolioItem, names []string) error {
	if len(names) == 0 {
		item.Tags = nil
		return s.profileRepo.SetPortfolioItemTags(ctx, item.ID, nil)
	}

	tags, err := s.profileRepo.GetOrCreateTags(ctx, names)
	if err != nil {
		return err
	}

	tagIDs := make([]uuid.UUID, len(tags))
	for i, t := range tags {
		tagIDs[i] = t.ID
	}

	if err := s.profileRepo.SetPortfolioItemTags(ctx, item.ID, tagIDs); err != nil {
		return err
	}

	// Reflect resolved tags back onto the entity for the response
	item.Tags = make([]user.PortfolioTag, len(tags))
	for i, t := range tags {
		item.Tags[i] = *t
	}
	return nil
}

func slugify(s string) string {
	return strings.ToLower(strings.NewReplacer(" ", "-", "_", "-").Replace(s))
}
