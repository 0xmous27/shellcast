package recorder

import (
	"database/sql"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/creack/pty"

	"github.com/0xmous27/shellcast/internal/highlight"
	"github.com/0xmous27/shellcast/internal/render"
	"github.com/0xmous27/shellcast/internal/storage"
	"github.com/0xmous27/shellcast/internal/strip"
	"github.com/0xmous27/shellcast/pkg/models"
)

type Recorder struct {
	db        *sql.DB
	sessionID string
	lastID    int
}

func New(db *sql.DB, sessionID string) *Recorder {
	return &Recorder{db: db, sessionID: sessionID}
}

func (r *Recorder) Run() error {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/bash"
	}

	c := exec.Command(shell)
	c.Env = append(os.Environ(), "SHELLCAST_SESSION="+r.sessionID)

	ptmx, err := pty.Start(c)
	if err != nil {
		return err
	}
	defer ptmx.Close()

	// Resize handling
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

	var inputBuf strings.Builder
	var outputBuf strings.Builder
	cmdStart := time.Now()
	gotFirstCmd := false

	// stdin → pty
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				return
			}
			ptmx.Write(buf[:n])

			for _, b := range buf[:n] {
				switch {
				case b == '\r' || b == '\n':
					input := strings.TrimSpace(inputBuf.String())
					if input != "" && gotFirstCmd {
						r.save(input, outputBuf.String(), cmdStart)
						outputBuf.Reset()
					}
					if input != "" {
						gotFirstCmd = true
					}
					inputBuf.Reset()
					cmdStart = time.Now()
				case b == 127 || b == 8: // backspace
					s := inputBuf.String()
					if len(s) > 0 {
						inputBuf.Reset()
						inputBuf.WriteString(s[:len(s)-1])
					}
				case b == 3 || b == 21: // Ctrl+C / Ctrl+U
					inputBuf.Reset()
				case b >= 32 && b < 127: // printable
					inputBuf.WriteByte(b)
				}
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
		if gotFirstCmd {
			outputBuf.Write(buf[:n])
		}
	}

	// Save last command
	if input := strings.TrimSpace(inputBuf.String()); input != "" && gotFirstCmd {
		r.save(input, outputBuf.String(), cmdStart)
	}

	c.Wait()
	return nil
}

func (r *Recorder) save(input, rawOutput string, start time.Time) {
	cleanOutput := strip.Clean(rawOutput)

	// Truncate if massive
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
		DurationMs:  float64(time.Since(start).Milliseconds()),
	}

	// Check #mark
	if isMark, tag := highlight.IsMark(input); isMark {
		cmd.Marked = true
		cmd.Tag = tag
		if r.lastID > 0 {
			storage.MarkCommand(r.db, r.lastID, tag)
		}
	}

	cmd.Highlight = highlight.IsHighlight(input, cleanOutput)
	storage.SaveCommand(r.db, cmd)
	r.lastID = cmd.ID

	// Take real screenshot of terminal (async to not block)
	go render.CaptureWindow(r.sessionID, cmd.ID)
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
