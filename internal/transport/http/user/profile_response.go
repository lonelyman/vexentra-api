package userhdl

import (
	"time"
	"vexentra-api/internal/modules/user"
	"vexentra-api/internal/modules/user/usersvc"
)

// ─────────────────────────────────────────────────────────────────────
//  Response types
// ─────────────────────────────────────────────────────────────────────

// ProfileResponse is the profile section of a user's public page.
type ProfileResponse struct {
	DisplayName  string               `json:"display_name"`
	Headline     string               `json:"headline"`
	Bio          string               `json:"bio"`
	Location     string               `json:"location"`
	AvatarFileID *string              `json:"avatar_file_id,omitempty"`
	AvatarURL    string               `json:"avatar_url"`
	SocialLinks  []SocialLinkResponse `json:"social_links"`
}

// SocialLinkResponse represents a single social link.
type SocialLinkResponse struct {
	ID         string `json:"id"`
	PlatformID string `json:"platform_id"`
	URL        string `json:"url"`
	SortOrder  int    `json:"sort_order"`
}

// SkillResponse represents a single skill entry.
type SkillResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Proficiency int    `json:"proficiency"` // 1–5
	SortOrder   int    `json:"sort_order"`
}

// ExperienceResponse represents a single work history entry.
type ExperienceResponse struct {
	ID          string     `json:"id"`
	Company     string     `json:"company"`
	Position    string     `json:"position"`
	Location    string     `json:"location"`
	Description string     `json:"description"`
	StartedAt   time.Time  `json:"started_at"`
	EndedAt     *time.Time `json:"ended_at"`
	IsCurrent   bool       `json:"is_current"`
	SortOrder   int        `json:"sort_order"`
}

// PortfolioTagResponse is the serialized form of a tag.
type PortfolioTagResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// PortfolioItemResponse represents a single portfolio/project entry.
type PortfolioItemResponse struct {
	ID              string                 `json:"id"`
	Title           string                 `json:"title"`
	Slug            string                 `json:"slug"`
	Summary         string                 `json:"summary"`
	Description     string                 `json:"description"`
	ContentMarkdown string                 `json:"content_markdown"`
	CoverImageURL   string                 `json:"cover_image_url"`
	DemoURL         string                 `json:"demo_url"`
	SourceURL       string                 `json:"source_url"`
	Status          string                 `json:"status"`
	Featured        bool                   `json:"featured"`
	StartedAt       *time.Time             `json:"started_at"`
	EndedAt         *time.Time             `json:"ended_at"`
	Tags            []PortfolioTagResponse `json:"tags"`
	CreatedAt       time.Time              `json:"created_at"`
}

// PublicUserSummary is the minimal user info included in the full profile view.
type PublicUserSummary struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

// FullProfileResponse is the response envelope for GET /api/v1/users/:id/profile.
type FullProfileResponse struct {
	User        PublicUserSummary       `json:"user"`
	Profile     *ProfileResponse        `json:"profile"` // nil if not set up yet
	Skills      []SkillResponse         `json:"skills"`
	Experiences []ExperienceResponse    `json:"experiences"`
	Portfolio   []PortfolioItemResponse `json:"portfolio"`
}

// ─────────────────────────────────────────────────────────────────────
//  Mappers — domain entity → response DTO
// ─────────────────────────────────────────────────────────────────────

func toProfileResponse(p *user.Profile) *ProfileResponse {
	if p == nil {
		return nil
	}
	links := make([]SocialLinkResponse, len(p.SocialLinks))
	for i, l := range p.SocialLinks {
		links[i] = SocialLinkResponse{
			ID:         l.ID.String(),
			PlatformID: l.PlatformID.String(),
			URL:        l.URL,
			SortOrder:  l.SortOrder,
		}
	}
	var avatarFileID *string
	if p.AvatarFileID != nil {
		id := p.AvatarFileID.String()
		avatarFileID = &id
	}
	return &ProfileResponse{
		DisplayName:  p.DisplayName,
		Headline:     p.Headline,
		Bio:          p.Bio,
		Location:     p.Location,
		AvatarFileID: avatarFileID,
		AvatarURL:    p.AvatarURL,
		SocialLinks:  links,
	}
}

func toSkillResponse(s *user.Skill) SkillResponse {
	return SkillResponse{
		ID:          s.ID.String(),
		Name:        s.Name,
		Category:    s.Category,
		Proficiency: s.Proficiency,
		SortOrder:   s.SortOrder,
	}
}

func toExperienceResponse(e *user.Experience) ExperienceResponse {
	return ExperienceResponse{
		ID:          e.ID.String(),
		Company:     e.Company,
		Position:    e.Position,
		Location:    e.Location,
		Description: e.Description,
		StartedAt:   e.StartedAt,
		EndedAt:     e.EndedAt,
		IsCurrent:   e.IsCurrent,
		SortOrder:   e.SortOrder,
	}
}

func toPortfolioTagResponse(t user.PortfolioTag) PortfolioTagResponse {
	return PortfolioTagResponse{
		ID:   t.ID.String(),
		Name: t.Name,
		Slug: t.Slug,
	}
}

func toPortfolioItemResponse(item *user.PortfolioItem) PortfolioItemResponse {
	tags := make([]PortfolioTagResponse, len(item.Tags))
	for i, t := range item.Tags {
		tags[i] = toPortfolioTagResponse(t)
	}
	return PortfolioItemResponse{
		ID:              item.ID.String(),
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
		StartedAt:       item.StartedAt,
		EndedAt:         item.EndedAt,
		Tags:            tags,
		CreatedAt:       item.CreatedAt,
	}
}

func toFullProfileResponse(result *usersvc.GetFullProfileResult) FullProfileResponse {
	skills := make([]SkillResponse, len(result.Skills))
	for i, s := range result.Skills {
		skills[i] = toSkillResponse(s)
	}

	exps := make([]ExperienceResponse, len(result.Experiences))
	for i, e := range result.Experiences {
		exps[i] = toExperienceResponse(e)
	}

	portfolio := make([]PortfolioItemResponse, len(result.Portfolio))
	for i, item := range result.Portfolio {
		portfolio[i] = toPortfolioItemResponse(item)
	}

	return FullProfileResponse{
		User: PublicUserSummary{
			ID:       result.User.ID.String(),
			Username: result.User.Username,
		},
		Profile:     toProfileResponse(result.Profile),
		Skills:      skills,
		Experiences: exps,
		Portfolio:   portfolio,
	}
}
