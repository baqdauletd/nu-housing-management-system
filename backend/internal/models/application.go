package models

import "time"

type Application struct {
	ID              int        `json:"id"`
	StudentID       int        `json:"student_id"`
	Year            int        `json:"year"`
	Major           string     `json:"major"`
	Gender          string     `json:"gender"`
	RoomPreference  string     `json:"room_preference,omitempty"`
	AdditionalInfo  string     `json:"additional_info,omitempty"`
	Status          string     `json:"status"`
	SubmittedAt     time.Time  `json:"submitted_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	RejectedReason  *string    `json:"rejected_reason,omitempty"`
	ReviewedBy      *int       `json:"reviewed_by,omitempty"`
	ReviewTimestamp *time.Time `json:"review_timestamp,omitempty"`
}
