package database

import (
	"database/sql"
	"errors"
	"nu-housing-management-system/backend/internal/models"
	"time"
)

func ensureApplicationIdentitySchema(db *sql.DB) error {
	statements := []string{
		`ALTER TABLE applications ADD COLUMN IF NOT EXISTS applicant_type VARCHAR(40)`,
		`ALTER TABLE applications ADD COLUMN IF NOT EXISTS student_number VARCHAR(40)`,
		`ALTER TABLE applications ADD COLUMN IF NOT EXISTS name_surname VARCHAR(255)`,
		`ALTER TABLE applications ADD COLUMN IF NOT EXISTS fio VARCHAR(255)`,
		`ALTER TABLE applications ADD COLUMN IF NOT EXISTS birth_date DATE`,
		`ALTER TABLE applications ADD COLUMN IF NOT EXISTS iin VARCHAR(20)`,
		`ALTER TABLE applications ADD COLUMN IF NOT EXISTS school VARCHAR(255)`,
		`ALTER TABLE applications ADD COLUMN IF NOT EXISTS level VARCHAR(80)`,
		`ALTER TABLE applications ADD COLUMN IF NOT EXISTS passport_number VARCHAR(80)`,
		`ALTER TABLE applications ADD COLUMN IF NOT EXISTS comments TEXT`,
	}

	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			return err
		}
	}
	return nil
}

func EnsureRuntimeSchema(db *sql.DB) error {
	if err := ensureApplicationIdentitySchema(db); err != nil {
		return err
	}
	return ensureRoomAllocationSchema(db)
}

func SubmitApplication(db *sql.DB, a models.Application) (int, error) {
	settings, err := GetSystemSettings(db)
	if err != nil {
		return 0, err
	}
	if !IsApplicationsOpen(settings, time.Now()) {
		return 0, errors.New("applications are currently closed")
	}

	query := `
		INSERT INTO applications
		(student_id, applicant_type, student_number, name_surname, fio, birth_date, iin, school, level, passport_number, comments, year, major, gender, room_preference, additional_info, status, submitted_at, updated_at)
		VALUES ($1, NULLIF($2, ''), NULLIF($3, ''), NULLIF($4, ''), NULLIF($5, ''), $6, NULLIF($7, ''), NULLIF($8, ''), NULLIF($9, ''), NULLIF($10, ''), NULLIF($11, ''), $12, $13, $14, $15, $16, 'pending', NOW(), NOW())
		RETURNING id
	`
	var id int
	err = db.QueryRow(
		query,
		a.StudentID,
		a.ApplicantType,
		a.StudentNumber,
		a.NameSurname,
		a.FIO,
		a.BirthDate,
		a.IIN,
		a.School,
		a.Level,
		a.PassportNumber,
		a.Comments,
		a.Year,
		a.Major,
		a.Gender,
		a.RoomPreference,
		a.AdditionalInfo,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	db.Exec(`INSERT INTO audit_logs (actor_id, action, entity, entity_id) VALUES ($1, 'submit', 'application', $2)`, a.StudentID, id)
	return id, nil
}

func UpdateStudentApplicationDetails(db *sql.DB, studentID int, a models.Application) (models.Application, error) {
	query := `
		UPDATE applications
		SET room_preference = $1,
		    additional_info = $2,
		    updated_at = NOW()
		WHERE id = $3
		  AND student_id = $4
		RETURNING id, student_id, COALESCE(student_number, ''), COALESCE(name_surname, ''), COALESCE(fio, ''), birth_date,
		          COALESCE(iin, ''), COALESCE(school, ''), COALESCE(level, ''), COALESCE(comments, ''),
		          year, major, gender, COALESCE(room_preference, ''), COALESCE(additional_info, ''),
		          COALESCE(applicant_type, 'local'), COALESCE(passport_number, ''), status, submitted_at, updated_at, rejected_reason, reviewed_by, review_timestamp
	`

	var updated models.Application
	err := db.QueryRow(
		query,
		a.RoomPreference,
		a.AdditionalInfo,
		a.ID,
		studentID,
	).Scan(
		&updated.ID,
		&updated.StudentID,
		&updated.StudentNumber,
		&updated.NameSurname,
		&updated.FIO,
		&updated.BirthDate,
		&updated.IIN,
		&updated.School,
		&updated.Level,
		&updated.Comments,
		&updated.Year,
		&updated.Major,
		&updated.Gender,
		&updated.RoomPreference,
		&updated.AdditionalInfo,
		&updated.ApplicantType,
		&updated.PassportNumber,
		&updated.Status,
		&updated.SubmittedAt,
		&updated.UpdatedAt,
		&updated.RejectedReason,
		&updated.ReviewedBy,
		&updated.ReviewTimestamp,
	)
	if err == sql.ErrNoRows {
		return updated, errors.New("application not found")
	}
	if err != nil {
		return updated, err
	}

	normalizeApplicationFIO(&updated)
	updated.DecisionReason = updated.RejectedReason
	db.Exec(`INSERT INTO audit_logs (actor_id, action, entity, entity_id) VALUES ($1, 'update', 'application', $2)`, studentID, updated.ID)
	return updated, nil
}

func GetApplicationByID(db *sql.DB, id int) (models.Application, error) {
	var a models.Application

	query := `
		SELECT id, student_id, COALESCE(student_number, ''), COALESCE(name_surname, ''), COALESCE(fio, ''), birth_date,
		       COALESCE(iin, ''), COALESCE(school, ''), COALESCE(level, ''), COALESCE(comments, ''),
		       year, major, gender, COALESCE(room_preference, ''), COALESCE(additional_info, ''),
		       COALESCE(applicant_type, 'local'), COALESCE(passport_number, ''), status, submitted_at, updated_at, rejected_reason, reviewed_by, review_timestamp
		FROM applications
		WHERE id = $1
	`

	err := db.QueryRow(query, id).Scan(
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
		&a.SubmittedAt,
		&a.UpdatedAt,
		&a.RejectedReason,
		&a.ReviewedBy,
		&a.ReviewTimestamp,
	)

	if err == sql.ErrNoRows {
		return a, errors.New("application not found")
	}
	normalizeApplicationFIO(&a)
	a.DecisionReason = a.RejectedReason
	return a, err
}

func GetApplicationsByStudent(db *sql.DB, studentID int) ([]models.Application, error) {
	query := `
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
		WHERE a.student_id = $1
		ORDER BY a.submitted_at DESC
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
		normalizeApplicationFIO(&a)
		a.DecisionReason = a.RejectedReason
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

func SetApplicationDecision(db *sql.DB, id int, status string, reason string) error {
	query := `
		UPDATE applications
		SET status = $1,
		    rejected_reason = NULLIF($2, ''),
		    reviewed_by = NULL,
		    review_timestamp = NOW(),
		    updated_at = NOW()
		WHERE id = $3
	`
	_, err := db.Exec(query, status, reason, id)
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
