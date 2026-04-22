package user

import (
	"time"

	"github.com/google/uuid"
)

// Profile stores extended personal information for a person — 1:1 with persons table.
type Profile struct {
	ID          uuid.UUID
	PersonID    uuid.UUID
	DisplayName string // ชื่อที่แสดงต่อสาธารณะ
	Headline    string // e.g. "Senior Go Engineer"
	Bio         string // multi-line about me
	Location    string // e.g. "Bangkok, TH"
	AvatarURL   string
	SocialLinks []SocialLink // loaded via 1:many join
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// SocialLink is a single social/contact link belonging to a person's profile.
// PlatformID is a foreign key referencing social_platforms.id.
type SocialLink struct {
	ID         uuid.UUID
	PersonID   uuid.UUID
	PlatformID uuid.UUID // FK → social_platforms.id
	URL        string
	SortOrder  int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Skill category constants.
const (
	SkillCategoryBackend  = "backend"
	SkillCategoryFrontend = "frontend"
	SkillCategoryDevOps   = "devops"
	SkillCategoryOther    = "other"
)

// Skill represents a technology or competency a person possesses.
type Skill struct {
	ID          uuid.UUID
	PersonID    uuid.UUID
	Name        string // e.g. "Go", "PostgreSQL"
	Category    string // backend | frontend | devops | other
	Proficiency int    // 1 (beginner) – 5 (expert)
	SortOrder   int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Experience represents a single work history entry on a person's timeline.
type Experience struct {
	ID          uuid.UUID
	PersonID    uuid.UUID
	Company     string
	Position    string
	Location    string // e.g. "Bangkok, TH" or "Remote"
	Description string
	StartedAt   time.Time
	EndedAt     *time.Time // nil = currently employed here
	IsCurrent   bool
	SortOrder   int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Portfolio item status constants.
const (
	PortfolioStatusDraft     = "draft"
	PortfolioStatusPublished = "published"
)

// PortfolioTag is a shared label applied to portfolio items across all users.
type PortfolioTag struct {
	ID   uuid.UUID
	Name string // human-readable, e.g. "Go"
	Slug string // URL-safe, e.g. "go"
}

// PortfolioItem represents a project or work showcase piece owned by a person.
type PortfolioItem struct {
	ID              uuid.UUID
	PersonID        uuid.UUID
	Title           string
	Slug            string
	Summary         string // one-liner for card views
	Description     string // short description
	ContentMarkdown string // full rich content for detail page
	CoverImageURL   string
	DemoURL         string
	SourceURL       string
	Status          string
	Featured        bool
	SortOrder       int
	StartedAt       *time.Time
	EndedAt         *time.Time
	Tags            []PortfolioTag
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}
