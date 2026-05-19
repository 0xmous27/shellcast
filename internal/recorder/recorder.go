package recorder

import (
	"bufio"
	"database/sql"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/creack/pty"

	"github.com/0xmous27/shellcast/internal/highlight"
	"github.com/0xmous27/shellcast/internal/render"
	"github.com/0xmous27/shellcast/internal/storage"
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

	// Unix socket for shell hook
	sockPath := filepath.Join(os.TempDir(), fmt.Sprintf("sc_%d.sock", os.Getpid()))
	os.Remove(sockPath)
	listener, err := net.Listen("unix", sockPath)
	if err != nil {
		return fmt.Errorf("socket: %w", err)
	}
	defer listener.Close()
	defer os.Remove(sockPath)

	// Write a small env file that gets sourced
	hookContent := fmt.Sprintf(`__sc_hook() { local cmd=$(HISTTIMEFORMAT= history 1 | sed 's/^[ ]*[0-9]*[ ]*//'); [ -n "$cmd" ] && echo "$cmd" | socat - UNIX-CONNECT:%s 2>/dev/null; }; PROMPT_COMMAND="__sc_hook;${PROMPT_COMMAND:-:}"`, sockPath)
	hookFile := filepath.Join(os.TempDir(), fmt.Sprintf("sc_hook_%d.sh", os.Getpid()))
	os.WriteFile(hookFile, []byte(hookContent), 0644)
	defer os.Remove(hookFile)

	// Start shell with BASH_ENV so hook loads automatically
	c := exec.Command(shell, "-i")
	env := os.Environ()
	env = append(env, "SHELLCAST_SESSION="+r.sessionID)
	// BASH_ENV is sourced for interactive shells when combined with --rcfile workaround
	// Instead, we use --init-file approach but source user's rc first
	c.Env = env

	ptmx, err := pty.Start(c)
	if err != nil {
		return fmt.Errorf("pty: %w", err)
	}
	defer ptmx.Close()

	// Resize
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
		return fmt.Errorf("raw: %w", err)
	}
	defer restoreTerminal(os.Stdin.Fd(), old)

	// Listen for commands
	go r.listenCommands(listener)

	// Inject hook after shell is ready
	go func() {
		time.Sleep(800 * time.Millisecond)
		// Source the hook file silently
		inject := fmt.Sprintf(" source %s\n", hookFile)
		ptmx.Write([]byte(inject))
	}()

	// stdin → pty
	go func() {
		io.Copy(ptmx, os.Stdin)
	}()

	// pty → stdout
	io.Copy(os.Stdout, ptmx)

	c.Wait()
	return nil
}

func (r *Recorder) listenCommands(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		go func(c net.Conn) {
			defer c.Close()
			scanner := bufio.NewScanner(c)
			for scanner.Scan() {
				r.saveCommand(scanner.Text())
			}
		}(conn)
	}
}

func (r *Recorder) saveCommand(input string) {
	input = strings.TrimSpace(input)
	if input == "" {
		return
	}
	// Skip our own hook commands
	if strings.HasPrefix(input, "source /tmp/sc_hook") || strings.HasPrefix(input, " source /tmp/sc_hook") {
		return
	}
	if strings.Contains(input, "__sc_hook") {
		return
	}

	cmd := &models.Command{
		SessionID:   r.sessionID,
		Input:       input,
		OutputClean: "",
		OutputRaw:   "",
		Timestamp:   time.Now(),
	}

	if isMark, tag := highlight.IsMark(input); isMark {
		cmd.Marked = true
		cmd.Tag = tag
		if r.lastID > 0 {
			storage.MarkCommand(r.db, r.lastID, tag)
		}
	}

	cmd.Highlight = highlight.IsHighlight(input, "")
	storage.SaveCommand(r.db, cmd)
	r.lastID = cmd.ID

	// Take screenshot async
	go render.CaptureWindow(r.sessionID, cmd.ID)
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
