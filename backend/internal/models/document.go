package models

import "time"

type Document struct {
	ID            int       `json:"id"`
	ApplicationID int       `json:"application_id"`
	Type          string    `json:"type"`
	FileURL       string    `json:"file_url"`
	UploadedAt    time.Time `json:"uploaded_at"`
}
