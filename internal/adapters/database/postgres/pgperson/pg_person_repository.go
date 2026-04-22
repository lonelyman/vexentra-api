package pgperson

import (
	"context"
	"errors"
	"time"
	"vexentra-api/internal/adapters/database/postgres/pgtx"
	"vexentra-api/internal/modules/person"
	"vexentra-api/pkg/custom_errors"
	"vexentra-api/pkg/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type personRepository struct {
	db     *gorm.DB
	logger logger.Logger
}

func NewPersonRepository(db *gorm.DB, l logger.Logger) person.PersonRepository {
	if l == nil {
		l = logger.Get()
	}
	return &personRepository{db: db, logger: l}
}

func (r *personRepository) Create(ctx context.Context, p *person.Person) error {
	id, err := uuid.NewV7()
	if err != nil {
		return custom_errors.NewInternalError("ไม่สามารถสร้าง ID ได้")
	}
	p.ID = id

	m := fromEntity(p)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		r.logger.Error("DB_CREATE_PERSON_ERROR", err)
		return custom_errors.NewInternalError("ไม่สามารถสร้าง person ได้")
	}
	return nil
}

func (r *personRepository) GetByID(ctx context.Context, id uuid.UUID) (*person.Person, error) {
	var m personModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, custom_errors.New(404, custom_errors.ErrNotFound, "ไม่พบข้อมูล person")
		}
		r.logger.Error("DB_GET_PERSON_ERROR", err)
		return nil, custom_errors.NewInternalError("เกิดข้อผิดพลาดในการดึงข้อมูล")
	}
	return m.ToEntity(), nil
}

func (r *personRepository) GetByLinkedUserID(ctx context.Context, userID uuid.UUID) (*person.Person, error) {
	var m personModel
	if err := r.db.WithContext(ctx).Where("linked_user_id = ?", userID).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // user hasn't been linked yet
		}
		r.logger.Error("DB_GET_PERSON_BY_USER_ERROR", err)
		return nil, custom_errors.NewInternalError("เกิดข้อผิดพลาดในการดึงข้อมูล")
	}
	return m.ToEntity(), nil
}

func (r *personRepository) GetByInviteEmail(ctx context.Context, email string) (*person.Person, error) {
	var m personModel
	if err := r.db.WithContext(ctx).Where("invite_email = ?", email).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.logger.Error("DB_GET_PERSON_BY_INVITE_EMAIL_ERROR", err)
		return nil, custom_errors.NewInternalError("เกิดข้อผิดพลาดในการดึงข้อมูล")
	}
	return m.ToEntity(), nil
}

func (r *personRepository) Update(ctx context.Context, p *person.Person) error {
	m := fromEntity(p)
	m.ID = p.ID
	if err := r.db.WithContext(ctx).Save(m).Error; err != nil {
		r.logger.Error("DB_UPDATE_PERSON_ERROR", err)
		return custom_errors.NewInternalError("ไม่สามารถอัปเดต person ได้")
	}
	return nil
}

func (r *personRepository) LinkUser(ctx context.Context, personID, userID uuid.UUID) error {
	if err := pgtx.DB(ctx, r.db).WithContext(ctx).
		Model(&personModel{}).
		Where("id = ?", personID).
		Update("linked_user_id", userID).Error; err != nil {
		r.logger.Error("DB_LINK_PERSON_USER_ERROR", err)
		return custom_errors.NewInternalError("ไม่สามารถเชื่อม person กับ user ได้")
	}
	return nil
}

func (r *personRepository) GetByInviteToken(ctx context.Context, token string) (*person.Person, error) {
	var m personModel
	if err := r.db.WithContext(ctx).Where("invite_token = ?", token).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.logger.Error("DB_GET_PERSON_BY_INVITE_TOKEN_ERROR", err)
		return nil, custom_errors.NewInternalError("เกิดข้อผิดพลาดในการดึงข้อมูล")
	}
	return m.ToEntity(), nil
}

func (r *personRepository) SetInviteToken(ctx context.Context, personID uuid.UUID, token string, expiresAt time.Time) error {
	if err := r.db.WithContext(ctx).
		Model(&personModel{}).
		Where("id = ?", personID).
		Updates(map[string]any{
			"invite_token":            token,
			"invite_token_expires_at": expiresAt,
		}).Error; err != nil {
		r.logger.Error("DB_SET_INVITE_TOKEN_ERROR", err)
		return custom_errors.NewInternalError("ไม่สามารถสร้าง invite token ได้")
	}
	return nil
}

func (r *personRepository) ClearInviteToken(ctx context.Context, personID uuid.UUID) error {
	if err := r.db.WithContext(ctx).
		Model(&personModel{}).
		Where("id = ?", personID).
		Updates(map[string]any{
			"invite_token":            nil,
			"invite_token_expires_at": nil,
		}).Error; err != nil {
		r.logger.Error("DB_CLEAR_INVITE_TOKEN_ERROR", err)
		return custom_errors.NewInternalError("ไม่สามารถล้าง invite token ได้")
	}
	return nil
}

func (r *personRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&personModel{}, "id = ?", id).Error; err != nil {
		r.logger.Error("DB_DELETE_PERSON_ERROR", err)
		return custom_errors.NewInternalError("ไม่สามารถลบ person ได้")
	}
	return nil
}
