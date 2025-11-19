package database

import (
	"database/sql"
	"nu-housing-management-system/backend/internal/models"
)

////////////////////////////////////////////////////////////
// HOUSING STAFF QUERIES (UPDATED)
////////////////////////////////////////////////////////////

func HousingListApplications(db *sql.DB) ([]models.Application, error) {
	query := `
		SELECT id, student_id, year, major, gender, room_preference, additional_info,
		       status, submitted_at, updated_at, rejected_reason, reviewed_by, review_timestamp
		FROM applications
		ORDER BY submitted_at DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []models.Application

	for rows.Next() {
		var a models.Application
		if err := rows.Scan(
			&a.ID,
			&a.StudentID,
			&a.Year,
			&a.Major,
			&a.Gender,
			&a.RoomPreference,
			&a.AdditionalInfo,
			&a.Status,
			&a.SubmittedAt,
			&a.UpdatedAt,
			&a.RejectedReason,
			&a.ReviewedBy,
			&a.ReviewTimestamp,
		); err != nil {
			return nil, err
		}
		apps = append(apps, a)
	}

	return apps, nil
}

func HousingGetApplication(db *sql.DB, id int) (models.Application, error) {
	return GetApplicationByID(db, id)
}

func HousingApprove(db *sql.DB, id int, reviewerID int) error {
	return UpdateApplicationStatus(db, id, "approved", reviewerID)
}

func HousingReject(db *sql.DB, id int, reason string, reviewerID int) error {
	return RejectApplication(db, id, reason, reviewerID)
}
