package database

import (
	"database/sql"
	"errors"
)

// GetRoleIDByName returns role id for a role name (e.g., "student", "housing", "admin")
func GetRoleIDByName(db *sql.DB, name string) (int, error) {
	var id int
	err := db.QueryRow(`SELECT id FROM roles WHERE name = $1`, name).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, errors.New("role not found")
	}
	return id, err
}

func GetRoleNameByID(db *sql.DB, id int) (string, error) {
	var name string
	err := db.QueryRow(`SELECT name FROM roles WHERE id = $1`, id).Scan(&name)
	if err == sql.ErrNoRows {
		return "", errors.New("role not found")
	}
	return name, err
}
