package render

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func ScreenshotDir(sessionID string) string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".shellcast", "screenshots", sessionID)
	os.MkdirAll(dir, 0755)
	return dir
}

// CaptureWindow takes a real screenshot of the focused terminal window (Linux X11).
func CaptureWindow(sessionID string, cmdID int) (string, error) {
	dir := ScreenshotDir(sessionID)
	filename := filepath.Join(dir, fmt.Sprintf("cmd_%04d.png", cmdID))

	// Small delay to let output render on screen
	time.Sleep(150 * time.Millisecond)

	// Try scrot (most common on Kali/Debian)
	if run("scrot", "-u", "-o", filename) == nil {
		return filename, nil
	}

	// Try maim + xdotool
	if wid := getActiveWindow(); wid != "" {
		if run("maim", "-i", wid, filename) == nil {
			return filename, nil
		}
	}

	// Try import (ImageMagick) + xdotool
	if wid := getActiveWindow(); wid != "" {
		if run("import", "-window", wid, filename) == nil {
			return filename, nil
		}
	}

	return "", fmt.Errorf("screenshot failed — install scrot: sudo apt install scrot")
}

func getActiveWindow() string {
	out, err := exec.Command("xdotool", "getactivewindow").Output()
	if err != nil || len(out) == 0 {
		return ""
	}
	return string(out[:len(out)-1])
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}
