package render

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func ScreenshotDir(sessionID string) string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".shellcast", "screenshots", sessionID)
	os.MkdirAll(dir, 0755)
	return dir
}

// CaptureWindow takes a real screenshot of the terminal window.
// On native Linux: uses scrot/maim/import (real pixel screenshot).
// On WSL: uses PowerShell to capture the Windows Terminal window.
func CaptureWindow(sessionID string, cmdID int) (string, error) {
	dir := ScreenshotDir(sessionID)
	filename := filepath.Join(dir, fmt.Sprintf("cmd_%04d.png", cmdID))

	// Small delay to let output render
	time.Sleep(200 * time.Millisecond)

	if isWSL() {
		return captureWSL(filename)
	}
	return captureLinux(filename)
}

// captureLinux uses native X11 screenshot tools (real PrtScr-quality screenshots)
func captureLinux(filename string) (string, error) {
	// scrot: most common on Kali/Debian
	if run("scrot", "-u", "-o", filename) == nil {
		return filename, nil
	}
	// maim + xdotool
	if wid := getActiveWindow(); wid != "" {
		if run("maim", "-i", wid, filename) == nil {
			return filename, nil
		}
	}
	// import (ImageMagick)
	if wid := getActiveWindow(); wid != "" {
		if run("import", "-window", wid, filename) == nil {
			return filename, nil
		}
	}
	return "", fmt.Errorf("no screenshot tool found")
}

// captureWSL uses PowerShell to screenshot the active window on Windows
func captureWSL(filename string) (string, error) {
	// Convert WSL path to Windows path for saving
	absPath, _ := filepath.Abs(filename)

	// PowerShell script to capture the foreground window
	ps := fmt.Sprintf(`
Add-Type @"
using System;
using System.Runtime.InteropServices;
using System.Drawing;
using System.Drawing.Imaging;
public class Screenshot {
    [DllImport("user32.dll")]
    public static extern IntPtr GetForegroundWindow();
    [DllImport("user32.dll")]
    public static extern bool GetWindowRect(IntPtr hWnd, out RECT rect);
    [StructLayout(LayoutKind.Sequential)]
    public struct RECT { public int Left, Top, Right, Bottom; }
    public static void Capture(string path) {
        IntPtr hwnd = GetForegroundWindow();
        RECT r;
        GetWindowRect(hwnd, out r);
        int w = r.Right - r.Left;
        int h = r.Bottom - r.Top;
        if (w <= 0 || h <= 0) return;
        Bitmap bmp = new Bitmap(w, h);
        Graphics g = Graphics.FromImage(bmp);
        g.CopyFromScreen(r.Left, r.Top, 0, 0, new Size(w, h));
        bmp.Save(path, ImageFormat.Png);
    }
}
"@ -ReferencedAssemblies System.Drawing,System.Drawing.Common
[Screenshot]::Capture("%s")
`, toWindowsPath(absPath))

	cmd := exec.Command("powershell.exe", "-NoProfile", "-Command", ps)
	if err := cmd.Run(); err != nil {
		return "", err
	}

	// Check if file was created
	if _, err := os.Stat(filename); err == nil {
		return filename, nil
	}
	return "", fmt.Errorf("WSL screenshot failed")
}

func isWSL() bool {
	data, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}
	lower := strings.ToLower(string(data))
	return strings.Contains(lower, "microsoft") || strings.Contains(lower, "wsl")
}

func toWindowsPath(linuxPath string) string {
	// Convert /home/user/... to \\wsl$\distro\home\user\... or use wslpath
	out, err := exec.Command("wslpath", "-w", linuxPath).Output()
	if err != nil {
		// Fallback: try direct conversion
		return strings.ReplaceAll(linuxPath, "/", "\\")
	}
	return strings.TrimSpace(string(out))
}

func getActiveWindow() string {
	out, err := exec.Command("xdotool", "getactivewindow").Output()
	if err != nil || len(out) == 0 {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}
