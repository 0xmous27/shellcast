package strip

import (
	"regexp"
	"strings"
)

var (
	reOSC   = regexp.MustCompile(`\x1b\][^\x07\x1b]*(\x07|\x1b\\)`)
	reSixel = regexp.MustCompile(`(?s)\x1bP[^\x1b]*\x1b\\`)
	reANSI  = regexp.MustCompile(`\x1b\[[0-9;?]*[A-Za-z]`)
	reCtrl  = regexp.MustCompile(`[\x00-\x08\x0b\x0c\x0e-\x1f\x7f]`)

	// Common prompt patterns to strip from output
	rePrompt = regexp.MustCompile(`^[┌└│├].*[$#]\s*$`)
	reKali   = regexp.MustCompile(`^\s*[┌└]──\(.*\)-\[.*\]\s*$`)
	reBasic  = regexp.MustCompile(`^.*@.*[$#]\s*$`)
)

// Clean removes all escape sequences and prompt lines, returns human-readable text
func Clean(s string) string {
	s = reSixel.ReplaceAllString(s, "")
	s = reOSC.ReplaceAllString(s, "")
	s = reANSI.ReplaceAllString(s, "")
	s = reCtrl.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")

	// Filter lines
	lines := strings.Split(s, "\n")
	var out []string
	blanks := 0
	for _, l := range lines {
		l = strings.TrimRight(l, " \t")

		// Skip prompt lines
		if isPromptLine(l) {
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

func isPromptLine(l string) bool {
	if reKali.MatchString(l) {
		return true
	}
	if rePrompt.MatchString(l) {
		return true
	}
	// Lines that are just the echoed command (starts with the input we already have)
	// Skip lines that look like "└─$ <command>"
	if strings.Contains(l, "└─$") || strings.Contains(l, "└─#") {
		return true
	}
	return false
}

// CleanInput strips escape sequences that leak into captured input
func CleanInput(s string) string {
	s = reOSC.ReplaceAllString(s, "")
	s = reANSI.ReplaceAllString(s, "")
	s = reCtrl.ReplaceAllString(s, "")
	// Remove raw bracket sequences: [digits;digits;...letter
	reBracket := regexp.MustCompile(`\[[\d;?]*[A-Za-z]`)
	s = reBracket.ReplaceAllString(s, "")
	// Remove ]digits;... terminal responses
	reTermResp := regexp.MustCompile(`\][\d]+;[^\\]*\\?`)
	s = reTermResp.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}
