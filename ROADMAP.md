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
│   ├── config/
│   │   ├── favorites.go         # Favorites persistence helpers
│   │   └── favorites_test.go    # Config load/save/error tests
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

## v0.2.0 - Execution Plan

This section is the working plan for v0.2.0. We will implement one feature at a time, verify it, then update this document before moving to the next feature.

### Execution Order

1. Search/filter by player name
2. Configurable round score display (`strokes` vs `to par`)
3. Toggleable on-screen hotkey hints
4. Stable player identity + canonical/display rank plumbing
5. Favorite/pin players
6. Persist favorites to `~/.config/gstat/favorites.json`
7. Leaderboard change arrows
8. Visual indication for player score/standing updates

### Verification Loop (After Every Feature)

1. Run targeted tests for touched packages
2. Run `go test ./...`
3. Do a manual TUI smoke check
4. Update this section with completion notes, files touched, and follow-ups

### Feature 1 - Search/filter by player name

- Status: complete (Mar 14, 2026)
- Goal: type to filter leaderboard rows by player name without mutating canonical leaderboard data
- Dependencies: none

Concrete checklist:

- [x] Add filter state to `internal/model/leaderboard.go`
- [x] Implement vim-like search mode: `/` enters search, typed characters update the query, `backspace` deletes, `enter` exits search mode while keeping the filter, `esc` clears and exits
- [x] Add a helper that derives the visible filtered player slice from `tournament.Players`
- [x] Update leaderboard rendering to use the derived visible slice instead of the canonical slice directly
- [x] Clamp or reset `scrollPos` whenever the filter changes
- [x] Preserve the active filter across refreshes
- [x] Show active filter text and clear hint in `internal/ui/statusbar.go`
- [x] Render a readable empty-results state when no players match
- [x] Add/extend tests for match, no-match, clear, scroll bounds, and refresh while filtered
- [x] Update `README.md` keybindings/behavior if the final UX adds new user-facing keys

Acceptance criteria:

- Typing immediately narrows the visible leaderboard rows
- `/` initiates search mode and the active query is visible in the UI
- Clearing the filter restores the full leaderboard
- Scrolling still works while filtered
- Refreshing data does not clear the active filter
- No-match state is obvious and does not break layout

Session notes:

- Do not introduce row selection yet; this feature should stay focused on filtering only
- Keep `tournament.Players` as canonical data and build a derived filtered view
- While search mode is active, printable keys should edit the query instead of triggering normal navigation behavior

Completion notes:

- Files touched: `internal/model/leaderboard.go`, `internal/model/leaderboard_test.go`, `internal/ui/statusbar.go`, `internal/ui/render_test.go`, `README.md`
- Tests run: `go test ./internal/model ./internal/ui`, `go test ./...`
- Follow-up: if we later add row selection or fuzzy matching, keep them separate from this canonical filtered-view pipeline

### Feature 2 - Configurable round score display

- Status: complete (Mar 14, 2026)
- Goal: keep `TOT` relative to par and allow `R1-R4` to toggle between strokes and relative-to-par
- Dependencies: none

Checklist:

- [x] Add round score display mode state to the leaderboard model
- [x] Render round columns from either `RoundScore.Strokes` or `RoundScore.ToPar`
- [x] Add a keybinding to toggle display mode
- [x] Surface the active mode in the status bar/help text
- [x] Add rendering tests for both modes
- [x] Update `README.md` docs

Acceptance criteria:

- `TOT` always remains relative to par
- `R1-R4` switch cleanly between strokes and to-par
- Column alignment remains stable at normal and narrow widths

Completion notes:

- Files touched: `internal/model/leaderboard.go`, `internal/model/leaderboard_test.go`, `internal/ui/table.go`, `internal/ui/statusbar.go`, `internal/ui/render_test.go`, `README.md`
- Tests run: `go test ./internal/model ./internal/ui`, `go test ./...`
- Behavior: the app now defaults to round columns in to-par mode, with `t` toggling back to strokes
- Follow-up: Feature 4 can now add stable player identity without needing to revisit this round-mode plumbing

