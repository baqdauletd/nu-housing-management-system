package database

import (
	"database/sql"
	"nu-housing-management-system/backend/internal/models"
)

func HousingListApplications(db *sql.DB) ([]models.Application, error) {
	query := `
      SELECT id, student_id, COALESCE(fio, ''), year, major, gender, COALESCE(room_preference, ''), COALESCE(additional_info, ''),
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
			&a.FIO,
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
		normalizeApplicationFIO(&a)
		apps = append(apps, a)
	}

	return apps, nil
}

func HousingGetApplication(db *sql.DB, id int) (models.Application, error) {
	return GetApplicationByID(db, id)
}

func HousingApprove(db *sql.DB, id int, reviewerID int) error {
	err := UpdateApplicationStatus(db, id, "approved", reviewerID)
	if err != nil {
		return err
	}
	_, _ = db.Exec(`UPDATE applications SET rejected_reason = NULL WHERE id = $1`, id)
	db.Exec(`INSERT INTO audit_logs (actor_id, action, entity, entity_id) VALUES ($1, 'approve', 'application', $2)`, reviewerID, id)
	return nil
}

func HousingReject(db *sql.DB, id int, reason string, reviewerID int) error {
	err := RejectApplication(db, id, reason, reviewerID)
	if err != nil {
		return err
	}
	db.Exec(`INSERT INTO audit_logs (actor_id, action, entity, entity_id) VALUES ($1, 'reject', 'application', $2)`, reviewerID, id)
	return nil
}
