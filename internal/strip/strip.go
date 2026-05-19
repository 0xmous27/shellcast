package strip

import (
	"regexp"
	"strings"
)

var (
	reOSC    = regexp.MustCompile(`\x1b\][^\x07\x1b]*(\x07|\x1b\\)`)
	reSixel  = regexp.MustCompile(`(?s)\x1bP[^\x1b]*\x1b\\`)
	reANSI   = regexp.MustCompile(`\x1b\[[0-9;?]*[A-Za-z]`)
	reCtrl   = regexp.MustCompile(`[\x00-\x08\x0b\x0c\x0e-\x1f\x7f]`)
)

// Clean removes all escape sequences and returns human-readable text
func Clean(s string) string {
	s = reSixel.ReplaceAllString(s, "")
	s = reOSC.ReplaceAllString(s, "")
	s = reANSI.ReplaceAllString(s, "")
	s = reCtrl.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")

	// Collapse excessive blank lines
	lines := strings.Split(s, "\n")
	var out []string
	blanks := 0
	for _, l := range lines {
		l = strings.TrimRight(l, " \t")
		if l == "" {
			blanks++
			if blanks <= 1 {
				out = append(out, "")
			}
		} else {
			blanks = 0
			out = append(out, l)
		}
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}
