package render

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/0xmous27/shellcast/internal/parser"
)

const (
	fontSize = 15
	lineH    = 24
	padX     = 24
	padY     = 18
)

// GenerateProof renders command + output as a tight-cropped PNG.
// If outName is empty, saves to current directory as shellcast_proof_<id>.png
func GenerateProof(cmdID int64, input, outputRaw, outputClean, outName string) (string, error) {
	var pngFile string
	if outName != "" {
		pngFile = outName
	} else {
		pngFile = fmt.Sprintf("shellcast_proof_%d.png", cmdID)
	}

	// Temp HTML
	htmlFile := pngFile + ".tmp.html"

	// Clean: remove echoed command, trailing blanks
	cleanedOutput := parser.CleanForProof(input, outputClean)

	// Build content lines
	var lines []string
	lines = append(lines, fmt.Sprintf(`<span class="p">$ %s</span>`, escHTML(input)))
	if cleanedOutput != "" {
		for _, l := range strings.Split(cleanedOutput, "\n") {
			lines = append(lines, escHTML(l))
		}
	}

	body := strings.Join(lines, "\n")

	// Exact dimensions from actual content
	maxLen := 0
	for _, l := range lines {
		// Strip HTML tags for length calc
		plain := strings.ReplaceAll(l, `<span class="p">`, "")
		plain = strings.ReplaceAll(plain, `</span>`, "")
		plain = strings.ReplaceAll(plain, "&amp;", "&")
		plain = strings.ReplaceAll(plain, "&lt;", "<")
		plain = strings.ReplaceAll(plain, "&gt;", ">")
		if len(plain) > maxLen {
			maxLen = len(plain)
		}
	}
	width := padX*2 + int(float64(maxLen)*9.0) + 10
	if width < 400 {
		width = 400
	}
	lineCount := len(lines)
	height := padY*2 + (lineCount+3)*lineH

	html := fmt.Sprintf(`<!DOCTYPE html><html><head><style>
*{margin:0;padding:0;box-sizing:border-box}
html,body{background:#0d1117;overflow:hidden;width:%dpx;height:%dpx}
pre{
  font-family:'JetBrains Mono','Fira Code','Cascadia Code','SF Mono','Consolas',monospace;
  font-size:15px;
  line-height:24px;
  color:#c9d1d9;
  padding:18px 24px;
  margin:0;
  white-space:pre;
  -webkit-font-smoothing:antialiased;
}
.p{color:#50fa7b;font-weight:bold}
</style></head><body><pre>%s</pre></body></html>`,
		width, height, body)

	os.WriteFile(htmlFile, []byte(html), 0644)

	// Render with Chromium — exact size, no scrollbars
	err := chromiumScreenshot(htmlFile, pngFile, width, height)
	os.Remove(htmlFile)
	if err != nil {
		return "", err
	}
	return pngFile, nil
}

func chromiumScreenshot(htmlFile, pngFile string, w, h int) error {
	abs, _ := filepath.Abs(htmlFile)
	size := fmt.Sprintf("%d,%d", w, h)
	browsers := []string{"chromium", "google-chrome", "chromium-browser"}
	for _, b := range browsers {
		if path, _ := exec.LookPath(b); path != "" {
			return exec.Command(path,
				"--headless", "--no-sandbox", "--disable-gpu",
				"--hide-scrollbars",
				"--force-device-scale-factor=2",
				"--screenshot="+pngFile,
				"--window-size="+size,
				"file://"+abs,
			).Run()
		}
	}
	return fmt.Errorf("chromium not found")
}

func escHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
