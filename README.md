# gstat

Live golf tournament leaderboard in your terminal.

```
 THE PLAYERS Championship                    Round 2 - Play Complete
 Mar 12, 2026 - Mar 15, 2026
 ─────────────────────────────────────────────────────────────────────
    POS   CHG  PLAYER                   CTRY    TOT   R1   R2   R3   R4  THRU
 ─────────────────────────────────────────────────────────────────────────────
>   1     ^1   Ludvig Åberg             SWE    -12   69   63    -    -     F
    T2    ˅1   Xander Schauffele        USA    -10   69   65    -    -     F
 *  T2    E    Scottie Scheffler        USA    -10   67   67    -    -     F
      4    ^3   Cameron Young           USA     -9   68   67    -    -     F
    T5    E    Corey Conners            CAN     -8   69   67    -    -    12
    T5    ˅2   Justin Thomas            USA     -8   68   68    -    -     F
 ──────────────────────────── CUT ────────────────────────────────
     CUT   Adam Schenk              USA     +3   77   70    -    -     F

 Updated 10:32 PM  Next 28s              / search  f favorite  F favorites  q quit
```

## Install

```bash
go install github.com/nickkoul/gstat@latest
```

Or build from source:

```bash
git clone https://github.com/nickkoul/gstat.git
cd gstat
go build -o gstat .
./gstat
```

## Features

- Live PGA Tour leaderboard from ESPN
- Auto-refreshes every 30 seconds
- Color-coded scores (green for under par, red for over par)
- Tied position indicators (T1, T2, etc.)
- Cut line separator with dimmed styling for eliminated players
- WD (withdrawn) player detection
- Vim-style player search with `/`
- Round columns default to to-par; press `t` to toggle strokes view
- Select rows with vim-style navigation, bold favorites with `f`, toggle a favorites-only view with `F`, and keep favorites across restarts
- `CHG` shows position diff versus the previous round as `^n`, `˅n`, or `E`
- Live refresh markers flag changed rows as `!` (score), `^` (standing), or `+` (both)
- Toggle an expanded hotkey help panel with `?`
- Scrollable with vim-style keybindings

## Keybindings

| Key | Action |
|-----|--------|
| `j` / `↓` | Move selection down |
| `k` / `↑` | Move selection up |
| `Ctrl+d` / `PgDn` | Move selection half page down |
| `Ctrl+u` / `PgUp` | Move selection half page up |
| `g` / `Home` | Jump selection to top |
| `G` / `End` | Jump selection to bottom |
| `/` | Start player search |
| `f` | Toggle favorite on the selected player |
| `F` | Toggle favorites-only view |
| `t` | Toggle round columns between strokes and to par |
| `?` | Toggle expanded hotkey help |
| `Enter` | Apply current search |
| `Esc` | Clear search |
| `r` | Force refresh |
| `q` / `Ctrl+C` | Quit |

## Data Source

Uses ESPN's public (undocumented) API. No API key required.

## Tech Stack

- [Go](https://go.dev)
- [Bubble Tea v2](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss v2](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [Bubbles v2](https://github.com/charmbracelet/bubbles) - TUI components

## License

MIT
