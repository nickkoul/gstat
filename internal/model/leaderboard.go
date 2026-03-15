package model

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/nickkoul/gstat/internal/config"
	"github.com/nickkoul/gstat/internal/espn"
	"github.com/nickkoul/gstat/internal/ui"
)

const (
	defaultRefreshInterval = 30 * time.Second
	scrollPadding          = 3 // rows of padding when scrolling near edges
)

type roundScoreDisplayMode string

type playerChange string

type playerUpdate uint8

const (
	roundScoreDisplayStrokes roundScoreDisplayMode = "strokes"
	roundScoreDisplayToPar   roundScoreDisplayMode = "to par"

	playerChangeNone playerChange = ""
	playerChangeEven playerChange = "E"
)

const (
	playerUpdateNone  playerUpdate = 0
	playerUpdateScore playerUpdate = 1 << iota
	playerUpdateStanding
)

type playerSnapshot struct {
	ID              string
	CanonicalRank   int
	DisplayPosition int
	TotalScore      string
	Status          string
}

// Model is the main Bubble Tea model for the leaderboard view.
type Model struct {
	// Data
	tournament     *espn.Tournament
	client         *espn.Client
	favoritesStore favoritesStore
	lastUpdate     time.Time
	lastError      string
	favoritesErr   string
	changes        map[string]playerChange
	updates        map[string]playerUpdate
	snapshot       map[string]playerSnapshot

	// UI state
	width         int
	height        int
	scrollPos     int
	favorites     map[string]bool
	selectedID    string
	filterQuery   string
	favoritesOnly bool
	searchMode    bool
	showDetail    bool
	detailRound   int
	roundMode     roundScoreDisplayMode
	showHelp      bool

	// Refresh
	refreshInterval time.Duration
	nextRefreshAt   time.Time
	loading         bool
}

type favoritesStore interface {
	Load() (map[string]bool, error)
	Save(map[string]bool) error
}

// New creates a new leaderboard model.
func New() Model {
	return Model{
		client:          espn.NewClient(),
		favoritesStore:  config.NewFavoritesStore(),
		favorites:       make(map[string]bool),
		refreshInterval: defaultRefreshInterval,
		loading:         true,
		detailRound:     1,
		roundMode:       roundScoreDisplayToPar,
	}
}

// Init initializes the model. It kicks off the first data fetch.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.fetchData(),
		m.loadFavorites(),
		tickCmd(time.Second), // 1-second tick for countdown display
	)
}

// Update handles messages and updates the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.syncVisibleState()
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case DataFetchedMsg:
		m.changes = computeRoundChanges(msg.Tournament)
		m.updates = computeRefreshUpdates(m.snapshot, msg.Tournament)
		m.snapshot = buildPlayerSnapshot(msg.Tournament)
		m.tournament = msg.Tournament
		m.lastUpdate = msg.FetchedAt
		m.lastError = ""
		m.loading = false
		m.nextRefreshAt = time.Now().Add(m.refreshInterval)
		m.syncVisibleState()
		return m, nil

	case DataErrorMsg:
		m.lastError = msg.Err.Error()
		m.loading = false
		m.nextRefreshAt = time.Now().Add(m.refreshInterval)
		m.syncVisibleState()
		return m, nil

	case FavoritesLoadedMsg:
		m.favorites = copyFavorites(msg.Favorites)
		m.favoritesErr = formatFavoritesError("load", msg.Err)
		m.syncVisibleState()
		return m, nil

	case TickMsg:
		var cmds []tea.Cmd

		// Check if it's time to refresh
		if !m.nextRefreshAt.IsZero() && time.Now().After(m.nextRefreshAt) {
			m.loading = true
			cmds = append(cmds, m.fetchData())
		}

		// Always schedule the next tick for countdown updates
		cmds = append(cmds, tickCmd(time.Second))
		return m, tea.Batch(cmds...)
	}

	return m, nil
}

// View renders the full UI.
func (m Model) View() tea.View {
	var v tea.View
	v.AltScreen = true

	if m.width == 0 {
		v.SetContent("Initializing...")
		return v
	}

	content := m.renderContent()
	v.SetContent(content)
	return v
}

