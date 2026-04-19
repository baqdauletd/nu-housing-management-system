package models

import "time"

type Application struct {
	ID              int        `json:"id"`
	StudentID       int        `json:"student_id"`
	ApplicantType   string     `json:"applicant_type,omitempty"`
	StudentNumber   string     `json:"student_number,omitempty"`
	NameSurname     string     `json:"name_surname,omitempty"`
	FIO             string     `json:"fio,omitempty"`
	BirthDate       *time.Time `json:"birth_date,omitempty"`
	IIN             string     `json:"iin,omitempty"`
	School          string     `json:"school,omitempty"`
	Level           string     `json:"level,omitempty"`
	PassportNumber  string     `json:"passport_number,omitempty"`
	Comments        string     `json:"comments,omitempty"`
	Year            int        `json:"year"`
	Major           string     `json:"major"`
	Gender          string     `json:"gender"`
	RoomPreference  string     `json:"room_preference,omitempty"`
	AdditionalInfo  string     `json:"additional_info,omitempty"`
	Status          string     `json:"status"`
	PaymentStatus   *string    `json:"payment_status,omitempty"`
	PaidAt          *time.Time `json:"paid_at,omitempty"`
	SubmittedAt     time.Time  `json:"submitted_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	RejectedReason  *string    `json:"rejected_reason,omitempty"`
	DecisionReason  *string    `json:"decision_reason,omitempty"`
	ReviewedBy      *int       `json:"reviewed_by,omitempty"`
	ReviewTimestamp *time.Time `json:"review_timestamp,omitempty"`
}
