package usersvc

import (
	"context"
	"strings"

	"vexentra-api/internal/modules/socialplatform"
	"vexentra-api/internal/modules/user"
	"vexentra-api/pkg/custom_errors"
	"vexentra-api/pkg/logger"

	"github.com/google/uuid"
)

// GetFullProfileResult bundles all profile data for GET /api/v1/persons/:id/profile.
type GetFullProfileResult struct {
	User        *user.User
	Profile     *user.Profile // nil if the person has not set up a profile yet
	Skills      []*user.Skill
	Experiences []*user.Experience
	Portfolio   []*user.PortfolioItem
}

type ProfileService interface {
	// GetFullProfile returns the complete profile for any person.
	// viewerIsOwner=true shows draft portfolio items; false shows published only.
	GetFullProfile(ctx context.Context, personID uuid.UUID, viewerIsOwner bool) (*GetFullProfileResult, error)

	UpsertProfile(ctx context.Context, personID uuid.UUID, p *user.Profile) error
	AdminUpsertProfile(ctx context.Context, userID uuid.UUID, p *user.Profile) error
	AdminAddSkill(ctx context.Context, userID uuid.UUID, s *user.Skill) error
	AdminAddExperience(ctx context.Context, userID uuid.UUID, e *user.Experience) error
	AdminUpdateExperience(ctx context.Context, userID, expID uuid.UUID, e *user.Experience) error
	AdminRemoveExperience(ctx context.Context, userID, expID uuid.UUID) error
	AdminAddPortfolioItem(ctx context.Context, userID uuid.UUID, item *user.PortfolioItem, tagNames []string) error

	AddSkill(ctx context.Context, personID uuid.UUID, s *user.Skill) error
	RemoveSkill(ctx context.Context, skillID, personID uuid.UUID) error

	AddExperience(ctx context.Context, personID uuid.UUID, e *user.Experience) error
	UpdateExperience(ctx context.Context, expID, personID uuid.UUID, e *user.Experience) error
	RemoveExperience(ctx context.Context, expID, personID uuid.UUID) error

	AddPortfolioItem(ctx context.Context, personID uuid.UUID, item *user.PortfolioItem, tagNames []string) error
	UpdatePortfolioItem(ctx context.Context, itemID, personID uuid.UUID, item *user.PortfolioItem, tagNames []string) error
	RemovePortfolioItem(ctx context.Context, itemID, personID uuid.UUID) error

	ListSocialLinks(ctx context.Context, personID uuid.UUID) ([]*user.SocialLink, error)
	UpsertSocialLink(ctx context.Context, personID, platformID uuid.UUID, url string, sortOrder int) (*user.SocialLink, error)
	DeleteSocialLink(ctx context.Context, linkID, personID uuid.UUID) error
}

type profileService struct {
	userRepo           user.UserRepository
	profileRepo        user.ProfileRepository
	socialPlatformRepo socialplatform.SocialPlatformRepository
	logger             logger.Logger
}

func NewProfileService(userRepo user.UserRepository, profileRepo user.ProfileRepository, socialPlatformRepo socialplatform.SocialPlatformRepository, l logger.Logger) ProfileService {
	if l == nil {
		l = logger.Get()
	}
	return &profileService{
		userRepo:           userRepo,
		profileRepo:        profileRepo,
		socialPlatformRepo: socialPlatformRepo,
		logger:             l,
	}
}

func (s *profileService) GetFullProfile(ctx context.Context, personID uuid.UUID, viewerIsOwner bool) (*GetFullProfileResult, error) {
	u, err := s.userRepo.GetByPersonID(ctx, personID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, custom_errors.New(404, custom_errors.ErrNotFound, "ไม่พบข้อมูลผู้ใช้งาน")
	}

	profile, err := s.profileRepo.GetProfileByPersonID(ctx, personID)
	if err != nil {
		return nil, err
	}

	skills, err := s.profileRepo.ListSkillsByPersonID(ctx, personID)
	if err != nil {
		return nil, err
	}

	exps, err := s.profileRepo.ListExperiencesByPersonID(ctx, personID)
	if err != nil {
		return nil, err
	}

	// Non-owners see only published portfolio items
	portfolio, err := s.profileRepo.ListPortfolioByPersonID(ctx, personID, !viewerIsOwner)
	if err != nil {
		return nil, err
	}

	if profile != nil {
		links, err := s.profileRepo.ListSocialLinks(ctx, personID)
		if err != nil {
			return nil, err
		}
		profile.SocialLinks = make([]user.SocialLink, len(links))
		for i, l := range links {
			profile.SocialLinks[i] = *l
		}
	}

	return &GetFullProfileResult{
		User:        u,
		Profile:     profile,
		Skills:      skills,
		Experiences: exps,
		Portfolio:   portfolio,
	}, nil
}

func (s *profileService) UpsertProfile(ctx context.Context, personID uuid.UUID, p *user.Profile) error {
	p.PersonID = personID
	return s.profileRepo.UpsertProfile(ctx, p)
}

