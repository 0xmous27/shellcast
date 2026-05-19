# ⚡ ShellCast

Terminal session evidence recorder for pentesters. Record commands + output, search history, generate professional terminal-style PNG screenshots and markdown reports.

## The Problem

You finish a pentest session and need to write a report. You forgot to screenshot. You can't remember exact commands. Terminal history is gone.

## The Solution

```bash
shellcast start engagement-1
# ... normal pentesting ...
#mark privesc
exit

shellcast show              # command timeline
shellcast search "root"     # find moments
shellcast proof 12          # generate PNG screenshot
shellcast export > report.md
```

---

## Install

```bash
git clone https://github.com/0xmous27/shellcast
cd shellcast
go build -o shellcast ./cmd/
sudo mv shellcast /usr/local/bin/
```

**Requirements:** Go 1.21+ and `gcc` (for SQLite).

---

## Usage

### Record

```bash
shellcast start <name>
```

Your shell works 100% normally. No modified prompt, no lag, no broken behavior. ShellCast records silently via PTY.

### Bookmark (during session)

```bash
#mark initial-access
#mark privesc
#mark domain-admin
```

Tags the **previous** command. Doesn't execute in shell.

### Review

```bash
shellcast show              # full timeline
shellcast highlights        # auto-detected interesting commands
shellcast marks             # bookmarked only
shellcast search "password" # search input + output
shellcast sessions          # list all sessions
```

### Export

```bash
shellcast export > report.md        # markdown report
shellcast proof <cmd-id>            # PNG terminal screenshot
```

### Retroactive Mark

```bash
shellcast mark 15 "creds-found"
```

---

## PNG Proof Screenshots

`shellcast proof <id>` generates a **professional terminal-style PNG**:

- Dark background
- Monospace font
- Window chrome (macOS-style dots)
- Green prompt, white output
- Ready for reports, Medium posts, Twitter

Output: `~/shellcast/proofs/proof_<id>.png`

---

## Auto-Highlighting

Flags commands containing:
`whoami`, `root`, `password`, `ssh`, `sqlmap`, `nuclei`, `linpeas`, `sudo`, `nmap`, `hashcat`, `hydra`, `meterpreter`, `dump`, `cred`...

Ignores noise: `ls`, `cd`, `pwd`, `clear`, `history`

Also highlights commands with significant output (>5 lines).

---

## Storage

```
~/.shellcast/
└── shellcast.db        # SQLite — commands, output, timestamps
~/shellcast/
└── proofs/             # Generated PNG screenshots
```

Everything local. No cloud. No telemetry.

---

## How It Works

1. `shellcast start` spawns your shell inside a PTY
2. Stdin keystrokes are captured to reconstruct commands (Enter, backspace, Ctrl+C handled)
3. PTY output is captured and stored (both raw + cleaned)
4. On exit, session is saved
5. `shellcast proof` renders stored output as terminal-style PNG using Go image libraries

**No shell modification.** No aliases, no PROMPT_COMMAND, no rcfile injection. Pure PTY passthrough.

---

## Testing

```bash
# Build
go build -o shellcast ./cmd/

# Record
rm -f ~/.shellcast/shellcast.db
./shellcast start test

# Inside session:
echo hello
ls /tmp
cat /etc/passwd
id
whoami
#mark test-proof
exit

# Review:
./shellcast show
./shellcast highlights
./shellcast marks
./shellcast search "passwd"

# Generate proof PNG:
./shellcast proof 3

# Export markdown:
./shellcast export > report.md
```

---

## Roadmap

- [ ] Per-command exit code detection
- [ ] Arrow key / escape sequence filtering in input
- [ ] Zsh/Fish support
- [ ] ANSI color rendering in PNG proofs
- [ ] `shellcast replay` — terminal playback
- [ ] Encrypted session export

---

## License

MIT
