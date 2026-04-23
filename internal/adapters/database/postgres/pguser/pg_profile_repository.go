package pguser

import (
	"context"
	"errors"
	"strings"
	"vexentra-api/internal/modules/user"
	"vexentra-api/pkg/custom_errors"
	"vexentra-api/pkg/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type profileRepository struct {
	db     *gorm.DB
	logger logger.Logger
}

func NewProfileRepository(db *gorm.DB, l logger.Logger) user.ProfileRepository {
	if l == nil {
		l = logger.Get()
	}
	return &profileRepository{db: db, logger: l}
}

// ─────────────────────────────────────────────────────────────────────
//  Profile
// ─────────────────────────────────────────────────────────────────────

func (r *profileRepository) GetProfileByPersonID(ctx context.Context, personID uuid.UUID) (*user.Profile, error) {
	var m profileModel
	if err := r.db.WithContext(ctx).Where("person_id = ?", personID).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // no profile yet — not an error
		}
		r.logger.Error("DB_GET_PROFILE_ERROR", err)
		return nil, err
	}
	return m.ToEntity(), nil
}

func (r *profileRepository) UpsertProfile(ctx context.Context, p *user.Profile) error {
	var existing profileModel
	err := r.db.WithContext(ctx).Where("person_id = ?", p.PersonID).First(&existing).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// First time — create
		id, genErr := uuid.NewV7()
		if genErr != nil {
			return custom_errors.NewInternalError("ไม่สามารถสร้าง ID ได้")
		}
		p.ID = id
		m := &profileModel{
			ID:          p.ID,
			PersonID:    p.PersonID,
			DisplayName: p.DisplayName,
			Headline:    p.Headline,
			Bio:         p.Bio,
			Location:    p.Location,
			AvatarURL:   p.AvatarURL,
		}
		if createErr := r.db.WithContext(ctx).Create(m).Error; createErr != nil {
			r.logger.Error("DB_CREATE_PROFILE_ERROR", createErr)
			return createErr
		}
		p.CreatedAt = m.CreatedAt
		p.UpdatedAt = m.UpdatedAt
		return nil
	}

	if err != nil {
		r.logger.Error("DB_UPSERT_PROFILE_CHECK_ERROR", err)
		return err
	}

	// Update — use map to overwrite even empty-string fields
	p.ID = existing.ID
	result := r.db.WithContext(ctx).Model(&existing).Updates(map[string]any{
		"display_name": p.DisplayName,
		"headline":     p.Headline,
		"bio":          p.Bio,
		"location":     p.Location,
		"avatar_url":   p.AvatarURL,
	})
	if result.Error != nil {
		r.logger.Error("DB_UPDATE_PROFILE_ERROR", result.Error)
		return result.Error
	}
	p.CreatedAt = existing.CreatedAt
	p.UpdatedAt = existing.UpdatedAt
	return nil
}

// ─────────────────────────────────────────────────────────────────────
//  Social Links
// ─────────────────────────────────────────────────────────────────────

func (r *profileRepository) ListSocialLinks(ctx context.Context, personID uuid.UUID) ([]*user.SocialLink, error) {
	var models []socialLinkModel
	if err := r.db.WithContext(ctx).
		Where("person_id = ?", personID).
		Order("sort_order ASC, created_at ASC").
		Find(&models).Error; err != nil {
		r.logger.Error("DB_LIST_SOCIAL_LINKS_ERROR", err)
		return nil, err
	}
	links := make([]*user.SocialLink, len(models))
	for i := range models {
		links[i] = models[i].ToEntity()
	}
	return links, nil
}

