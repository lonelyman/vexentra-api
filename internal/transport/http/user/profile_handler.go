package userhdl

import (
	"vexentra-api/internal/modules/user"
	"vexentra-api/internal/modules/user/usersvc"
	"vexentra-api/internal/transport/http/presenter"
	"vexentra-api/pkg/auth"
	"vexentra-api/pkg/custom_errors"
	"vexentra-api/pkg/logger"
	"vexentra-api/pkg/validation"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type ProfileHandler struct {
	svc              usersvc.ProfileService
	showcasePersonID string
	validate         *validator.Validate
	logger           logger.Logger
}

func NewProfileHandler(svc usersvc.ProfileService, showcasePersonID string, l logger.Logger) *ProfileHandler {
	if l == nil {
		l = logger.Get()
	}
	return &ProfileHandler{svc: svc, showcasePersonID: showcasePersonID, validate: validation.New(), logger: l}
}

// GetShowcase — GET /api/v1/showcase
// Public: returns the full profile of the pre-configured showcase person (APP_SHOWCASE_PERSON_ID).
// No login required; only published portfolio items are returned.
func (h *ProfileHandler) GetShowcase(c fiber.Ctx) error {
	if h.showcasePersonID == "" {
		return presenter.RenderError(c, custom_errors.New(404, custom_errors.ErrNotFound, "ยังไม่ได้ตั้งค่า showcase person"))
	}

	targetID, err := uuid.Parse(h.showcasePersonID)
	if err != nil {
		return presenter.RenderError(c, custom_errors.New(500, custom_errors.ErrInternal, "APP_SHOWCASE_PERSON_ID ไม่ใช่ UUID ที่ถูกต้อง"))
	}

	result, svcErr := h.svc.GetFullProfile(c.Context(), targetID, false)
	if svcErr != nil {
		return presenter.RenderError(c, svcErr)
	}

	return presenter.RenderItem(c, toFullProfileResponse(result))
}

// GetMyProfile — GET /api/v1/me/profile
// Protected: returns the full profile of the currently authenticated user.
func (h *ProfileHandler) GetMyProfile(c fiber.Ctx) error {
	personID, err := ownerPersonID(c)
	if err != nil {
		return presenter.RenderError(c, err)
	}

	result, svcErr := h.svc.GetFullProfile(c.Context(), personID, true)
	if svcErr != nil {
		return presenter.RenderError(c, svcErr)
	}

	return presenter.RenderItem(c, toFullProfileResponse(result))
}

// GetPublicProfile — GET /api/v1/users/:id/profile
// Protected: any logged-in user can view any profile.
// The :id param is the person_id. Owner sees draft portfolio items too.
func (h *ProfileHandler) GetPublicProfile(c fiber.Ctx) error {
	targetPersonID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrInvalidFormat, "person ID ไม่ถูกต้อง"))
	}

	claims := auth.GetClaims(c)
	viewerIsOwner := claims != nil && claims.GetPersonID() == targetPersonID.String()

	result, svcErr := h.svc.GetFullProfile(c.Context(), targetPersonID, viewerIsOwner)
	if svcErr != nil {
		return presenter.RenderError(c, svcErr)
	}

	return presenter.RenderItem(c, toFullProfileResponse(result))
}

// UpsertProfile — PUT /api/v1/me/profile
func (h *ProfileHandler) UpsertProfile(c fiber.Ctx) error {
	personID, err := ownerPersonID(c)
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
	}

	if svcErr := h.svc.UpsertProfile(c.Context(), personID, p); svcErr != nil {
		return presenter.RenderError(c, svcErr)
	}

	return presenter.RenderItem(c, toProfileResponse(p))
}

// AddSkill — POST /api/v1/me/skills
func (h *ProfileHandler) AddSkill(c fiber.Ctx) error {
	personID, err := ownerPersonID(c)
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

	if svcErr := h.svc.AddSkill(c.Context(), personID, s); svcErr != nil {
		return presenter.RenderError(c, svcErr)
	}

	return presenter.RenderItem(c, toSkillResponse(s), fiber.StatusCreated)
}

// AdminAddSkill — POST /api/v1/users/:id/skills (admin only)
func (h *ProfileHandler) AdminAddSkill(c fiber.Ctx) error {
	userID, parseErr := uuid.Parse(c.Params("id"))
	if parseErr != nil {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrInvalidFormat, "user ID ไม่ถูกต้อง"))
	}

	var req AddSkillRequest
	if bindErr := c.Bind().JSON(&req); bindErr != nil {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrInvalidFormat, "รูปแบบข้อมูลไม่ถูกต้อง"))
	}
	if vResult := validation.Validate(h.validate, &req); !vResult.IsValid {
		return custom_errors.New(400, custom_errors.ErrValidation, "ข้อมูลไม่ถูกต้อง", vResult.Errors)
	}

	s := &user.Skill{
		Name:        req.Name,
		Category:    req.Category,
		Proficiency: req.Proficiency,
		SortOrder:   req.SortOrder,
	}

	if svcErr := h.svc.AdminAddSkill(c.Context(), userID, s); svcErr != nil {
		return presenter.RenderError(c, svcErr)
	}

	return presenter.RenderItem(c, toSkillResponse(s), fiber.StatusCreated)
}

