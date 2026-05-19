# ⚡ ShellCast

**Terminal session evidence recorder for pentesters.**

Record your terminal sessions, reconstruct clean command history, and generate professional report-ready PNG screenshots — all from a single binary.

> "Never forget to screenshot again."

---

## Example Output

**`shellcast proof 1`** — `whoami`:

![whoami](examples/example_whoami.png)

**`shellcast proof 2`** — `ls -la /etc/ | head -10`:

![ls](examples/example_ls.png)

**`shellcast proof 4`** — `ps aux | head -15`:

![ps](examples/example_ps.png)

**`shellcast proof 5`** — `cat /nonexistent` (error):

![error](examples/example_error.png)

---

## Installation

### Requirements

- **Go 1.21+** (to build)
- **gcc** (for SQLite CGo binding)
- **Chromium** (for PNG proof generation)
- **socat** (optional, for future shell integration features)

### Build

```bash
git clone https://github.com/0xmous27/shellcast.git
cd shellcast
go build -o shellcast ./cmd/
```

### Install system-wide

```bash
sudo mv shellcast /usr/local/bin/
```

### Install dependencies

**Debian/Ubuntu/Kali:**
```bash
sudo apt install chromium
```

**Arch:**
```bash
sudo pacman -S chromium
```

### Verify

```bash
shellcast version
# shellcast v0.1.0
```

---

## Quick Start

```bash
# 1. Start recording
shellcast start my-engagement

# 2. Do your thing (shell works 100% normally)
nmap -p- 10.10.10.5
gobuster dir -u http://10.10.10.5 -w /usr/share/wordlists/common.txt
sqlmap -u "http://10.10.10.5/page?id=1" --os-shell
#mark shell-access
whoami
#mark root-proof
exit

# 3. Review
shellcast show

# 4. Generate evidence
shellcast proof 3-6

# 5. Export report
shellcast export > report.md
```

---

## Usage

### Record a Session

```bash
shellcast start <engagement-name>
```

Your shell works **100% normally** — colors, prompt, aliases, everything. ShellCast records silently via PTY. No shell modification, no hooks, no lag.

Type `exit` to stop recording.

### Bookmark Important Moments

During your session, type:

```bash
#mark <tag>
```

This bookmarks the **previous** command. It doesn't execute in the shell — it's intercepted by ShellCast.

Examples:
```bash
#mark initial-access
#mark privesc
#mark domain-admin
#mark creds-found
```

### Review Commands

```bash
shellcast show                  # full command timeline
shellcast highlights            # auto-highlighted security commands
shellcast marks                 # bookmarked commands only
shellcast search "password"     # search input + output
shellcast sessions              # list all recorded sessions
```

### Generate PNG Proof Screenshots

```bash
shellcast proof 5               # single command
shellcast proof 3-8             # range of commands
```

Output: `~/shellcast/proofs/proof_<id>.png`

Each PNG contains:
- Green `$ command` prompt
- White command output
- Dark background
- Monospace font (JetBrains Mono / Fira Code)
- Tight-cropped to content — no extra space
- 2x retina resolution

### Export Markdown Report

```bash
shellcast export > report.md
```

Generates a structured markdown report with all highlighted/bookmarked commands, ready for GitHub, Notion, or any markdown editor.

### Retroactive Bookmarking

Forgot to `#mark` during the session? Do it after:

```bash
shellcast mark 12 "privesc"
shellcast mark 15 "creds-found"
```

---

## Auto-Highlighting

ShellCast automatically flags commands containing security-relevant keywords:

`whoami` · `root` · `password` · `ssh` · `sqlmap` · `nuclei` · `linpeas` · `sudo` · `nmap` · `hashcat` · `hydra` · `meterpreter` · `dump` · `cred` · `privesc` · `bloodhound` · `mimikatz`

Also highlights commands with significant output (>5 lines).

Ignores noise: `ls` · `cd` · `pwd` · `clear` · `history`

---

## Multiple Terminals

Each terminal gets its own session. Just start ShellCast in each:

```bash
# Tab 1
shellcast start recon

# Tab 2
shellcast start exploit

# Tab 3
shellcast start privesc
```

Review all sessions:
```bash
shellcast sessions
```

---

## Storage

```
~/.shellcast/
└── shellcast.db            # SQLite — commands, output, timestamps

~/shellcast/
└── proofs/                 # Generated PNG screenshots
    ├── proof_1.png
    ├── proof_2.png
    └── ...
```

Everything local. No cloud. No telemetry. No network calls.

---

## How It Works

1. `shellcast start` spawns your shell inside a PTY (transparent wrapper)
2. Keystrokes are captured to reconstruct commands (Enter, backspace, Ctrl+C)
3. PTY output is stored — both raw (with ANSI) and cleaned (readable text)
4. `#mark` commands are intercepted and tag the previous command
5. `shellcast proof` renders stored output as terminal-styled PNG via headless Chromium
6. Everything persists in SQLite for later search/export

**No shell modification.** No `.bashrc` changes, no aliases, no `PROMPT_COMMAND` injection. Pure PTY passthrough.

---

## Use Cases

| Who | Why |
|-----|-----|
| **OSCP/CPTS students** | Evidence screenshots for exam reports |
| **Bug bounty hunters** | Proof of exploitation for submissions |
| **Red teamers** | Client engagement evidence |
| **CTF players** | Writeup screenshots |
| **Pentesters** | Report appendix generation |
| **Security researchers** | Blog post terminal screenshots |

---

## Command Reference

| Command | Description |
|---------|-------------|
| `shellcast start <name>` | Start recording a session |
| `shellcast show` | Show command timeline |
| `shellcast highlights` | Show auto-highlighted commands |
| `shellcast marks` | Show bookmarked commands |
| `shellcast search <query>` | Search commands and output |
| `shellcast proof <id\|range>` | Generate PNG screenshot(s) |
| `shellcast export` | Export markdown report |
| `shellcast sessions` | List all sessions |
| `shellcast mark <id> [tag]` | Mark a command retroactively |
| `shellcast version` | Show version |

---

## Roadmap

- [ ] Per-command exit code tracking
- [ ] Arrow key / escape sequence filtering improvements
- [ ] Zsh and Fish shell support
- [ ] ANSI color rendering in PNG proofs
- [ ] `shellcast replay` — terminal session playback
- [ ] Proof themes (dark, light, Dracula, Nord)
- [ ] Copy proof to clipboard
- [ ] PDF report generation
- [ ] Encrypted session export for team sharing

---

## License

MIT

---

## Author

[@0xmous27](https://github.com/0xmous27)

> *"The best evidence is the evidence you didn't have to remember to collect."*
