package models

import "time"

type DormInventoryRoommate struct {
	ApplicationID int    `json:"application_id"`
	StudentID     int    `json:"student_id"`
	NuID          string `json:"nu_id"`
	FIO           string `json:"fio"`
	Email         string `json:"email"`
	BedNumber     int    `json:"bed_number"`
}

type DormInventoryRow struct {
	ApplicationID  int                    `json:"application_id"`
	StudentID      int                    `json:"student_id"`
	NuID           string                 `json:"nu_id"`
	FIO            string                 `json:"fio"`
	Email          string                 `json:"email"`
	Gender         string                 `json:"gender"`
	Major          string                 `json:"major"`
	Block          int                    `json:"block"`
	RoomNumber     int                    `json:"room_number"`
	BedNumber      int                    `json:"bed_number"`
	Capacity       int                    `json:"capacity"`
	OccupancyCount int                    `json:"occupancy_count"`
	FromWhen       time.Time              `json:"from_when"`
	TillWhen       *time.Time             `json:"till_when,omitempty"`
	Roommates      []DormInventoryRoommate `json:"roommates"`
}
