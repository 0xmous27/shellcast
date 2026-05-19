package models

import "time"

type Session struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	StartedAt time.Time `json:"started_at"`
	EndedAt   time.Time `json:"ended_at,omitempty"`
	Commands  int       `json:"commands"`
}

type Command struct {
	ID          int       `json:"id"`
	SessionID   string    `json:"session_id"`
	Input       string    `json:"input"`
	OutputRaw   string    `json:"output_raw"`
	OutputClean string    `json:"output_clean"`
	ExitCode    int       `json:"exit_code"`
	Timestamp   time.Time `json:"timestamp"`
	DurationMs  float64   `json:"duration_ms"`
	Marked      bool      `json:"marked"`
	Tag         string    `json:"tag,omitempty"`
	Highlight   bool      `json:"highlight"`
}