// renderContent builds the full screen content string.
func (m Model) renderContent() string {
	var s string

	// Header
	s += ui.RenderHeader(m.tournament, m.width)
	s += "\n"

	if m.showHelp {
		s += ui.RenderHelpPanel(m.width, m.searchMode, string(m.roundMode), m.favoritesOnly, m.showDetail)
		s += "\n"
	}

	if m.loading && m.tournament == nil {
		s += "\n"
		styles := ui.DefaultStyles()
		s += styles.Loading.Render("  Fetching leaderboard data...")
		s += "\n"
	} else if m.tournament != nil {
		s += m.renderLeaderboard()
	} else if m.lastError != "" {
		s += "\n"
		styles := ui.DefaultStyles()
		s += styles.Error.Render(fmt.Sprintf("  Error: %s", m.lastError))
		s += "\n\n"
		s += styles.StatusDim.Render("  Press r to retry, q to quit")
		s += "\n"
	}

	// Status bar at the bottom
	var nextRefresh time.Duration
	if !m.nextRefreshAt.IsZero() {
		nextRefresh = time.Until(m.nextRefreshAt)
		if nextRefresh < 0 {
			nextRefresh = 0
		}
	}

	// Only show the error in the status bar if we also have tournament data
	// (i.e., a refresh failed but we still have stale data to show)
	statusErr := m.statusError()
	s += ui.RenderStatusBar(m.lastUpdate, nextRefresh, m.width, statusErr, m.filterQuery, m.searchMode, m.showHelp, string(m.roundMode), m.favoritesOnly, m.showDetail)

	return s
}

// renderLeaderboard renders the player table.
func (m Model) renderLeaderboard() string {
	if m.tournament == nil || len(m.tournament.Players) == 0 {
		styles := ui.DefaultStyles()
		return styles.StatusDim.Render("  No players to display\n")
	}

	t := m.tournament
	players := m.visiblePlayers()
	if len(players) == 0 {
		styles := ui.DefaultStyles()
		if m.favoritesOnly {
			if m.filterQuery != "" {
				return styles.StatusDim.Render(fmt.Sprintf("  No favorites match %q\n", m.filterQuery))
			}
			return styles.StatusDim.Render("  No favorite players selected\n")
		}
		if m.filterQuery != "" {
			return styles.StatusDim.Render(fmt.Sprintf("  No players match %q\n", m.filterQuery))
		}
		return styles.StatusDim.Render("  No players to display\n")
	}

	// Determine total rounds to show
	totalRounds := maxRounds(t)

	var s string
	styles := ui.DefaultStyles()
	if m.favoritesOnly {
		s += styles.StatusValue.Render("  Favorites only")
		s += "\n"
	}

	// Table header
	s += ui.RenderTableHeader(m.width, totalRounds)
	s += "\n"

	// Visible rows after reserving space for header, table chrome,
	// scroll indicator, and the two-line status bar block.
	visibleRows := m.visibleRows()

	// Determine the cut line position
	cutLine := findCutLine(players)

	// Calculate visible range based on scroll position
	startIdx := m.scrollPos
	endIdx := startIdx + visibleRows
	if endIdx > len(players) {
		endIdx = len(players)
	}

	// Render visible rows
	for i := startIdx; i < endIdx; i++ {
		p := players[i]

		// Insert cut line if applicable
		if cutLine > 0 && i == cutLine && i > startIdx {
			s += ui.RenderCutLine(m.width)
			s += "\n"
		}

		s += ui.RenderPlayerRow(
			p,
			i,
			m.width,
			totalRounds,
			cutLine,
			m.roundMode == roundScoreDisplayToPar,
			string(m.changeFor(p.ID)),
			m.scoreUpdated(p.ID),
			m.standingUpdated(p.ID),
			p.ID == m.selectedID,
			m.isFavorite(p.ID),
		)
		s += "\n"

		if m.showDetail && p.ID == m.selectedID {
			s += ui.RenderPlayerDetail(p, m.width, m.detailRound)
			s += "\n"
		}
	}

	// Scroll indicator
	if len(players) > visibleRows {
		styles := ui.DefaultStyles()
		indicator := fmt.Sprintf("  Showing %d-%d of %d players",
			startIdx+1, endIdx, len(players))
		s += styles.StatusDim.Render(indicator)
		s += "\n"
	}

	return s
}

