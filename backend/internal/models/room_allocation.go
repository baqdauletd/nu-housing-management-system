package models

import "time"

type RoomAllocation struct {
	ID            int       `json:"id"`
	ApplicationID int       `json:"application_id"`
	StudentID     int       `json:"student_id"`
	Block         int       `json:"block"`
	RoomNumber    int       `json:"room_number"`
	BedNumber     int       `json:"bed_number"`
	CreatedAt     time.Time `json:"created_at"`
}
