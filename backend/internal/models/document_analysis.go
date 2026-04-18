package models

import "time"

type DocumentAnalysis struct {
	ID                   int       `json:"id"`
	DocumentID           int       `json:"document_id"`
	ApplicationID        int       `json:"application_id"`
	ExpectedType         string    `json:"expected_type"`
	DetectedCategory     string    `json:"detected_category"`
	Status               string    `json:"status"`
	HasAstanaProperty    bool      `json:"has_astana_property"`
	HasAstanaResidence   bool      `json:"has_astana_residence"`
	HasAstanaEmployment  bool      `json:"has_astana_employment"`
	MatchedApplicantFIO  bool      `json:"matched_applicant_fio"`
	ExtractedFIO         string    `json:"extracted_fio"`
	Issues               []string  `json:"issues"`
	ReasoningSummary     string    `json:"reasoning_summary"`
	ExtractedTextPreview string    `json:"extracted_text_preview"`
	RawAIJSON            string    `json:"raw_ai_json,omitempty"`
	AnalyzedAt           time.Time `json:"analyzed_at"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}
