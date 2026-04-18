package database

import (
	"database/sql"
	"encoding/json"
	"time"

	"nu-housing-management-system/backend/internal/models"
)

func GetSystemSettings(db *sql.DB) (models.SystemSettings, error) {
	settingsID, err := ensureSystemSettingsRow(db)
	if err != nil {
		return models.SystemSettings{}, err
	}

	var (
		settings              models.SystemSettings
		openDate              sql.NullTime
		closeDate             sql.NullTime
		requiredDocumentsJSON []byte
	)

	query := `
		SELECT id, applications_enabled, application_open, application_close, COALESCE(required_documents, '[]'::jsonb)
		FROM system_settings
		WHERE id = $1
	`
	err = db.QueryRow(query, settingsID).Scan(
		&settings.ID,
		&settings.ApplicationsEnabled,
		&openDate,
		&closeDate,
		&requiredDocumentsJSON,
	)
	if err != nil {
		return models.SystemSettings{}, err
	}

	if openDate.Valid {
		date := normalizeDate(openDate.Time)
		settings.ApplicationOpen = &date
	}
	if closeDate.Valid {
		date := normalizeDate(closeDate.Time)
		settings.ApplicationClose = &date
	}
	if err := json.Unmarshal(requiredDocumentsJSON, &settings.RequiredDocuments); err != nil {
		return models.SystemSettings{}, err
	}
	if settings.RequiredDocuments == nil {
		settings.RequiredDocuments = []string{}
	}

	return settings, nil
}

func UpdateSystemSettings(db *sql.DB, settings models.SystemSettings) (models.SystemSettings, error) {
	settingsID, err := ensureSystemSettingsRow(db)
	if err != nil {
		return models.SystemSettings{}, err
	}

	requiredDocumentsJSON, err := json.Marshal(settings.RequiredDocuments)
	if err != nil {
		return models.SystemSettings{}, err
	}

	query := `
		UPDATE system_settings
		SET applications_enabled = $1,
		    application_open = $2,
		    application_close = $3,
		    required_documents = $4
		WHERE id = $5
	`
	if _, err := db.Exec(
		query,
		settings.ApplicationsEnabled,
		settings.ApplicationOpen,
		settings.ApplicationClose,
		requiredDocumentsJSON,
		settingsID,
	); err != nil {
		return models.SystemSettings{}, err
	}

	return GetSystemSettings(db)
}

func IsApplicationsOpen(settings models.SystemSettings, now time.Time) bool {
	if !settings.ApplicationsEnabled {
		return false
	}

	currentDate := normalizeDate(now)
	if settings.ApplicationOpen != nil && currentDate.Before(normalizeDate(*settings.ApplicationOpen)) {
		return false
	}
	if settings.ApplicationClose != nil && currentDate.After(normalizeDate(*settings.ApplicationClose)) {
		return false
	}

	return true
}

func ensureSystemSettingsRow(db *sql.DB) (int, error) {
	if err := ensureSystemSettingsSchema(db); err != nil {
		return 0, err
	}

	var settingsID int

	query := `
		WITH existing AS (
			SELECT id
			FROM system_settings
			ORDER BY id ASC
			LIMIT 1
		),
		inserted AS (
			INSERT INTO system_settings (applications_enabled, required_documents)
			SELECT TRUE, '[]'::jsonb
			WHERE NOT EXISTS (SELECT 1 FROM existing)
			RETURNING id
		)
		SELECT COALESCE(
			(SELECT id FROM existing),
			(SELECT id FROM inserted)
		)
	`

	err := db.QueryRow(query).Scan(&settingsID)
	return settingsID, err
}

func ensureSystemSettingsSchema(db *sql.DB) error {
	statements := []string{
		`
		CREATE TABLE IF NOT EXISTS system_settings (
			id SERIAL PRIMARY KEY,
			application_open DATE,
			application_close DATE,
			required_documents JSONB DEFAULT '[]'::jsonb
		)
		`,
		`
		ALTER TABLE system_settings
		ADD COLUMN IF NOT EXISTS applications_enabled BOOLEAN NOT NULL DEFAULT TRUE
		`,
		`
		ALTER TABLE system_settings
		ADD COLUMN IF NOT EXISTS application_open DATE
		`,
		`
		ALTER TABLE system_settings
		ADD COLUMN IF NOT EXISTS application_close DATE
		`,
		`
		ALTER TABLE system_settings
		ADD COLUMN IF NOT EXISTS required_documents JSONB DEFAULT '[]'::jsonb
		`,
		`
		ALTER TABLE system_settings
		ALTER COLUMN required_documents SET DEFAULT '[]'::jsonb
		`,
	}

	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			return err
		}
	}

	return nil
}

func normalizeDate(value time.Time) time.Time {
	year, month, day := value.UTC().Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
