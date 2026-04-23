package pgfile

import (
	"time"
	"vexentra-api/internal/modules/file"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type fileModel struct {
	ID               uuid.UUID      `gorm:"type:uuid;primaryKey"`
	OwnerType        string         `gorm:"column:owner_type"`
	OwnerID          uuid.UUID      `gorm:"type:uuid;column:owner_id"`
	Category         string         `gorm:"column:category"`
	ObjectKey        string         `gorm:"column:object_key"`
	OriginalFilename string         `gorm:"column:original_filename"`
	MIMEType         string         `gorm:"column:mime_type"`
	SizeBytes        int64          `gorm:"column:size_bytes"`
	SHA256           string         `gorm:"column:sha256"`
	ETag             string         `gorm:"column:etag"`
	Visibility       string         `gorm:"column:visibility"`
	ProcessingStatus string         `gorm:"column:processing_status"`
	ProcessingError  string         `gorm:"column:processing_error"`
	Metadata         datatypes.JSON `gorm:"column:metadata"`
	CreatedBy        uuid.UUID      `gorm:"type:uuid;column:created_by"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (fileModel) TableName() string { return "files" }

func (m *fileModel) ToEntity() *file.File {
	return &file.File{
		ID:               m.ID,
		OwnerType:        m.OwnerType,
		OwnerID:          m.OwnerID,
		Category:         m.Category,
		ObjectKey:        m.ObjectKey,
		OriginalFilename: m.OriginalFilename,
		MIMEType:         m.MIMEType,
		SizeBytes:        m.SizeBytes,
		SHA256:           m.SHA256,
		ETag:             m.ETag,
		Visibility:       m.Visibility,
		ProcessingStatus: m.ProcessingStatus,
		ProcessingError:  m.ProcessingError,
		CreatedBy:        m.CreatedBy,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}
}

type uploadSessionModel struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID           uuid.UUID  `gorm:"type:uuid;column:user_id"`
	PersonID         uuid.UUID  `gorm:"type:uuid;column:person_id"`
	Intent           string     `gorm:"column:intent"`
	TempObjectKey    string     `gorm:"column:temp_object_key"`
	OriginalFilename string     `gorm:"column:original_filename"`
	ExpectedMIME     string     `gorm:"column:expected_mime"`
	ExpectedMaxSize  int64      `gorm:"column:expected_max_size"`
	Status           string     `gorm:"column:status"`
	ExpiresAt        time.Time  `gorm:"column:expires_at"`
	CompletedAt      *time.Time `gorm:"column:completed_at"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (uploadSessionModel) TableName() string { return "upload_sessions" }

func (m *uploadSessionModel) ToEntity() *file.UploadSession {
	return &file.UploadSession{
		ID:               m.ID,
		UserID:           m.UserID,
		PersonID:         m.PersonID,
		Intent:           m.Intent,
		TempObjectKey:    m.TempObjectKey,
		OriginalFilename: m.OriginalFilename,
		ExpectedMIME:     m.ExpectedMIME,
		ExpectedMaxSize:  m.ExpectedMaxSize,
		Status:           m.Status,
		ExpiresAt:        m.ExpiresAt,
		CompletedAt:      m.CompletedAt,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}
}