// RemoveSkill — DELETE /api/v1/me/skills/:skillID
func (h *ProfileHandler) RemoveSkill(c fiber.Ctx) error {
	personID, err := ownerPersonID(c)
	if err != nil {
		return presenter.RenderError(c, err)
	}

	skillID, parseErr := uuid.Parse(c.Params("skillID"))
	if parseErr != nil {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrInvalidFormat, "skill ID ไม่ถูกต้อง"))
	}

	if svcErr := h.svc.RemoveSkill(c.Context(), skillID, personID); svcErr != nil {
		return presenter.RenderError(c, svcErr)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// AddExperience — POST /api/v1/me/experiences
func (h *ProfileHandler) AddExperience(c fiber.Ctx) error {
	personID, err := ownerPersonID(c)
	if err != nil {
		return presenter.RenderError(c, err)
	}

	var req AddExperienceRequest
	if bindErr := c.Bind().JSON(&req); bindErr != nil {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrInvalidFormat, "รูปแบบข้อมูลไม่ถูกต้อง"))
	}

	e := req.ToEntity()
	if svcErr := h.svc.AddExperience(c.Context(), personID, e); svcErr != nil {
		return presenter.RenderError(c, svcErr)
	}

	return presenter.RenderItem(c, toExperienceResponse(e), fiber.StatusCreated)
}

// AdminAddExperience — POST /api/v1/users/:id/experiences (admin only)
func (h *ProfileHandler) AdminAddExperience(c fiber.Ctx) error {
	userID, parseErr := uuid.Parse(c.Params("id"))
	if parseErr != nil {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrInvalidFormat, "user ID ไม่ถูกต้อง"))
	}

	var req AddExperienceRequest
	if bindErr := c.Bind().JSON(&req); bindErr != nil {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrInvalidFormat, "รูปแบบข้อมูลไม่ถูกต้อง"))
	}
	if vResult := validation.Validate(h.validate, &req); !vResult.IsValid {
		return custom_errors.New(400, custom_errors.ErrValidation, "ข้อมูลไม่ถูกต้อง", vResult.Errors)
	}

	e := req.ToEntity()
	if svcErr := h.svc.AdminAddExperience(c.Context(), userID, e); svcErr != nil {
		return presenter.RenderError(c, svcErr)
	}

	return presenter.RenderItem(c, toExperienceResponse(e), fiber.StatusCreated)
}

// AdminUpdateExperience — PUT /api/v1/users/:id/experiences/:expID (admin only)
func (h *ProfileHandler) AdminUpdateExperience(c fiber.Ctx) error {
	userID, parseUserErr := uuid.Parse(c.Params("id"))
	if parseUserErr != nil {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrInvalidFormat, "user ID ไม่ถูกต้อง"))
	}
	expID, parseExpErr := uuid.Parse(c.Params("expID"))
	if parseExpErr != nil {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrInvalidFormat, "experience ID ไม่ถูกต้อง"))
	}

	var req AddExperienceRequest
	if bindErr := c.Bind().JSON(&req); bindErr != nil {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrInvalidFormat, "รูปแบบข้อมูลไม่ถูกต้อง"))
	}
	if vResult := validation.Validate(h.validate, &req); !vResult.IsValid {
		return custom_errors.New(400, custom_errors.ErrValidation, "ข้อมูลไม่ถูกต้อง", vResult.Errors)
	}

	e := req.ToEntity()
	if svcErr := h.svc.AdminUpdateExperience(c.Context(), userID, expID, e); svcErr != nil {
		return presenter.RenderError(c, svcErr)
	}

	return presenter.RenderItem(c, toExperienceResponse(e))
}

