# Development Guide

## Prerequisites

- **Go 1.26+** - install via [go.dev/dl](https://go.dev/dl/) or `brew install go`

No other dependencies are needed. All Go modules are fetched automatically.

## Quick Start

```bash
git clone https://github.com/nickkoul/gstat.git
cd gstat
go run .
```

## Build

```bash
# Build binary
go build -o gstat .

# Run it
./gstat

# Install to your $GOPATH/bin
go install .
```

## Project Structure

```
gstat/
├── main.go                          # Entry point, creates Bubble Tea program
├── internal/
│   ├── espn/
│   │   ├── types.go                 # ESPN API JSON structs + domain types
│   │   ├── client.go                # HTTP client, parsing, business logic
│   │   └── helpers.go               # Date parsing utilities
│   ├── model/
│   │   ├── leaderboard.go           # Bubble Tea model (Init/Update/View)
│   │   └── messages.go              # Custom message types (tea.Msg)
│   └── ui/
│       ├── styles.go                # Lip Gloss color palette + style defs
│       ├── header.go                # Tournament header renderer
│       ├── table.go                 # Leaderboard table + row renderer
│       ├── statusbar.go             # Bottom status bar renderer
│       └── render_test.go           # Visual rendering tests
├── go.mod
├── go.sum
├── .gitignore
├── LICENSE
├── README.md
├── ROADMAP.md                       # Feature roadmap + session history
└── DEVELOPMENT.md                   # This file
```

## Architecture

### Bubble Tea (Elm Architecture)

The app follows the [Elm Architecture](https://guide.elm-lang.org/architecture/) pattern via [Bubble Tea v2](https://github.com/charmbracelet/bubbletea):

```
Init() → fetches data + starts tick timer
  ↓
Update(msg) → handles keyboard input, data responses, tick events
  ↓
View() → renders the full screen as a tea.View (declarative, v2-style)
```

The single model lives in `internal/model/leaderboard.go` and orchestrates everything.

### Data Flow

```
ESPN API  →  espn.Client.FetchLeaderboard()  →  espn.Tournament  →  model.Update()  →  ui.Render*()
   ↑                                                                       |
   └───────────────── 30s tick ────────────────────────────────────────────┘
```

1. `espn.Client` makes an HTTP GET to the ESPN scoreboard endpoint
2. Raw JSON is unmarshalled into `espn.ScoreboardResponse` structs
3. `parseTournament()` converts raw data into simplified `Tournament` + `Player` types
4. Business logic runs: tied position calculation, cut detection, WD detection
5. The Bubble Tea model receives the data via `DataFetchedMsg`
6. `ui.Render*()` functions produce styled strings using Lip Gloss

### Package Responsibilities

| Package | Responsibility |
|---------|---------------|
| `espn` | ESPN API communication, JSON types, data parsing, domain logic |
| `model` | Bubble Tea model, keyboard handling, refresh scheduling, view orchestration |
| `ui` | Pure rendering functions (header, table rows, status bar), styles |

Key design principle: `ui` functions are **pure** -- they take data in and return strings. No side effects, no state. This makes them easy to test.

## ESPN API

The app uses ESPN's undocumented public API. No API key or auth is needed.

### Endpoint

```
GET https://site.api.espn.com/apis/site/v2/sports/golf/pga/scoreboard
```

### Key Data Paths

```
events[0].name                                          → tournament name
events[0].competitions[0].status.type.detail            → "Round 2 - In Progress"
events[0].competitions[0].competitors[n].order          → leaderboard position
events[0].competitions[0].competitors[n].score          → total to par ("-12")
events[0].competitions[0].competitors[n].athlete        → player name, country flag
events[0].competitions[0].competitors[n].linescores[r]  → round data (strokes, to-par)
events[0].competitions[0].competitors[n].linescores[r].linescores[h] → hole-by-hole
leagues[0].calendar[]                                   → season schedule (48 events)
```

### Quirks & Edge Cases

- **Date format**: `2006-01-02T15:04Z` (no seconds) -- needs custom parsing
- **Tied positions**: The API's `order` field is unique (1, 2, 3...) even for ties. We calculate ties by grouping players with the same `score` value
- **Cut detection**: Players who made the cut get an R3 linescore entry (even an empty placeholder with just `period: 3`). Players without an R3 entry after round 2 is complete missed the cut
- **WD detection**: Withdrawn players have a round where `value` is holes played (e.g., `4.0` for 4 holes) rather than total strokes (60-90 range), and no hole-by-hole sub-array
- **Hole order**: Hole-by-hole data is in playing order, not hole-number order. Players starting on the back nine have `period` values: 10, 11, ..., 18, 1, 2, ..., 9

### Testing API Responses

Useful for debugging:

```bash
# Fetch and pretty-print the full response
curl -s 'https://site.api.espn.com/apis/site/v2/sports/golf/pga/scoreboard' | python3 -m json.tool | head -100

# Show top 5 players
curl -s 'https://site.api.espn.com/apis/site/v2/sports/golf/pga/scoreboard' | \
  python3 -c "
import sys, json
d = json.load(sys.stdin)
for c in d['events'][0]['competitions'][0]['competitors'][:5]:
    print(f\"#{c['order']} {c['athlete']['displayName']} {c['score']}\")
"
```

## Testing

See [TESTING.md](TESTING.md) for the full testing guidelines, philosophy, and conventions.

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run a specific package
go test ./internal/espn/
go test ./internal/ui/

# Run a specific test
go test ./internal/espn/ -run TestCalculateTiedPositions

# Coverage report
go test -cover ./...
```

### Test coverage targets

| Package | Coverage | Notes |
|---------|----------|-------|
| `internal/espn` | ~77% | Uncovered: HTTP client methods (intentionally skip network calls) |
| `internal/ui` | ~95% | All render functions tested with content assertions |
| `internal/model` | 0% | Bubble Tea model requires TTY -- skip |

### Test fixtures

Real ESPN API responses are saved in `internal/espn/testdata/` for offline integration testing. See [TESTING.md](TESTING.md) for how to capture new fixtures.

## Code Quality

```bash
# Static analysis
go vet ./...

# Build check
go build ./...

# Full check (vet + tests + build)
go vet ./... && go test ./... && go build ./...
```

## Adding a New Feature

General workflow:

1. Check `ROADMAP.md` for the current state and planned features
2. If the feature needs new data from ESPN, update `internal/espn/types.go` with new structs and `internal/espn/client.go` with parsing logic
3. If it's a new view/screen, add a new model in `internal/model/` and rendering functions in `internal/ui/`
4. If it's enhancing the existing leaderboard, modify `internal/model/leaderboard.go` (state + key handling) and the relevant `internal/ui/` renderer
5. Update `ROADMAP.md` to mark progress

### Adding a New Keybinding

1. Add the case to `handleKey()` in `internal/model/leaderboard.go`
2. Update the status bar hints in `internal/ui/statusbar.go`
3. Update the keybindings table in `README.md`

### Adding a New Column to the Leaderboard

1. Add a constant for column width in `internal/ui/table.go`
2. Update `RenderTableHeader()` to include the column header
3. Update `RenderPlayerRow()` to render the column data
4. If the column needs new data, update `espn.Player` in `types.go` and `parsePlayer()` in `client.go`

## Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| `charm.land/bubbletea/v2` | v2.0.2 | TUI framework (Elm architecture) |
| `charm.land/lipgloss/v2` | v2.0.2 | Terminal styling (colors, borders, layout) |

Bubble Tea v2 was released Feb 24, 2026. Key differences from v1:
- `View()` returns `tea.View` struct (declarative) instead of `string`
- Import path uses `charm.land/` vanity domain instead of `github.com/charmbracelet/`
- Cursed Renderer for better performance
- Native `AltScreen`, `MouseMode`, `WindowTitle` via View fields

Docs: [charm.land/blog/v2](https://charm.land/blog/v2/) | [Upgrade guide](https://github.com/charmbracelet/bubbletea/blob/main/UPGRADE.md)