func (r *profileRepository) UpsertSocialLink(ctx context.Context, l *user.SocialLink) error {
	// one platform per person — upsert by (person_id, platform_id)
	var existing socialLinkModel
	err := r.db.WithContext(ctx).
		Where("person_id = ? AND platform_id = ?", l.PersonID, l.PlatformID).
		First(&existing).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		id, genErr := uuid.NewV7()
		if genErr != nil {
			return custom_errors.NewInternalError("ไม่สามารถสร้าง ID ได้")
		}
		l.ID = id
		m := &socialLinkModel{
			ID:         l.ID,
			PersonID:   l.PersonID,
			PlatformID: l.PlatformID,
			URL:        l.URL,
			SortOrder:  l.SortOrder,
		}
		if createErr := r.db.WithContext(ctx).Create(m).Error; createErr != nil {
			r.logger.Error("DB_CREATE_SOCIAL_LINK_ERROR", createErr)
			return createErr
		}
		return nil
	}
	if err != nil {
		r.logger.Error("DB_UPSERT_SOCIAL_LINK_CHECK_ERROR", err)
		return err
	}

	// Update existing
	l.ID = existing.ID
	result := r.db.WithContext(ctx).Model(&existing).Updates(map[string]any{
		"url":        l.URL,
		"sort_order": l.SortOrder,
	})
	if result.Error != nil {
		r.logger.Error("DB_UPDATE_SOCIAL_LINK_ERROR", result.Error)
		return result.Error
	}
	return nil
}

