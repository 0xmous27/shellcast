package main

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/0xmous27/shellcast/internal/export"
	"github.com/0xmous27/shellcast/internal/recorder"
	"github.com/0xmous27/shellcast/internal/render"
	"github.com/0xmous27/shellcast/internal/storage"
	"github.com/0xmous27/shellcast/pkg/models"
)

const version = "0.1.0"

const usage = `shellcast — terminal session evidence recorder

USAGE:
  shellcast start <name>        Start recording a session
  shellcast show                Show command timeline
  shellcast highlights          Show highlighted commands
  shellcast marks               Show bookmarked commands
  shellcast search <query>      Search commands and output
  shellcast proof <cmd-id>      Generate PNG screenshot
  shellcast export              Export markdown report
  shellcast sessions            List all sessions
  shellcast mark <id> [tag]     Mark a command retroactively
  shellcast version             Show version

DURING SESSION:
  #mark <tag>                   Bookmark the previous command
`

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		return
	}
	switch os.Args[1] {
	case "start":
		doStart()
	case "show":
		doShow()
	case "highlights":
		doHighlights()
	case "marks":
		doMarks()
	case "search":
		doSearch()
	case "proof":
		doProof()
	case "export":
		doExport()
	case "sessions":
		doSessions()
	case "mark":
		doMark()
	case "version", "-v":
		fmt.Printf("shellcast v%s\n", version)
	default:
		fmt.Print(usage)
	}
}

func fatal(msg string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "\033[31m✕ "+msg+"\033[0m\n", a...)
	os.Exit(1)
}

func openDB() *sql.DB {
	db, err := storage.Open()
	if err != nil {
		fatal("database: %v", err)
	}
	return db
}

func latestSessionID(db *sql.DB) int64 {
	s, err := storage.GetLatestSession(db)
	if err != nil {
		fatal("no sessions found")
	}
	return s.ID
}

func doStart() {
	name := "session"
	if len(os.Args) > 2 {
		name = os.Args[2]
	}
	db, err := storage.Open()
	if err != nil {
		fatal("db: %v", err)
	}
	defer db.Close()

	sid, err := storage.CreateSession(db, name)
	if err != nil {
		fatal("create session: %v", err)
	}

	fmt.Printf("\033[32m⚡ ShellCast recording: %s\033[0m\n", name)
	fmt.Printf("\033[90m   Type '#mark <tag>' to bookmark. 'exit' to stop.\033[0m\n\n")

	rec := recorder.New(db, sid)
	if err := rec.Run(); err != nil {
		fatal("recorder: %v", err)
	}

	storage.EndSession(db, sid)
	cmds, _ := storage.GetCommands(db, sid)
	fmt.Printf("\n\033[32m✓ Session saved: %s (%d commands)\033[0m\n", name, len(cmds))
}

func doShow() {
	db, _ := storage.Open()
	defer db.Close()
	cmds, _ := storage.GetCommands(db, latestSessionID(db))
	printCmds(cmds)
}

func doHighlights() {
	db, _ := storage.Open()
	defer db.Close()
	cmds, _ := storage.GetHighlights(db, latestSessionID(db))
	if len(cmds) == 0 {
		fmt.Println("No highlights.")
		return
	}
	printCmds(cmds)
}

func doMarks() {
	db, _ := storage.Open()
	defer db.Close()
	cmds, _ := storage.GetMarks(db, latestSessionID(db))
	if len(cmds) == 0 {
		fmt.Println("No bookmarks. Use '#mark <tag>' during a session.")
		return
	}
	printCmds(cmds)
}

func doSearch() {
	if len(os.Args) < 3 {
		fatal("usage: shellcast search <query>")
	}
	db, _ := storage.Open()
	defer db.Close()
	cmds, _ := storage.SearchCommands(db, latestSessionID(db), os.Args[2])
	if len(cmds) == 0 {
		fmt.Printf("No results for \"%s\"\n", os.Args[2])
		return
	}
	printCmds(cmds)
}

func doProof() {
	if len(os.Args) < 3 {
		fatal("usage: shellcast proof <cmd-id>")
	}
	id, err := strconv.ParseInt(os.Args[2], 10, 64)
	if err != nil {
		fatal("invalid id")
	}
	db, _ := storage.Open()
	defer db.Close()
	cmds, _ := storage.GetCommands(db, latestSessionID(db))
	for _, c := range cmds {
		if c.ID == id {
			path, err := render.GenerateProofFile(c.ID, c.Input, c.OutputRaw, c.OutputClean)
			if err != nil {
				fatal("render: %v", err)
			}
			fmt.Printf("\033[32m✓ Proof saved: %s\033[0m\n", path)
			return
		}
	}
	fatal("command #%d not found", id)
}

func doExport() {
	db, _ := storage.Open()
	defer db.Close()
	sid := latestSessionID(db)
	sess, _ := storage.GetSession(db, sid)
	cmds, _ := storage.GetHighlights(db, sid)
	if len(cmds) == 0 {
		cmds, _ = storage.GetCommands(db, sid)
	}
	fmt.Print(export.ToMarkdown(sess, cmds))
}

func doSessions() {
	db, _ := storage.Open()
	defer db.Close()
	sessions, _ := storage.ListSessions(db)
	if len(sessions) == 0 {
		fmt.Println("No sessions. Run: shellcast start <name>")
		return
	}
	fmt.Printf("\033[1m%-6s %-20s %-20s\033[0m\n", "ID", "NAME", "STARTED")
	for _, s := range sessions {
		fmt.Printf("%-6d %-20s %-20s\n", s.ID, s.Name, s.StartedAt.Format("2006-01-02 15:04"))
	}
}

func doMark() {
	if len(os.Args) < 3 {
		fatal("usage: shellcast mark <cmd-id> [tag]")
	}
	id, _ := strconv.ParseInt(os.Args[2], 10, 64)
	tag := "marked"
	if len(os.Args) > 3 {
		tag = strings.Join(os.Args[3:], " ")
	}
	db, _ := storage.Open()
	defer db.Close()
	storage.MarkCommand(db, id, tag)
	fmt.Printf("✓ Command #%d marked: \"%s\"\n", id, tag)
}

func printCmds(cmds []models.Command) {
	for _, c := range cmds {
		icon := " "
		if c.Marked {
			icon = "\033[33m🔖\033[0m"
		} else if c.Highlight {
			icon = "\033[36m⚡\033[0m"
		}
		tag := ""
		if c.Tag != "" {
			tag = fmt.Sprintf("  \033[33m[%s]\033[0m", c.Tag)
		}
		fmt.Printf("%s \033[90m#%-4d %s\033[0m%s\n", icon, c.ID, c.Timestamp.Format("15:04:05"), tag)
		fmt.Printf("  \033[32m$\033[0m %s\n", c.Input)
		if c.OutputClean != "" {
			lines := strings.Split(c.OutputClean, "\n")
			max := 5
			if len(lines) < max {
				max = len(lines)
			}
			for _, l := range lines[:max] {
				fmt.Printf("  \033[90m%s\033[0m\n", l)
			}
			if len(lines) > 5 {
				fmt.Printf("  \033[90m(%d more lines)\033[0m\n", len(lines)-5)
			}
		}
		fmt.Println()
	}
}
