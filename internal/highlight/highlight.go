package highlight

import "strings"

var keywords = []string{
	"whoami", "root", "password", "ssh", "sqlmap", "ffuf", "nuclei",
	"linpeas", "sudo", "token", "shell", "nmap", "hashcat", "john",
	"hydra", "meterpreter", "reverse", "dump", "cred", "privesc",
	"bloodhound", "mimikatz", "secretsdump",
}

var noise = map[string]bool{
	"ls": true, "cd": true, "pwd": true, "clear": true, "history": true,
	"exit": true, "fg": true, "bg": true, "jobs": true, "cls": true,
}

func IsHighlight(input, cleanOutput string) bool {
	base := strings.Fields(strings.TrimSpace(input))
	if len(base) > 0 && noise[base[0]] {
		return false
	}
	lower := strings.ToLower(input + " " + cleanOutput)
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	if strings.Count(cleanOutput, "\n") > 5 {
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
