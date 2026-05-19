package highlight

import "strings"

var keywords = []string{
	"whoami", "id", "root", "admin", "password", "passwd", "shadow",
	"shell", "reverse", "nc ", "ncat", "netcat", "bash -i",
	"dump", "extract", "secret", "token", "key", "cred",
	"ssh ", "rdp", "pivot", "tunnel", "proxychains",
	"exploit", "payload", "msfconsole", "meterpreter",
	"sudo", "suid", "capability", "privesc",
	"flag", "proof", "user.txt", "root.txt",
	"nmap", "gobuster", "sqlmap", "hydra", "hashcat", "john",
}

var noise = map[string]bool{
	"ls": true, "cd": true, "pwd": true, "clear": true, "cls": true,
	"history": true, "exit": true, "fg": true, "bg": true, "jobs": true,
}

func IsHighlight(input, outputClean string) bool {
	cmd := strings.TrimSpace(input)
	base := strings.Fields(cmd)
	if len(base) > 0 && noise[base[0]] {
		// Exception: cat sensitive files
		if strings.Contains(cmd, "shadow") || strings.Contains(cmd, "passwd") || strings.Contains(cmd, "flag") {
			return true
		}
		return false
	}

	lower := strings.ToLower(cmd + " " + outputClean)
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}

	// Significant output
	lines := strings.Count(outputClean, "\n")
	if lines > 5 && lines < 200 {
		return true
	}
	return false
}

func IsMark(input string) (bool, string) {
	t := strings.TrimSpace(input)
	if strings.HasPrefix(t, "#mark") {
		tag := strings.TrimSpace(strings.TrimPrefix(t, "#mark"))
		if tag == "" {
			tag = "marked"
		}
		return true, tag
	}
	return false, ""
}