// AdminRemoveExperience — DELETE /api/v1/users/:id/experiences/:expID (admin only)
func (h *ProfileHandler) AdminRemoveExperience(c fiber.Ctx) error {
	userID, parseUserErr := uuid.Parse(c.Params("id"))
	if parseUserErr != nil {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrInvalidFormat, "user ID ไม่ถูกต้อง"))
	}
	expID, parseExpErr := uuid.Parse(c.Params("expID"))
	if parseExpErr != nil {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrInvalidFormat, "experience ID ไม่ถูกต้อง"))
	}

	if svcErr := h.svc.AdminRemoveExperience(c.Context(), userID, expID); svcErr != nil {
		return presenter.RenderError(c, svcErr)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// UpdateExperience — PUT /api/v1/me/experiences/:expID
func (h *ProfileHandler) UpdateExperience(c fiber.Ctx) error {
	personID, err := ownerPersonID(c)
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
	if svcErr := h.svc.UpdateExperience(c.Context(), expID, personID, e); svcErr != nil {
		return presenter.RenderError(c, svcErr)
	}

	return presenter.RenderItem(c, toExperienceResponse(e))
}

// RemoveExperience — DELETE /api/v1/me/experiences/:expID
func (h *ProfileHandler) RemoveExperience(c fiber.Ctx) error {
	personID, err := ownerPersonID(c)
	if err != nil {
		return presenter.RenderError(c, err)
	}

	expID, parseErr := uuid.Parse(c.Params("expID"))
	if parseErr != nil {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrInvalidFormat, "experience ID ไม่ถูกต้อง"))
	}

	if svcErr := h.svc.RemoveExperience(c.Context(), expID, personID); svcErr != nil {
		return presenter.RenderError(c, svcErr)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// AddPortfolioItem — POST /api/v1/me/portfolio
func (h *ProfileHandler) AddPortfolioItem(c fiber.Ctx) error {
	personID, err := ownerPersonID(c)
	if err != nil {
		return presenter.RenderError(c, err)
	}

	var req AddPortfolioItemRequest
	if bindErr := c.Bind().JSON(&req); bindErr != nil {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrInvalidFormat, "รูปแบบข้อมูลไม่ถูกต้อง"))
	}
	if vResult := validation.Validate(h.validate, &req); !vResult.IsValid {
		return custom_errors.New(400, custom_errors.ErrValidation, "ข้อมูลไม่ถูกต้อง", vResult.Errors)
	}

	item := req.ToEntity()
	if svcErr := h.svc.AddPortfolioItem(c.Context(), personID, item, req.Tags); svcErr != nil {
		return presenter.RenderError(c, svcErr)
	}

	return presenter.RenderItem(c, toPortfolioItemResponse(item), fiber.StatusCreated)
}

// AdminAddPortfolioItem — POST /api/v1/users/:id/portfolio (admin only)
func (h *ProfileHandler) AdminAddPortfolioItem(c fiber.Ctx) error {
	userID, parseErr := uuid.Parse(c.Params("id"))
	if parseErr != nil {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrInvalidFormat, "user ID ไม่ถูกต้อง"))
	}

	var req AddPortfolioItemRequest
	if bindErr := c.Bind().JSON(&req); bindErr != nil {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrInvalidFormat, "รูปแบบข้อมูลไม่ถูกต้อง"))
	}
	if vResult := validation.Validate(h.validate, &req); !vResult.IsValid {
		return custom_errors.New(400, custom_errors.ErrValidation, "ข้อมูลไม่ถูกต้อง", vResult.Errors)
	}

	item := req.ToEntity()
	if svcErr := h.svc.AdminAddPortfolioItem(c.Context(), userID, item, req.Tags); svcErr != nil {
		return presenter.RenderError(c, svcErr)
	}

	return presenter.RenderItem(c, toPortfolioItemResponse(item), fiber.StatusCreated)
}

// UpdatePortfolioItem — PUT /api/v1/me/portfolio/:itemID
func (h *ProfileHandler) UpdatePortfolioItem(c fiber.Ctx) error {
	personID, err := ownerPersonID(c)
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
	if vResult := validation.Validate(h.validate, &req); !vResult.IsValid {
		return custom_errors.New(400, custom_errors.ErrValidation, "ข้อมูลไม่ถูกต้อง", vResult.Errors)
	}

	item := req.ToEntity()
	if svcErr := h.svc.UpdatePortfolioItem(c.Context(), itemID, personID, item, req.Tags); svcErr != nil {
		return presenter.RenderError(c, svcErr)
	}

	return presenter.RenderItem(c, toPortfolioItemResponse(item))
}

