package models

import "time"

type User struct {
	ID           int       `json:"id"`
	NuID         string    `json:"nu_id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash,omitempty"`
	RoleID       int       `json:"role_id"`
	Phone        string    `json:"phone,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Helper field — not stored in DB, filled from JOIN or manual lookup
	RoleName string `json:"role_name,omitempty"`
}
