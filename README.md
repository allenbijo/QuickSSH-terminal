# QuickSSH Terminal

Terminal-resident equivalent of [QuickSSH](https://github.com/allenbijo/QuickSSH) (the Electron version). Same SSH port-forward tunnels, same JSON config file - but a polished, idle-zero-CPU TUI built on Bubbletea + Lipgloss instead of Electron + React.

```
  /$$$$$$            /$$           /$$        /$$$$$$   /$$$$$$  /$$   /$$
 /$$__  $$          |__/          | $$       /$$__  $$ /$$__  $$| $$  | $$
| $$  \ $$ /$$   /$$ /$$  /$$$$$$$| $$   /$$| $$  \__/| $$  \__/| $$  | $$
| $$  | $$| $$  | $$| $$ /$$_____/| $$  /$$/|  $$$$$$ |  $$$$$$ | $$$$$$$$
| $$  | $$| $$  | $$| $$| $$      | $$$$$$/  \____  $$ \____  $$| $$__  $$
| $$/$$ $$| $$  | $$| $$| $$      | $$_  $$  /$$  \ $$ /$$  \ $$| $$  | $$
|  $$$$$$/|  $$$$$$/| $$|  $$$$$$$| $$ \  $$|  $$$$$$/|  $$$$$$/| $$  | $$
 \____ $$$ \______/ |__/ \_______/|__/  \__/ \______/  \______/ |__/  |__/
      \__/
```

```
  prod-db        *  2/2   9000 -> db:5432 +1
  staging-api    o
  bastion        *  1/1   2222 -> web:22
  analytics      !  0/1   connect: refused
  dev-tunnel     ~  0/1   connecting
```

> Looking for the GUI version? See **[allenbijo/QuickSSH](https://github.com/allenbijo/QuickSSH)** - the original Electron app this TUI mirrors. Both share the same `config.json`, so aliases created in one show up in the other.

## Why a TUI port?

* No Chromium/Node runtime in memory.
* Lives over SSH, inside tmux, on a Raspberry Pi - anywhere you have a terminal.
* Renders only when state changes (Bubbletea is Msg-driven, not FPS-driven), so idle CPU is 0 %.
* Reads and writes the **same `config.json`** as the Electron app, so aliases, settings, and SSH host names stay in sync across both.

## Install

### Pre-built binary (no Go required)

**macOS / Linux**
```bash
curl -fsSL https://github.com/allenbijo/QuickSSH-terminal/releases/latest/download/install.sh | sh
```

**Windows (PowerShell)**
```powershell
irm https://github.com/allenbijo/QuickSSH-terminal/releases/latest/download/install.ps1 | iex
```

**Homebrew (macOS / Linux)**
```bash
brew install allenbijo/tap/qsh
```

**winget (Windows)**
```powershell
winget install allenbijo.qsh
```

**Debian / Ubuntu / Mint** (download `.deb` from [latest release](https://github.com/allenbijo/QuickSSH-terminal/releases/latest))
```bash
curl -fsSL -o qsh.deb \
  "$(curl -fsSL https://api.github.com/repos/allenbijo/QuickSSH-terminal/releases/latest \
    | grep browser_download_url | grep amd64.deb | cut -d '"' -f4)"
sudo apt install ./qsh.deb
```

**Fedora / RHEL / openSUSE**
```bash
curl -fsSL -o qsh.rpm \
  "$(curl -fsSL https://api.github.com/repos/allenbijo/QuickSSH-terminal/releases/latest \
    | grep browser_download_url | grep x86_64.rpm | cut -d '"' -f4)"
sudo rpm -i qsh.rpm
```

**Alpine**
```bash
curl -fsSL -o qsh.apk \
  "$(curl -fsSL https://api.github.com/repos/allenbijo/QuickSSH-terminal/releases/latest \
    | grep browser_download_url | grep x86_64.apk | cut -d '"' -f4)"
sudo apk add --allow-untrusted ./qsh.apk
```

**Arch / Manjaro / EndeavourOS** (via AUR)
```bash
yay -S qsh-bin       # or paru -S qsh-bin
```

> Whichever channel you use, the resulting command is **`qsh`**. Run it from any terminal.

**Manual** - grab the archive for your platform from the [Releases page](https://github.com/allenbijo/QuickSSH-terminal/releases/latest), unzip, drop the binary anywhere on your `PATH`.

### From source

The build needs **Go 1.22 or newer**.

**One-shot via `go install`**

```bash
go install github.com/allenbijo/QuickSSH-terminal/cmd/qsh@latest
```

This places `qsh` in `$(go env GOBIN)` (defaults to `~/go/bin`). Make sure that directory is on your `PATH`.

**Manual build**

```bash
git clone https://github.com/allenbijo/QuickSSH-terminal
cd QuickSSH-terminal
go mod tidy
go build -o qsh ./cmd/qsh
```

Cross-compile examples:

```powershell
# from Windows for Linux
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o qsh ./cmd/qsh

# from Linux/macOS for Windows
GOOS=windows GOARCH=amd64 go build -o qsh.exe ./cmd/qsh
```

### Installing Go (if you need it)

| OS | Command |
|---|---|
| Windows (winget) | `winget install --id GoLang.Go --silent --accept-source-agreements --accept-package-agreements` |
| Windows (MSI) | Download `go1.22.x.windows-amd64.msi` from <https://go.dev/dl/> |
| macOS | `brew install go` |
| Linux (Debian) | `sudo apt install golang-go` (may be older; for latest grab the tarball from <https://go.dev/dl/>) |

Verify with `go version` - you should see `go1.22.x` or newer.

## Run

After installing via any of the channels above, just type:

```bash
qsh
```

The TUI opens fullscreen. On first launch with no aliases yet, the empty state prompts you to press **n** to create one or **s** to point at your SSH config.

If you built manually (not installed), the binary lives in the build directory:

```bash
./qsh                 # Linux/macOS
.\qsh.exe             # Windows
```

Other supported invocations:

```bash
qsh --version         # print version
qsh --help            # show usage
```

## Keys

| Screen   | Key             | Action                              |
|----------|-----------------|-------------------------------------|
| List     | up / down       | move selection                      |
| List     | enter / space   | toggle tunnel on/off                |
| List     | `h`             | open an interactive `ssh` session   |
| List     | `e`             | edit selected alias                 |
| List     | `n`             | new alias                           |
| List     | `d`             | delete (with confirmation)          |
| List     | `r`             | reconnect (disconnect + connect)    |
| List     | `s`             | open settings                       |
| List     | `q` / `ctrl+c`  | quit (asks if tunnels are running)  |
| Editor   | tab / shift+tab | move between fields                 |
| Editor   | `ctrl+a`        | add a new forward                   |
| Editor   | `ctrl+x`        | remove the focused forward          |
| Editor   | left / right    | cycle SSH host (when on host field) |
| Editor   | `ctrl+s`        | save                                |
| Editor   | `esc`           | cancel                              |
| Settings | tab             | move between fields                 |
| Settings | `ctrl+s`        | save                                |

## How it works

* **Storage** - `internal/store` reads and writes the exact same `config.json` produced by [electron-store](https://github.com/sindresorhus/electron-store) in the Electron app. Paths:
  * Windows: `%APPDATA%\Quick SSH\config.json`
  * macOS: `~/Library/Application Support/Quick SSH/config.json`
  * Linux: `~/.config/Quick SSH/config.json`

  Writes are atomic (`tmp` + rename) and guarded by a cross-platform file lock so simultaneous Electron + TUI use is safe.

* **Tunnels** - `internal/ssh/tunnel.go` spawns the system `ssh` binary with **byte-identical arguments** to the Electron app:

  ```
  ssh -N -o ExitOnForwardFailure=yes -o ServerAliveInterval=15 \
         -o ServerAliveCountMax=3 -o StrictHostKeyChecking=accept-new \
         -o BatchMode=no -L <local>:<host>:<port> [-p P] [-l U] [-i KEY] [-J JUMP] <hostname>
  ```

  The 2.5 s connection heuristic, 10 s health-check, and SIGTERM-then-SIGKILL teardown mirror the original. On Windows it uses `taskkill /T /F /PID` since SIGTERM isn't POSIX there.

* **Interactive sessions** - Pressing `h` on an alias calls `tea.ExecProcess(exec.Command("ssh", "-t", "-t", host))`, which suspends the TUI, hands the user a real native SSH session (full color, mouse, resize, scrollback, nested TUIs), and resumes the TUI on exit. There is no PTY emulation layer.

* **Rendering** - Bubbletea only redraws on a `Msg`. The manager fans tunnel status changes into a `chan StatusEvent` that the TUI consumes via a recursive `tea.Cmd`; the health-check goroutine emits a `Msg` only when a status actually changes. Idle CPU is **0 %**.

## Project layout

```
cmd/qsh/main.go              program entry
internal/tui/                 TUI screens, theme, keys
internal/ssh/                 tunnel + manager + config parser + session handoff
internal/store/               JSON store, types, electron-store-compatible paths
```

## Requirements

* Go 1.22+ (only if building from source)
* An `ssh` client on `PATH` (OpenSSH on Linux/macOS/Windows 10+; on older Windows install Git Bash or the OpenSSH optional feature)
* A terminal that supports 256 colors and UTF-8 (Windows Terminal, iTerm2, kitty, alacritty, GNOME Terminal - all fine)

## Related

* **[allenbijo/QuickSSH](https://github.com/allenbijo/QuickSSH)** - the Electron / desktop version. Shares the same config file.
* **[charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea)** - the TUI framework powering this app.

## License

MIT. See [LICENSE](LICENSE).
