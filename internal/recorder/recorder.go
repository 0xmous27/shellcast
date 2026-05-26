package recorder

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/creack/pty"

	"github.com/0xmous27/shellcast/internal/highlight"
	"github.com/0xmous27/shellcast/internal/parser"
	"github.com/0xmous27/shellcast/internal/render"
	"github.com/0xmous27/shellcast/internal/storage"
	"github.com/0xmous27/shellcast/pkg/models"
)

type Recorder struct {
	db        *sql.DB
	sessionID int64
	lastID    int64
}

func New(db *sql.DB, sessionID int64) *Recorder {
	return &Recorder{db: db, sessionID: sessionID}
}

func (r *Recorder) Run() error {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/bash"
	}

	c := exec.Command(shell)
	c.Env = append(os.Environ(), fmt.Sprintf("SHELLCAST_SESSION=%d", r.sessionID))

	ptmx, err := pty.Start(c)
	if err != nil {
		return err
	}
	defer ptmx.Close()

	// Resize
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			pty.InheritSize(os.Stdin, ptmx)
		}
	}()
	ch <- syscall.SIGWINCH

	// Raw mode
	old, err := setRaw(os.Stdin.Fd())
	if err != nil {
		return err
	}
	defer restoreTerminal(os.Stdin.Fd(), old)

	var mu sync.Mutex
	var outputBuf strings.Builder // accumulates ALL output between Enter presses
	var pendingCmd string         // last detected command
	var cmdStart time.Time
	cmdCount := 0
	enterCount := 0

	// stdin → pty (just forward + count Enters)
	go func() {
		buf := make([]byte, 256)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				return
			}
			ptmx.Write(buf[:n])
			for _, b := range buf[:n] {
				if b == '\r' || b == '\n' {
					mu.Lock()
					enterCount++
					mu.Unlock()
				}
			}
		}
	}()

	// Process output: detect commands by finding prompt+command in the stream
	// After each Enter, we wait for output to settle, then extract the command
	// from the prompt line that appeared just before the output
	go func() {
		lastEnter := 0
		for {
			time.Sleep(300 * time.Millisecond)
			mu.Lock()
			if enterCount > lastEnter {
				lastEnter = enterCount
				mu.Unlock()
				// Wait for output to settle
				time.Sleep(200 * time.Millisecond)

				mu.Lock()
				raw := outputBuf.String()
				// Try to extract command from the accumulated output
				cmd := extractCommandFromOutput(raw)
				if cmd != "" && cmd != pendingCmd {
					// Save previous command
					if pendingCmd != "" && cmdCount > 0 {
						// Output for previous command = everything up to this new prompt
						r.saveCmd(pendingCmd, raw, cmdStart)
					}
					pendingCmd = cmd
					outputBuf.Reset()
					cmdStart = time.Now()
					cmdCount++
				}
				mu.Unlock()
			} else {
				mu.Unlock()
			}
		}
	}()

	// pty → stdout + capture
	buf := make([]byte, 8192)
	for {
		n, err := ptmx.Read(buf)
		if err != nil {
			if err != io.EOF {
				// ignore
			}
			break
		}
		os.Stdout.Write(buf[:n])
		mu.Lock()
		outputBuf.Write(buf[:n])
		mu.Unlock()
	}

	// Save last command
	mu.Lock()
	if pendingCmd != "" && cmdCount > 0 {
		r.saveCmd(pendingCmd, outputBuf.String(), cmdStart)
	}
	mu.Unlock()

	c.Wait()
	return nil
}

// extractCommandFromOutput finds the most recent command from terminal output.
// It looks for prompt patterns ($ or #) followed by the command text.
func extractCommandFromOutput(raw string) string {
	clean := parser.Clean(raw)
	lines := strings.Split(clean, "\n")

	// Scan from the end to find the last prompt+command line
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		// Look for "$ command" or "# command" pattern
		if idx := strings.LastIndex(line, "$ "); idx >= 0 {
			cmd := strings.TrimSpace(line[idx+2:])
			if cmd != "" && !isPromptOnly(cmd) {
				return cmd
			}
		}
		if idx := strings.LastIndex(line, "# "); idx >= 0 {
			cmd := strings.TrimSpace(line[idx+2:])
			if cmd != "" && !isPromptOnly(cmd) {
				return cmd
			}
		}
	}
	return ""
}

func isPromptOnly(s string) bool {
	// Filter out things that look like prompt fragments, not commands
	s = strings.TrimSpace(s)
	if s == "" {
		return true
	}
	// If it's just escape remnants
	if len(s) < 2 {
		return true
	}
	return false
}

func (r *Recorder) saveCmd(input, rawOutput string, start time.Time) {
	input = parser.CleanInput(input)
	if input == "" {
		return
	}

	cleanOutput := parser.Clean(rawOutput)

	if len(rawOutput) > 100000 {
		rawOutput = rawOutput[:100000]
	}
	if len(cleanOutput) > 50000 {
		cleanOutput = cleanOutput[:50000]
	}

	cmd := &models.Command{
		SessionID:   r.sessionID,
		Input:       input,
		OutputRaw:   rawOutput,
		OutputClean: cleanOutput,
		Timestamp:   start,
		DurationMs:  time.Since(start).Milliseconds(),
	}

	if isMark, tag := highlight.IsMark(input); isMark {
		cmd.Marked = true
		cmd.Tag = tag
		if r.lastID > 0 {
			storage.MarkCommand(r.db, r.lastID, tag)
		}
	}

	cmd.Highlight = highlight.IsHighlight(input, cleanOutput)
	id, _ := storage.SaveCommand(r.db, cmd)
	r.lastID = id

	go render.GenerateProof(id, input, rawOutput, cleanOutput, "")
}

// Terminal raw mode
type termState syscall.Termios

func setRaw(fd uintptr) (*termState, error) {
	var t syscall.Termios
	if _, _, e := syscall.Syscall(syscall.SYS_IOCTL, fd, syscall.TCGETS, uintptr(unsafe.Pointer(&t))); e != 0 {
		return nil, e
	}
	old := t
	t.Iflag &^= syscall.IGNBRK | syscall.BRKINT | syscall.PARMRK | syscall.ISTRIP | syscall.INLCR | syscall.IGNCR | syscall.ICRNL | syscall.IXON
	t.Oflag &^= syscall.OPOST
	t.Lflag &^= syscall.ECHO | syscall.ECHONL | syscall.ICANON | syscall.ISIG | syscall.IEXTEN
	t.Cflag &^= syscall.CSIZE | syscall.PARENB
	t.Cflag |= syscall.CS8
	if _, _, e := syscall.Syscall(syscall.SYS_IOCTL, fd, syscall.TCSETS, uintptr(unsafe.Pointer(&t))); e != 0 {
		return nil, e
	}
	return (*termState)(&old), nil
}

func restoreTerminal(fd uintptr, state *termState) {
	syscall.Syscall(syscall.SYS_IOCTL, fd, syscall.TCSETS, uintptr(unsafe.Pointer(state)))
}
