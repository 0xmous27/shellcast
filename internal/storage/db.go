package storage

import (
	"database/sql"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
	"github.com/0xmous27/shellcast/pkg/models"
)

func Open() (*sql.DB, error) {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".shellcast")
	os.MkdirAll(dir, 0755)
	db, err := sql.Open("sqlite", filepath.Join(dir, "shellcast.db"))
	if err != nil {
		return nil, err
	}
	db.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			id INTEGER PRIMARY KEY,
			name TEXT,
			started_at DATETIME,
			ended_at DATETIME
		);
		CREATE TABLE IF NOT EXISTS commands (
			id INTEGER PRIMARY KEY,
			session_id INTEGER,
			input TEXT,
			output_raw TEXT,
			output_clean TEXT,
			exit_code INTEGER DEFAULT 0,
			timestamp DATETIME,
			duration_ms INTEGER DEFAULT 0,
			marked INTEGER DEFAULT 0,
			tag TEXT DEFAULT '',
			highlight INTEGER DEFAULT 0
		);
	`)
	return db, nil
}

func CreateSession(db *sql.DB, name string) (int64, error) {
	res, err := db.Exec("INSERT INTO sessions (name, started_at) VALUES (?, ?)", name, time.Now())
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func EndSession(db *sql.DB, id int64) {
	db.Exec("UPDATE sessions SET ended_at = ? WHERE id = ?", time.Now(), id)
}

func SaveCommand(db *sql.DB, cmd *models.Command) (int64, error) {
	res, err := db.Exec(
		`INSERT INTO commands (session_id, input, output_raw, output_clean, exit_code, timestamp, duration_ms, marked, tag, highlight)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		cmd.SessionID, cmd.Input, cmd.OutputRaw, cmd.OutputClean, cmd.ExitCode,
		cmd.Timestamp, cmd.DurationMs, cmd.Marked, cmd.Tag, cmd.Highlight)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func MarkCommand(db *sql.DB, id int64, tag string) {
	db.Exec("UPDATE commands SET marked = 1, tag = ? WHERE id = ?", tag, id)
}

func GetLatestSession(db *sql.DB) (*models.Session, error) {
	s := &models.Session{}
	var end sql.NullTime
	err := db.QueryRow("SELECT id, name, started_at, ended_at FROM sessions ORDER BY id DESC LIMIT 1").
		Scan(&s.ID, &s.Name, &s.StartedAt, &end)
	if end.Valid {
		s.EndedAt = end.Time
	}
	return s, err
}

func GetSession(db *sql.DB, id int64) (*models.Session, error) {
	s := &models.Session{}
	var end sql.NullTime
	err := db.QueryRow("SELECT id, name, started_at, ended_at FROM sessions WHERE id = ?", id).
		Scan(&s.ID, &s.Name, &s.StartedAt, &end)
	if end.Valid {
		s.EndedAt = end.Time
	}
	return s, err
}

func ListSessions(db *sql.DB) ([]models.Session, error) {
	rows, err := db.Query("SELECT id, name, started_at, ended_at FROM sessions ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Session
	for rows.Next() {
		var s models.Session
		var end sql.NullTime
		rows.Scan(&s.ID, &s.Name, &s.StartedAt, &end)
		if end.Valid {
			s.EndedAt = end.Time
		}
		out = append(out, s)
	}
	return out, nil
}

func GetCommands(db *sql.DB, sessionID int64) ([]models.Command, error) {
	return queryCmd(db, "session_id = ?", sessionID)
}

func GetHighlights(db *sql.DB, sessionID int64) ([]models.Command, error) {
	return queryCmd(db, "session_id = ? AND (highlight = 1 OR marked = 1)", sessionID)
}

func GetMarks(db *sql.DB, sessionID int64) ([]models.Command, error) {
	return queryCmd(db, "session_id = ? AND marked = 1", sessionID)
}

func SearchCommands(db *sql.DB, sessionID int64, q string) ([]models.Command, error) {
	return queryCmd(db, "session_id = ? AND (input LIKE ? OR output_clean LIKE ?)", sessionID, "%"+q+"%", "%"+q+"%")
}

func queryCmd(db *sql.DB, where string, args ...interface{}) ([]models.Command, error) {
	rows, err := db.Query("SELECT id, session_id, input, output_raw, output_clean, exit_code, timestamp, duration_ms, marked, tag, highlight FROM commands WHERE "+where+" ORDER BY id", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Command
	for rows.Next() {
		var c models.Command
		rows.Scan(&c.ID, &c.SessionID, &c.Input, &c.OutputRaw, &c.OutputClean, &c.ExitCode, &c.Timestamp, &c.DurationMs, &c.Marked, &c.Tag, &c.Highlight)
		out = append(out, c)
	}
	return out, nil
}
