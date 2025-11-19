package database

import (
	"database/sql"
	"errors"
	"nu-housing-management-system/backend/internal/models"
	"time"
)

////////////////////////////////////////////////////////////
// USER QUERIES
////////////////////////////////////////////////////////////

// CreateUser inserts a new user (used for Register + AdminCreateUser)
func CreateUser(db *sql.DB, u models.User) error {
	query := `
		INSERT INTO users (name, email, password_hash, role, created_at)
		VALUES ($1, $2, $3, $4, NOW())
	`
	_, err := db.Exec(query, u.Name, u.Email, u.PasswordHash, u.Role)
	return err
}

// GetUserByEmail is important for Login
func GetUserByEmail(db *sql.DB, email string) (models.User, error) {
	var user models.User

	query := `
		SELECT id, name, email, password_hash, role, created_at
		FROM users
		WHERE email = $1
	`
	err := db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return user, errors.New("user not found")
		}
		return user, err
	}

	return user, nil
}

// GetUserByID is used for profile (GetProfile)
func GetUserByID(db *sql.DB, id int) (models.User, error) {
	var user models.User

	query := `
		SELECT id, name, email, role, created_at
		FROM users
		WHERE id = $1
	`
	err := db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Role,
		&user.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return user, errors.New("user not found")
		}
		return user, err
	}

	return user, nil
}

// UpdateUser modifies allowed user fields
func UpdateUser(db *sql.DB, u models.User) error {
	query := `
		UPDATE users
		SET name = $1,
			email = $2
		WHERE id = $3
	`
	_, err := db.Exec(query, u.Name, u.Email, u.ID)
	return err
}

// DeleteUser removes a user (AdminDeleteUser)
func DeleteUser(db *sql.DB, userID int) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := db.Exec(query, userID)
	return err
}

// ListUsers returns all users (AdminListUsers)
func ListUsers(db *sql.DB) ([]models.User, error) {
	query := `
		SELECT id, name, email, role, created_at
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

		err := rows.Scan(
			&u.ID,
			&u.Name,
			&u.Email,
			&u.Role,
			&u.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		users = append(users, u)
	}

	return users, nil
}

// ValidateUserCredentials is for Login
func ValidateUserCredentials(db *sql.DB, email string) (models.User, error) {
	var user models.User

	query := `
		SELECT id, name, email, password_hash, role, created_at
		FROM users
		WHERE email = $1
	`

	err := db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
	)

	if err != nil {
		return user, err
	}

	return user, nil
}

////////////////////////////////////////////////////////////
// OPTIONAL: FOR SEEDING / TEST PURPOSES
////////////////////////////////////////////////////////////

func SeedAdminUser(db *sql.DB) error {
	_, err := db.Exec(`
		INSERT INTO users (name, email, password_hash, role, created_at)
		VALUES ('Admin', 'admin@example.com', 'adminpass', 'admin', $1)
	`, time.Now())

	return err
}
