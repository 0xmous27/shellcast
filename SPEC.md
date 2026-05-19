# ShellCast — Project Spec

## Core Objective

A CLI tool that **passively records your terminal session** and lets you **retroactively extract report-ready evidence** (screenshots, formatted commands, markdown) without breaking your workflow.

**The problem:** You finish a 2-hour pentest session and need to write a report. You forgot to screenshot. You can't remember exact commands. Scrolling terminal history is painful.

**The solution:** `shellcast start` → do your thing → `exit` → pull screenshots and evidence after the fact.

---

## What "Output" Means

The tool captures your terminal session and produces:

1. **PNG screenshots** — rendered terminal output (like what you see on screen), not raw escape codes
2. **Formatted markdown** — clean command + output blocks for reports
3. **Searchable history** — find the exact moment something happened

The stored data should look like what a HUMAN sees in the terminal, not raw PTY garbage.

---

## Key Features (Priority Order)

### 1. Record Session (transparent shell wrapper)
```bash
shellcast start <engagement-name>
# your normal shell appears — do anything
exit
```
- Wraps your shell via PTY
- Records everything silently
- Your terminal picture, colors, prompt — all display normally
- Zero friction, zero workflow change

### 2. Bookmark Moments (during session)
```bash
#mark privesc          # type as a command — tags that moment
```
- One-word inline bookmarking
- Tags the PREVIOUS command (the one that mattered)

### 3. Review After Session
```bash
shellcast show              # timeline of all commands (clean text)
shellcast highlights        # auto-detected interesting moments
shellcast marks             # only bookmarked moments
shellcast search "root"     # find by keyword in input OR output
```

### 4. Export Evidence
```bash
shellcast proof 14-18           # export commands 14-18 as PNG screenshot
shellcast proof --search "root" # screenshot the moment you got root
shellcast export                # full markdown report (highlights only)
shellcast export --full         # everything
```

### 5. Smart Filtering (automatic)
- Auto-highlight: commands with security keywords (whoami, root, password, shell, dump, ssh...)
- Auto-ignore noise: ls, cd, pwd, clear, history
- Significant output detection (>5 lines)
- Skip commands that errored with typos

### 6. Live Bookmarking Hotkey (stretch goal)
- `Ctrl+Shift+B` during session = silent bookmark

---

## Output Format Requirements

### `shellcast show` — Clean readable text
```
⚡ #1   02:14:03
  $ nmap -p- 10.10.10.5
  PORT    STATE SERVICE
  22/tcp  open  ssh
  80/tcp  open  http
  ... (3 more lines)

  #2   02:14:45
  $ gobuster dir -u http://10.10.10.5 -w /usr/share/wordlists/common.txt
  /admin (Status: 200)
  /login (Status: 302)

🔖 #3   02:15:10  [sqli]
  $ sqlmap -u "http://10.10.10.5/page?id=1" --os-shell
  os-shell> whoami
  www-data
```

- NO escape codes, NO raw PTY data, NO image/sixel data
- Clean command + clean output (as a human would read it)
- Markers for highlights (⚡) and bookmarks (🔖)
- Truncated output (show first 5 lines, indicate more)

### `shellcast proof` — PNG Screenshot
- Renders terminal output as an image (dark background, monospace font)
- Looks like an actual terminal screenshot
- Suitable for pasting into a pentest report (PDF/Word)

### `shellcast export` — Markdown
```markdown
## Initial Access — SQLi
**Time:** 2026-05-19 02:15:10
**Tag:** sqli

\```bash
$ sqlmap -u "http://10.10.10.5/page?id=1" --os-shell
os-shell> whoami
www-data
\```
```

---

## Technical Approach

### Recording
- PTY wrapper (like `script` but structured)
- Store RAW terminal data (with escapes) for screenshot rendering
- Store CLEANED text (stripped escapes) for search/show/export
- SQLite database per user (~/.shellcast/shellcast.db)

### Screenshot Generation (PNG)
- Render stored terminal output to PNG using a terminal renderer
- Options: use Go image library to draw monospace text on dark background
- OR: use a library like `go-ansi` to parse ANSI colors and render with color

### Escape Stripping (for text output)
- Remove ANSI color/cursor codes
- Remove OSC sequences (terminal integration markers)
- Remove sixel/image data (terminal pictures)
- Keep only printable text + newlines

### Input Detection
- Track Enter keypresses to delimit commands
- Handle backspace, Ctrl+C, Ctrl+U
- Skip shell startup output (prompt picture, motd, etc.)
- Detect actual command vs prompt echo

---

## Architecture

```
shellcast (single binary)
├── cmd/main.go              ← CLI commands
├── internal/
│   ├── recorder/            ← PTY wrapper, input/output capture
│   ├── storage/             ← SQLite (sessions, commands)
│   ├── highlight/           ← Auto-detect interesting commands
│   ├── strip/               ← ANSI/OSC/sixel escape removal
│   ├── render/              ← PNG screenshot generation
│   └── export/              ← Markdown/report generation
└── pkg/models/              ← Data structures
```

### Database Schema
```sql
sessions: id, name, started_at, ended_at, commands_count
commands: id, session_id, input, output_raw, output_clean, exit_code,
          timestamp, duration_ms, marked, tag, highlight
```

- `output_raw` = full PTY output (for screenshot rendering with colors)
- `output_clean` = stripped text (for search, show, export)

---

## What Makes This Different

| Existing Tool | What it does | Why it's not enough |
|---|---|---|
| `script` | Records raw terminal | Unreadable dump, no search, no screenshots |
| `asciinema` | Replayable recording | Can't extract screenshots, not for reports |
| Manual screenshots | Proof for report | Breaks flow, easy to forget |
| CherryTree/Obsidian | Manual notes | You have to remember to paste |
| Ghostwriter | Report platform | Doesn't capture anything automatically |

ShellCast = **automatic capture** + **retroactive evidence extraction** + **report-ready output**

---

## Non-Goals (for v1)

- NOT a terminal emulator (wraps your existing shell)
- NOT a keylogger (runs on YOUR machine, records YOUR session)
- NOT cloud-based (everything local, SQLite)
- NOT real-time collaboration (single user)
- NOT AI-powered (rule-based highlighting, no LLM)

---

## Success Criteria

1. `shellcast start` works transparently — user's terminal looks and behaves 100% normal
2. `shellcast show` outputs clean, readable text (no escape garbage)
3. `shellcast proof` generates a PNG that looks like a real terminal screenshot
4. `shellcast search` finds commands by keyword in input or output
5. `shellcast export` produces copy-paste-ready markdown for reports
6. Works with any shell (bash, zsh, fish)
7. Handles terminal pictures/sixel graphics without corrupting stored data
8. Single static binary, no runtime dependencies
