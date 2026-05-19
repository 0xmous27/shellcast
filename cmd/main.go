package main

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/0xmous27/shellcast/internal/export"
	"github.com/0xmous27/shellcast/internal/recorder"
	"github.com/0xmous27/shellcast/internal/render"
	"github.com/0xmous27/shellcast/internal/storage"
	"github.com/0xmous27/shellcast/pkg/models"
)

const version = "0.1.0"

const usage = `shellcast — terminal session recorder for pentesters

USAGE:
  shellcast start <name>           Start recording a session
  shellcast show [session-id]      Show all commands (clean text)
  shellcast highlights [id]        Show auto-highlighted commands
  shellcast marks [id]             Show bookmarked commands
  shellcast search <query> [id]    Search commands by input/output
  shellcast proof <from-to> [id]   Export commands as PNG screenshot
  shellcast export [id]            Export session as markdown
  shellcast sessions               List all sessions
  shellcast mark <cmd-id> [tag]    Mark a command retroactively
  shellcast version                Show version

DURING SESSION:
  #mark <tag>                      Bookmark the current moment

EXAMPLES:
  shellcast start client-x
  shellcast show
  shellcast highlights
  shellcast proof 3-7
  shellcast search "root"
  shellcast export > report.md
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
	case "version", "-v", "--version":
		fmt.Printf("shellcast v%s\n", version)
	default:
		fmt.Print(usage)
	}
}

func fatal(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "\033[31m✕ "+msg+"\033[0m\n", args...)
	os.Exit(1)
}

func openDB() *sql.DB {
	db, err := storage.Open()
	if err != nil {
		fatal("database error: %v", err)
	}
	return db
}

func resolveSession(db *sql.DB) string {
	if len(os.Args) > 2 && !strings.HasPrefix(os.Args[2], "-") {
		// Check if it's a flag-like arg (for proof command, it could be a range)
		if os.Args[1] != "proof" && os.Args[1] != "search" && os.Args[1] != "mark" {
			return os.Args[2]
		}
	}
	// For commands with positional args, session is arg[3]
	if (os.Args[1] == "proof" || os.Args[1] == "search") && len(os.Args) > 3 {
		return os.Args[3]
	}
	s, err := storage.GetLatestSession(db)
	if err != nil {
		fatal("no sessions found. Run: shellcast start <name>")
	}
	return s.ID
}

func doStart() {
	name := "session"
	if len(os.Args) > 2 {
		name = os.Args[2]
	}
	db := openDB()
	defer db.Close()

	sess, err := storage.CreateSession(db, name)
	if err != nil {
		fatal("failed to create session: %v", err)
	}

	fmt.Printf("\033[32m⚡ ShellCast recording: %s\033[0m\n", sess.ID)
	fmt.Printf("\033[90m   Type '#mark <tag>' to bookmark. Type 'exit' to stop.\033[0m\n\n")

	rec := recorder.New(db, sess.ID)
	if err := rec.Run(); err != nil {
		fatal("recorder error: %v", err)
	}

	storage.EndSession(db, sess.ID)
	fmt.Printf("\n\033[32m✓ Session saved: %s (%d commands)\033[0m\n", sess.ID, countCmds(db, sess.ID))
}

func countCmds(db *sql.DB, sid string) int {
	cmds, _ := storage.GetCommands(db, sid)
	return len(cmds)
}

func doShow() {
	db := openDB()
	defer db.Close()
	sid := resolveSession(db)
	cmds, _ := storage.GetCommands(db, sid)
	printCmds(cmds)
}

func doHighlights() {
	db := openDB()
	defer db.Close()
	sid := resolveSession(db)
	cmds, _ := storage.GetHighlights(db, sid)
	if len(cmds) == 0 {
		fmt.Println("No highlights found.")
		return
	}
	printCmds(cmds)
}

func doMarks() {
	db := openDB()
	defer db.Close()
	sid := resolveSession(db)
	cmds, _ := storage.GetMarks(db, sid)
	if len(cmds) == 0 {
		fmt.Println("No bookmarks. Use '#mark <tag>' during a session.")
		return
	}
	printCmds(cmds)
}

func doSearch() {
	if len(os.Args) < 3 {
		fatal("usage: shellcast search <query> [session-id]")
	}
	query := os.Args[2]
	db := openDB()
	defer db.Close()
	sid := resolveSession(db)
	cmds, _ := storage.SearchCommands(db, sid, query)
	if len(cmds) == 0 {
		fmt.Printf("No results for \"%s\"\n", query)
		return
	}
	printCmds(cmds)
}

func doProof() {
	if len(os.Args) < 3 {
		fatal("usage: shellcast proof <cmd-id|from-to|--marks> [session-id]")
	}
	arg := os.Args[2]
	db := openDB()
	defer db.Close()

	var sid string
	if len(os.Args) > 3 {
		sid = os.Args[3]
	} else {
		s, err := storage.GetLatestSession(db)
		if err != nil {
			fatal("no sessions found")
		}
		sid = s.ID
	}

	dir := render.ScreenshotDir(sid)

	if arg == "--marks" || arg == "-m" {
		cmds, _ := storage.GetMarks(db, sid)
		if len(cmds) == 0 {
			fmt.Println("No marked screenshots.")
			return
		}
		for _, c := range cmds {
			file := fmt.Sprintf("%s/cmd_%04d.png", dir, c.ID)
			if _, err := os.Stat(file); err == nil {
				fmt.Printf("🔖 #%d [%s] → %s\n", c.ID, c.Tag, file)
			}
		}
		return
	}

	if arg == "--all" || arg == "-a" {
		entries, _ := os.ReadDir(dir)
		if len(entries) == 0 {
			fmt.Println("No screenshots found.")
			return
		}
		for _, e := range entries {
			fmt.Println(filepath.Join(dir, e.Name()))
		}
		return
	}

	// Single ID or range
	if strings.Contains(arg, "-") {
		parts := strings.Split(arg, "-")
		from, _ := strconv.Atoi(parts[0])
		to, _ := strconv.Atoi(parts[1])
		for i := from; i <= to; i++ {
			file := fmt.Sprintf("%s/cmd_%04d.png", dir, i)
			if _, err := os.Stat(file); err == nil {
				fmt.Printf("#%d → %s\n", i, file)
			}
		}
	} else {
		id, _ := strconv.Atoi(arg)
		file := fmt.Sprintf("%s/cmd_%04d.png", dir, id)
		if _, err := os.Stat(file); err == nil {
			fmt.Printf("#%d → %s\n", id, file)
			// Try to open with default viewer
			exec.Command("xdg-open", file).Start()
		} else {
			fatal("screenshot not found: %s", file)
		}
	}
}

func doExport() {
	db := openDB()
	defer db.Close()
	sid := resolveSession(db)
	sess, _ := storage.GetSession(db, sid)
	cmds, _ := storage.GetHighlights(db, sid)
	if len(cmds) == 0 {
		cmds, _ = storage.GetCommands(db, sid)
	}
	fmt.Print(export.ToMarkdown(sess, cmds))
}

func doSessions() {
	db := openDB()
	defer db.Close()
	sessions, _ := storage.ListSessions(db)
	if len(sessions) == 0 {
		fmt.Println("No sessions yet. Run: shellcast start <name>")
		return
	}
	fmt.Printf("\033[1m%-30s %-20s %-6s\033[0m\n", "SESSION", "STARTED", "CMDS")
	for _, s := range sessions {
		fmt.Printf("%-30s %-20s %-6d\n", s.ID, s.StartedAt.Format("2006-01-02 15:04"), s.Commands)
	}
}

func doMark() {
	if len(os.Args) < 3 {
		fatal("usage: shellcast mark <cmd-id> [tag]")
	}
	id, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fatal("invalid command ID")
	}
	tag := "marked"
	if len(os.Args) > 3 {
		tag = strings.Join(os.Args[3:], " ")
	}
	db := openDB()
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

		tagStr := ""
		if c.Tag != "" {
			tagStr = fmt.Sprintf("  \033[33m[%s]\033[0m", c.Tag)
		}

		fmt.Printf("%s \033[90m#%-4d %s\033[0m%s\n", icon, c.ID, c.Timestamp.Format("15:04:05"), tagStr)
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
				fmt.Printf("  \033[90m... (%d more lines)\033[0m\n", len(lines)-5)
			}
		}
		fmt.Println()
	}
}
