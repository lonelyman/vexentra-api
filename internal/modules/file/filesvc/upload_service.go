package filesvc

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"
	"vexentra-api/internal/config"
	"vexentra-api/internal/modules/file"
	"vexentra-api/internal/modules/user"
	"vexentra-api/pkg/custom_errors"
	"vexentra-api/pkg/logger"
	"vexentra-api/pkg/storage"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PresignResult struct {
	UploadSessionID string            `json:"upload_session_id"`
	UploadURL       string            `json:"upload_url"`
	UploadHeaders   map[string]string `json:"upload_headers"`
	ExpiresAt       time.Time         `json:"expires_at"`
}

type CompleteResult struct {
	FileID string `json:"file_id"`
}

type Service interface {
	PresignProfileImage(ctx context.Context, caller user.Caller, filename, mimeType string, sizeBytes int64, targetPersonID *uuid.UUID) (*PresignResult, *custom_errors.AppError)
	CompleteProfileImage(ctx context.Context, caller user.Caller, sessionID uuid.UUID) (*CompleteResult, *custom_errors.AppError)
	GetFileURL(ctx context.Context, caller user.Caller, fileID uuid.UUID) (string, *custom_errors.AppError)
	DeleteFile(ctx context.Context, caller user.Caller, fileID uuid.UUID) *custom_errors.AppError
}

type service struct {
	db          *gorm.DB
	cfg         config.StorageConfig
	repo        file.Repository
	profileRepo user.ProfileRepository
	storage     storage.ObjectStorage
	log         logger.Logger
}

func NewService(
	db *gorm.DB,
	cfg config.StorageConfig,
	repo file.Repository,
	profileRepo user.ProfileRepository,
	objectStorage storage.ObjectStorage,
	log logger.Logger,
) Service {
	return &service{
		db:          db,
		cfg:         cfg,
		repo:        repo,
		profileRepo: profileRepo,
		storage:     objectStorage,
		log:         log,
	}
}

func (s *service) PresignProfileImage(ctx context.Context, caller user.Caller, filename, mimeType string, sizeBytes int64, targetPersonID *uuid.UUID) (*PresignResult, *custom_errors.AppError) {
	mimeType = strings.TrimSpace(strings.ToLower(mimeType))
	filename = strings.TrimSpace(filename)
	if filename == "" {
		filename = "upload"
	}
	if sizeBytes <= 0 {
		return nil, custom_errors.NewBadRequestError(custom_errors.ErrValidation, "ขนาดไฟล์ไม่ถูกต้อง")
	}
	if sizeBytes > s.cfg.HardMaxFileSize {
		return nil, custom_errors.NewBadRequestError(custom_errors.ErrValidation, "ไฟล์เกินเพดานระบบ 30MB")
	}
	if sizeBytes > s.cfg.ProfileMaxImageSize {
		return nil, custom_errors.NewBadRequestError(custom_errors.ErrValidation, "รูปโปรไฟล์ต้องไม่เกิน 5MB")
	}
	if !isAllowedImageMIME(mimeType) {
		return nil, custom_errors.NewBadRequestError(custom_errors.ErrValidation, "รองรับเฉพาะ JPEG, PNG, WEBP")
	}

	sessionPersonID := caller.PersonID
	if targetPersonID != nil {
		if !caller.IsAdmin() {
			return nil, custom_errors.New(403, custom_errors.ErrForbidden, "เฉพาะ admin เท่านั้นที่อัปโหลดแทนผู้อื่นได้")
		}
		sessionPersonID = *targetPersonID
	}

	sessionID, err := uuid.NewV7()
	if err != nil {
		return nil, custom_errors.NewInternalError("ไม่สามารถสร้าง session ได้")
	}
	tempObjectKey := fmt.Sprintf("uploads/tmp/%s", sessionID.String())
	expiresAt := time.Now().UTC().Add(s.cfg.PresignTTL)

	uploadSession := &file.UploadSession{
		ID:               sessionID,
		UserID:           caller.UserID,
		PersonID:         sessionPersonID,
		Intent:           file.IntentProfileImage,
		TempObjectKey:    tempObjectKey,
		OriginalFilename: sanitizeFilename(filename),
		ExpectedMIME:     mimeType,
		ExpectedMaxSize:  s.cfg.ProfileMaxImageSize,
		Status:           file.UploadStatusPending,
		ExpiresAt:        expiresAt,
	}

	if err := s.repo.CreateUploadSession(ctx, uploadSession); err != nil {
		s.log.Error("CREATE_UPLOAD_SESSION_ERROR", err)
		return nil, custom_errors.NewInternalError("ไม่สามารถสร้าง upload session ได้")
	}

	presigned, err := s.storage.PresignPut(ctx, tempObjectKey, mimeType, s.cfg.PresignTTL)
	if err != nil {
		s.log.Error("PRESIGN_UPLOAD_ERROR", err)
		return nil, custom_errors.NewInternalError("ไม่สามารถสร้าง URL อัปโหลดได้")
	}

	return &PresignResult{
		UploadSessionID: sessionID.String(),
		UploadURL:       presigned.URL,
		UploadHeaders:   presigned.Headers,
		ExpiresAt:       expiresAt,
	}, nil
}

