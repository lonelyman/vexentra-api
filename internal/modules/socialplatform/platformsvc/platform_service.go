package platformsvc

import (
	"context"

	"vexentra-api/internal/modules/socialplatform"
	"vexentra-api/pkg/custom_errors"
	"vexentra-api/pkg/logger"

	"github.com/google/uuid"
)

type SocialPlatformService interface {
	List(ctx context.Context) ([]*socialplatform.SocialPlatform, error)
	ListOffset(ctx context.Context, limit, offset int) ([]*socialplatform.SocialPlatform, int64, error)
	Create(ctx context.Context, p *socialplatform.SocialPlatform) (*socialplatform.SocialPlatform, error)
	Update(ctx context.Context, id uuid.UUID, p *socialplatform.SocialPlatform) (*socialplatform.SocialPlatform, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type socialPlatformService struct {
	repo   socialplatform.SocialPlatformRepository
	logger logger.Logger
}

func NewSocialPlatformService(repo socialplatform.SocialPlatformRepository, l logger.Logger) SocialPlatformService {
	return &socialPlatformService{repo: repo, logger: l}
}

func (s *socialPlatformService) List(ctx context.Context) ([]*socialplatform.SocialPlatform, error) {
	return s.repo.List(ctx)
}

func (s *socialPlatformService) ListOffset(ctx context.Context, limit, offset int) ([]*socialplatform.SocialPlatform, int64, error) {
	return s.repo.ListOffset(ctx, limit, offset)
}

func (s *socialPlatformService) Create(ctx context.Context, p *socialplatform.SocialPlatform) (*socialplatform.SocialPlatform, error) {
	if p.Key == "" {
		return nil, custom_errors.New(400, custom_errors.ErrInvalidFormat, "key ของ platform ต้องไม่ว่าง")
	}
	if p.Name == "" {
		return nil, custom_errors.New(400, custom_errors.ErrInvalidFormat, "ชื่อ platform ต้องไม่ว่าง")
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *socialPlatformService) Update(ctx context.Context, id uuid.UUID, p *socialplatform.SocialPlatform) (*socialplatform.SocialPlatform, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, custom_errors.New(404, custom_errors.ErrNotFound, "ไม่พบ platform นี้")
	}

	if p.Name != "" {
		existing.Name = p.Name
	}
	if p.IconURL != "" {
		existing.IconURL = p.IconURL
	}
	existing.SortOrder = p.SortOrder
	existing.IsActive = p.IsActive

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *socialPlatformService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