### Feature 3 - Toggleable on-screen hotkey hints [COMPLETE]

- Status: complete (Mar 14, 2026)
- Goal: make available hotkeys discoverable in-app with a toggleable help view that does not interfere with normal leaderboard use
- Dependencies: builds on the existing keybinding/status bar foundation

Checklist:

- [x] Add a dedicated help-panel render path for available hotkeys and their actions
- [x] Toggle the help panel with a keyboard shortcut
- [x] Keep a compact always-visible hint that the full help can be opened
- [x] Show context-aware hints for normal mode vs search mode
- [x] Reduce visible leaderboard rows while the help panel is open so layout remains stable
- [x] Add rendering/model tests for help visibility and content
- [x] Update `README.md` docs

Acceptance criteria:

- Users can discover the major hotkeys without leaving the app
- The help view can be opened and closed repeatedly without breaking scrolling or search
- Search-mode help reflects search-specific controls
- The leaderboard remains readable with the help view open

Completion notes:

- Files touched: `internal/model/leaderboard.go`, `internal/model/leaderboard_test.go`, `internal/ui/help.go`, `internal/ui/statusbar.go`, `internal/ui/styles.go`, `internal/ui/render_test.go`, `README.md`
- Tests run: `go test ./internal/model ./internal/ui`, `go test ./...`
- Behavior: when expanded hints are hidden, the status bar now keeps `? show hints` visible and reserves enough space for the full status line on shorter terminals
- Follow-up: Feature 4 still starts the stable-ID plumbing needed for favorites and movement tracking

### Feature 4 - Stable player identity + canonical/display rank plumbing [COMPLETE]

- Status: complete (Mar 14, 2026)
- Goal: introduce stable player identity and preserve separate canonical rank vs tied display rank
- Dependencies: required before Features 5-8

Checklist:

- [x] Carry ESPN athlete ID into the simplified `espn.Player` model
- [x] Preserve canonical/raw leaderboard order separately from tied display position
- [x] Keep tie rendering behavior unchanged in the UI
- [x] Add tests for ID parsing and tie/rank behavior

Acceptance criteria:

- Existing leaderboard output remains visually the same
- Player identity is stable across refreshes
- Tie rendering still shows `Tn` correctly

Completion notes:

- Files touched: `internal/espn/types.go`, `internal/espn/client.go`, `internal/espn/client_test.go`, `internal/ui/table.go`, `internal/ui/render_test.go`, `internal/model/leaderboard_test.go`
- Tests run: `go test ./internal/espn`, `go test ./internal/ui`, `go test ./internal/model`, `go test ./...`
- Behavior: players now carry a stable ESPN athlete ID plus separate `CanonicalRank` and `DisplayPosition` fields, so tie rendering stays the same while canonical order remains available for future features
- Follow-up: Features 5-8 can now key persistent/comparison state off stable player IDs and canonical rank without depending on visible row order

### Feature 5 - Favorite/pin players [COMPLETE]

- Status: complete (Mar 14, 2026)
- Goal: highlight favorite players without disturbing leaderboard order, plus support a favorites-only view
- Dependencies: Feature 4

Checklist:

- [x] Add selected-row/cursor state
- [x] Add favorite state keyed by stable player ID
- [x] Add a keybinding to toggle favorite on the selected player
- [x] Keep favorites in canonical leaderboard order during the normal view
- [x] Add a favorites-only toggle that filters to favorites while preserving their tournament ranking
- [x] Add favorite marker styling in the table
- [x] Add tests for toggle/reorder/render behavior

Acceptance criteria:

- Selection is predictable
- Toggling favorite updates the row immediately
- Favorites remain in place during the normal view without breaking tie/cut rendering
- Favorites-only view shows just favorited players with their tournament positions intact

