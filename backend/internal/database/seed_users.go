package database

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"

	"github.com/lib/pq"
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
	id := 0
	for attempt := 0; attempt < 20; attempt++ {
		candidateNuID := nuID
		if candidateNuID == "" || attempt > 0 {
			candidateNuID = generateRandomSubmissionNuID()
		}

		id, err = CreateUser(db, models.User{
			NuID:         candidateNuID,
			Email:        email,
			PasswordHash: "",
			RoleID:       roleID,
			Phone:        &phone,
		})
		if err == nil {
			break
		}

		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" && pqErr.Constraint == "users_nu_id_key" {
			continue
		}

		return models.User{}, false, err
	}
	if err != nil {
		return models.User{}, false, fmt.Errorf("failed to create OAuth user with unique nu_id after retries: %w", err)
	}

	user, err = GetUserByID(db, id)
	if err != nil {
		return models.User{}, false, err
	}

	return user, true, nil
}

func generateRandomSubmissionNuID() string {
	const (
		minSubmissionYear = 2018
		maxSubmissionYear = 2024
	)

	yearSpan := int64(maxSubmissionYear - minSubmissionYear + 1)
	yearOffset, err := rand.Int(rand.Reader, big.NewInt(yearSpan))
	if err != nil {
		return "202400000"
	}

	suffix, err := rand.Int(rand.Reader, big.NewInt(100000))
	if err != nil {
		return "202400000"
	}

	year := minSubmissionYear + int(yearOffset.Int64())
	return fmt.Sprintf("%04d%05d", year, suffix.Int64())
}
