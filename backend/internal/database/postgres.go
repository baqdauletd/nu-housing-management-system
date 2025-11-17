package database

import (
    "database/sql"
    _ "github.com/lib/pq"
    "nu-housing-management-system/backend/internal/config" 
)

func ConnectPostgres(cfg *config.Config) (*sql.DB, error) {
    db, err := sql.Open("postgres", cfg.PostgresURL)
    if err != nil {
        return nil, err
    }
    return db, db.Ping()
}
