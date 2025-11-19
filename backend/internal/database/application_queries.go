package database

import (
	"database/sql"
	"errors"
	"nu-housing-management-system/backend/internal/models"
)

////////////////////////////////////////////////////////////
// APPLICATION QUERIES (UPDATED)
////////////////////////////////////////////////////////////

func SubmitApplication(db *sql.DB, a models.Application) (int, error) {
	query := `
		INSERT INTO applications 
		(student_id, year, major, gender, room_preference, additional_info, status, submitted_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, 'pending', NOW(), NOW())
		RETURNING id
	`

	var id int
	err := db.QueryRow(query, a.StudentID, a.Year, a.Major, a.Gender, a.RoomPreference, a.AdditionalInfo).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func GetApplicationByID(db *sql.DB, id int) (models.Application, error) {
	var a models.Application

	query := `
		SELECT id, student_id, year, major, gender, room_preference, additional_info,
		       status, submitted_at, updated_at, rejected_reason, reviewed_by, review_timestamp
		FROM applications
		WHERE id = $1
	`

	err := db.QueryRow(query, id).Scan(
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
	)

	if err == sql.ErrNoRows {
		return a, errors.New("application not found")
	}
	return a, err
}

func GetApplicationsByStudent(db *sql.DB, studentID int) ([]models.Application, error) {
	query := `
		SELECT id, student_id, year, major, gender, room_preference, additional_info,
		       status, submitted_at, updated_at, rejected_reason, reviewed_by, review_timestamp
		FROM applications
		WHERE student_id = $1
		ORDER BY submitted_at DESC
	`

	rows, err := db.Query(query, studentID)
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

func UpdateApplicationStatus(db *sql.DB, id int, status string, reviewerID int) error {
	query := `
		UPDATE applications
		SET status = $1,
		    reviewed_by = $2,
		    review_timestamp = NOW(),
		    updated_at = NOW()
		WHERE id = $3
	`
	_, err := db.Exec(query, status, reviewerID, id)
	return err
}

func RejectApplication(db *sql.DB, id int, reason string, reviewerID int) error {
	query := `
		UPDATE applications
		SET status = 'rejected',
		    rejected_reason = $1,
		    reviewed_by = $2,
		    review_timestamp = NOW(),
		    updated_at = NOW()
		WHERE id = $3
	`
	_, err := db.Exec(query, reason, reviewerID, id)
	return err
}