// RemovePortfolioItem — DELETE /api/v1/me/portfolio/:itemID
func (h *ProfileHandler) RemovePortfolioItem(c fiber.Ctx) error {
	personID, err := ownerPersonID(c)
	if err != nil {
		return presenter.RenderError(c, err)
	}

	itemID, parseErr := uuid.Parse(c.Params("itemID"))
	if parseErr != nil {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrInvalidFormat, "portfolio item ID ไม่ถูกต้อง"))
	}

	if svcErr := h.svc.RemovePortfolioItem(c.Context(), itemID, personID); svcErr != nil {
		return presenter.RenderError(c, svcErr)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// UpsertSocialLink — PUT /api/v1/me/social-links/:platformID
func (h *ProfileHandler) UpsertSocialLink(c fiber.Ctx) error {
	personID, err := ownerPersonID(c)
	if err != nil {
		return presenter.RenderError(c, err)
	}

	platformID, parseErr := uuid.Parse(c.Params("platformID"))
	if parseErr != nil {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrInvalidFormat, "platform ID ไม่ถูกต้อง"))
	}

	var req UpsertSocialLinkRequest
	if bindErr := c.Bind().JSON(&req); bindErr != nil {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrInvalidFormat, "รูปแบบข้อมูลไม่ถูกต้อง"))
	}
	if vResult := validation.Validate(h.validate, &req); !vResult.IsValid {
		return custom_errors.New(400, custom_errors.ErrValidation, "ข้อมูลไม่ถูกต้อง", vResult.Errors)
	}

	l, svcErr := h.svc.UpsertSocialLink(c.Context(), personID, platformID, req.URL, req.SortOrder)
	if svcErr != nil {
		return presenter.RenderError(c, svcErr)
	}

	return presenter.RenderItem(c, SocialLinkResponse{
		ID:         l.ID.String(),
		PlatformID: l.PlatformID.String(),
		URL:        l.URL,
		SortOrder:  l.SortOrder,
	})
}

// DeleteSocialLink — DELETE /api/v1/me/social-links/:linkID
func (h *ProfileHandler) DeleteSocialLink(c fiber.Ctx) error {
	personID, err := ownerPersonID(c)
	if err != nil {
		return presenter.RenderError(c, err)
	}

	linkID, parseErr := uuid.Parse(c.Params("linkID"))
	if parseErr != nil {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrInvalidFormat, "social link ID ไม่ถูกต้อง"))
	}

	if svcErr := h.svc.DeleteSocialLink(c.Context(), linkID, personID); svcErr != nil {
		return presenter.RenderError(c, svcErr)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// GetSocialLinks — GET /api/v1/me/social-links
func (h *ProfileHandler) GetSocialLinks(c fiber.Ctx) error {
	personID, appErr := ownerPersonID(c)
	if appErr != nil {
		return presenter.RenderError(c, appErr)
	}

	links, svcErr := h.svc.ListSocialLinks(c.Context(), personID)
	if svcErr != nil {
		return presenter.RenderError(c, svcErr)
	}

	resp := make([]SocialLinkResponse, len(links))
	for i, l := range links {
		resp[i] = SocialLinkResponse{
			ID:         l.ID.String(),
			PlatformID: l.PlatformID.String(),
			URL:        l.URL,
			SortOrder:  l.SortOrder,
		}
	}
	return presenter.RenderList(c, resp)
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

// ownerPersonID extracts the authenticated user's Person UUID from the request context.
// Used for all profile/portfolio/skill/experience operations (which key on person_id).
func ownerPersonID(c fiber.Ctx) (uuid.UUID, *custom_errors.AppError) {
	claims := auth.GetClaims(c)
	if claims == nil {
		return uuid.Nil, custom_errors.New(401, custom_errors.ErrUnauthorized, "กรุณาเข้าสู่ระบบ")
	}
	id, err := uuid.Parse(claims.GetPersonID())
	if err != nil {
		return uuid.Nil, custom_errors.New(401, custom_errors.ErrUnauthorized, "Token มี PersonID ไม่ถูกต้อง")
	}
	return id, nil
}

func (h *ProfileHandler) AdminUpsertProfile(c fiber.Ctx) error {
	targetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return presenter.RenderError(c, custom_errors.New(400, "INVALID_ID", "รูปแบบ User ID ไม่ถูกต้อง"))
	}
	req := new(UpsertProfileRequest)
	if bindErr := c.Bind().Body(req); bindErr != nil {
		return presenter.RenderError(c, custom_errors.New(400, "INVALID_JSON", "รูปแบบ JSON ไม่ถูกต้อง"))
	}
	if vResult := validation.Validate(h.validate, req); !vResult.IsValid {
		return presenter.RenderError(c, custom_errors.New(400, custom_errors.ErrValidation, "ข้อมูลไม่ถูกต้อง", vResult.Errors))
	}
	p := &user.Profile{
		DisplayName: req.DisplayName,
		Headline:    req.Headline,
		Bio:         req.Bio,
		Location:    req.Location,
		AvatarURL:   req.AvatarURL,
	}
	if svcErr := h.svc.AdminUpsertProfile(c.Context(), targetID, p); svcErr != nil {
		return presenter.RenderError(c, svcErr)
	}
	h.logger.Info("Admin updated profile", "targetUserID", targetID)
	return presenter.RenderItem(c, fiber.Map{"message": "อัปเดตโปรไฟล์สำเร็จ"})
}
