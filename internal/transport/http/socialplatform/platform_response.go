package socialplatformhdl

import "vexentra-api/internal/modules/socialplatform"

// SocialPlatformResponse is the serialized form of a social platform record.
type SocialPlatformResponse struct {
	ID        string `json:"id"`
	Key       string `json:"key"`
	Name      string `json:"name"`
	IconURL   string `json:"icon_url"`
	SortOrder int    `json:"sort_order"`
	IsActive  bool   `json:"is_active"`
}

func toResponse(p *socialplatform.SocialPlatform) SocialPlatformResponse {
	return SocialPlatformResponse{
		ID:        p.ID.String(),
		Key:       p.Key,
		Name:      p.Name,
		IconURL:   p.IconURL,
		SortOrder: p.SortOrder,
		IsActive:  p.IsActive,
	}
}
