package userhdl

import (
	"vexentra-api/internal/modules/user"
	"vexentra-api/internal/modules/user/usersvc"
	"vexentra-api/internal/transport/http/presenter"
	"vexentra-api/pkg/auth"
	"vexentra-api/pkg/custom_errors"
	"vexentra-api/pkg/logger"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type ProfileHandler struct {
	svc            usersvc.ProfileService
	showcaseUserID string
	logger         logger.Logger
}

func NewProfileHandler(svc usersvc.ProfileService, showcaseUserID string, l logger.Logger) *ProfileHandler {
	if l == nil {
		l = logger.Get()
	}
	return &ProfileHandler{svc: svc, showcaseUserID: showcaseUserID, logger: l}
}

// GetShowcase — GET /api/v1/showcase
// Public: returns the full profile of the pre-configured showcase user (APP_SHOWCASE_USER_ID).
// No login required; only published portfolio items are returned.
func (h *ProfileHandler) GetShowcase(c fiber.Ctx) error {
	if h.showcaseUserID == "" {
		return presenter.RenderError(c, custom_errors.New(404, custom_errors.ErrNotFound, "ยังไม่ได้ตั้งค่า showcase user"))
	}

	targetID, err := uuid.Parse(h.showcaseUserID)
	if err != nil {
		return presenter.RenderError(c, custom_errors.New(500, custom_errors.ErrInternal, "APP_SHOWCASE_USER_ID ไม่ใช่ UUID ที่ถูกต้อง"))
	}

	result, svcErr := h.svc.GetFullProfile(c.Context(), targetID, false)
	if svcErr != nil {
		return presenter.RenderError(c, svcErr)
	}

	return presenter.RenderItem(c, toFullProfileResponse(result))
}

// GetPublicProfile — GET /api/v1/users/:id/profile
// Protected: any logged-in user can view any profile.
// The owner additionally sees draft portfolio items.
func (h *ProfileHandler) GetPublicProfile(c fiber.Ctx) error {
	targetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrInvalidFormat, "user ID ไม่ถูกต้อง"))
	}

	claims := auth.GetClaims(c)
	viewerIsOwner := claims != nil && claims.GetUserID() == targetID.String()

	result, svcErr := h.svc.GetFullProfile(c.Context(), targetID, viewerIsOwner)
	if svcErr != nil {
		return presenter.RenderError(c, svcErr)
	}

	return presenter.RenderItem(c, toFullProfileResponse(result))
}

// UpsertProfile — PUT /api/v1/me/profile
func (h *ProfileHandler) UpsertProfile(c fiber.Ctx) error {
	userID, err := ownerID(c)
	if err != nil {
		return presenter.RenderError(c, err)
	}

	var req UpsertProfileRequest
	if bindErr := c.Bind().JSON(&req); bindErr != nil {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrInvalidFormat, "รูปแบบข้อมูลไม่ถูกต้อง"))
	}

	p := &user.Profile{
		DisplayName: req.DisplayName,
		Headline:    req.Headline,
		Bio:         req.Bio,
		Location:    req.Location,
		AvatarURL:   req.AvatarURL,
		WebsiteURL:  req.WebsiteURL,
		GitHubURL:   req.GitHubURL,
		LinkedInURL: req.LinkedInURL,
		TwitterURL:  req.TwitterURL,
	}

	if svcErr := h.svc.UpsertProfile(c.Context(), userID, p); svcErr != nil {
		return presenter.RenderError(c, svcErr)
	}

	return presenter.RenderItem(c, toProfileResponse(p))
}

// AddSkill — POST /api/v1/me/skills
func (h *ProfileHandler) AddSkill(c fiber.Ctx) error {
	userID, err := ownerID(c)
	if err != nil {
		return presenter.RenderError(c, err)
	}

	var req AddSkillRequest
	if bindErr := c.Bind().JSON(&req); bindErr != nil {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrInvalidFormat, "รูปแบบข้อมูลไม่ถูกต้อง"))
	}

	s := &user.Skill{
		Name:        req.Name,
		Category:    req.Category,
		Proficiency: req.Proficiency,
		SortOrder:   req.SortOrder,
	}

	if svcErr := h.svc.AddSkill(c.Context(), userID, s); svcErr != nil {
		return presenter.RenderError(c, svcErr)
	}

	return presenter.RenderItem(c, toSkillResponse(s), fiber.StatusCreated)
}

// RemoveSkill — DELETE /api/v1/me/skills/:skillID
func (h *ProfileHandler) RemoveSkill(c fiber.Ctx) error {
	userID, err := ownerID(c)
	if err != nil {
		return presenter.RenderError(c, err)
	}

	skillID, parseErr := uuid.Parse(c.Params("skillID"))
	if parseErr != nil {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrInvalidFormat, "skill ID ไม่ถูกต้อง"))
	}

	if svcErr := h.svc.RemoveSkill(c.Context(), skillID, userID); svcErr != nil {
		return presenter.RenderError(c, svcErr)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// AddExperience — POST /api/v1/me/experiences
func (h *ProfileHandler) AddExperience(c fiber.Ctx) error {
	userID, err := ownerID(c)
	if err != nil {
		return presenter.RenderError(c, err)
	}

	var req AddExperienceRequest
	if bindErr := c.Bind().JSON(&req); bindErr != nil {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrInvalidFormat, "รูปแบบข้อมูลไม่ถูกต้อง"))
	}

	e := req.ToEntity()
	if svcErr := h.svc.AddExperience(c.Context(), userID, e); svcErr != nil {
		return presenter.RenderError(c, svcErr)
	}

	return presenter.RenderItem(c, toExperienceResponse(e), fiber.StatusCreated)
}

