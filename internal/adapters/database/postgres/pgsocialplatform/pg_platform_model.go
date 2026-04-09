package pgsocialplatform

import (
	"time"

	"vexentra-api/internal/modules/socialplatform"

	"github.com/google/uuid"
)

type socialPlatformModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Key       string    `gorm:"uniqueIndex;not null"`
	Name      string    `gorm:"not null"`
	IconURL   string
	SortOrder int  `gorm:"default:0"`
	IsActive  bool `gorm:"default:true"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (socialPlatformModel) TableName() string {
	return "social_platforms"
}

func (m *socialPlatformModel) ToEntity() *socialplatform.SocialPlatform {
	return &socialplatform.SocialPlatform{
		ID:        m.ID,
		Key:       m.Key,
		Name:      m.Name,
		IconURL:   m.IconURL,
		SortOrder: m.SortOrder,
		IsActive:  m.IsActive,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
