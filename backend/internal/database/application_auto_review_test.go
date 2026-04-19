package database

import (
	"testing"
	"time"

	"nu-housing-management-system/backend/internal/analysis"
	"nu-housing-management-system/backend/internal/models"
)

func TestDetermineAutomatedDecisionRejectsOnAstanaEvidence(t *testing.T) {
	analyses := buildAnalyses(string(analysis.StatusPassed), "ok")
	analyses[0].Status = string(analysis.StatusFailed)
	analyses[0].ReasoningSummary = "The document indicates real estate ownership in Astana."

	decision := DetermineAutomatedDecision(models.Application{}, analyses)
	if decision.Status != "rejected" {
		t.Fatalf("status = %q, want rejected", decision.Status)
	}
}

func TestDetermineAutomatedDecisionManualReviewOnInconclusiveDocument(t *testing.T) {
	analyses := buildAnalyses(string(analysis.StatusPassed), "ok")
	analyses[0].Status = string(analysis.StatusManualReview)
	analyses[0].ReasoningSummary = "The uploaded PDF does not clearly contain the required information."

	decision := DetermineAutomatedDecision(models.Application{}, analyses)
	if decision.Status != "pending" {
		t.Fatalf("status = %q, want pending", decision.Status)
	}
	if decision.Reason == "" {
		t.Fatal("expected pending reason")
	}
}

func TestDetermineAutomatedDecisionApprovesWhenAllRequiredDocsPass(t *testing.T) {
	decision := DetermineAutomatedDecision(models.Application{}, buildAnalyses(string(analysis.StatusPassed), "ok"))
	if decision.Status != "approved" {
		t.Fatalf("status = %q, want approved", decision.Status)
	}
}

func TestDetermineAutomatedDecisionApprovesInternationalApplicantWithPassport(t *testing.T) {
	decision := DetermineAutomatedDecision(models.Application{ApplicantType: "international"}, []models.DocumentAnalysis{
		{
			DocumentID:       1,
			ExpectedType:     string(analysis.DocumentTypePassport),
			Status:           string(analysis.StatusPassed),
			ReasoningSummary: "ok",
			AnalyzedAt:       time.Now().UTC(),
		},
	})
	if decision.Status != "approved" {
		t.Fatalf("status = %q, want approved", decision.Status)
	}
}

func TestDetermineAutomatedDecisionRejectsInternationalApplicantOnFailedPassport(t *testing.T) {
	decision := DetermineAutomatedDecision(models.Application{ApplicantType: "international"}, []models.DocumentAnalysis{
		{
			DocumentID:       1,
			ExpectedType:     string(analysis.DocumentTypePassport),
			Status:           string(analysis.StatusFailed),
			ReasoningSummary: "Kazakhstan citizenship detected.",
			AnalyzedAt:       time.Now().UTC(),
		},
	})
	if decision.Status != "rejected" {
		t.Fatalf("status = %q, want rejected", decision.Status)
	}
}

func buildAnalyses(status string, reason string) []models.DocumentAnalysis {
	now := time.Now().UTC()
	types := []string{
		string(analysis.DocumentTypePropertySelf),
		string(analysis.DocumentTypePropertyMother),
		string(analysis.DocumentTypePropertyFather),
		string(analysis.DocumentTypeResidenceSelf),
		string(analysis.DocumentTypeResidenceMother),
		string(analysis.DocumentTypeResidenceFather),
		string(analysis.DocumentTypeWorkMother),
		string(analysis.DocumentTypeWorkFather),
	}

	analyses := make([]models.DocumentAnalysis, 0, len(types))
	for idx, docType := range types {
		analyses = append(analyses, models.DocumentAnalysis{
			DocumentID:       idx + 1,
			ExpectedType:     docType,
			Status:           status,
			ReasoningSummary: reason,
			AnalyzedAt:       now.Add(time.Duration(idx) * time.Second),
		})
	}
	return analyses
}