Completion notes:

- Files touched: `internal/model/leaderboard.go`, `internal/model/leaderboard_test.go`, `internal/ui/table.go`, `internal/ui/styles.go`, `internal/ui/statusbar.go`, `internal/ui/help.go`, `internal/ui/render_test.go`, `README.md`
- Tests run: `go test ./internal/model`, `go test ./internal/ui`, `go test ./...`
- Behavior: the leaderboard now keeps a selected row, toggles favorites with `f`, leaves favorites in canonical order during the normal view, and toggles a favorites-only filtered view with `F` that preserves tournament positions
- Follow-up: Feature 6 can persist the in-memory favorites map directly because it is already keyed by stable player ID

### Feature 6 - Persist favorites

- Status: complete (Mar 14, 2026)
- Goal: persist favorites across app restarts
- Dependencies: Feature 5

Checklist:

- [x] Add config helper for reading/writing favorites
- [x] Use `os.UserConfigDir()` and store under `gstat/favorites.json`
- [x] Create config directory if missing
- [x] Save atomically
- [x] Handle missing/corrupt/unreadable config gracefully
- [x] Surface persistence errors non-fatally in the UI
- [x] Add tests for load/save/error cases

Acceptance criteria:

- Favorites survive restart
- Corrupt config does not crash the app
- Persistence failure does not break in-memory favorites

Completion notes:

- Files touched: `internal/config/favorites.go`, `internal/config/favorites_test.go`, `internal/model/leaderboard.go`, `internal/model/leaderboard_test.go`, `internal/model/messages.go`, `README.md`, `ROADMAP.md`, `DEVELOPMENT.md`
- Tests run: `go test ./internal/config ./internal/model ./internal/ui`, `go test ./...`
- Behavior: favorites now load from `~/.config/gstat/favorites.json` (or the platform-equivalent `os.UserConfigDir()` path) during startup and save atomically whenever a favorite is toggled
- UI behavior: corrupted or unreadable favorites config no longer blocks the app; persistence failures stay non-fatal and surface in the status bar while in-memory favorites continue working
- Follow-up: Feature 7 can reuse the same stable player IDs and refresh lifecycle without touching favorites persistence

### Feature 7 - Leaderboard change arrows

- Status: planned
- Goal: show when a player has moved up, down, or is newly tracked since the previous refresh
- Dependencies: Feature 4

Checklist:

- [ ] Store previous refresh snapshot keyed by stable player ID
- [ ] Compute movement metadata on each successful refresh
- [ ] Compare canonical rank, not filtered/pinned display order
- [ ] Render up/down/new indicators in the table
- [ ] Handle first load and missing previous data cleanly
- [ ] Add tests for up/down/same/new cases

Acceptance criteria:

- Arrows only appear after at least one refresh
- Filtering/pinning does not affect movement calculation
- Tie groups do not create noisy false signals

### Feature 8 - Visual indication for player score/standing updates

- Status: planned
- Goal: make player changes easy to notice when score or standing changes on refresh
- Dependencies: reuse Feature 7 snapshot plumbing

Checklist:

- [ ] Reuse previous-refresh snapshot data
- [ ] Detect score changes and standing changes separately
- [ ] Add transient row-level change metadata
- [ ] Add subtle highlight or marker that coexists with favorite and movement styling
- [ ] Decide and implement decay/reset behavior for the indicator
- [ ] Add tests for changed vs unchanged rows

Acceptance criteria:

- Updated players are visually noticeable
- The table remains readable when many players change
- Indicators clear/reset predictably

### Session Resume Rules

- Do not start a new feature until the current feature passes tests and manual smoke checks
- Preserve `tournament.Players` as canonical data; filter/pin views should be derived views
- Use stable player ID for anything persisted or compared across refreshes
- Compute movement/update state from canonical rank, not filtered or pinned order

## Backlog (Future Versions)

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
