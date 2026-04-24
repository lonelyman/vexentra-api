package pgfile

import (
	"context"
	"errors"
	"time"
	"vexentra-api/internal/modules/file"
	"vexentra-api/pkg/logger"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type repository struct {
	db     *gorm.DB
	logger logger.Logger
}

func NewRepository(db *gorm.DB, l logger.Logger) file.Repository {
	if l == nil {
		l = logger.Get()
	}
	return &repository{db: db, logger: l}
}

func (r *repository) WithTx(tx *gorm.DB) file.Repository {
	return &repository{db: tx, logger: r.logger}
}

func (r *repository) CreateUploadSession(ctx context.Context, s *file.UploadSession) error {
	m := &uploadSessionModel{
		ID:               s.ID,
		UserID:           s.UserID,
		PersonID:         s.PersonID,
		Intent:           s.Intent,
		TempObjectKey:    s.TempObjectKey,
		OriginalFilename: s.OriginalFilename,
		ExpectedMIME:     s.ExpectedMIME,
		ExpectedMaxSize:  s.ExpectedMaxSize,
		Status:           s.Status,
		ExpiresAt:        s.ExpiresAt,
	}
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *repository) GetUploadSessionByID(ctx context.Context, id uuid.UUID) (*file.UploadSession, error) {
	var m uploadSessionModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return m.ToEntity(), nil
}

func (r *repository) MarkUploadSessionCompleted(ctx context.Context, id uuid.UUID, completedAt time.Time) error {
	return r.db.WithContext(ctx).Model(&uploadSessionModel{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":       file.UploadStatusCompleted,
			"completed_at": completedAt,
		}).Error
}

func (r *repository) CreateFile(ctx context.Context, f *file.File) error {
	m := &fileModel{
		ID:               f.ID,
		OwnerType:        f.OwnerType,
		OwnerID:          f.OwnerID,
		Category:         f.Category,
		ObjectKey:        f.ObjectKey,
		OriginalFilename: f.OriginalFilename,
		MIMEType:         f.MIMEType,
		SizeBytes:        f.SizeBytes,
		SHA256:           f.SHA256,
		ETag:             f.ETag,
		Visibility:       f.Visibility,
		ProcessingStatus: f.ProcessingStatus,
		ProcessingError:  f.ProcessingError,
		Metadata:         datatypes.JSON([]byte("{}")),
		CreatedBy:        f.CreatedBy,
	}
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *repository) GetFileByID(ctx context.Context, id uuid.UUID) (*file.File, error) {
	var m fileModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return m.ToEntity(), nil
}

func (r *repository) FindFileByObjectKey(ctx context.Context, objectKey string) (*file.File, error) {
	var m fileModel
	if err := r.db.WithContext(ctx).Where("object_key = ?", objectKey).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return m.ToEntity(), nil
}

func (r *repository) SoftDeleteFile(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&fileModel{}).Error
}
