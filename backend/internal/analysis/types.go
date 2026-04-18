package analysis

import (
	"context"
	"encoding/json"
	"strings"
	"time"
)

type DocumentType string

const (
	DocumentTypePropertySelf    DocumentType = "property_egov_self"
	DocumentTypePropertyMother  DocumentType = "property_egov_mother"
	DocumentTypePropertyFather  DocumentType = "property_egov_father"
	DocumentTypeResidenceSelf   DocumentType = "residence_egov_self"
	DocumentTypeResidenceMother DocumentType = "residence_egov_mother"
	DocumentTypeResidenceFather DocumentType = "residence_egov_father"
	DocumentTypeWorkMother      DocumentType = "work_mother"
	DocumentTypeWorkFather      DocumentType = "work_father"
)

type Category string

const (
	CategoryProperty  Category = "property"
	CategoryResidence Category = "residence"
	CategoryWork      Category = "work"
	CategoryUnknown   Category = "unknown"
)

type Status string

const (
	StatusPassed       Status = "passed"
	StatusFailed       Status = "failed"
	StatusManualReview Status = "manual_review"
)

type ApplicationContext struct {
	ApplicationID  int
	StudentID      int
	StudentEmail   string
	StudentNuID    string
	StudentFIO     string
	Year           int
	Major          string
	Gender         string
	RoomPreference string
	AdditionalInfo string
	SubmittedAt    time.Time
	UploadTime     time.Time
}

type Request struct {
	DocumentID   int
	ExpectedType DocumentType
	PDFBytes     []byte
	Application  ApplicationContext
}

type Result struct {
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
}

type aiResponse struct {
	DetectedCategory     string   `json:"detected_category"`
	HasAstanaProperty    bool     `json:"has_astana_property"`
	HasAstanaResidence   bool     `json:"has_astana_residence"`
	HasAstanaEmployment  bool     `json:"has_astana_employment"`
	ContainsRelevantInfo bool     `json:"contains_relevant_info"`
	MatchedApplicantFIO  bool     `json:"matched_applicant_fio"`
	ExtractedFIO         string   `json:"extracted_fio"`
	Status               string   `json:"status"`
	Issues               []string `json:"issues"`
	ReasoningSummary     string   `json:"reasoning_summary"`
	ExtractedTextPreview string   `json:"extracted_text_preview"`
}

type Analyzer interface {
	AnalyzeDocument(ctx context.Context, req Request) (Result, error)
}

func ExpectedCategory(docType DocumentType) Category {
	switch docType {
	case DocumentTypePropertySelf, DocumentTypePropertyMother, DocumentTypePropertyFather:
		return CategoryProperty
	case DocumentTypeResidenceSelf, DocumentTypeResidenceMother, DocumentTypeResidenceFather:
		return CategoryResidence
	case DocumentTypeWorkMother, DocumentTypeWorkFather:
		return CategoryWork
	default:
		return CategoryUnknown
	}
}

func RequiresApplicantNameMatch(docType DocumentType) bool {
	switch docType {
	case DocumentTypePropertySelf, DocumentTypeResidenceSelf:
		return true
	default:
		return false
	}
}

func NormalizeDocumentType(value string) (DocumentType, bool) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = strings.ReplaceAll(normalized, "-", "_")
	normalized = strings.ReplaceAll(normalized, " ", "_")

	switch normalized {
	case "property_self":
		normalized = string(DocumentTypePropertySelf)
	case "property_mother":
		normalized = string(DocumentTypePropertyMother)
	case "property_father":
		normalized = string(DocumentTypePropertyFather)
	case "residence_self":
		normalized = string(DocumentTypeResidenceSelf)
	case "residence_mother":
		normalized = string(DocumentTypeResidenceMother)
	case "residence_father":
		normalized = string(DocumentTypeResidenceFather)
	}

	switch DocumentType(normalized) {
	case DocumentTypePropertySelf,
		DocumentTypePropertyMother,
		DocumentTypePropertyFather,
		DocumentTypeResidenceSelf,
		DocumentTypeResidenceMother,
		DocumentTypeResidenceFather,
		DocumentTypeWorkMother,
		DocumentTypeWorkFather:
		return DocumentType(normalized), true
	default:
		return "", false
	}
}

func NormalizeCategory(value string) Category {
	switch Category(strings.ToLower(strings.TrimSpace(value))) {
	case CategoryProperty:
		return CategoryProperty
	case CategoryResidence:
		return CategoryResidence
	case CategoryWork:
		return CategoryWork
	default:
		return CategoryUnknown
	}
}

