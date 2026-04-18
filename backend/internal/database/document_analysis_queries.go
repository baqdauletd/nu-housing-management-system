package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	"nu-housing-management-system/backend/internal/analysis"
	"nu-housing-management-system/backend/internal/models"
)

func UpsertDocumentAnalysis(db *sql.DB, result analysis.Result) error {
	issuesJSON, err := json.Marshal(result.Issues)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO document_extractions (
			document_id, detected_doc_type, extracted_fio, raw_text, confidence, model_name,
			requires_manual_review, contains_astana, registration_in_astana, workplace_in_astana, property_in_astana,
			processing_status, extraction_error_type, extraction_error_message, warnings, city_mentions, created_at, updated_at
		)
		VALUES (
			$1, $2, NULLIF($3, ''), $4, $5, $6,
			$7, $8, $9, $10, $11,
			$12, $13, $14, $15, '[]'::jsonb, NOW(), NOW()
		)
		ON CONFLICT (document_id) DO UPDATE SET
			detected_doc_type = EXCLUDED.detected_doc_type,
			extracted_fio = EXCLUDED.extracted_fio,
			raw_text = EXCLUDED.raw_text,
			confidence = EXCLUDED.confidence,
			model_name = EXCLUDED.model_name,
			requires_manual_review = EXCLUDED.requires_manual_review,
			contains_astana = EXCLUDED.contains_astana,
			registration_in_astana = EXCLUDED.registration_in_astana,
			workplace_in_astana = EXCLUDED.workplace_in_astana,
			property_in_astana = EXCLUDED.property_in_astana,
			processing_status = EXCLUDED.processing_status,
			extraction_error_type = EXCLUDED.extraction_error_type,
			extraction_error_message = EXCLUDED.extraction_error_message,
			warnings = EXCLUDED.warnings,
			updated_at = NOW()
	`

	processingStatus := "completed"
	if result.Status == string(analysis.StatusManualReview) {
		processingStatus = "manual_review"
	}
	errorType := ""
	if len(result.Issues) > 0 {
		errorType = result.Issues[0]
	}

	_, err = db.Exec(
		query,
		result.DocumentID,
		result.DetectedCategory,
		strings.ToValidUTF8(result.ExtractedFIO, ""),
		strings.ToValidUTF8(result.ExtractedTextPreview, ""),
		1.0,
		"openai:"+result.ExpectedType,
		result.Status == string(analysis.StatusManualReview),
		result.HasAstanaProperty || result.HasAstanaResidence || result.HasAstanaEmployment,
		result.HasAstanaResidence,
		result.HasAstanaEmployment,
		result.HasAstanaProperty,
		processingStatus,
		errorType,
		strings.ToValidUTF8(result.ReasoningSummary, ""),
		issuesJSON,
	)
	return err
}

func ListDocumentAnalysesByApplicationID(db *sql.DB, applicationID int) ([]models.DocumentAnalysis, error) {
	query := `
		SELECT de.id,
		       de.document_id,
		       d.application_id,
		       d.type,
		       COALESCE(NULLIF(de.detected_doc_type, ''), 'unknown'),
		       CASE
		           WHEN de.processing_status <> 'completed' OR de.requires_manual_review THEN 'manual_review'
		           WHEN de.property_in_astana OR de.registration_in_astana OR de.workplace_in_astana THEN 'failed'
		           ELSE 'passed'
		       END AS status,
		       de.property_in_astana,
		       de.registration_in_astana,
		       de.workplace_in_astana,
		       FALSE,
		       COALESCE(de.extracted_fio, ''),
		       COALESCE(de.warnings, '[]'::jsonb),
		       COALESCE(
		           NULLIF(de.extraction_error_message, ''),
		           CASE
		               WHEN de.property_in_astana THEN 'The document indicates real estate ownership in Astana.'
		               WHEN de.registration_in_astana THEN 'The document indicates residence registration in Astana.'
		               WHEN de.workplace_in_astana THEN 'The document indicates employment in Astana.'
		               WHEN de.processing_status <> 'completed' OR de.requires_manual_review THEN 'Manual review is required because the document is unrelated or inconclusive.'
		               ELSE 'The document passed the Astana verification checks.'
		           END
		       ) AS reasoning_summary,
		       LEFT(COALESCE(de.raw_text, ''), 240) AS extracted_text_preview,
		       '' AS raw_ai_json,
		       de.updated_at,
		       de.created_at,
		       de.updated_at
		FROM document_extractions de
		JOIN documents d ON d.id = de.document_id
		WHERE d.application_id = $1
		ORDER BY de.updated_at DESC, de.document_id DESC
	`

	rows, err := db.Query(query, applicationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var analyses []models.DocumentAnalysis
	for rows.Next() {
		analysisRow, err := scanDocumentAnalysisRow(rows)
		if err != nil {
			return nil, err
		}
		analyses = append(analyses, analysisRow)
	}
	return analyses, rows.Err()
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanDocumentAnalysisRow(scanner rowScanner) (models.DocumentAnalysis, error) {
	var analysisRow models.DocumentAnalysis
	var issuesJSON []byte

	err := scanner.Scan(
		&analysisRow.ID,
		&analysisRow.DocumentID,
		&analysisRow.ApplicationID,
		&analysisRow.ExpectedType,
		&analysisRow.DetectedCategory,
		&analysisRow.Status,
		&analysisRow.HasAstanaProperty,
		&analysisRow.HasAstanaResidence,
		&analysisRow.HasAstanaEmployment,
		&analysisRow.MatchedApplicantFIO,
		&analysisRow.ExtractedFIO,
		&issuesJSON,
		&analysisRow.ReasoningSummary,
		&analysisRow.ExtractedTextPreview,
		&analysisRow.RawAIJSON,
		&analysisRow.AnalyzedAt,
		&analysisRow.CreatedAt,
		&analysisRow.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return analysisRow, errors.New("document analysis not found")
		}
		return analysisRow, err
	}
	if len(issuesJSON) > 0 {
		if err := json.Unmarshal(issuesJSON, &analysisRow.Issues); err != nil {
			return analysisRow, err
		}
	}
	if analysisRow.Issues == nil {
		analysisRow.Issues = []string{}
	}
	return analysisRow, nil
}
