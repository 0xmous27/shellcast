package parser

import (
	"os"
	"path/filepath"
	"strings"
)

// LoadExcludePatterns reads user-defined exclude patterns from ~/.shellcast/exclude
func LoadExcludePatterns() []string {
	home, _ := os.UserHomeDir()
	data, err := os.ReadFile(filepath.Join(home, ".shellcast", "exclude"))
	if err != nil {
		return nil
	}
	var patterns []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			patterns = append(patterns, line)
		}
	}
	return patterns
}

// ShouldExcludeLine checks if a line matches any user-defined exclude pattern
func ShouldExcludeLine(line string, patterns []string) bool {
	for _, p := range patterns {
		if strings.Contains(line, p) {
			return true
		}
	}
	return false
}