// UpdateExperience — PUT /api/v1/me/experiences/:expID
func (h *ProfileHandler) UpdateExperience(c fiber.Ctx) error {
	userID, err := ownerID(c)
	if err != nil {
		return presenter.RenderError(c, err)
	}

	expID, parseErr := uuid.Parse(c.Params("expID"))
	if parseErr != nil {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrInvalidFormat, "experience ID ไม่ถูกต้อง"))
	}

	var req AddExperienceRequest
	if bindErr := c.Bind().JSON(&req); bindErr != nil {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrInvalidFormat, "รูปแบบข้อมูลไม่ถูกต้อง"))
	}

	e := req.ToEntity()
	if svcErr := h.svc.UpdateExperience(c.Context(), expID, userID, e); svcErr != nil {
		return presenter.RenderError(c, svcErr)
	}

	return presenter.RenderItem(c, toExperienceResponse(e))
}

// RemoveExperience — DELETE /api/v1/me/experiences/:expID
func (h *ProfileHandler) RemoveExperience(c fiber.Ctx) error {
	userID, err := ownerID(c)
	if err != nil {
		return presenter.RenderError(c, err)
	}

	expID, parseErr := uuid.Parse(c.Params("expID"))
	if parseErr != nil {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrInvalidFormat, "experience ID ไม่ถูกต้อง"))
	}

	if svcErr := h.svc.RemoveExperience(c.Context(), expID, userID); svcErr != nil {
		return presenter.RenderError(c, svcErr)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// AddPortfolioItem — POST /api/v1/me/portfolio
func (h *ProfileHandler) AddPortfolioItem(c fiber.Ctx) error {
	userID, err := ownerID(c)
	if err != nil {
		return presenter.RenderError(c, err)
	}

	var req AddPortfolioItemRequest
	if bindErr := c.Bind().JSON(&req); bindErr != nil {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrInvalidFormat, "รูปแบบข้อมูลไม่ถูกต้อง"))
	}

	item := req.ToEntity()
	if svcErr := h.svc.AddPortfolioItem(c.Context(), userID, item, req.Tags); svcErr != nil {
		return presenter.RenderError(c, svcErr)
	}

	return presenter.RenderItem(c, toPortfolioItemResponse(item), fiber.StatusCreated)
}

// UpdatePortfolioItem — PUT /api/v1/me/portfolio/:itemID
func (h *ProfileHandler) UpdatePortfolioItem(c fiber.Ctx) error {
	userID, err := ownerID(c)
	if err != nil {
		return presenter.RenderError(c, err)
	}

	itemID, parseErr := uuid.Parse(c.Params("itemID"))
	if parseErr != nil {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrInvalidFormat, "portfolio item ID ไม่ถูกต้อง"))
	}

	var req AddPortfolioItemRequest
	if bindErr := c.Bind().JSON(&req); bindErr != nil {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrInvalidFormat, "รูปแบบข้อมูลไม่ถูกต้อง"))
	}

	item := req.ToEntity()
	if svcErr := h.svc.UpdatePortfolioItem(c.Context(), itemID, userID, item, req.Tags); svcErr != nil {
		return presenter.RenderError(c, svcErr)
	}

	return presenter.RenderItem(c, toPortfolioItemResponse(item))
}

// RemovePortfolioItem — DELETE /api/v1/me/portfolio/:itemID
func (h *ProfileHandler) RemovePortfolioItem(c fiber.Ctx) error {
	userID, err := ownerID(c)
	if err != nil {
		return presenter.RenderError(c, err)
	}

	itemID, parseErr := uuid.Parse(c.Params("itemID"))
	if parseErr != nil {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrInvalidFormat, "portfolio item ID ไม่ถูกต้อง"))
	}

	if svcErr := h.svc.RemovePortfolioItem(c.Context(), itemID, userID); svcErr != nil {
		return presenter.RenderError(c, svcErr)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ownerID extracts the authenticated user's UUID from the request context.
// Returns an AppError if the token is missing or malformed (should not happen on protected routes).
func ownerID(c fiber.Ctx) (uuid.UUID, *custom_errors.AppError) {
	claims := auth.GetClaims(c)
	if claims == nil {
		return uuid.Nil, custom_errors.New(401, custom_errors.ErrUnauthorized, "กรุณาเข้าสู่ระบบ")
	}
	id, err := uuid.Parse(claims.GetUserID())
	if err != nil {
		return uuid.Nil, custom_errors.New(401, custom_errors.ErrUnauthorized, "Token มี UserID ไม่ถูกต้อง")
	}
	return id, nil
}
