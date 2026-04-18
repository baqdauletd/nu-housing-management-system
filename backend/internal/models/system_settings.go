package models

import "time"

type SystemSettings struct {
	ID                  int
	ApplicationsEnabled bool
	ApplicationOpen     *time.Time
	ApplicationClose    *time.Time
	RequiredDocuments   []string
}
