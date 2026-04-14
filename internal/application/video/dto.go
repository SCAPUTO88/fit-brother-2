package video

type RequestUploadURLInput struct {
	ExerciseID string `json:"exercise_id" binding:"required,uuid"`
	Filename   string `json:"filename" binding:"required"`
	MimeType   string `json:"mime_type" binding:"required"`
}

type RequestUploadURLOutput struct {
	PresignedURL string `json:"presigned_url"`
	R2Key        string `json:"r2_key"`
	ExpiresIn    int    `json:"expires_in_seconds"`
}

type ConfirmUploadInput struct {
	ExerciseID string `json:"exercise_id" binding:"required,uuid"`
	R2Key      string `json:"r2_key" binding:"required"`
	Filename   string `json:"filename" binding:"required"`
	SizeBytes  int64  `json:"size_bytes"`
	MimeType   string `json:"mime_type"`
}

type ConfirmUploadOutput struct {
	VideoID string `json:"video_id"`
}

type GetSignedURLOutput struct {
	SignedURL string `json:"signed_url"`
	ExpiresIn int    `json:"expires_in_seconds"`
}
