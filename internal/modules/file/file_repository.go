package file

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	WithTx(tx *gorm.DB) Repository

	CreateUploadSession(ctx context.Context, s *UploadSession) error
	GetUploadSessionByID(ctx context.Context, id uuid.UUID) (*UploadSession, error)
	MarkUploadSessionCompleted(ctx context.Context, id uuid.UUID, completedAt time.Time) error

	CreateFile(ctx context.Context, f *File) error
	GetFileByID(ctx context.Context, id uuid.UUID) (*File, error)
	FindFileByObjectKey(ctx context.Context, objectKey string) (*File, error)
	SoftDeleteFile(ctx context.Context, id uuid.UUID) error
}
