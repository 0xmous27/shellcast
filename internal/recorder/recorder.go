package recorder

import (
	"database/sql"
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
	"github.com/0xmous27/shellcast/internal/storage"
	"github.com/0xmous27/shellcast/pkg/models"
)

type Recorder struct {
	db        *sql.DB
	sessionID int64
	lastCmdID int64
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
	c.Env = os.Environ()

	ptmx, err := pty.Start(c)
	if err != nil {
		return err
	}
	defer ptmx.Close()

	// Handle resize
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGWINCH)
	go func() {
		for range sigCh {
			pty.InheritSize(os.Stdin, ptmx)
		}
	}()
	sigCh <- syscall.SIGWINCH

	// Raw mode
	old, err := setRaw(os.Stdin.Fd())
	if err != nil {
		return err
	}
	defer restoreTerminal(os.Stdin.Fd(), old)

	var mu sync.Mutex
	var inputLine []byte      // current line being typed
	var outputBuf []byte      // output since last command
	var pendingCmd string     // command waiting for output
	var cmdStart time.Time    // when pending command was submitted
	promptReady := false      // have we seen at least one prompt cycle

	// stdin → pty (capture input keystrokes)
	go func() {
		buf := make([]byte, 256)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				return
			}
			// Always forward to PTY immediately
			ptmx.Write(buf[:n])

			mu.Lock()
			for _, b := range buf[:n] {
				switch {
				case b == '\r' || b == '\n':
					cmd := strings.TrimSpace(string(inputLine))
					inputLine = inputLine[:0]

					if cmd == "" {
						mu.Unlock()
						goto next
					}

					// Save previous command now (we have its output)
					if pendingCmd != "" && promptReady {
						r.saveCmd(pendingCmd, string(outputBuf), cmdStart)
					}

					// This new command becomes pending
					pendingCmd = cmd
					outputBuf = outputBuf[:0]
					cmdStart = time.Now()
					promptReady = true

				case b == 127 || b == 8: // backspace/delete
					if len(inputLine) > 0 {
						inputLine = inputLine[:len(inputLine)-1]
					}
				case b == 3: // Ctrl+C — cancel current input
					inputLine = inputLine[:0]
				case b == 21: // Ctrl+U — clear line
					inputLine = inputLine[:0]
				case b == 0x1b: // ESC — start of escape sequence, skip
					// We'll handle this by ignoring until we get a letter
					// For now just don't add ESC to input
				case b >= 32 && b < 127: // printable ASCII
					inputLine = append(inputLine, b)
				}
			}
			mu.Unlock()
		next:
		}
	}()

	// pty → stdout (capture output)
	buf := make([]byte, 8192)
	for {
		n, err := ptmx.Read(buf)
		if err != nil {
			if err != io.EOF {
				// ignore
			}
			break
		}
		// Always display to user
		os.Stdout.Write(buf[:n])

		// Capture output for pending command
		mu.Lock()
		if pendingCmd != "" {
			outputBuf = append(outputBuf, buf[:n]...)
		}
		mu.Unlock()
	}

	// Save final pending command
	mu.Lock()
	if pendingCmd != "" && promptReady {
		r.saveCmd(pendingCmd, string(outputBuf), cmdStart)
	}
	mu.Unlock()

	c.Wait()
	return nil
}

func (r *Recorder) saveCmd(input, rawOutput string, start time.Time) {
	cleanOutput := parser.Clean(rawOutput)
	duration := time.Since(start).Milliseconds()

	// Clean input — remove terminal response sequences that leak through
	input = parser.CleanInput(input)
	if input == "" {
		return
	}

	// Handle #mark
	if isMark, tag := highlight.IsMark(input); isMark {
		if r.lastCmdID > 0 {
			storage.MarkCommand(r.db, r.lastCmdID, tag)
		}
		return // don't store the #mark itself as a command
	}

	cmd := &models.Command{
		SessionID:   r.sessionID,
		Input:       input,
		OutputRaw:   rawOutput,
		OutputClean: cleanOutput,
		Timestamp:   start,
		DurationMs:  duration,
		Highlight:   highlight.IsHighlight(input, cleanOutput),
	}

	id, _ := storage.SaveCommand(r.db, cmd)
	r.lastCmdID = id
}

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
