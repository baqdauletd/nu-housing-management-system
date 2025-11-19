package database

import (
	"database/sql"
	"errors"
	"nu-housing-management-system/backend/internal/models"
)

////////////////////////////////////////////////////////////
// DOCUMENT QUERIES (UPDATED)
////////////////////////////////////////////////////////////

func InsertDocument(db *sql.DB, doc models.Document) (int, error) {
	query := `
		INSERT INTO documents (application_id, type, file_url, uploaded_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id
	`
	var id int
	err := db.QueryRow(query, doc.ApplicationID, doc.Type, doc.FileURL).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func GetDocument(db *sql.DB, id int) (models.Document, error) {
	var d models.Document

	query := `
		SELECT id, application_id, type, file_url, uploaded_at
		FROM documents
		WHERE id = $1
	`

	err := db.QueryRow(query, id).Scan(
		&d.ID,
		&d.ApplicationID,
		&d.Type,
		&d.FileURL,
		&d.UploadedAt,
	)

	if err == sql.ErrNoRows {
		return d, errors.New("document not found")
	}
	return d, err
}

func GetDocumentsByApplication(db *sql.DB, appID int) ([]models.Document, error) {
	query := `
		SELECT id, application_id, type, file_url, uploaded_at
		FROM documents
		WHERE application_id = $1
	`

	rows, err := db.Query(query, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []models.Document
	for rows.Next() {
		var d models.Document
		if err := rows.Scan(&d.ID, &d.ApplicationID, &d.Type, &d.FileURL, &d.UploadedAt); err != nil {
			return nil, err
		}
		docs = append(docs, d)
	}
	return docs, nil
}
