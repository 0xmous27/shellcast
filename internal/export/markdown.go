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
	sb.WriteString(fmt.Sprintf("**Session:** %s  \n", session.ID))
	sb.WriteString(fmt.Sprintf("**Started:** %s  \n", session.StartedAt.Format("2006-01-02 15:04:05")))
	if !session.EndedAt.IsZero() {
		dur := session.EndedAt.Sub(session.StartedAt)
		sb.WriteString(fmt.Sprintf("**Duration:** %s  \n", dur.Round(time.Second)))
	}
	sb.WriteString(fmt.Sprintf("**Commands:** %d  \n\n---\n\n", len(commands)))

	for _, cmd := range commands {
		marker := ""
		if cmd.Marked && cmd.Tag != "" {
			marker = fmt.Sprintf(" `[%s]`", cmd.Tag)
		}
		sb.WriteString(fmt.Sprintf("### #%d — %s%s\n\n", cmd.ID, cmd.Timestamp.Format("15:04:05"), marker))
		sb.WriteString("```bash\n")
		sb.WriteString(fmt.Sprintf("$ %s\n", cmd.Input))
		if cmd.OutputClean != "" {
			out := cmd.OutputClean
			lines := strings.Split(out, "\n")
			if len(lines) > 30 {
				out = strings.Join(lines[:30], "\n") + "\n... [truncated]"
			}
			sb.WriteString(out + "\n")
		}
		sb.WriteString("```\n\n")
	}
	return sb.String()
}
