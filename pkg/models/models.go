package models

import "time"

type Session struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	StartedAt time.Time `json:"started_at"`
	EndedAt   time.Time `json:"ended_at,omitempty"`
}

type Command struct {
	ID          int64     `json:"id"`
	SessionID   int64     `json:"session_id"`
	Input       string    `json:"input"`
	OutputRaw   string    `json:"output_raw"`
	OutputClean string    `json:"output_clean"`
	ExitCode    int       `json:"exit_code"`
	Timestamp   time.Time `json:"timestamp"`
	DurationMs  int64     `json:"duration_ms"`
	Marked      bool      `json:"marked"`
	Tag         string    `json:"tag,omitempty"`
	Highlight   bool      `json:"highlight"`
}
