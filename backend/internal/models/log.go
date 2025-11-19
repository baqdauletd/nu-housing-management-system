package models

import "time"

type LogEntry struct {
	ID        int       `json:"id"`
	ActorID   *int      `json:"actor_id,omitempty"`
	Action    string    `json:"action"`
	Entity    string    `json:"entity"`
	EntityID  *int      `json:"entity_id,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}
