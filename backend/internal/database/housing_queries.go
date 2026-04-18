package database

import (
	"database/sql"
	"fmt"
	"nu-housing-management-system/backend/internal/models"
	"strings"
)

type HousingApplicationFilters struct {
	Status string
	Year   *int
	Gender string
	Major  string
	Search string
}

func HousingListApplications(db *sql.DB, filters HousingApplicationFilters) ([]models.Application, error) {
	baseQuery := `
      SELECT id, student_id, COALESCE(fio, ''), year, major, gender, COALESCE(room_preference, ''), COALESCE(additional_info, ''),
             status, submitted_at, updated_at, rejected_reason, reviewed_by, review_timestamp
      FROM applications
   `

	var conditions []string
	var args []any
	argIndex := 1

	if status := strings.TrimSpace(filters.Status); status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, status)
		argIndex++
	}

	if filters.Year != nil {
		conditions = append(conditions, fmt.Sprintf("year = $%d", argIndex))
		args = append(args, *filters.Year)
		argIndex++
	}

	if gender := strings.TrimSpace(filters.Gender); gender != "" {
		conditions = append(conditions, fmt.Sprintf("LOWER(gender) = LOWER($%d)", argIndex))
		args = append(args, gender)
		argIndex++
	}

	if major := strings.TrimSpace(filters.Major); major != "" {
		conditions = append(conditions, fmt.Sprintf("major ILIKE $%d", argIndex))
		args = append(args, "%"+major+"%")
		argIndex++
	}

	if search := strings.TrimSpace(filters.Search); search != "" {
		conditions = append(conditions, fmt.Sprintf(`(
			CAST(student_id AS TEXT) ILIKE $%d OR
			CAST(id AS TEXT) ILIKE $%d OR
			COALESCE(fio, '') ILIKE $%d OR
			COALESCE(additional_info, '') ILIKE $%d OR
			major ILIKE $%d
		)`, argIndex, argIndex, argIndex, argIndex, argIndex))
		args = append(args, "%"+search+"%")
		argIndex++
	}

	query := baseQuery
	if len(conditions) > 0 {
		query += "\nWHERE " + strings.Join(conditions, "\n  AND ")
	}
	query += "\nORDER BY submitted_at DESC"

	rows, err := db.Query(query, args...)
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