// handleKey processes keyboard input.
func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.Key()

	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}

	if m.searchMode {
		return m.handleSearchKey(key)
	}

	if key.Text == "/" {
		m.searchMode = true
		return m, nil
	}

	switch msg.String() {
	case "q":
		return m, tea.Quit

	case "up", "k":
		m.moveSelection(-1)
		return m, nil

	case "down", "j":
		m.moveSelection(1)
		return m, nil

	case "ctrl+u", "pgup":
		step := max(1, m.visibleRows()/2)
		m.moveSelection(-step)
		return m, nil

	case "ctrl+d", "pgdown":
		step := max(1, m.visibleRows()/2)
		m.moveSelection(step)
		return m, nil

	case "home", "g":
		m.selectFirstVisible()
		return m, nil

	case "end", "G":
		m.selectLastVisible()
		return m, nil

	case "f":
		m.toggleFavoriteSelected()
		return m, nil

	case "F":
		m.favoritesOnly = !m.favoritesOnly
		m.syncVisibleState()
		return m, nil

	case "r":
		m.loading = true
		m.lastError = ""
		return m, m.fetchData()

	case "enter":
		m.toggleDetail()
		return m, nil

	case "esc":
		if m.showDetail {
			m.showDetail = false
			return m, nil
		}

	case "tab":
		if m.showDetail {
			m.cycleDetailRound(1)
			return m, nil
		}

	case "t":
		m.toggleRoundMode()
		return m, nil

	case "?":
		m.showHelp = !m.showHelp
		return m, nil
	}

	return m, nil
}

func (m Model) handleSearchKey(key tea.Key) (tea.Model, tea.Cmd) {
	switch key.Code {
	case tea.KeyEscape:
		m.searchMode = false
		m.setFilterQuery("")
		return m, nil
	case tea.KeyEnter:
		m.searchMode = false
		return m, nil
	case tea.KeyBackspace, tea.KeyDelete:
		query := []rune(m.filterQuery)
		if len(query) > 0 {
			m.setFilterQuery(string(query[:len(query)-1]))
		}
		return m, nil
	}

	if key.Text != "" {
		m.setFilterQuery(m.filterQuery + key.Text)
	}

	return m, nil
}

// fetchData creates a command that fetches tournament data from ESPN.
func (m Model) fetchData() tea.Cmd {
	return func() tea.Msg {
		tournament, err := m.client.FetchLeaderboard()
		if err != nil {
			return DataErrorMsg{Err: err}
		}
		return DataFetchedMsg{
			Tournament: tournament,
			FetchedAt:  time.Now(),
		}
	}
}

func (m Model) loadFavorites() tea.Cmd {
	if m.favoritesStore == nil {
		return nil
	}

	return func() tea.Msg {
		favorites, err := m.favoritesStore.Load()
		return FavoritesLoadedMsg{
			Favorites: favorites,
			Err:       err,
		}
	}
}

// maxRounds determines the maximum number of rounds to display columns for.
func maxRounds(t *espn.Tournament) int {
	max := 4 // always show 4 round columns for a standard tournament
	if t.Round > 0 && t.Round <= 4 {
		return max
	}
	return max
}

// findCutLine returns the index of the first player who missed the cut.
// Returns -1 if no cut line exists.
func findCutLine(players []espn.Player) int {
	for i, p := range players {
		if p.Status == "CUT" {
			return i
		}
	}
	return -1
}

func (m Model) filteredPlayers() []espn.Player {
	if m.tournament == nil {
		return nil
	}
	return filterPlayers(m.tournament.Players, m.filterQuery)
}

func (m Model) visiblePlayers() []espn.Player {
	players := m.filteredPlayers()
	if !m.favoritesOnly {
		return players
	}
	return filterFavoritePlayers(players, m.favorites)
}

func (m Model) isFavorite(playerID string) bool {
	return m.favorites[playerID]
}

func (m Model) changeFor(playerID string) playerChange {
	if len(m.changes) == 0 {
		return playerChangeNone
	}
	return m.changes[playerID]
}

func (m Model) updateFor(playerID string) playerUpdate {
	if len(m.updates) == 0 {
		return playerUpdateNone
	}
	return m.updates[playerID]
}