func (s *profileService) AddSkill(ctx context.Context, personID uuid.UUID, skill *user.Skill) error {
	skill.PersonID = personID
	return s.profileRepo.CreateSkill(ctx, skill)
}

func (s *profileService) RemoveSkill(ctx context.Context, skillID, personID uuid.UUID) error {
	return s.profileRepo.DeleteSkill(ctx, skillID, personID)
}

func (s *profileService) AddExperience(ctx context.Context, personID uuid.UUID, e *user.Experience) error {
	e.PersonID = personID
	return s.profileRepo.CreateExperience(ctx, e)
}

func (s *profileService) UpdateExperience(ctx context.Context, expID, personID uuid.UUID, e *user.Experience) error {
	e.ID = expID
	e.PersonID = personID
	return s.profileRepo.UpdateExperience(ctx, e)
}

func (s *profileService) RemoveExperience(ctx context.Context, expID, personID uuid.UUID) error {
	return s.profileRepo.DeleteExperience(ctx, expID, personID)
}

func (s *profileService) AddPortfolioItem(ctx context.Context, personID uuid.UUID, item *user.PortfolioItem, tagNames []string) error {
	item.PersonID = personID
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

func (s *profileService) UpdatePortfolioItem(ctx context.Context, itemID, personID uuid.UUID, item *user.PortfolioItem, tagNames []string) error {
	// Verify ownership
	existing, err := s.profileRepo.GetPortfolioItemByID(ctx, itemID)
	if err != nil {
		return err
	}
	if existing.PersonID != personID {
		return custom_errors.New(403, custom_errors.ErrForbidden, "ไม่มีสิทธิ์แก้ไข portfolio item นี้")
	}

	item.ID = itemID
	item.PersonID = personID
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

func (s *profileService) RemovePortfolioItem(ctx context.Context, itemID, personID uuid.UUID) error {
	return s.profileRepo.DeletePortfolioItem(ctx, itemID, personID)
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

func (s *profileService) ListSocialLinks(ctx context.Context, personID uuid.UUID) ([]*user.SocialLink, error) {
	return s.profileRepo.ListSocialLinks(ctx, personID)
}

func (s *profileService) UpsertSocialLink(ctx context.Context, personID, platformID uuid.UUID, url string, sortOrder int) (*user.SocialLink, error) {
	p, err := s.socialPlatformRepo.GetByID(ctx, platformID)
	if err != nil {
		return nil, err
	}
	if p == nil || !p.IsActive {
		return nil, custom_errors.New(400, custom_errors.ErrInvalidFormat, "platform ไม่รองรับหรือไม่ได้เปิดใช้งาน")
	}
	l := &user.SocialLink{
		PersonID:   personID,
		PlatformID: platformID,
		URL:        url,
		SortOrder:  sortOrder,
	}
	if err := s.profileRepo.UpsertSocialLink(ctx, l); err != nil {
		return nil, err
	}
	return l, nil
}

func (s *profileService) DeleteSocialLink(ctx context.Context, linkID, personID uuid.UUID) error {
	return s.profileRepo.DeleteSocialLink(ctx, linkID, personID)
}

func (s *profileService) AdminUpsertProfile(ctx context.Context, userID uuid.UUID, p *user.Profile) error {
	personID, err := s.personIDByUserID(ctx, userID)
	if err != nil {
		return err
	}
	p.PersonID = personID
	return s.profileRepo.UpsertProfile(ctx, p)
}

func (s *profileService) AdminAddSkill(ctx context.Context, userID uuid.UUID, skill *user.Skill) error {
	personID, err := s.personIDByUserID(ctx, userID)
	if err != nil {
		return err
	}
	skill.PersonID = personID
	return s.profileRepo.CreateSkill(ctx, skill)
}

func (s *profileService) AdminAddExperience(ctx context.Context, userID uuid.UUID, e *user.Experience) error {
	personID, err := s.personIDByUserID(ctx, userID)
	if err != nil {
		return err
	}
	e.PersonID = personID
	return s.profileRepo.CreateExperience(ctx, e)
}

func (s *profileService) AdminUpdateExperience(ctx context.Context, userID, expID uuid.UUID, e *user.Experience) error {
	personID, err := s.personIDByUserID(ctx, userID)
	if err != nil {
		return err
	}
	e.ID = expID
	e.PersonID = personID
	return s.profileRepo.UpdateExperience(ctx, e)
}

func (s *profileService) AdminRemoveExperience(ctx context.Context, userID, expID uuid.UUID) error {
	personID, err := s.personIDByUserID(ctx, userID)
	if err != nil {
		return err
	}
	return s.profileRepo.DeleteExperience(ctx, expID, personID)
}

func (s *profileService) AdminAddPortfolioItem(ctx context.Context, userID uuid.UUID, item *user.PortfolioItem, tagNames []string) error {
	personID, err := s.personIDByUserID(ctx, userID)
	if err != nil {
		return err
	}
	item.PersonID = personID
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

func (s *profileService) personIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil || u == nil {
		return uuid.Nil, custom_errors.NewNotFoundError("USER_NOT_FOUND", "ไม่พบผู้ใช้งาน")
	}
	return u.PersonID, nil
}
