package database

import (
	"database/sql"

	"nu-housing-management-system/backend/internal/models"
)

// EnsureOAuthUser returns the existing user by email or creates a default student account.
func EnsureOAuthUser(db *sql.DB, email string, nuID string) (models.User, bool, error) {
	user, err := GetUserByEmail(db, email)
	if err == nil {
		return user, false, nil
	}
	if err.Error() != "user not found" {
		return models.User{}, false, err
	}

	roleID, err := GetRoleIDByName(db, "student")
	if err != nil {
		return models.User{}, false, err
	}

	phone := ""
	id, err := CreateUser(db, models.User{
		NuID:         nuID,
		Email:        email,
		PasswordHash: "",
		RoleID:       roleID,
		Phone:        &phone,
	})
	if err != nil {
		return models.User{}, false, err
	}

	user, err = GetUserByID(db, id)
	if err != nil {
		return models.User{}, false, err
	}

	return user, true, nil
}
