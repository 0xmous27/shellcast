# ⚡ ShellCast

Terminal session recorder for pentesters. Automatically captures **real screenshots** of your terminal after every command. Export evidence for CPTS reports, bug bounty writeups, Medium posts, or client deliverables.

## The Problem

You finish a 2-hour pentest session and need to write a report. You forgot to screenshot. You can't remember exact commands. You're scrolling terminal history trying to reconstruct what happened.

## The Solution

```bash
shellcast start client-x    # wraps your shell, records + screenshots everything
# ... do your normal pentesting ...
#mark privesc               # bookmark important moments
exit                        # stop recording

shellcast show              # see all commands captured
shellcast proof 4           # open screenshot of command #4
shellcast proof --marks     # list bookmarked screenshots
shellcast export > report.md  # markdown report
```

Every screenshot is a **real PNG of your actual terminal window** — not rendered text, not fake terminal graphics. The same quality as pressing PrtScr.

---

## Installation

### Build from source

```bash
git clone https://github.com/0xmous27/shellcast
cd shellcast
go build -o shellcast ./cmd/
sudo mv shellcast /usr/local/bin/
```

### Dependencies

**Required (all platforms):**
```bash
sudo apt install socat
```

**For screenshots on native Linux (Kali, Ubuntu, Parrot with desktop):**
```bash
sudo apt install scrot xdotool
```

**For screenshots on WSL:**
- No extra install needed — uses PowerShell to capture the Windows Terminal window automatically.

---

## Platform Support

| Platform | Command Recording | Screenshots | Quality |
|----------|------------------|-------------|---------|
| **Linux desktop** (Kali, Ubuntu, Parrot) | ✅ | ✅ Real PrtScr-quality | Best — native X11 capture |
| **WSL2** (Windows Terminal) | ✅ | ✅ Via PowerShell | Good — captures Windows Terminal window |
| **SSH session** | ✅ | ❌ No display | Commands only, no screenshots |

---

## Usage

### Record a Session

```bash
shellcast start <engagement-name>
```

Your normal shell appears. Everything works exactly as before — your prompt, colors, terminal picture, aliases, all of it. ShellCast runs invisibly in the background.

### Bookmark Moments (during session)

```bash
#mark initial-access     # tags the previous command as important
#mark privesc
#mark lateral-move
```

### End Session

```bash
exit
```

### Review After Session

```bash
shellcast sessions              # list all recorded sessions
shellcast show [session-id]     # all commands captured
shellcast highlights [id]       # auto-highlighted interesting commands
shellcast marks [id]            # bookmarked commands only
shellcast search "password"     # search by keyword
shellcast mark 12 "creds"       # retroactively mark command #12
```

### Export Evidence

```bash
shellcast proof 4               # open screenshot #4
shellcast proof 3-7             # list screenshots for range
shellcast proof --marks         # bookmarked screenshots
shellcast proof --all           # all screenshots
shellcast export > report.md    # markdown report
```

---

## Auto-Highlighting

ShellCast automatically flags interesting commands:

- **Security keywords:** `whoami`, `root`, `password`, `shell`, `dump`, `ssh`, `sudo`, `nmap`, `sqlmap`, `hashcat`...
- **Skips noise:** `ls`, `cd`, `pwd`, `clear`, `history`

---

## How It Works

1. `shellcast start` spawns your shell inside a PTY (transparent wrapper)
2. A lightweight hook (`PROMPT_COMMAND`) reports each command via unix socket
3. After each command, a real screenshot of your terminal window is saved
4. Commands are stored in SQLite with timestamps
5. On `exit`, session is closed and all data persists

```
~/.shellcast/
├── shellcast.db                    # SQLite (commands, sessions)
└── screenshots/
    └── <session-id>/
        ├── cmd_0001.png            # real terminal screenshot
        ├── cmd_0002.png
        └── ...
```

Everything stays local. Nothing leaves your machine.

---

## Testing

### On WSL

```bash
rm -f ~/.shellcast/shellcast.db
./shellcast start test

# Run commands:
echo hello
ls
id
cat /etc/passwd
#mark test-bookmark
exit

# Check results:
./shellcast show              # should list all commands cleanly
./shellcast highlights        # security-relevant ones flagged
./shellcast marks             # bookmarked ones
./shellcast proof --all       # screenshots (captured via PowerShell)
```

### On Native Linux (Kali desktop)

```bash
sudo apt install scrot xdotool socat
./shellcast start engagement-1

# Do your pentesting...
nmap -p- 10.10.10.5
gobuster dir -u http://10.10.10.5 -w /usr/share/wordlists/common.txt
sqlmap -u "http://10.10.10.5/page?id=1" --os-shell
#mark shell-access
exit

# Evidence ready:
./shellcast proof --marks     # real terminal screenshots
./shellcast export > report.md
ls ~/.shellcast/screenshots/  # PNG files ready for your report
```

---

## Use Cases

- **CPTS/OSCP reports** — real terminal screenshots, timestamped
- **Bug bounty writeups** — exact commands with proof
- **Medium/blog posts** — professional terminal screenshots
- **Client deliverables** — evidence appendix
- **Twitter/X** — share wins with real screenshots
- **Team knowledge** — searchable command history per engagement

---

## Roadmap

- [ ] Per-command output capture (store what each command printed)
- [ ] `shellcast replay` — replay session in terminal
- [ ] `shellcast proof --copy` — copy screenshot to clipboard
- [ ] Auto-chapters (detect recon → exploit → post-ex phases)
- [ ] Zsh/Fish shell support
- [ ] Encrypted session export for team sharing

---

## License

MIT