func (m Model) scoreUpdated(playerID string) bool {
	return m.updateFor(playerID)&playerUpdateScore != 0
}

func (m Model) standingUpdated(playerID string) bool {
	return m.updateFor(playerID)&playerUpdateStanding != 0
}

func filterPlayers(players []espn.Player, query string) []espn.Player {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return players
	}

	filtered := make([]espn.Player, 0, len(players))
	for _, p := range players {
		if strings.Contains(strings.ToLower(p.Name), query) {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

func filterFavoritePlayers(players []espn.Player, favorites map[string]bool) []espn.Player {
	if len(players) == 0 || len(favorites) == 0 {
		return nil
	}

	favoritesOnly := make([]espn.Player, 0, len(players))
	for _, p := range players {
		if favorites[p.ID] {
			favoritesOnly = append(favoritesOnly, p)
		}
	}
	return favoritesOnly
}

func copyFavorites(favorites map[string]bool) map[string]bool {
	if len(favorites) == 0 {
		return make(map[string]bool)
	}

	copy := make(map[string]bool, len(favorites))
	for playerID, favorite := range favorites {
		if favorite {
			copy[playerID] = true
		}
	}
	return copy
}

func computeRoundChanges(tournament *espn.Tournament) map[string]playerChange {
	if tournament == nil || len(tournament.Players) == 0 || tournament.Round <= 1 {
		return nil
	}

	previousPositions := previousRoundPositions(tournament)
	if len(previousPositions) == 0 {
		return nil
	}

	changes := make(map[string]playerChange, len(tournament.Players))

	for _, p := range tournament.Players {
		if p.Status != "" {
			continue
		}
		previousPosition, ok := previousPositions[p.ID]
		if !ok || p.DisplayPosition <= 0 {
			continue
		}

		diff := previousPosition - p.DisplayPosition
		switch {
		case diff > 0:
			changes[p.ID] = playerChange(fmt.Sprintf("+%d", diff))
		case diff < 0:
			changes[p.ID] = playerChange(fmt.Sprintf("%d", diff))
		default:
			changes[p.ID] = playerChangeEven
		}
	}

	return changes
}

func buildPlayerSnapshot(tournament *espn.Tournament) map[string]playerSnapshot {
	if tournament == nil || len(tournament.Players) == 0 {
		return nil
	}

	snapshot := make(map[string]playerSnapshot, len(tournament.Players))
	for _, p := range tournament.Players {
		if p.ID == "" {
			continue
		}
		snapshot[p.ID] = playerSnapshot{
			ID:              p.ID,
			CanonicalRank:   p.CanonicalRank,
			DisplayPosition: p.DisplayPosition,
			TotalScore:      p.TotalScore,
			Status:          p.Status,
		}
	}

	if len(snapshot) == 0 {
		return nil
	}

	return snapshot
}

func computeRefreshUpdates(previous map[string]playerSnapshot, tournament *espn.Tournament) map[string]playerUpdate {
	if len(previous) == 0 || tournament == nil || len(tournament.Players) == 0 {
		return nil
	}

	updates := make(map[string]playerUpdate)
	for _, p := range tournament.Players {
		prev, ok := previous[p.ID]
		if !ok {
			continue
		}

		var update playerUpdate
		if prev.TotalScore != p.TotalScore {
			update |= playerUpdateScore
		}
		if prev.CanonicalRank != p.CanonicalRank || prev.DisplayPosition != p.DisplayPosition || prev.Status != p.Status {
			update |= playerUpdateStanding
		}

		if update != playerUpdateNone {
			updates[p.ID] = update
		}
	}

	if len(updates) == 0 {
		return nil
	}

	return updates
}

type roundStanding struct {
	ID            string
	PreviousTotal int
	CurrentRank   int
}

func previousRoundPositions(tournament *espn.Tournament) map[string]int {
	standings := make([]roundStanding, 0, len(tournament.Players))
	for _, p := range tournament.Players {
		if p.Status != "" {
			continue
		}

		previousTotal, ok := previousRoundTotal(p, tournament.Round)
		if !ok {
			continue
		}

		standings = append(standings, roundStanding{
			ID:            p.ID,
			PreviousTotal: previousTotal,
			CurrentRank:   p.CanonicalRank,
		})
	}

	if len(standings) == 0 {
		return nil
	}

	slices.SortFunc(standings, func(a, b roundStanding) int {
		if a.PreviousTotal != b.PreviousTotal {
			return a.PreviousTotal - b.PreviousTotal
		}
		return a.CurrentRank - b.CurrentRank
	})

	positions := make(map[string]int, len(standings))
	for i := 0; i < len(standings); {
		start := i + 1
		score := standings[i].PreviousTotal
		j := i
		for j < len(standings) && standings[j].PreviousTotal == score {
			positions[standings[j].ID] = start
			j++
		}
		i = j
	}

	return positions
}

func previousRoundTotal(player espn.Player, currentRound int) (int, bool) {
	total, ok := parseRelativeScore(player.TotalScore)
	if !ok {
		return 0, false
	}

	currentRoundIndex := currentRound - 1
	if currentRoundIndex <= 0 || currentRoundIndex >= len(player.Rounds) {
		return total, true
	}

	currentRoundScore := player.Rounds[currentRoundIndex]
	if !currentRoundScore.Played || currentRoundScore.ToPar == "" {
		return total, true
	}

	roundTotal, ok := parseRelativeScore(currentRoundScore.ToPar)
	if !ok {
		return total, true
	}

	return total - roundTotal, true
}

func parseRelativeScore(score string) (int, bool) {
	score = strings.TrimSpace(strings.ToUpper(score))
	if score == "" || score == "-" {
		return 0, false
	}
	if score == "E" {
		return 0, true
	}

	value, err := strconv.Atoi(score)
	if err != nil {
		return 0, false
	}
	return value, true
}

func playerIndexByID(players []espn.Player, playerID string) int {
	if playerID == "" {
		return -1
	}
	for i, p := range players {
		if p.ID == playerID {
			return i
		}
	}
	return -1
}

func (m *Model) setFilterQuery(query string) {
	m.filterQuery = query
	m.scrollPos = 0
	m.syncVisibleState()
}

func (m *Model) clampScroll() {
	players := m.visiblePlayers()
	visibleRows := m.visibleRows()

	maxScroll := len(players) - visibleRows
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.scrollPos > maxScroll {
		m.scrollPos = maxScroll
	}
	if m.scrollPos < 0 {
		m.scrollPos = 0
	}
}

func (m *Model) syncVisibleState() {
	players := m.visiblePlayers()
	if len(players) == 0 {
		m.selectedID = ""
		m.scrollPos = 0
		m.showDetail = false
		return
	}

	if playerIndexByID(players, m.selectedID) < 0 {
		m.selectedID = players[0].ID
	}
	m.normalizeDetailRound()

	m.clampScroll()
	m.ensureSelectionVisible()
}

func (m *Model) ensureSelectionVisible() {
	players := m.visiblePlayers()
	selectedIdx := playerIndexByID(players, m.selectedID)
	if selectedIdx < 0 {
		return
	}

	visibleRows := m.visibleRows()
	maxScroll := len(players) - visibleRows
	if maxScroll < 0 {
		maxScroll = 0
	}

	upperBound := m.scrollPos + scrollPadding
	lowerBound := m.scrollPos + visibleRows - scrollPadding - 1
	if lowerBound < m.scrollPos {
		lowerBound = m.scrollPos
	}

	if selectedIdx < upperBound {
		m.scrollPos = selectedIdx - scrollPadding
	} else if selectedIdx > lowerBound {
		m.scrollPos = selectedIdx - visibleRows + scrollPadding + 1
	}

	if m.scrollPos < 0 {
		m.scrollPos = 0
	}
	if m.scrollPos > maxScroll {
		m.scrollPos = maxScroll
	}
}

func (m *Model) moveSelection(delta int) {
	players := m.visiblePlayers()
	if len(players) == 0 {
		return
	}

	selectedIdx := playerIndexByID(players, m.selectedID)
	if selectedIdx < 0 {
		selectedIdx = 0
	}

	selectedIdx += delta
	if selectedIdx < 0 {
		selectedIdx = 0
	}
	if selectedIdx >= len(players) {
		selectedIdx = len(players) - 1
	}

	m.selectedID = players[selectedIdx].ID
	m.normalizeDetailRound()
	m.ensureSelectionVisible()
}

func (m *Model) selectFirstVisible() {
	players := m.visiblePlayers()
	if len(players) == 0 {
		return
	}
	m.selectedID = players[0].ID
	m.normalizeDetailRound()
	m.ensureSelectionVisible()
}

func (m *Model) selectLastVisible() {
	players := m.visiblePlayers()
	if len(players) == 0 {
		return
	}
	m.selectedID = players[len(players)-1].ID
	m.normalizeDetailRound()
	m.ensureSelectionVisible()
}

func (m *Model) toggleFavoriteSelected() {
	players := m.visiblePlayers()
	selectedIdx := playerIndexByID(players, m.selectedID)
	if selectedIdx < 0 {
		return
	}

	playerID := players[selectedIdx].ID
	if m.isFavorite(playerID) {
		delete(m.favorites, playerID)
	} else {
		m.favorites[playerID] = true
	}

	m.syncVisibleState()
	if m.favoritesStore == nil {
		m.favoritesErr = ""
		return
	}
	if err := m.favoritesStore.Save(copyFavorites(m.favorites)); err != nil {
		m.favoritesErr = formatFavoritesError("save", err)
		return
	}
	m.favoritesErr = ""
}

func (m Model) visibleRows() int {
	visibleRows := m.height - 11 - m.helpPanelHeight() - m.detailPanelHeight()
	if visibleRows < 1 {
		visibleRows = 1
	}
	return visibleRows
}

func (m Model) helpPanelHeight() int {
	if !m.showHelp {
		return 0
	}
	return ui.HelpPanelLineCount(m.searchMode, m.showDetail) + 1
}

func (m Model) detailPanelHeight() int {
	if !m.showDetail {
		return 0
	}
	player, ok := m.selectedPlayer()
	if !ok {
		return 0
	}
	return ui.DetailPanelHeight(player, m.width, m.detailRound)
}

func (m *Model) toggleRoundMode() {
	if m.roundMode == roundScoreDisplayToPar {
		m.roundMode = roundScoreDisplayStrokes
		return
	}
	m.roundMode = roundScoreDisplayToPar
}

func (m Model) statusError() string {
	errors := make([]string, 0, 2)
	if m.lastError != "" && m.tournament != nil {
		errors = append(errors, m.lastError)
	}
	if m.favoritesErr != "" {
		errors = append(errors, m.favoritesErr)
	}
	return strings.Join(errors, " | ")
}

func formatFavoritesError(action string, err error) string {
	if err == nil {
		return ""
	}
	return fmt.Sprintf("Favorites %s failed: %v", action, err)
}

func (m Model) selectedPlayer() (espn.Player, bool) {
	players := m.visiblePlayers()
	selectedIdx := playerIndexByID(players, m.selectedID)
	if selectedIdx < 0 {
		return espn.Player{}, false
	}
	return players[selectedIdx], true
}

func (m *Model) toggleDetail() {
	if m.showDetail {
		m.showDetail = false
		return
	}

	player, ok := m.selectedPlayer()
	if !ok {
		return
	}

	m.showDetail = true
	m.detailRound = defaultDetailRound(player)
	m.syncVisibleState()
}

func (m *Model) cycleDetailRound(delta int) {
	if !m.showDetail {
		return
	}

	player, ok := m.selectedPlayer()
	if !ok {
		return
	}

	totalRounds := max(4, len(player.Rounds))
	next := m.detailRound + delta
	if next > totalRounds {
		next = 1
	}
	if next < 1 {
		next = totalRounds
	}
	if next == m.detailRound {
		return
	}
	m.detailRound = next
	m.syncVisibleState()
}

func (m *Model) normalizeDetailRound() {
	if !m.showDetail {
		return
	}
	player, ok := m.selectedPlayer()
	if !ok {
		return
	}
	totalRounds := max(4, len(player.Rounds))
	if m.detailRound < 1 || m.detailRound > totalRounds {
		m.detailRound = defaultDetailRound(player)
	}
}

func defaultDetailRound(player espn.Player) int {
	for i := len(player.Rounds) - 1; i >= 0; i-- {
		round := player.Rounds[i]
		if round.Played || len(round.Holes) > 0 {
			return i + 1
		}
	}
	if len(player.Rounds) > 0 {
		return 1
	}
	return 1
}
