# gstat Roadmap

## Architecture

- **Language**: Go 1.26+
- **TUI Framework**: Bubble Tea v2 + Lip Gloss v2 + Bubbles v2 (charmbracelet)
- **Data Source**: ESPN hidden API (`site.api.espn.com`) - no auth required, free JSON
- **Refresh**: 30-second polling interval (configurable in code)
- **Binary**: Single static binary, no config files needed

## Project Structure

```
gstat/
├── main.go                      # Entry point
├── internal/
│   ├── espn/
│   │   ├── client.go            # HTTP client, data parsing, tied position calc
│   │   ├── client_test.go       # Business logic + fixture integration tests
│   │   ├── types.go             # JSON response structs + simplified domain types
│   │   ├── helpers.go           # Date parsing helpers
│   │   └── testdata/
│   │       └── scoreboard.json  # Real ESPN API response fixture
│   ├── model/
│   │   ├── leaderboard.go       # Main Bubble Tea model (Init/Update/View)
│   │   └── messages.go          # Custom tea.Msg types
│   └── ui/
│       ├── styles.go            # Centralized Lip Gloss color palette + styles
│       ├── header.go            # Tournament header bar rendering
│       ├── table.go             # Leaderboard table + player row rendering
│       ├── statusbar.go         # Bottom status bar rendering
│       └── render_test.go       # UI rendering tests with content assertions
```

## ESPN API Notes

- **Endpoint**: `https://site.api.espn.com/apis/site/v2/sports/golf/pga/scoreboard`
- **Date format**: `2006-01-02T15:04Z` (no seconds)
- **Leaderboard**: `events[0].competitions[0].competitors[]` sorted by `order`
- **Tied positions**: Not provided by API; calculated from score groupings
- **Cut detection**: Players who made the cut have an R3+ linescore entry (even empty placeholder)
- **WD detection**: Round `value` < 30 with no hole-by-hole data = holes played, not strokes
- **Hole-by-hole**: Available at `competitors[n].linescores[r].linescores[h]`
- **Season calendar**: Available at `leagues[0].calendar[]` (48 events for 2026)

## v0.1.0 - MVP Leaderboard [COMPLETE]

- [x] Go module init with Bubble Tea v2 deps
- [x] ESPN API client with HTTP fetch + JSON parsing
- [x] Domain types (Tournament, Player, RoundScore)
- [x] Date parsing for ESPN's shortened ISO 8601 format
- [x] Tied position calculation from score groupings
- [x] Cut/WD player detection
- [x] Bubble Tea v2 model with declarative View
- [x] Alt-screen full TUI with auto-refresh (30s)
- [x] Tournament header bar (name, round status, dates)
- [x] Leaderboard table (POS, PLAYER, CTRY, TOT, R1-R4, THRU)
- [x] Color-coded scores (green=under, red=over, yellow=even)
- [x] Cut line separator
- [x] Dimmed styling for CUT/WD players
- [x] Status bar with last update, countdown, keybind hints
- [x] Keyboard: j/k scroll, pgup/pgdn, g/G home/end, r refresh, q quit
- [x] Error handling (network failures, no active tournament)
- [x] .gitignore, README, ROADMAP

## Backlog (Future Versions)

### v0.2.0 - Leaderboard QoL
- [ ] Search/filter: type to filter leaderboard by player name
- [ ] Favorite/pin players: highlight favorites, group at top of leaderboard
- [ ] Persist favorites to `~/.config/gstat/favorites.json`
- [ ] Leaderboard change arrows: show position movement (up/down/new) since last refresh

### v0.3.0 - Player Detail View
- [ ] Select a player (Enter) to expand inline hole-by-hole scorecard
- [ ] Show each hole: hole number, par, strokes, score type (eagle/birdie/par/bogey/double+)
- [ ] Color-coded score types in the detail view
- [ ] Collapse detail view (Enter or Esc) and continue scrolling
- [ ] Round selector within detail view (tab between R1/R2/R3/R4)

### v0.4.0 - Release & Distribution
- [ ] Add `.goreleaser.yml` (cross-compile for macOS arm64/amd64, Linux arm64/amd64, Windows)
- [ ] Add `.github/workflows/release.yml` (GitHub Actions triggered on git tag push)
- [ ] Create `nickkoul/homebrew-tap` repo on GitHub
- [ ] Configure Homebrew formula generation in GoReleaser config
- [ ] Tag and publish first release
- [ ] Verify install methods: `brew install nickkoul/tap/gstat`, direct download, `go install`
- [ ] Document install methods in README

### v0.5.0 - Tournament Selector
- [ ] Tournament picker from ESPN season calendar (48 events)
- [ ] View past tournament results
- [ ] Show tournament as "not started" / "in progress" / "completed" states

### Future Ideas
- [ ] Multiple tour support (European Tour, LIV, LPGA)
- [ ] Notifications when pinned players make eagle/hole-in-one
- [ ] Course info display (par, yardage per hole)
- [ ] Historical comparison (player's previous results at this course)
- [ ] Configurable refresh interval via CLI flag
- [ ] Mouse support for clicking on players
- [ ] Filter by country
- [ ] Sort by different columns
- [ ] Export leaderboard to markdown/CSV

## Completed Sessions

### Session 1 (Mar 13, 2026)
- Built entire v0.1.0 MVP from scratch
- Researched ESPN hidden API, analyzed JSON response structure
- Chose Bubble Tea v2 (released Feb 24, 2026) as TUI framework
- Verified live data working with THE PLAYERS Championship 2026
- All features tested and working: leaderboard, tied positions, cut detection, auto-refresh