func (s *service) CompleteProfileImage(ctx context.Context, caller user.Caller, sessionID uuid.UUID) (*CompleteResult, *custom_errors.AppError) {
	session, err := s.repo.GetUploadSessionByID(ctx, sessionID)
	if err != nil {
		s.log.Error("GET_UPLOAD_SESSION_ERROR", err)
		return nil, custom_errors.NewInternalError("ไม่สามารถอ่าน upload session ได้")
	}
	if session == nil {
		return nil, custom_errors.NewNotFoundError(custom_errors.ErrNotFound, "ไม่พบ upload session")
	}
	if session.UserID != caller.UserID {
		return nil, custom_errors.New(403, custom_errors.ErrForbidden, "ไม่มีสิทธิ์ใช้ upload session นี้")
	}
	if session.PersonID != caller.PersonID && !caller.IsAdmin() {
		return nil, custom_errors.New(403, custom_errors.ErrForbidden, "ไม่มีสิทธิ์อัปโหลดให้ผู้ใช้นี้")
	}
	if session.Intent != file.IntentProfileImage {
		return nil, custom_errors.NewBadRequestError(custom_errors.ErrValidation, "intent ไม่รองรับ")
	}

	now := time.Now().UTC()
	if now.After(session.ExpiresAt) {
		return nil, custom_errors.NewBadRequestError(custom_errors.ErrValidation, "upload session หมดอายุแล้ว")
	}

	fileID := session.ID
	permanentObjectKey := fmt.Sprintf("profiles/%s/%s", session.PersonID.String(), fileID.String())

	if existing, findErr := s.repo.FindFileByObjectKey(ctx, permanentObjectKey); findErr == nil && existing != nil {
		return &CompleteResult{FileID: existing.ID.String()}, nil
	}

	obj, statErr := s.storage.StatObject(ctx, session.TempObjectKey)
	if statErr != nil {
		s.log.Error("STAT_TEMP_OBJECT_ERROR", statErr)
		return nil, custom_errors.NewBadRequestError(custom_errors.ErrValidation, "ไม่พบไฟล์ที่อัปโหลด")
	}

	if obj.Size <= 0 || obj.Size > session.ExpectedMaxSize || obj.Size > s.cfg.HardMaxFileSize {
		return nil, custom_errors.NewBadRequestError(custom_errors.ErrValidation, "ขนาดไฟล์ไม่ถูกต้อง")
	}
	sniffedMIME, sniffErr := s.detectObjectMIME(ctx, session.TempObjectKey)
	if sniffErr != nil {
		s.log.Error("SNIFF_OBJECT_MIME_ERROR", sniffErr)
		return nil, custom_errors.NewBadRequestError(custom_errors.ErrValidation, "ไม่สามารถอ่านประเภทไฟล์ได้")
	}
	if !isAllowedImageMIME(sniffedMIME) {
		return nil, custom_errors.NewBadRequestError(custom_errors.ErrValidation, "ประเภทไฟล์ไม่รองรับ")
	}

	checksum, hashErr := s.calculateSHA256(ctx, session.TempObjectKey)
	if hashErr != nil {
		s.log.Error("HASH_OBJECT_ERROR", hashErr)
		return nil, custom_errors.NewInternalError("ไม่สามารถตรวจสอบไฟล์ได้")
	}

	if copyErr := s.storage.CopyObject(ctx, session.TempObjectKey, permanentObjectKey, sniffedMIME); copyErr != nil {
		s.log.Error("COPY_OBJECT_ERROR", copyErr)
		return nil, custom_errors.NewInternalError("ไม่สามารถย้ายไฟล์เข้า storage หลักได้")
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)
		txProfileRepo := s.profileRepo.WithTx(tx)
		fileEntity := &file.File{
			ID:               fileID,
			OwnerType:        file.OwnerTypePerson,
			OwnerID:          session.PersonID,
			Category:         file.CategoryProfileImage,
			ObjectKey:        permanentObjectKey,
			OriginalFilename: session.OriginalFilename,
			MIMEType:         sniffedMIME,
			SizeBytes:        obj.Size,
			SHA256:           checksum,
			ETag:             obj.ETag,
			Visibility:       file.VisibilityPrivate,
			ProcessingStatus: "ready",
			CreatedBy:        caller.UserID,
		}
		if createErr := txRepo.CreateFile(ctx, fileEntity); createErr != nil {
			return createErr
		}
		if markErr := txRepo.MarkUploadSessionCompleted(ctx, session.ID, now); markErr != nil {
			return markErr
		}
		if avatarErr := txProfileRepo.SetProfileAvatarFileID(ctx, session.PersonID, fileID); avatarErr != nil {
			return avatarErr
		}
		return nil
	})
	if err != nil {
		s.log.Error("COMPLETE_UPLOAD_TX_ERROR", err)
		return nil, custom_errors.NewInternalError("บันทึกไฟล์ไม่สำเร็จ")
	}

	_ = s.storage.RemoveObject(ctx, session.TempObjectKey)

	return &CompleteResult{FileID: fileID.String()}, nil
}

