package export

import (
	"fmt"
	"strings"
	"time"

	"github.com/0xmous27/shellcast/pkg/models"
)

func ToMarkdown(session *models.Session, commands []models.Command) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# ShellCast Report: %s\n\n", session.Name))
	sb.WriteString(fmt.Sprintf("**Started:** %s  \n", session.StartedAt.Format("2006-01-02 15:04:05")))
	if !session.EndedAt.IsZero() {
		sb.WriteString(fmt.Sprintf("**Duration:** %s  \n", session.EndedAt.Sub(session.StartedAt).Round(time.Second)))
	}
	sb.WriteString(fmt.Sprintf("**Commands:** %d  \n\n---\n\n", len(commands)))

	for _, cmd := range commands {
		tag := ""
		if cmd.Marked && cmd.Tag != "" {
			tag = fmt.Sprintf(" `[%s]`", cmd.Tag)
		}
		sb.WriteString(fmt.Sprintf("## Command: %s%s\n\n", cmd.Input, tag))
		sb.WriteString(fmt.Sprintf("**Time:** %s  \n\n", cmd.Timestamp.Format("2006-01-02 15:04:05")))
		sb.WriteString("```\n")
		sb.WriteString(fmt.Sprintf("$ %s\n", cmd.Input))
		if cmd.OutputClean != "" {
			lines := strings.Split(cmd.OutputClean, "\n")
			if len(lines) > 40 {
				sb.WriteString(strings.Join(lines[:40], "\n"))
				sb.WriteString(fmt.Sprintf("\n... [%d more lines]\n", len(lines)-40))
			} else {
				sb.WriteString(cmd.OutputClean + "\n")
			}
		}
		sb.WriteString("```\n\n")
	}
	return sb.String()
}
