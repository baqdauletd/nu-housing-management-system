package database

import (
	"database/sql"
	"errors"
	"nu-housing-management-system/backend/internal/models"
)

////////////////////////////////////////////////////////////
// DOCUMENT QUERIES (UPDATED)
////////////////////////////////////////////////////////////

func SaveDocumentReplacingType(db *sql.DB, doc models.Document) (int, error) {
	updateQuery := `
		UPDATE documents
		SET file_url = $1,
		    original_filename = $2,
		    content_type = $3,
		    uploaded_at = NOW()
		WHERE id = (
			SELECT id
			FROM documents
			WHERE application_id = $4
			  AND type = $5
			ORDER BY uploaded_at DESC, id DESC
			LIMIT 1
		)
		RETURNING id
	`

	var id int
	err := db.QueryRow(
		updateQuery,
		doc.FileURL,
		doc.OriginalFilename,
		doc.ContentType,
		doc.ApplicationID,
		doc.Type,
	).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}

	query := `
		INSERT INTO documents (
			application_id,
			type,
			file_url,
			original_filename,
			content_type,
			uploaded_at
		)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id
	`
	err = db.QueryRow(
		query,
		doc.ApplicationID,
		doc.Type,
		doc.FileURL,
		doc.OriginalFilename,
		doc.ContentType,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func GetDocument(db *sql.DB, id int) (models.Document, error) {
	var d models.Document

	query := `
		SELECT id, application_id, type, file_url, original_filename, content_type, uploaded_at
		FROM documents
		WHERE id = $1
	`

	err := db.QueryRow(query, id).Scan(
		&d.ID,
		&d.ApplicationID,
		&d.Type,
		&d.FileURL,
		&d.OriginalFilename,
		&d.ContentType,
		&d.UploadedAt,
	)

	if err == sql.ErrNoRows {
		return d, errors.New("document not found")
	}
	return d, err
}

func GetDocumentsByApplication(db *sql.DB, appID int) ([]models.Document, error) {
	query := `
		SELECT DISTINCT ON (type) id, application_id, type, file_url, original_filename, content_type, uploaded_at
		FROM documents
		WHERE application_id = $1
		ORDER BY type, uploaded_at DESC, id DESC
	`

	rows, err := db.Query(query, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []models.Document
	for rows.Next() {
		var d models.Document
		if err := rows.Scan(
			&d.ID,
			&d.ApplicationID,
			&d.Type,
			&d.FileURL,
			&d.OriginalFilename,
			&d.ContentType,
			&d.UploadedAt,
		); err != nil {
			return nil, err
		}
		docs = append(docs, d)
	}
	return docs, nil
}
