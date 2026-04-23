package filehdl

type PresignUploadRequest struct {
	Intent         string `json:"intent" validate:"required,oneof=profile_image"`
	Filename       string `json:"filename" validate:"required"`
	MIMEType       string `json:"mime_type" validate:"required"`
	SizeBytes      int64  `json:"size_bytes" validate:"required,gt=0"`
	TargetPersonID string `json:"target_person_id" validate:"omitempty,uuid"`
}

type CompleteUploadRequest struct {
	UploadSessionID string `json:"upload_session_id" validate:"required,uuid"`
}
