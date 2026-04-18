package database

import (
	"strings"

	"nu-housing-management-system/backend/internal/models"
)

func normalizeApplicationFIO(app *models.Application) {
	if strings.TrimSpace(app.FIO) != "" {
		return
	}
	app.FIO = extractFIOFromAdditionalInfo(app.AdditionalInfo)
}

func ExtractFIOForSubmission(additionalInfo string) string {
	return extractFIOFromAdditionalInfo(additionalInfo)
}

func extractFIOFromAdditionalInfo(additionalInfo string) string {
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
			strings.HasPrefix(lower, "full name -"):
			if idx := strings.IndexAny(trimmed, ":-"); idx >= 0 && idx+1 < len(trimmed) {
				return strings.TrimSpace(trimmed[idx+1:])
			}
		}
	}

	return ""
}
