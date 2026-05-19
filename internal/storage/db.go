package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/0xmous27/shellcast/pkg/models"
)

func DBPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".shellcast", "shellcast.db")
}

func Open() (*sql.DB, error) {
	p := DBPath()
	os.MkdirAll(filepath.Dir(p), 0755)
	db, err := sql.Open("sqlite3", p+"?_journal_mode=WAL")
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			started_at DATETIME NOT NULL,
			ended_at DATETIME,
			commands INTEGER DEFAULT 0
		);
		CREATE TABLE IF NOT EXISTS commands (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL,
			input TEXT NOT NULL,
			output_raw TEXT,
			output_clean TEXT,
			exit_code INTEGER DEFAULT 0,
			timestamp DATETIME NOT NULL,
			duration_ms REAL DEFAULT 0,
			marked INTEGER DEFAULT 0,
			tag TEXT DEFAULT '',
			highlight INTEGER DEFAULT 0,
			FOREIGN KEY (session_id) REFERENCES sessions(id)
		);
		CREATE INDEX IF NOT EXISTS idx_cmd_session ON commands(session_id);
		CREATE INDEX IF NOT EXISTS idx_cmd_marked ON commands(session_id, marked);
	`)
	return db, err
}

func CreateSession(db *sql.DB, name string) (*models.Session, error) {
	id := fmt.Sprintf("%s-%d", name, time.Now().Unix())
	s := &models.Session{ID: id, Name: name, StartedAt: time.Now()}
	_, err := db.Exec("INSERT INTO sessions (id, name, started_at) VALUES (?,?,?)", s.ID, s.Name, s.StartedAt)
	return s, err
}

func EndSession(db *sql.DB, id string) error {
	_, err := db.Exec("UPDATE sessions SET ended_at=? WHERE id=?", time.Now(), id)
	return err
}

func SaveCommand(db *sql.DB, cmd *models.Command) error {
	res, err := db.Exec(
		`INSERT INTO commands (session_id,input,output_raw,output_clean,exit_code,timestamp,duration_ms,marked,tag,highlight)
		 VALUES (?,?,?,?,?,?,?,?,?,?)`,
		cmd.SessionID, cmd.Input, cmd.OutputRaw, cmd.OutputClean, cmd.ExitCode,
		cmd.Timestamp, cmd.DurationMs, cmd.Marked, cmd.Tag, cmd.Highlight)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	cmd.ID = int(id)
	_, _ = db.Exec("UPDATE sessions SET commands=commands+1 WHERE id=?", cmd.SessionID)
	return nil
}

func MarkCommand(db *sql.DB, id int, tag string) error {
	_, err := db.Exec("UPDATE commands SET marked=1, tag=? WHERE id=?", tag, id)
	return err
}

func GetLatestSession(db *sql.DB) (*models.Session, error) {
	s := &models.Session{}
	var end sql.NullTime
	err := db.QueryRow("SELECT id,name,started_at,ended_at,commands FROM sessions ORDER BY started_at DESC LIMIT 1").
		Scan(&s.ID, &s.Name, &s.StartedAt, &end, &s.Commands)
	if end.Valid {
		s.EndedAt = end.Time
	}
	return s, err
}

func GetSession(db *sql.DB, id string) (*models.Session, error) {
	s := &models.Session{}
	var end sql.NullTime
	err := db.QueryRow("SELECT id,name,started_at,ended_at,commands FROM sessions WHERE id=?", id).
		Scan(&s.ID, &s.Name, &s.StartedAt, &end, &s.Commands)
	if end.Valid {
		s.EndedAt = end.Time
	}
	return s, err
}

func ListSessions(db *sql.DB) ([]models.Session, error) {
	rows, err := db.Query("SELECT id,name,started_at,ended_at,commands FROM sessions ORDER BY started_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Session
	for rows.Next() {
		var s models.Session
		var end sql.NullTime
		rows.Scan(&s.ID, &s.Name, &s.StartedAt, &end, &s.Commands)
		if end.Valid {
			s.EndedAt = end.Time
		}
		out = append(out, s)
	}
	return out, nil
}

func queryCommands(db *sql.DB, where string, args ...interface{}) ([]models.Command, error) {
	q := "SELECT id,session_id,input,output_raw,output_clean,exit_code,timestamp,duration_ms,marked,tag,highlight FROM commands WHERE " + where + " ORDER BY id"
	rows, err := db.Query(q, args...)
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

func GetCommands(db *sql.DB, sessionID string) ([]models.Command, error) {
	return queryCommands(db, "session_id=?", sessionID)
}

func GetHighlights(db *sql.DB, sessionID string) ([]models.Command, error) {
	return queryCommands(db, "session_id=? AND (highlight=1 OR marked=1)", sessionID)
}

func GetMarks(db *sql.DB, sessionID string) ([]models.Command, error) {
	return queryCommands(db, "session_id=? AND marked=1", sessionID)
}

func SearchCommands(db *sql.DB, sessionID, query string) ([]models.Command, error) {
	return queryCommands(db, "session_id=? AND (input LIKE ? OR output_clean LIKE ?)", sessionID, "%"+query+"%", "%"+query+"%")
}

func GetCommandRange(db *sql.DB, sessionID string, from, to int) ([]models.Command, error) {
	return queryCommands(db, "session_id=? AND id>=? AND id<=?", sessionID, from, to)
}
