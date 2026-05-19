package render

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func ProofDir() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, "shellcast", "proofs")
	os.MkdirAll(dir, 0755)
	return dir
}

// GenerateProofFile renders command + clean output as terminal-styled PNG.
func GenerateProofFile(cmdID int64, input, outputRaw, outputClean string) (string, error) {
	dir := ProofDir()
	htmlFile := filepath.Join(dir, fmt.Sprintf("proof_%d.html", cmdID))
	pngFile := filepath.Join(dir, fmt.Sprintf("proof_%d.png", cmdID))

	body := fmt.Sprintf(`<span style="color:#50fa7b;font-weight:bold">$ %s</span>
%s`, escHTML(input), escHTML(outputClean))

	fullHTML := wrapTerminalHTML(body)
	os.WriteFile(htmlFile, []byte(fullHTML), 0644)

	// Convert HTML → PNG using headless Chromium
	if err := htmlToPNG(htmlFile, pngFile); err != nil {
		// Keep HTML as fallback
		return htmlFile, nil
	}
	os.Remove(htmlFile)
	return pngFile, nil
}

func htmlToPNG(htmlFile, pngFile string) error {
	abs, _ := filepath.Abs(htmlFile)
	browsers := []string{"chromium", "google-chrome", "chromium-browser"}
	for _, b := range browsers {
		if path, _ := exec.LookPath(b); path != "" {
			return exec.Command(path, "--headless", "--no-sandbox",
				"--screenshot="+pngFile, "--window-size=900,600",
				"file://"+abs).Run()
		}
	}
	return fmt.Errorf("no browser found for PNG conversion")
}

func wrapTerminalHTML(body string) string {
	return fmt.Sprintf(`<!DOCTYPE html><html><head><style>
body { background:#0d1117; color:#c9d1d9; font-family:'JetBrains Mono','Fira Code','Cascadia Code','Consolas',monospace; font-size:14px; line-height:1.5; margin:0; padding:0; }
.window { background:#161b22; border-radius:10px; margin:16px; overflow:hidden; border:1px solid #30363d; box-shadow:0 8px 40px rgba(0,0,0,0.5); }
.titlebar { background:#21262d; padding:10px 16px; display:flex; align-items:center; gap:8px; border-bottom:1px solid #30363d; }
.dot { width:12px; height:12px; border-radius:50%%; display:inline-block; }
.red { background:#ff5f56; } .yellow { background:#ffbd2e; } .green { background:#27c93f; }
.content { padding:16px 20px; white-space:pre-wrap; word-wrap:break-word; }
</style></head><body><div class="window"><div class="titlebar">
<span class="dot red"></span><span class="dot yellow"></span><span class="dot green"></span>
</div><div class="content">%s</div></div></body></html>`, body)
}

func escHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
