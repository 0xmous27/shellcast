# ShellCast — SESSION FILE
# Last updated: 2026-05-20
# Drop this at the start of any new session to continue without re-explaining.
# ⚠️ KEEP LOCAL ONLY — in .gitignore

---

## PROJECT IDENTITY

- **Name:** ShellCast
- **Repo:** https://github.com/0xmous27/shellcast (public)
- **Stack:** Go, SQLite (pure Go via modernc.org/sqlite), Chromium (headless PNG)
- **Purpose:** Terminal session evidence recorder for pentesters
- **Install:** `go install -v github.com/0xmous27/shellcast/cmd@latest`
- **Binary:** Single static binary, no CGo, no runtime deps except Chromium for `shot`

---

## WHAT'S BUILT (v0.1.0)

### Commands (11 total)
- `shellcast start <name>` — PTY shell wrapper, records session
- `shellcast show [session-id]` — command timeline (clean text)
- `shellcast highlights` — auto-flagged security commands
- `shellcast marks` — bookmarked commands
- `shellcast search <query>` — search input + output
- `shellcast shot <id|range>` — generate PNG screenshot
- `shellcast shot 3 -o file.png` — custom filename
- `shellcast export` — markdown report
- `shellcast sessions` — list all sessions
- `shellcast mark <id> [tag]` — retroactive bookmark
- `shellcast delete <session-id>` — remove session
- `#mark <tag>` — inline bookmark during session (bash comment, silent)

### Architecture
```
shellcast/
├── cmd/main.go                  ← CLI entry (11 commands)
├── internal/
│   ├── recorder/recorder.go     ← PTY wrapper, keystroke capture
│   ├── storage/db.go            ← SQLite CRUD
│   ├── parser/clean.go          ← ANSI strip + CleanInput + CleanForProof
│   ├── highlight/highlight.go   ← Keyword detection
│   ├── render/png.go            ← HTML→PNG via headless Chromium
│   └── export/markdown.go       ← Markdown report
├── pkg/models/models.go         ← Session + Command structs
├── examples/                    ← 4 example PNGs for README
├── README.md, LICENSE, .gitignore
└── go.mod, go.sum
```

### PNG Proof Generation
- Renders command + clean output as HTML
- Chromium headless screenshots it at 2x retina
- Dark bg (#0d1117), green prompt (#50fa7b), white text (#c9d1d9)
- Tight-cropped to content, no extra space
- Saves to current directory

### Storage
- `~/.shellcast/shellcast.db` (SQLite)
- Tables: `sessions` (id, name, started_at, ended_at) + `commands` (id, session_id, input, output_raw, output_clean, exit_code, timestamp, duration_ms, marked, tag, highlight)

---

## KNOWN ISSUES

1. **First command input has escape junk** — terminal response sequences (`]10;rgb:...`, `[8;30;120t`) leak into the first captured command. `CleanInput` strips most but not all.

2. **Bracketed paste mode** — `[200~` and `[201~` sequences appear in input when pasting. Fixed in `CleanInput` regex but needs testing.

3. **Output includes next prompt** — the output for a command includes the prompt that appears after it. Parser strips Kali-style prompts (`┌──(`, `└─$`) but won't work on custom prompts.

4. **No exit code capture** — spec requires it but not implemented.

5. **Arrow keys in input** — if user presses up arrow to recall history, the escape sequence bytes may leak into input buffer.

---

## NEXT SESSION — PRIORITY TASKS

### 1. Fix Command Input Capture (Critical)
The pure PTY keystroke approach is unreliable. Two options:
- **Option A:** Accept shell integration (PROMPT_COMMAND hook) — captures commands perfectly but modifies shell
- **Option B:** Post-process approach — record raw session, reconstruct commands after by detecting prompt patterns

Recommendation: Option A with opt-in flag (`shellcast start --hook`) for reliable capture, keep pure PTY as default.

### 2. Fix Output Attribution
- Strip the echoed command from output (partially done via `CleanForProof`)
- Strip prompt lines for non-Kali prompts (detect `$` or `#` at end of line)

### 3. PNG Quality Polish
- Current rendering works well
- Consider: add subtle rounded corners, slight shadow
- Consider: optional `--theme` flag (dark/light/dracula)

### 4. Community Launch
- Record 30-second demo GIF
- Post on r/netsec, r/oscp, Twitter with GIF
- Add GitHub topics: `pentest`, `terminal`, `screenshot`, `evidence`, `oscp`
- Add GitHub Actions CI (build + test on push)

---

## BACKLOG (Lower Priority)

- [ ] Per-command exit code detection
- [ ] `shellcast replay` — terminal playback
- [ ] Multiple proof themes
- [ ] Copy proof to clipboard (`--clip` flag)
- [ ] PDF report generation
- [ ] Auto-chapters (detect recon → exploit → post-ex phases)
- [ ] Zsh/Fish shell support
- [ ] ANSI color rendering in PNG proofs
- [ ] Encrypted session export
- [ ] GitHub Actions CI/CD
- [ ] Homebrew formula

---

## CREDENTIALS / LOCAL INFO

- **Git identity:** `0xmous27` / `117016087+0xmous27@users.noreply.github.com`
- **Project path:** `/home/oxmous/shellcast`
- **Push:** `git add -A && git commit -m "..." && git push`

---

## COMPLETED (2026-05-20)

- [x] Pure PTY recorder (no shell modification)
- [x] Command detection from keystrokes
- [x] SQLite storage (pure Go, no CGo)
- [x] ANSI/OSC/sixel stripping
- [x] Auto-highlight engine
- [x] #mark bookmarking
- [x] show/highlights/marks/search commands
- [x] Markdown export
- [x] PNG proof generation (Chromium headless, 2x retina)
- [x] Tight-crop PNG (no extra space)
- [x] CleanInput (strips terminal responses, bracketed paste)
- [x] CleanForProof (removes echoed command)
- [x] Custom output filename (-o flag)
- [x] Session delete command
- [x] Show accepts session-id
- [x] One-line install (go install)
- [x] Clean README with examples
- [x] Pushed to GitHub

---

> *"The best evidence is the evidence you didn't have to remember to collect."*
