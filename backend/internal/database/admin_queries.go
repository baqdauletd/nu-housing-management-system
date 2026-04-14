package database

import (
   "database/sql"
   "nu-housing-management-system/backend/internal/models"
)

func AdminSystemLogs(db *sql.DB) ([]models.LogEntry, error) {
   query := `
      SELECT al.id, al.actor_id, u.email, u.nu_id, al.action, al.entity, al.entity_id, al.timestamp
      FROM audit_logs al
      LEFT JOIN users u ON u.id = al.actor_id
      ORDER BY al.timestamp DESC
   `

   rows, err := db.Query(query)
   if err != nil {
      return nil, err
   }
   defer rows.Close()

   logs := make([]models.LogEntry, 0)

   for rows.Next() {
      var l models.LogEntry
      if err := rows.Scan(
         &l.ID,
         &l.ActorID,
         &l.ActorEmail,
         &l.ActorNuID,
         &l.Action,
         &l.Entity,
         &l.EntityID,
         &l.Timestamp,
      ); err != nil {
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