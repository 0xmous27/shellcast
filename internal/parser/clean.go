package parser

import (
	"regexp"
	"strings"
)

var (
	reANSI  = regexp.MustCompile(`\x1b\[[0-9;?]*[A-Za-z]`)
	reOSC   = regexp.MustCompile(`\x1b\][^\x07\x1b]*(\x07|\x1b\\)`)
	reSixel = regexp.MustCompile(`(?s)\x1bP[^\x1b]*\x1b\\`)
	reCtrl  = regexp.MustCompile(`[\x00-\x08\x0b\x0c\x0e-\x1f\x7f]`)
)

// Clean converts raw PTY output into readable searchable text.
func Clean(raw string) string {
	s := reSixel.ReplaceAllString(raw, "")
	s = reOSC.ReplaceAllString(s, "")
	s = reANSI.ReplaceAllString(s, "")
	s = reCtrl.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")

	lines := strings.Split(s, "\n")
	var out []string
	blanks := 0
	for _, l := range lines {
		l = strings.TrimRight(l, " \t")
		// Skip prompt lines
		if strings.Contains(l, "└─$") || strings.Contains(l, "└─#") {
			continue
		}
		if strings.HasPrefix(l, "┌──(") {
			continue
		}
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

// CleanInput removes terminal response sequences from captured input
func CleanInput(s string) string {
	// Remove ]digits;...\ sequences (terminal query responses)
	reResp := regexp.MustCompile(`\][\d]+;[^\\]*\\`)
	s = reResp.ReplaceAllString(s, "")
	// Remove [digits;...letter sequences (CSI responses without ESC)
	reBracket := regexp.MustCompile(`\[[\d;?]*[A-Za-z]`)
	s = reBracket.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}
