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
      SELECT a.id, a.student_id, COALESCE(a.student_number, ''), COALESCE(a.name_surname, ''), COALESCE(a.fio, ''), a.birth_date,
             COALESCE(a.iin, ''), COALESCE(a.school, ''), COALESCE(a.level, ''), COALESCE(a.comments, ''),
             a.year, a.major, a.gender, COALESCE(a.room_preference, ''), COALESCE(a.additional_info, ''),
             COALESCE(a.applicant_type, 'local'), COALESCE(a.passport_number, ''), a.status,
             latest_payment.status, latest_payment.paid_at,
             a.submitted_at, a.updated_at, a.rejected_reason, a.reviewed_by, a.review_timestamp
      FROM applications a
      LEFT JOIN LATERAL (
          SELECT p.status, p.paid_at
          FROM payments p
          WHERE p.application_id = a.id
          ORDER BY p.created_at DESC, p.id DESC
          LIMIT 1
      ) latest_payment ON TRUE
   `

	var conditions []string
	var args []any
	argIndex := 1

	if status := strings.TrimSpace(filters.Status); status != "" {
		conditions = append(conditions, fmt.Sprintf("a.status = $%d", argIndex))
		args = append(args, status)
		argIndex++
	}

	if filters.Year != nil {
		conditions = append(conditions, fmt.Sprintf("a.year = $%d", argIndex))
		args = append(args, *filters.Year)
		argIndex++
	}

	if gender := strings.TrimSpace(filters.Gender); gender != "" {
		conditions = append(conditions, fmt.Sprintf("LOWER(a.gender) = LOWER($%d)", argIndex))
		args = append(args, gender)
		argIndex++
	}

	if major := strings.TrimSpace(filters.Major); major != "" {
		conditions = append(conditions, fmt.Sprintf("a.major ILIKE $%d", argIndex))
		args = append(args, "%"+major+"%")
		argIndex++
	}

	if search := strings.TrimSpace(filters.Search); search != "" {
		conditions = append(conditions, fmt.Sprintf(`(
			CAST(a.student_id AS TEXT) ILIKE $%d OR
			CAST(a.id AS TEXT) ILIKE $%d OR
			COALESCE(a.fio, '') ILIKE $%d OR
			COALESCE(a.name_surname, '') ILIKE $%d OR
			COALESCE(a.student_number, '') ILIKE $%d OR
			COALESCE(a.additional_info, '') ILIKE $%d OR
			a.major ILIKE $%d
		)`, argIndex, argIndex, argIndex, argIndex, argIndex, argIndex, argIndex))
		args = append(args, "%"+search+"%")
		argIndex++
	}

	query := baseQuery
	if len(conditions) > 0 {
		query += "\nWHERE " + strings.Join(conditions, "\n  AND ")
	}
	query += "\nORDER BY a.submitted_at DESC"

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
			&a.StudentNumber,
			&a.NameSurname,
			&a.FIO,
			&a.BirthDate,
			&a.IIN,
			&a.School,
			&a.Level,
			&a.Comments,
			&a.Year,
			&a.Major,
			&a.Gender,
			&a.RoomPreference,
			&a.AdditionalInfo,
			&a.ApplicantType,
			&a.PassportNumber,
			&a.Status,
			&a.PaymentStatus,
			&a.PaidAt,
			&a.SubmittedAt,
			&a.UpdatedAt,
			&a.RejectedReason,
			&a.ReviewedBy,
			&a.ReviewTimestamp,
		); err != nil {
			return nil, err
		}
		normalizeApplicationIdentity(&a)
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
