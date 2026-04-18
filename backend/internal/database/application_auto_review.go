package database

import (
	"database/sql"
	"fmt"
	"strings"

	"nu-housing-management-system/backend/internal/analysis"
	"nu-housing-management-system/backend/internal/models"
)

type AutomatedDecision struct {
	Status string
	Reason string
}

var requiredAutomatedDocumentTypes = []string{
	string(analysis.DocumentTypePropertySelf),
	string(analysis.DocumentTypePropertyMother),
	string(analysis.DocumentTypePropertyFather),
	string(analysis.DocumentTypeResidenceSelf),
	string(analysis.DocumentTypeResidenceMother),
	string(analysis.DocumentTypeResidenceFather),
	string(analysis.DocumentTypeWorkMother),
	string(analysis.DocumentTypeWorkFather),
}

func DetermineAutomatedDecision(app models.Application, analyses []models.DocumentAnalysis) AutomatedDecision {
	if app.ReviewedBy != nil {
		return AutomatedDecision{}
	}

	latestByType := make(map[string]models.DocumentAnalysis, len(analyses))
	for _, docAnalysis := range analyses {
		docType, ok := analysis.NormalizeDocumentType(docAnalysis.ExpectedType)
		if !ok {
			continue
		}
		docTypeKey := string(docType)
		current, exists := latestByType[docTypeKey]
		if !exists || docAnalysis.AnalyzedAt.After(current.AnalyzedAt) {
			docAnalysis.ExpectedType = docTypeKey
			latestByType[docTypeKey] = docAnalysis
		}
	}

	for _, requiredType := range requiredAutomatedDocumentTypes {
		docAnalysis, ok := latestByType[requiredType]
		if !ok {
			continue
		}
		switch strings.ToLower(strings.TrimSpace(docAnalysis.Status)) {
		case string(analysis.StatusFailed):
			return AutomatedDecision{Status: "rejected", Reason: docAnalysis.ReasoningSummary}
		}
	}

	missing := make([]string, 0, len(requiredAutomatedDocumentTypes))
	for _, requiredType := range requiredAutomatedDocumentTypes {
		if _, ok := latestByType[requiredType]; !ok {
			missing = append(missing, requiredType)
		}
	}
	if len(missing) > 0 {
		return AutomatedDecision{
			Status: "pending",
			Reason: fmt.Sprintf("Manual review required because required documents are still missing: %s.", strings.Join(missing, ", ")),
		}
	}

	for _, requiredType := range requiredAutomatedDocumentTypes {
		docAnalysis := latestByType[requiredType]
		switch strings.ToLower(strings.TrimSpace(docAnalysis.Status)) {
		case string(analysis.StatusManualReview), "":
			return AutomatedDecision{
				Status: "pending",
				Reason: fmt.Sprintf("Manual review required for %s: %s", requiredType, strings.TrimSpace(docAnalysis.ReasoningSummary)),
			}
		}
	}

	return AutomatedDecision{
		Status: "approved",
		Reason: "Automatically approved because all required property, residence, and work documents show no Astana disqualifying evidence.",
	}
}

func ApplyAutomatedDecision(db *sql.DB, applicationID int) error {
	app, err := GetApplicationByID(db, applicationID)
	if err != nil {
		return err
	}

	analyses, err := ListDocumentAnalysesByApplicationID(db, applicationID)
	if err != nil {
		return err
	}

	decision := DetermineAutomatedDecision(app, analyses)
	if decision.Status == "" {
		return nil
	}
	return SetApplicationDecision(db, applicationID, decision.Status, decision.Reason)
}
