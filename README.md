# ⚡ ShellCast

Terminal session recorder for pentesters. Record everything, screenshot every command, export evidence for reports.

## The Problem

You pop a shell, run 50 commands, escalate privileges... then 3 hours later you're writing the report and can't remember what you ran or forgot to screenshot.

## The Solution

```bash
shellcast start client-x    # wraps your shell, records + screenshots everything
# ... do your normal pentesting ...
#mark privesc               # bookmark important moments
exit                        # stop recording

shellcast show              # see all commands (clean text)
shellcast highlights        # auto-detected interesting commands
shellcast proof 4           # open screenshot of command #4
shellcast proof --marks     # list all bookmarked screenshots
shellcast export > report.md  # markdown report
```

## Install

```bash
git clone https://github.com/0xmous27/shellcast
cd shellcast
go build -o shellcast ./cmd/
sudo mv shellcast /usr/local/bin/
```

### Requirements (Linux)

```bash
sudo apt install scrot xdotool
```

## How It Works

1. `shellcast start` wraps your shell via PTY — your terminal looks and behaves 100% normal
2. After every command, it silently takes a **real screenshot** of your terminal window
3. It also stores the clean text output for search and markdown export
4. `#mark <tag>` bookmarks the current moment (tags the previous command)
5. After the session, browse screenshots, search commands, export evidence

## Usage

### Record
```bash
shellcast start <engagement-name>
```

### During Session
```bash
#mark initial-access     # bookmark this moment
#mark privesc
#mark lateral-move
```

### After Session
```bash
shellcast sessions              # list all sessions
shellcast show [session-id]     # all commands (clean text)
shellcast highlights [id]       # auto-highlighted commands
shellcast marks [id]            # bookmarked commands only
shellcast search "password"     # search input and output
shellcast mark 12 "creds"       # retroactively mark command #12
```

### Evidence Export
```bash
shellcast proof 4               # open screenshot #4
shellcast proof 3-7             # list screenshots for commands 3-7
shellcast proof --marks         # list bookmarked screenshots
shellcast proof --all           # list all screenshots
shellcast export > report.md    # markdown report (highlights)
```

## Auto-Highlighting

ShellCast automatically detects interesting commands:
- Security keywords: `whoami`, `root`, `password`, `shell`, `dump`, `ssh`, `sudo`, `nmap`, `sqlmap`...
- Commands with significant output (>5 lines)
- Skips noise: `ls`, `cd`, `pwd`, `clear`, `history`

## Storage

```
~/.shellcast/
├── shellcast.db                    # SQLite (commands, sessions)
└── screenshots/
    └── <session-id>/
        ├── cmd_0001.png            # real terminal screenshot
        ├── cmd_0002.png
        └── ...
```

Everything local. Nothing leaves your machine.

## Use Cases

- **CPTS/OSCP reports** — real terminal screenshots ready to paste
- **Bug bounty writeups** — evidence with exact commands
- **Medium/blog posts** — professional terminal screenshots
- **Client deliverables** — timestamped proof of exploitation
- **Twitter/X posts** — share your wins with real screenshots

## Roadmap

- [ ] `shellcast replay` — replay session in terminal (like asciinema)
- [ ] `shellcast proof --copy` — copy screenshot to clipboard
- [ ] Auto-chapters (detect recon → exploit → post-ex phases)
- [ ] Team sharing (encrypted session export)
- [ ] Tmux pane support

## License

MIT
