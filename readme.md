# PolyClaude

Schedule multiple Claude Code Pro accounts to minimize rate-limit downtime.

## The Problem

Claude Code Pro ($20/mo) starts a 5-hour usage cycle with your first prompt. With heavy usage, you hit the rate limit well before the cycle resets. Upgrading to Max ($100/mo) is a 5x price jump with no middle ground.

A cheaper alternative: run multiple Pro accounts and rotate when one is rate-limited. Two accounts at $40/mo give ~2x capacity at less than half the cost of Max.

But naive rotation wastes capacity. **When** you activate an account matters as much as **which** one you use. A single throwaway prompt sent hours before you start coding can unlock an entire extra usage cycle.

PolyClaude computes the optimal pre-activation schedule for your accounts and installs cron jobs to execute it automatically.

## How It Works

Given your number of accounts, coding window, and average time before hitting the rate limit, PolyClaude:

1. Computes the maximum usable cycles per account (including pre-activation)
2. Generates an optimal block schedule using your chosen strategy
3. Calculates the exact pre-activation time for each account
4. Installs cron jobs that send a throwaway prompt (`claude -p "say hi"`) at each pre-activation time to start the 5-hour cycle timer

When you sit down to code, your accounts' cycles are already aligned with your coding window.

### Example

3 accounts, 1 hour average before rate limit, coding 09:00-17:00 (8h window):

**Without PolyClaude:** Each account gets ~2 cycles. 6 hours of coding, 2 hours waiting.

**With PolyClaude (bunch strategy):** Pre-activation unlocks a 3rd cycle per account. 8 continuous hours of coding, zero downtime:

```
Pre-activation:  Account A at 05:00, Account B at 06:00, Account C at 09:00

09  10  11  12  13  14  15  16  17
[─A─][─B─][─C─][─A─][─B─][─C─][─A─][─B─]
          8 hours continuous coding
```

## Strategies

Both strategies achieve the same total coding time. They differ in how cooldown is distributed:

**Spread** — Minimizes the longest interruption. Coding blocks are evenly spaced with short gaps between them. Best for incremental tasks, debugging, and code reviews where short pauses are tolerable.

```
[─A─] 6min [─B─] 6min [─C─] 6min [─A─] 6min [─B─] ...
```

**Bunch** — Maximizes the longest unbroken coding stretch. All blocks are packed together, with cooldown pushed to the end. Best for deep-focus work, building features end-to-end, and complex refactors.

```
[─A─][─B─][─C─][─A─][─B─][─C─][─A─][─B─][─C─]  1h cooldown
```

## Installation

Requires [Go 1.26+](https://go.dev/dl/) and the [Claude CLI](https://docs.anthropic.com/en/docs/claude-code).

```sh
go install github.com/armanjr/polyclaude@latest
```

Or build from source:

```sh
git clone https://github.com/armanjr/polyclaude.git
cd polyclaude
go build -o polyclaude .
```

## Usage

```sh
# Interactive setup wizard
polyclaude

# Preview the full wizard without writing any files or cron jobs
polyclaude --dry-run
```

The wizard walks you through:

1. **Verify** — Checks that the Claude CLI is installed
2. **Configure** — Home directory, number of accounts, avg dev time, coding window, weekdays, strategy
3. **Login** — Guides you through `claude /login` for each account in an isolated config directory
4. **Sanity check** — Runs `claude -p "say hi"` per account to verify auth
5. **Schedule** — Displays the computed schedule with pre-activation times and block timeline
6. **Cron** — Installs cron jobs (with confirmation)

### Re-running

If an existing configuration is found at `~/.polyclaude/config.yaml`, you'll be prompted to start fresh or exit. Re-running is safe and idempotent — cron entries are managed between `# BEGIN polyclaude` / `# END polyclaude` markers.

### Inspecting

```sh
# View installed cron jobs
crontab -l

# Check config
cat ~/.polyclaude/config.yaml
```

## Configuration

All configuration is stored in `~/.polyclaude/config.yaml`:

```yaml
home_dir: /Users/you/.polyclaude
num_accounts: 3
avg_dev_time: 1.0          # hours before hitting rate limit
start_time: "09:00"        # 24h format
end_time: "17:00"
weekdays: [mon, tue, wed, thu, fri]
strategy: bunch            # "spread" or "bunch"
claude_path: /usr/local/bin/claude
accounts:
  - name: bright-falcon
    dir: /Users/you/.polyclaude/accounts/bright-falcon
  - name: calm-eagle
    dir: /Users/you/.polyclaude/accounts/calm-eagle
  - name: keen-otter
    dir: /Users/you/.polyclaude/accounts/keen-otter
```

Each account gets an isolated directory used as `CLAUDE_CONFIG_DIR`, so multiple Claude logins coexist without conflict.

## Requirements

- **Go 1.26+** for building
- **Claude CLI** (`npm install -g @anthropic-ai/claude-code`)
- **cron** — standard on Linux; available on macOS (ensure cron has Full Disk Access in System Settings > Privacy & Security)
- **Machine must be awake** when cron jobs fire. If your machine sleeps before the earliest pre-activation time, the prompt won't be sent and that account won't get its extra cycle.

## Project Structure

```
polyclaude/
├── main.go                          # Entrypoint: --dry-run flag, launch TUI
├── internal/
│   ├── config/                      # Config struct, YAML load/save, validators
│   ├── scheduler/                   # Core math: kMax, blocks, spread/bunch, pre-activation
│   ├── cron/                        # Crontab read/update/write with marker-based splicing
│   ├── accounts/                    # Directory creation, login verification, sanity checks
│   └── tui/                         # Bubbletea v2 interactive wizard (8 steps)
├── math.md                          # Full mathematical derivation
└── go.mod
```

## The Math

The scheduling problem is a constrained interval-packing problem with periodic renewal. The full derivation — including proofs of optimality for both strategies — is in [`math.md`](math.md). Key equations:

| Formula | Meaning |
|---|---|
| k_max = floor((W + 2(T-x)) / T) | Max usable cycles per account |
| B = n * k_max | Total coding blocks |
| L = min(W, n * k_max * x) | Total coding hours |
| g* = max(D/(B+1), g_cycle) | Optimal gap (spread) |
| k_c = floor((T-x)/(T-nx)) + 1 | Consecutive cycles before cooldown (bunch) |

Where T=5h (fixed cycle length), x=avg dev time, W=coding window, n=number of accounts.

## License

MIT
