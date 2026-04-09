package socialplatformhdl

// CreateSocialPlatformRequest is the body for POST /api/v1/social-platforms.
type CreateSocialPlatformRequest struct {
	Key       string `json:"key"  validate:"required"`
	Name      string `json:"name" validate:"required"`
	IconURL   string `json:"icon_url"`
	SortOrder int    `json:"sort_order"`
	IsActive  bool   `json:"is_active"`
}

// UpdateSocialPlatformRequest is the body for PUT /api/v1/social-platforms/:id.
type UpdateSocialPlatformRequest struct {
	Name      string `json:"name" validate:"required"`
	IconURL   string `json:"icon_url"`
	SortOrder int    `json:"sort_order"`
	IsActive  *bool  `json:"is_active"` // pointer — allows explicit false
}
