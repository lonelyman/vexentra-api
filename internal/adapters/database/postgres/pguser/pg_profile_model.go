package pguser

import (
	"time"
	"vexentra-api/internal/modules/user"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ─────────────────────────────────────────────────────────────────────
//  Profile Model — 1:1 with users
// ─────────────────────────────────────────────────────────────────────

type profileModel struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID      uuid.UUID `gorm:"type:uuid;uniqueIndex;not null"`
	DisplayName string    `gorm:"size:100;column:display_name"`
	Headline    string    `gorm:"column:headline"`
	Bio         string    `gorm:"column:bio;type:text"`
	Location    string    `gorm:"column:location"`
	AvatarURL   string    `gorm:"column:avatar_url"`
	WebsiteURL  string    `gorm:"column:website_url"`
	GitHubURL   string    `gorm:"column:github_url"`
	LinkedInURL string    `gorm:"column:linkedin_url"`
	TwitterURL  string    `gorm:"column:twitter_url"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (profileModel) TableName() string { return "profiles" }

func (m *profileModel) ToEntity() *user.Profile {
	return &user.Profile{
		ID:          m.ID,
		UserID:      m.UserID,
		DisplayName: m.DisplayName,
		Headline:    m.Headline,
		Bio:         m.Bio,
		Location:    m.Location,
		AvatarURL:   m.AvatarURL,
		WebsiteURL:  m.WebsiteURL,
		GitHubURL:   m.GitHubURL,
		LinkedInURL: m.LinkedInURL,
		TwitterURL:  m.TwitterURL,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// ─────────────────────────────────────────────────────────────────────
//  Skill Model
// ─────────────────────────────────────────────────────────────────────

type skillModel struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID      uuid.UUID `gorm:"type:uuid;index;not null"`
	Name        string    `gorm:"column:name;not null"`
	Category    string    `gorm:"column:category;default:'other'"`
	Proficiency int       `gorm:"column:proficiency;default:1"`
	SortOrder   int       `gorm:"column:sort_order;default:0"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (skillModel) TableName() string { return "skills" }

func (m *skillModel) ToEntity() *user.Skill {
	return &user.Skill{
		ID:          m.ID,
		UserID:      m.UserID,
		Name:        m.Name,
		Category:    m.Category,
		Proficiency: m.Proficiency,
		SortOrder:   m.SortOrder,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// ─────────────────────────────────────────────────────────────────────
//  Experience Model
// ─────────────────────────────────────────────────────────────────────

type experienceModel struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID      uuid.UUID  `gorm:"type:uuid;index;not null"`
	Company     string     `gorm:"column:company;not null"`
	Position    string     `gorm:"column:position;not null"`
	Location    string     `gorm:"size:100;column:location"`
	Description string     `gorm:"column:description;type:text"`
	StartedAt   time.Time  `gorm:"column:started_at"`
	EndedAt     *time.Time `gorm:"column:ended_at"`
	IsCurrent   bool       `gorm:"column:is_current;default:false"`
	SortOrder   int        `gorm:"column:sort_order;default:0"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (experienceModel) TableName() string { return "experiences" }

func (m *experienceModel) ToEntity() *user.Experience {
	return &user.Experience{
		ID:          m.ID,
		UserID:      m.UserID,
		Company:     m.Company,
		Position:    m.Position,
		Location:    m.Location,
		Description: m.Description,
		StartedAt:   m.StartedAt,
		EndedAt:     m.EndedAt,
		IsCurrent:   m.IsCurrent,
		SortOrder:   m.SortOrder,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// ─────────────────────────────────────────────────────────────────────
//  Portfolio Tag Model (global pool, shared across all users)
// ─────────────────────────────────────────────────────────────────────

type portfolioTagModel struct {
	ID   uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name string    `gorm:"column:name;uniqueIndex;not null"`
	Slug string    `gorm:"column:slug;uniqueIndex;not null"`
}

func (portfolioTagModel) TableName() string { return "portfolio_tags" }

func (m *portfolioTagModel) ToEntity() *user.PortfolioTag {
	return &user.PortfolioTag{
		ID:   m.ID,
		Name: m.Name,
		Slug: m.Slug,
	}
}

// ─────────────────────────────────────────────────────────────────────
//  Portfolio Item Model
// ─────────────────────────────────────────────────────────────────────

type portfolioItemModel struct {
	ID              uuid.UUID           `gorm:"type:uuid;primaryKey"`
	UserID          uuid.UUID           `gorm:"type:uuid;index;not null"`
	Title           string              `gorm:"column:title;not null"`
	Slug            string              `gorm:"column:slug;not null"`
	Summary         string              `gorm:"column:summary"`
	Description     string              `gorm:"column:description"`
	ContentMarkdown string              `gorm:"column:content_markdown;type:text"`
	CoverImageURL   string              `gorm:"column:cover_image_url"`
	DemoURL         string              `gorm:"column:demo_url"`
	SourceURL       string              `gorm:"column:source_url"`
	Status          string              `gorm:"column:status;default:'draft'"`
	Featured        bool                `gorm:"column:featured;default:false"`
	SortOrder       int                 `gorm:"column:sort_order;default:0"`
	StartedAt       *time.Time          `gorm:"column:started_at"`
	EndedAt         *time.Time          `gorm:"column:ended_at"`
	Tags            []portfolioTagModel `gorm:"many2many:portfolio_item_tags;joinForeignKey:PortfolioItemID;joinReferences:TagID"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

func (portfolioItemModel) TableName() string { return "portfolio_items" }

func (m *portfolioItemModel) ToEntity() *user.PortfolioItem {
	var deletedAt *time.Time
	if m.DeletedAt.Valid {
		t := m.DeletedAt.Time
		deletedAt = &t
	}
	tags := make([]user.PortfolioTag, len(m.Tags))
	for i := range m.Tags {
		tags[i] = *m.Tags[i].ToEntity()
	}
	return &user.PortfolioItem{
		ID:              m.ID,
		UserID:          m.UserID,
		Title:           m.Title,
		Slug:            m.Slug,
		Summary:         m.Summary,
		Description:     m.Description,
		ContentMarkdown: m.ContentMarkdown,
		CoverImageURL:   m.CoverImageURL,
		DemoURL:         m.DemoURL,
		SourceURL:       m.SourceURL,
		Status:          m.Status,
		Featured:        m.Featured,
		SortOrder:       m.SortOrder,
		StartedAt:       m.StartedAt,
		EndedAt:         m.EndedAt,
		Tags:            tags,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
		DeletedAt:       deletedAt,
	}
}
