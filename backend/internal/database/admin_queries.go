package database

import (
	"database/sql"
	"nu-housing-management-system/backend/internal/models"
)

////////////////////////////////////////////////////////////
// ADMIN QUERIES (UPDATED)
////////////////////////////////////////////////////////////

func AdminSystemLogs(db *sql.DB) ([]models.LogEntry, error) {
	query := `
		SELECT id, actor_id, action, entity, entity_id, timestamp
		FROM audit_logs
		ORDER BY timestamp DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.LogEntry

	for rows.Next() {
		var l models.LogEntry
		if err := rows.Scan(&l.ID, &l.ActorID, &l.Action, &l.Entity, &l.EntityID, &l.Timestamp); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}

func AdminStats(db *sql.DB) (models.Stats, error) {
	var stats models.Stats

	_ = db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&stats.Users)
	_ = db.QueryRow(`SELECT COUNT(*) FROM applications`).Scan(&stats.Applications)
	_ = db.QueryRow(`SELECT COUNT(*) FROM applications WHERE status='approved'`).Scan(&stats.Approved)

	return stats, nil
}