func (s *service) GetFileURL(ctx context.Context, caller user.Caller, fileID uuid.UUID) (string, *custom_errors.AppError) {
	f, err := s.repo.GetFileByID(ctx, fileID)
	if err != nil {
		return "", custom_errors.NewInternalError("ไม่สามารถอ่านไฟล์ได้")
	}
	if f == nil {
		return "", custom_errors.NewNotFoundError(custom_errors.ErrNotFound, "ไม่พบไฟล์")
	}
	if f.OwnerType == file.OwnerTypePerson && f.OwnerID != caller.PersonID && !caller.IsAdmin() {
		return "", custom_errors.New(403, custom_errors.ErrForbidden, "ไม่มีสิทธิ์เข้าถึงไฟล์นี้")
	}
	url, presignErr := s.storage.PresignGet(ctx, f.ObjectKey, s.cfg.PresignTTL)
	if presignErr != nil {
		return "", custom_errors.NewInternalError("ไม่สามารถสร้าง URL สำหรับไฟล์ได้")
	}
	return url, nil
}

func (s *service) DeleteFile(ctx context.Context, caller user.Caller, fileID uuid.UUID) *custom_errors.AppError {
	f, err := s.repo.GetFileByID(ctx, fileID)
	if err != nil {
		return custom_errors.NewInternalError("ไม่สามารถอ่านไฟล์ได้")
	}
	if f == nil {
		return custom_errors.NewNotFoundError(custom_errors.ErrNotFound, "ไม่พบไฟล์")
	}
	if f.OwnerType == file.OwnerTypePerson && f.OwnerID != caller.PersonID && !caller.IsAdmin() {
		return custom_errors.New(403, custom_errors.ErrForbidden, "ไม่มีสิทธิ์ลบไฟล์นี้")
	}
	if err := s.repo.SoftDeleteFile(ctx, fileID); err != nil {
		return custom_errors.NewInternalError("ลบข้อมูลไฟล์ไม่สำเร็จ")
	}
	_ = s.storage.RemoveObject(ctx, f.ObjectKey)
	return nil
}

func (s *service) calculateSHA256(ctx context.Context, objectKey string) (string, error) {
	reader, err := s.storage.GetObject(ctx, objectKey)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	h := sha256.New()
	if _, err = io.Copy(h, reader); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func (s *service) detectObjectMIME(ctx context.Context, objectKey string) (string, error) {
	reader, err := s.storage.GetObject(ctx, objectKey)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	buf := make([]byte, 512)
	n, readErr := io.ReadFull(reader, buf)
	if readErr != nil && readErr != io.EOF && readErr != io.ErrUnexpectedEOF {
		return "", readErr
	}
	detected := strings.ToLower(http.DetectContentType(bytes.TrimSpace(buf[:n])))
	switch detected {
	case "image/jpeg", "image/png", "image/webp":
		return detected, nil
	default:
		return detected, nil
	}
}

func isAllowedImageMIME(mime string) bool {
	switch strings.TrimSpace(strings.ToLower(mime)) {
	case "image/jpeg", "image/png", "image/webp":
		return true
	default:
		return false
	}
}

func sanitizeFilename(name string) string {
	base := filepath.Base(strings.TrimSpace(name))
	if base == "." || base == "/" || base == "" {
		return "upload"
	}
	return base
}
