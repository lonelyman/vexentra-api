package userhdl

import (
	"time"
	"vexentra-api/internal/modules/user"
)

// UpsertProfileRequest is the body for PUT /api/v1/me/profile.
// All fields are optional — omitted fields clear their stored value.
type UpsertProfileRequest struct {
	DisplayName string `json:"display_name"`
	Headline    string `json:"headline"`
	Bio         string `json:"bio"`
	Location    string `json:"location"`
}

// UpsertSocialLinkRequest is the body for PUT /api/v1/me/social-links/:platform.
type UpsertSocialLinkRequest struct {
	URL       string `json:"url"       validate:"required,url"`
	SortOrder int    `json:"sort_order"`
}

// AddSkillRequest is the body for POST /api/v1/me/skills.
type AddSkillRequest struct {
	Name        string `json:"name"        validate:"required"`
	Category    string `json:"category"    validate:"required,oneof=backend frontend devops other"`
	Proficiency int    `json:"proficiency" validate:"required,min=1,max=5"`
	SortOrder   int    `json:"sort_order"`
}

// AddExperienceRequest is the body for POST /api/v1/me/experiences
// and PUT /api/v1/me/experiences/:expID.
type AddExperienceRequest struct {
	Company     string     `json:"company"     validate:"required"`
	Position    string     `json:"position"    validate:"required"`
	Location    string     `json:"location"`
	Description string     `json:"description"`
	StartedAt   time.Time  `json:"started_at"  validate:"required"`
	EndedAt     *time.Time `json:"ended_at"`
	IsCurrent   bool       `json:"is_current"`
	SortOrder   int        `json:"sort_order"`
}

func (r *AddExperienceRequest) ToEntity() *user.Experience {
	return &user.Experience{
		Company:     r.Company,
		Position:    r.Position,
		Location:    r.Location,
		Description: r.Description,
		StartedAt:   r.StartedAt,
		EndedAt:     r.EndedAt,
		IsCurrent:   r.IsCurrent,
		SortOrder:   r.SortOrder,
	}
}

// AddPortfolioItemRequest is the body for POST /api/v1/me/portfolio
// and PUT /api/v1/me/portfolio/:itemID.
type AddPortfolioItemRequest struct {
	Title           string     `json:"title"           validate:"required"`
	Slug            string     `json:"slug"`
	Summary         string     `json:"summary"`
	Description     string     `json:"description"`
	ContentMarkdown string     `json:"content_markdown"`
	CoverImageURL   string     `json:"cover_image_url" validate:"omitempty,url"`
	DemoURL         string     `json:"demo_url"        validate:"omitempty,url"`
	SourceURL       string     `json:"source_url"      validate:"omitempty,url"`
	Status          string     `json:"status"          validate:"omitempty,oneof=draft published"`
	Featured        bool       `json:"featured"`
	SortOrder       int        `json:"sort_order"`
	StartedAt       *time.Time `json:"started_at"`
	EndedAt         *time.Time `json:"ended_at"`
	Tags            []string   `json:"tags"`
}

func (r *AddPortfolioItemRequest) ToEntity() *user.PortfolioItem {
	return &user.PortfolioItem{
		Title:           r.Title,
		Slug:            r.Slug,
		Summary:         r.Summary,
		Description:     r.Description,
		ContentMarkdown: r.ContentMarkdown,
		CoverImageURL:   r.CoverImageURL,
		DemoURL:         r.DemoURL,
		SourceURL:       r.SourceURL,
		Status:          r.Status,
		Featured:        r.Featured,
		SortOrder:       r.SortOrder,
		StartedAt:       r.StartedAt,
		EndedAt:         r.EndedAt,
	}
}