func NormalizeStatus(value string) Status {
	switch Status(strings.ToLower(strings.TrimSpace(value))) {
	case StatusPassed:
		return StatusPassed
	case StatusFailed:
		return StatusFailed
	case StatusManualReview:
		return StatusManualReview
	default:
		return StatusManualReview
	}
}

func BuildPreview(text string) string {
	text = NormalizeExtractedText(text)
	if len(text) <= 240 {
		return text
	}
	return strings.TrimSpace(text[:240])
}

func NormalizeExtractedText(text string) string {
	text = strings.ToValidUTF8(text, "")
	return strings.Join(strings.Fields(text), " ")
}

func ManualReviewResult(req Request, preview string, summary string, issues ...string) Result {
	return Result{
		DocumentID:           req.DocumentID,
		ApplicationID:        req.Application.ApplicationID,
		ExpectedType:         string(req.ExpectedType),
		DetectedCategory:     string(CategoryUnknown),
		Status:               string(StatusManualReview),
		Issues:               uniqueIssues(issues),
		ReasoningSummary:     summary,
		ExtractedTextPreview: preview,
		AnalyzedAt:           time.Now().UTC(),
	}
}

func PostProcessResult(req Request, extractedText string, aiResult aiResponse, rawAIJSON string) Result {
	preview := BuildPreview(extractedText)
	category := NormalizeCategory(aiResult.DetectedCategory)
	status := NormalizeStatus(aiResult.Status)
	issues := uniqueIssues(aiResult.Issues)
	reasoning := strings.TrimSpace(aiResult.ReasoningSummary)
	reasoning = strings.ToValidUTF8(reasoning, "")
	expectedCategory := ExpectedCategory(req.ExpectedType)
	if reasoning == "" {
		reasoning = "Automated document check completed."
	}

	if category == CategoryUnknown && expectedCategory != CategoryUnknown {
		category = expectedCategory
	}

	relevantAndSafe := status == StatusPassed &&
		!aiResult.HasAstanaProperty &&
		!aiResult.HasAstanaResidence &&
		!aiResult.HasAstanaEmployment &&
		category != CategoryUnknown

	if (!aiResult.ContainsRelevantInfo || category == CategoryUnknown) && !relevantAndSafe {
		status = StatusManualReview
		issues = append(issues, "missing_relevant_information")
		if strings.TrimSpace(aiResult.ReasoningSummary) == "" {
			reasoning = "The uploaded PDF does not clearly contain property, residence, or work information required for automatic review."
		}
	}

	if aiResult.HasAstanaProperty {
		status = StatusFailed
		issues = append(issues, "astana_real_estate_detected")
		if reasoning == "Automated document check completed." {
			reasoning = "The document indicates real estate ownership in Astana."
		}
	}
	if aiResult.HasAstanaResidence {
		status = StatusFailed
		issues = append(issues, "astana_residence_detected")
		if reasoning == "Automated document check completed." {
			reasoning = "The document indicates residence registration in Astana."
		}
	}
	if aiResult.HasAstanaEmployment {
		status = StatusFailed
		issues = append(issues, "astana_employment_detected")
		if reasoning == "Automated document check completed." {
			reasoning = "The document indicates employment in Astana."
		}
	}

	if RequiresApplicantNameMatch(req.ExpectedType) {
		appFIO := NormalizeExtractedText(req.Application.StudentFIO)
		extractedFIO := NormalizeExtractedText(aiResult.ExtractedFIO)
		if appFIO == "" {
			status = StatusManualReview
			issues = append(issues, "missing_application_fio")
			reasoning = "Manual review required because the application FIO is missing and the self-document name cannot be compared."
		} else if extractedFIO == "" {
			status = StatusManualReview
			issues = append(issues, "missing_document_fio")
			reasoning = "Manual review required because the self-document does not contain a clear Cyrillic FIO to compare with the application."
		} else if !aiResult.MatchedApplicantFIO {
			status = StatusFailed
			issues = append(issues, "applicant_name_mismatch")
			reasoning = "The Cyrillic FIO in the applicant's own document does not match the FIO in the application."
		}
	}

	return Result{
		DocumentID:           req.DocumentID,
		ApplicationID:        req.Application.ApplicationID,
		ExpectedType:         string(req.ExpectedType),
		DetectedCategory:     string(category),
		Status:               string(status),
		HasAstanaProperty:    aiResult.HasAstanaProperty,
		HasAstanaResidence:   aiResult.HasAstanaResidence,
		HasAstanaEmployment:  aiResult.HasAstanaEmployment,
		MatchedApplicantFIO:  aiResult.MatchedApplicantFIO,
		ExtractedFIO:         NormalizeExtractedText(aiResult.ExtractedFIO),
		Issues:               uniqueIssues(issues),
		ReasoningSummary:     reasoning,
		ExtractedTextPreview: preview,
		RawAIJSON:            rawAIJSON,
		AnalyzedAt:           time.Now().UTC(),
	}
}

