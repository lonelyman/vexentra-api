package pguser

import (
	"time"
	"vexentra-api/internal/modules/user"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ─────────────────────────────────────────────────────────────────────
//  Profile Model — 1:1 with persons
// ─────────────────────────────────────────────────────────────────────

type profileModel struct {
	ID           uuid.UUID         `gorm:"type:uuid;primaryKey"`
	PersonID     uuid.UUID         `gorm:"type:uuid;uniqueIndex;not null"`
	DisplayName  string            `gorm:"size:100;column:display_name"`
	Headline     string            `gorm:"column:headline"`
	Bio          string            `gorm:"column:bio;type:text"`
	Location     string            `gorm:"column:location"`
	AvatarFileID *uuid.UUID        `gorm:"type:uuid;column:avatar_file_id"`
	SocialLinks  []socialLinkModel `gorm:"foreignKey:PersonID;references:PersonID"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (profileModel) TableName() string { return "profiles" }

func (m *profileModel) ToEntity() *user.Profile {
	links := make([]user.SocialLink, len(m.SocialLinks))
	for i, l := range m.SocialLinks {
		links[i] = *l.ToEntity()
	}
	return &user.Profile{
		ID:           m.ID,
		PersonID:     m.PersonID,
		DisplayName:  m.DisplayName,
		Headline:     m.Headline,
		Bio:          m.Bio,
		Location:     m.Location,
		AvatarFileID: m.AvatarFileID,
		SocialLinks:  links,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

// ─────────────────────────────────────────────────────────────────────
//  Social Link Model (1:many with profiles via person_id)
// ─────────────────────────────────────────────────────────────────────

// socialLinkModel maps to the social_links table.
// PlatformID is a FK referencing social_platforms.id — one platform per person.
type socialLinkModel struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	PersonID   uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_social_links_person_platform;not null"`
	PlatformID uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_social_links_person_platform;not null"`
	URL        string    `gorm:"size:512;column:url;not null"`
	SortOrder  int       `gorm:"column:sort_order;default:0"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (socialLinkModel) TableName() string { return "social_links" }

func (m *socialLinkModel) ToEntity() *user.SocialLink {
	return &user.SocialLink{
		ID:         m.ID,
		PersonID:   m.PersonID,
		PlatformID: m.PlatformID,
		URL:        m.URL,
		SortOrder:  m.SortOrder,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}
}

// ─────────────────────────────────────────────────────────────────────
//  Skill Model
// ─────────────────────────────────────────────────────────────────────

type skillModel struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	PersonID    uuid.UUID `gorm:"type:uuid;index;not null"`
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
		PersonID:    m.PersonID,
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
	PersonID    uuid.UUID  `gorm:"type:uuid;index;not null"`
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
		PersonID:    m.PersonID,
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
	PersonID        uuid.UUID           `gorm:"type:uuid;index;not null"`
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
		PersonID:        m.PersonID,
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