func (r *profileRepository) DeleteSocialLink(ctx context.Context, linkID, personID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND person_id = ?", linkID, personID).
		Delete(&socialLinkModel{})
	if result.Error != nil {
		r.logger.Error("DB_DELETE_SOCIAL_LINK_ERROR", result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		return custom_errors.New(404, custom_errors.ErrNotFound, "ไม่พบ social link นี้")
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────
//  Skills
// ─────────────────────────────────────────────────────────────────────

func (r *profileRepository) ListSkillsByPersonID(ctx context.Context, personID uuid.UUID) ([]*user.Skill, error) {
	var models []skillModel
	if err := r.db.WithContext(ctx).
		Where("person_id = ?", personID).
		Order("sort_order ASC, created_at ASC").
		Find(&models).Error; err != nil {
		r.logger.Error("DB_LIST_SKILLS_ERROR", err)
		return nil, err
	}
	skills := make([]*user.Skill, len(models))
	for i := range models {
		skills[i] = models[i].ToEntity()
	}
	return skills, nil
}

func (r *profileRepository) CreateSkill(ctx context.Context, s *user.Skill) error {
	id, err := uuid.NewV7()
	if err != nil {
		return custom_errors.NewInternalError("ไม่สามารถสร้าง ID ได้")
	}
	s.ID = id
	m := &skillModel{
		ID:          s.ID,
		PersonID:    s.PersonID,
		Name:        s.Name,
		Category:    s.Category,
		Proficiency: s.Proficiency,
		SortOrder:   s.SortOrder,
	}
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		r.logger.Error("DB_CREATE_SKILL_ERROR", err)
		return err
	}
	s.CreatedAt = m.CreatedAt
	s.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *profileRepository) UpdateSkill(ctx context.Context, s *user.Skill) error {
	updates := map[string]any{
		"name":        s.Name,
		"category":    s.Category,
		"proficiency": s.Proficiency,
		"sort_order":  s.SortOrder,
	}
	result := r.db.WithContext(ctx).
		Model(&skillModel{}).
		Where("id = ? AND person_id = ?", s.ID, s.PersonID).
		Updates(updates)
	if result.Error != nil {
		r.logger.Error("DB_UPDATE_SKILL_ERROR", result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		return custom_errors.NewNotFoundError(custom_errors.ErrNotFound, "ไม่พบ skill")
	}
	return nil
}

func (r *profileRepository) DeleteSkill(ctx context.Context, skillID, personID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND person_id = ?", skillID, personID).
		Delete(&skillModel{})
	if result.Error != nil {
		r.logger.Error("DB_DELETE_SKILL_ERROR", result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		return custom_errors.NewNotFoundError(custom_errors.ErrNotFound, "ไม่พบ skill")
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────
//  Experiences
// ─────────────────────────────────────────────────────────────────────

func (r *profileRepository) ListExperiencesByPersonID(ctx context.Context, personID uuid.UUID) ([]*user.Experience, error) {
	var models []experienceModel
	if err := r.db.WithContext(ctx).
		Where("person_id = ?", personID).
		Order("sort_order ASC, started_at DESC").
		Find(&models).Error; err != nil {
		r.logger.Error("DB_LIST_EXPERIENCES_ERROR", err)
		return nil, err
	}
	exps := make([]*user.Experience, len(models))
	for i := range models {
		exps[i] = models[i].ToEntity()
	}
	return exps, nil
}

func (r *profileRepository) CreateExperience(ctx context.Context, e *user.Experience) error {
	id, err := uuid.NewV7()
	if err != nil {
		return custom_errors.NewInternalError("ไม่สามารถสร้าง ID ได้")
	}
	e.ID = id
	m := &experienceModel{
		ID:          e.ID,
		PersonID:    e.PersonID,
		Company:     e.Company,
		Position:    e.Position,
		Location:    e.Location,
		Description: e.Description,
		StartedAt:   e.StartedAt,
		EndedAt:     e.EndedAt,
		IsCurrent:   e.IsCurrent,
		SortOrder:   e.SortOrder,
	}
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		r.logger.Error("DB_CREATE_EXPERIENCE_ERROR", err)
		return err
	}
	e.CreatedAt = m.CreatedAt
	e.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *profileRepository) UpdateExperience(ctx context.Context, e *user.Experience) error {
	// Use map to correctly persist bool false and nil pointer fields
	updates := map[string]any{
		"company":     e.Company,
		"position":    e.Position,
		"location":    e.Location,
		"description": e.Description,
		"started_at":  e.StartedAt,
		"ended_at":    e.EndedAt, // nil clears the column
		"is_current":  e.IsCurrent,
		"sort_order":  e.SortOrder,
	}
	result := r.db.WithContext(ctx).
		Model(&experienceModel{}).
		Where("id = ? AND person_id = ?", e.ID, e.PersonID).
		Updates(updates)
	if result.Error != nil {
		r.logger.Error("DB_UPDATE_EXPERIENCE_ERROR", result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		return custom_errors.NewNotFoundError(custom_errors.ErrNotFound, "ไม่พบ experience")
	}
	return nil
}

func (r *profileRepository) DeleteExperience(ctx context.Context, expID, personID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND person_id = ?", expID, personID).
		Delete(&experienceModel{})
	if result.Error != nil {
		r.logger.Error("DB_DELETE_EXPERIENCE_ERROR", result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		return custom_errors.NewNotFoundError(custom_errors.ErrNotFound, "ไม่พบ experience")
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────
//  Portfolio Items
// ─────────────────────────────────────────────────────────────────────

func (r *profileRepository) ListPortfolioByPersonID(ctx context.Context, personID uuid.UUID, publishedOnly bool) ([]*user.PortfolioItem, error) {
	var models []portfolioItemModel
	q := r.db.WithContext(ctx).
		Preload("Tags").
		Where("person_id = ?", personID).
		Order("sort_order ASC, created_at DESC")
	if publishedOnly {
		q = q.Where("status = ?", user.PortfolioStatusPublished)
	}
	if err := q.Find(&models).Error; err != nil {
		r.logger.Error("DB_LIST_PORTFOLIO_ERROR", err)
		return nil, err
	}
	items := make([]*user.PortfolioItem, len(models))
	for i := range models {
		items[i] = models[i].ToEntity()
	}
	return items, nil
}

func (r *profileRepository) GetPortfolioItemByID(ctx context.Context, itemID uuid.UUID) (*user.PortfolioItem, error) {
	var m portfolioItemModel
	if err := r.db.WithContext(ctx).Preload("Tags").First(&m, "id = ?", itemID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, custom_errors.NewNotFoundError(custom_errors.ErrNotFound, "ไม่พบ portfolio item")
		}
		r.logger.Error("DB_GET_PORTFOLIO_ITEM_ERROR", err)
		return nil, err
	}
	return m.ToEntity(), nil
}

func (r *profileRepository) CreatePortfolioItem(ctx context.Context, item *user.PortfolioItem) error {
	id, err := uuid.NewV7()
	if err != nil {
		return custom_errors.NewInternalError("ไม่สามารถสร้าง ID ได้")
	}
	item.ID = id
	m := &portfolioItemModel{
		ID:              item.ID,
		PersonID:        item.PersonID,
		Title:           item.Title,
		Slug:            item.Slug,
		Summary:         item.Summary,
		Description:     item.Description,
		ContentMarkdown: item.ContentMarkdown,
		CoverImageURL:   item.CoverImageURL,
		DemoURL:         item.DemoURL,
		SourceURL:       item.SourceURL,
		Status:          item.Status,
		Featured:        item.Featured,
		SortOrder:       item.SortOrder,
		StartedAt:       item.StartedAt,
		EndedAt:         item.EndedAt,
	}
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		r.logger.Error("DB_CREATE_PORTFOLIO_ITEM_ERROR", err)
		return err
	}
	item.CreatedAt = m.CreatedAt
	item.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *profileRepository) UpdatePortfolioItem(ctx context.Context, item *user.PortfolioItem) error {
	result := r.db.WithContext(ctx).
		Model(&portfolioItemModel{}).
		Where("id = ? AND person_id = ?", item.ID, item.PersonID).
		Updates(map[string]any{
			"title":            item.Title,
			"slug":             item.Slug,
			"summary":          item.Summary,
			"description":      item.Description,
			"content_markdown": item.ContentMarkdown,
			"cover_image_url":  item.CoverImageURL,
			"demo_url":         item.DemoURL,
			"source_url":       item.SourceURL,
			"status":           item.Status,
			"featured":         item.Featured,
			"sort_order":       item.SortOrder,
			"started_at":       item.StartedAt,
			"ended_at":         item.EndedAt,
		})
	if result.Error != nil {
		r.logger.Error("DB_UPDATE_PORTFOLIO_ITEM_ERROR", result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		return custom_errors.NewNotFoundError(custom_errors.ErrNotFound, "ไม่พบ portfolio item")
	}
	return nil
}

func (r *profileRepository) DeletePortfolioItem(ctx context.Context, itemID, personID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND person_id = ?", itemID, personID).
		Delete(&portfolioItemModel{})
	if result.Error != nil {
		r.logger.Error("DB_DELETE_PORTFOLIO_ITEM_ERROR", result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		return custom_errors.NewNotFoundError(custom_errors.ErrNotFound, "ไม่พบ portfolio item")
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────
//  Tags
// ─────────────────────────────────────────────────────────────────────

func (r *profileRepository) GetOrCreateTags(ctx context.Context, names []string) ([]*user.PortfolioTag, error) {
	tags := make([]*user.PortfolioTag, 0, len(names))
	for _, rawName := range names {
		name := strings.TrimSpace(rawName)
		if name == "" {
			continue
		}
		slug := strings.ToLower(strings.NewReplacer(" ", "-", "_", "-").Replace(name))

		var m portfolioTagModel
		err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&m).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			id, genErr := uuid.NewV7()
			if genErr != nil {
				return nil, custom_errors.NewInternalError("ไม่สามารถสร้าง ID ได้")
			}
			m = portfolioTagModel{ID: id, Name: name, Slug: slug}
			if createErr := r.db.WithContext(ctx).Create(&m).Error; createErr != nil {
				r.logger.Error("DB_CREATE_TAG_ERROR", createErr)
				return nil, createErr
			}
		} else if err != nil {
			r.logger.Error("DB_GET_TAG_ERROR", err)
			return nil, err
		}
		tags = append(tags, m.ToEntity())
	}
	return tags, nil
}

func (r *profileRepository) SetPortfolioItemTags(ctx context.Context, itemID uuid.UUID, tagIDs []uuid.UUID) error {
	tagModels := make([]portfolioTagModel, len(tagIDs))
	for i, id := range tagIDs {
		tagModels[i] = portfolioTagModel{ID: id}
	}
	return r.db.WithContext(ctx).
		Model(&portfolioItemModel{ID: itemID}).
		Association("Tags").
		Replace(tagModels)
}
