package database

import (
	"strings"

	"nu-housing-management-system/backend/internal/models"
)

func normalizeApplicationFIO(app *models.Application) {
	normalizeApplicationIdentity(app)
}

func normalizeApplicationIdentity(app *models.Application) {
	if strings.TrimSpace(app.ApplicantType) == "" {
		app.ApplicantType = extractApplicantTypeFromAdditionalInfo(app.AdditionalInfo)
	}
	if strings.TrimSpace(app.ApplicantType) == "" {
		app.ApplicantType = "local"
	}
	if strings.TrimSpace(app.FIO) != "" {
	} else {
		app.FIO = extractFIOFromAdditionalInfo(app.AdditionalInfo)
	}
	if strings.TrimSpace(app.PassportNumber) == "" {
		app.PassportNumber = extractPassportNumberFromAdditionalInfo(app.AdditionalInfo)
	}
}

func ExtractFIOForSubmission(additionalInfo string) string {
	return extractFIOFromAdditionalInfo(additionalInfo)
}

func ExtractPassportNumberForSubmission(additionalInfo string) string {
	return extractPassportNumberFromAdditionalInfo(additionalInfo)
}

func NormalizeApplicantTypeForSubmission(applicantType, additionalInfo string) string {
	normalized := strings.ToLower(strings.TrimSpace(applicantType))
	switch normalized {
	case "international", "intl", "foreign":
		return "international"
	case "local", "domestic", "":
		if normalized == "" {
			if parsed := extractApplicantTypeFromAdditionalInfo(additionalInfo); parsed != "" {
				return parsed
			}
		}
		return "local"
	default:
		if parsed := extractApplicantTypeFromAdditionalInfo(additionalInfo); parsed != "" {
			return parsed
		}
		return "local"
	}
}

func extractFIOFromAdditionalInfo(additionalInfo string) string {
	firstName := ""
	lastName := ""
	for _, line := range strings.Split(additionalInfo, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		lower := strings.ToLower(trimmed)
		switch {
		case strings.HasPrefix(lower, "фио:"),
			strings.HasPrefix(lower, "фио -"),
			strings.HasPrefix(lower, "fio:"),
			strings.HasPrefix(lower, "fio -"),
			strings.HasPrefix(lower, "full name:"),
			strings.HasPrefix(lower, "full name -"),
			strings.HasPrefix(lower, "name surname:"),
			strings.HasPrefix(lower, "name surname -"):
			if idx := strings.IndexAny(trimmed, ":-"); idx >= 0 && idx+1 < len(trimmed) {
				return strings.TrimSpace(trimmed[idx+1:])
			}
		case strings.HasPrefix(lower, "name:"),
			strings.HasPrefix(lower, "name -"),
			strings.HasPrefix(lower, "first name:"),
			strings.HasPrefix(lower, "first name -"),
			strings.HasPrefix(lower, "given name:"),
			strings.HasPrefix(lower, "given name -"):
			if idx := strings.IndexAny(trimmed, ":-"); idx >= 0 && idx+1 < len(trimmed) {
				firstName = strings.TrimSpace(trimmed[idx+1:])
			}
		case strings.HasPrefix(lower, "surname:"),
			strings.HasPrefix(lower, "surname -"),
			strings.HasPrefix(lower, "last name:"),
			strings.HasPrefix(lower, "last name -"),
			strings.HasPrefix(lower, "family name:"),
			strings.HasPrefix(lower, "family name -"):
			if idx := strings.IndexAny(trimmed, ":-"); idx >= 0 && idx+1 < len(trimmed) {
				lastName = strings.TrimSpace(trimmed[idx+1:])
			}
		}
	}

	if firstName != "" && lastName != "" {
		return strings.TrimSpace(firstName + " " + lastName)
	}
	if firstName != "" {
		return firstName
	}
	if lastName != "" {
		return lastName
	}

	return ""
}

func extractPassportNumberFromAdditionalInfo(additionalInfo string) string {
	for _, line := range strings.Split(additionalInfo, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		lower := strings.ToLower(trimmed)
		switch {
		case strings.HasPrefix(lower, "passport number:"),
			strings.HasPrefix(lower, "passport number -"),
			strings.HasPrefix(lower, "passport_no:"),
			strings.HasPrefix(lower, "passport_no -"),
			strings.HasPrefix(lower, "passport:"),
			strings.HasPrefix(lower, "passport -"):
			if idx := strings.IndexAny(trimmed, ":-"); idx >= 0 && idx+1 < len(trimmed) {
				return strings.TrimSpace(trimmed[idx+1:])
			}
		}
	}
	return ""
}

func extractApplicantTypeFromAdditionalInfo(additionalInfo string) string {
	for _, line := range strings.Split(additionalInfo, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		lower := strings.ToLower(trimmed)
		switch {
		case strings.HasPrefix(lower, "applicant type:"),
			strings.HasPrefix(lower, "applicant type -"),
			strings.HasPrefix(lower, "type:"),
			strings.HasPrefix(lower, "type -"):
			if idx := strings.IndexAny(trimmed, ":-"); idx >= 0 && idx+1 < len(trimmed) {
				value := strings.ToLower(strings.TrimSpace(trimmed[idx+1:]))
				switch value {
				case "international", "foreign", "intl":
					return "international"
				case "local", "domestic":
					return "local"
				}
			}
		}
	}
	return ""
}