func analysisSystemPrompt() string {
	return strings.Join([]string{
		"You validate university housing application PDF documents.",
		"Return strict JSON only that matches the required schema.",
		"Use the uploaded PDF file, any extracted PDF text, and the provided application data.",
		"Detect whether the document contains evidence of property in Astana, residence in Astana, or work in Astana.",
		"Only for the applicant's own documents, compare the Cyrillic FIO in the document against the application's FIO field.",
		"Do not compare names in parents' documents because there is nothing reliable to compare them to.",
		"If the document clearly refers to another city such as Shymkent, Almaty, Karaganda, or any city other than Astana, that is acceptable and should not be treated as a failure.",
		"Reject only when the document indicates Astana specifically, not just any city in Kazakhstan.",
		"If extracted text is empty because the PDF is scanned, inspect the PDF pages themselves before deciding manual_review.",
		"If the uploaded PDF is unrelated, unreadable, or missing the relevant information, set status to manual_review.",
		"If the document shows any Astana property, Astana residence, or Astana employment, set status to failed.",
		"If the document is relevant and clearly shows no Astana property, residence, or employment issue, set status to passed.",
		"List issues as short snake_case strings.",
	}, " ")
}

func analysisPayload(req Request, extractedText string) map[string]any {
	return map[string]any{
		"expected_document_type": string(req.ExpectedType),
		"application_data": map[string]any{
			"application_id":  req.Application.ApplicationID,
			"student_id":      req.Application.StudentID,
			"student_email":   req.Application.StudentEmail,
			"student_nu_id":   req.Application.StudentNuID,
			"student_fio":     req.Application.StudentFIO,
			"year":            req.Application.Year,
			"major":           req.Application.Major,
			"gender":          req.Application.Gender,
			"room_preference": req.Application.RoomPreference,
			"additional_info": req.Application.AdditionalInfo,
			"submitted_at":    req.Application.SubmittedAt.UTC().Format(time.RFC3339),
		},
		"extracted_pdf_text": extractedText,
		"review_rule":        "Documents mentioning cities other than Astana are acceptable. Only Astana property, Astana residence, or Astana employment should disqualify the application.",
	}
}

func analysisJSONSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			"detected_category": map[string]any{
				"type": "string",
				"enum": []string{"property", "residence", "work", "unknown"},
			},
			"has_astana_property":    map[string]any{"type": "boolean"},
			"has_astana_residence":   map[string]any{"type": "boolean"},
			"has_astana_employment":  map[string]any{"type": "boolean"},
			"contains_relevant_info": map[string]any{"type": "boolean"},
			"matched_applicant_fio":  map[string]any{"type": "boolean"},
			"extracted_fio":          map[string]any{"type": "string"},
			"status": map[string]any{
				"type": "string",
				"enum": []string{"passed", "failed", "manual_review"},
			},
			"issues": map[string]any{
				"type":  "array",
				"items": map[string]any{"type": "string"},
			},
			"reasoning_summary":      map[string]any{"type": "string"},
			"extracted_text_preview": map[string]any{"type": "string"},
		},
		"required": []string{
			"detected_category",
			"has_astana_property",
			"has_astana_residence",
			"has_astana_employment",
			"contains_relevant_info",
			"matched_applicant_fio",
			"extracted_fio",
			"status",
			"issues",
			"reasoning_summary",
			"extracted_text_preview",
		},
	}
}

func MustJSON(value any) string {
	bytes, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(bytes)
}

func uniqueIssues(issues []string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(issues))
	for _, issue := range issues {
		normalized := strings.ToLower(strings.TrimSpace(issue))
		normalized = strings.ReplaceAll(normalized, " ", "_")
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	return result
}
