package file

import (
	"time"

	"github.com/google/uuid"
)

const (
	IntentProfileImage = "profile_image"

	OwnerTypePerson = "person"

	CategoryProfileImage = "profile_image"

	VisibilityPrivate = "private"

	UploadStatusPending   = "pending"
	UploadStatusCompleted = "completed"
	UploadStatusExpired   = "expired"
	UploadStatusCancelled = "cancelled"
)

type UploadSession struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	PersonID         uuid.UUID
	Intent           string
	TempObjectKey    string
	OriginalFilename string
	ExpectedMIME     string
	ExpectedMaxSize  int64
	Status           string
	ExpiresAt        time.Time
	CompletedAt      *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type File struct {
	ID               uuid.UUID
	OwnerType        string
	OwnerID          uuid.UUID
	Category         string
	ObjectKey        string
	OriginalFilename string
	MIMEType         string
	SizeBytes        int64
	SHA256           string
	ETag             string
	Visibility       string
	ProcessingStatus string
	ProcessingError  string
	CreatedBy        uuid.UUID
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
