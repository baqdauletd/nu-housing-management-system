package database

import (
	"database/sql"
	"errors"
	"nu-housing-management-system/backend/internal/models"
)

////////////////////////////////////////////////////////////
// USER QUERIES (UPDATED FOR SCHEMA)
////////////////////////////////////////////////////////////

// CreateUser inserts a new user (used for Register + AdminCreateUser)
func CreateUser(db *sql.DB, u models.User) (int, error) {
	query := `
		INSERT INTO users (nu_id, email, password_hash, role_id, phone, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id
	`
	var id int
	err := db.QueryRow(query, u.NuID, u.Email, u.PasswordHash, u.RoleID, u.Phone).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func GetUserByEmail(db *sql.DB, email string) (models.User, error) {
	var user models.User

	query := `
		SELECT id, nu_id, email, password_hash, role_id, phone, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	err := db.QueryRow(query, email).Scan(
		&user.ID,
		&user.NuID,
		&user.Email,
		&user.PasswordHash,
		&user.RoleID,
		&user.Phone,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return user, errors.New("user not found")
	}
	return user, err
}

func GetUserByID(db *sql.DB, id int) (models.User, error) {
	var user models.User

	query := `
		SELECT id, nu_id, email, role_id, phone, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	err := db.QueryRow(query, id).Scan(
		&user.ID,
		&user.NuID,
		&user.Email,
		&user.RoleID,
		&user.Phone,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return user, errors.New("user not found")
	}

	return user, err
}

func UpdateUser(db *sql.DB, u models.User) error {
	query := `
		UPDATE users
		SET email = $1,
		    phone = $2,
		    updated_at = NOW()
		WHERE id = $3
	`
	_, err := db.Exec(query, u.Email, u.Phone, u.ID)
	return err
}

func DeleteUser(db *sql.DB, userID int) error {
	_, err := db.Exec(`DELETE FROM users WHERE id = $1`, userID)
	return err
}

func ListUsers(db *sql.DB) ([]models.User, error) {
	query := `
		SELECT id, nu_id, email, role_id, phone, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
	`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []models.User{}
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.NuID, &u.Email, &u.RoleID, &u.Phone, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

// ValidateUserCredentials returns user with password_hash for login check
func ValidateUserCredentials(db *sql.DB, email string) (models.User, error) {
	return GetUserByEmail(db, email)
}
