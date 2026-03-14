package model

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/nickkoul/gstat/internal/espn"
	"github.com/nickkoul/gstat/internal/ui"
)

const (
	defaultRefreshInterval = 30 * time.Second
	scrollPadding          = 3 // rows of padding when scrolling near edges
)

// Model is the main Bubble Tea model for the leaderboard view.
type Model struct {
	// Data
	tournament  *espn.Tournament
	client      *espn.Client
	lastUpdate  time.Time
	lastError   string

	// UI state
	width       int
	height      int
	scrollPos   int
	showHelp    bool

	// Refresh
	refreshInterval time.Duration
	nextRefreshAt   time.Time
	loading         bool
}

// New creates a new leaderboard model.
func New() Model {
	return Model{
		client:          espn.NewClient(),
		refreshInterval: defaultRefreshInterval,
		loading:         true,
	}
}

// Init initializes the model. It kicks off the first data fetch.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.fetchData(),
		tickCmd(time.Second), // 1-second tick for countdown display
	)
}

// Update handles messages and updates the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case DataFetchedMsg:
		m.tournament = msg.Tournament
		m.lastUpdate = msg.FetchedAt
		m.lastError = ""
		m.loading = false
		m.nextRefreshAt = time.Now().Add(m.refreshInterval)
		return m, nil

	case DataErrorMsg:
		m.lastError = msg.Err.Error()
		m.loading = false
		m.nextRefreshAt = time.Now().Add(m.refreshInterval)
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
	statusErr := ""
	if m.lastError != "" && m.tournament != nil {
		statusErr = m.lastError
	}
	s += ui.RenderStatusBar(m.lastUpdate, nextRefresh, m.width, statusErr)

	return s
}

// renderLeaderboard renders the player table.
func (m Model) renderLeaderboard() string {
	if m.tournament == nil || len(m.tournament.Players) == 0 {
		styles := ui.DefaultStyles()
		return styles.StatusDim.Render("  No players to display\n")
	}

	t := m.tournament

	// Determine total rounds to show
	totalRounds := maxRounds(t)

	var s string

	// Table header
	s += ui.RenderTableHeader(m.width, totalRounds)
	s += "\n"

	// Visible rows (accounting for header, footer, etc.)
	// Header takes ~4 lines, status bar ~3 lines, table header ~2 lines
	visibleRows := m.height - 9
	if visibleRows < 5 {
		visibleRows = 5
	}

	// Determine the cut line position
	cutLine := findCutLine(t.Players)

	// Calculate visible range based on scroll position
	startIdx := m.scrollPos
	endIdx := startIdx + visibleRows
	if endIdx > len(t.Players) {
		endIdx = len(t.Players)
	}

	// Render visible rows
	for i := startIdx; i < endIdx; i++ {
		p := t.Players[i]

		// Insert cut line if applicable
		if cutLine > 0 && i == cutLine && i > startIdx {
			s += ui.RenderCutLine(m.width)
			s += "\n"
		}

		s += ui.RenderPlayerRow(p, i, m.width, totalRounds, cutLine)
		s += "\n"
	}

	// Scroll indicator
	if len(t.Players) > visibleRows {
		styles := ui.DefaultStyles()
		indicator := fmt.Sprintf("  Showing %d-%d of %d players",
			startIdx+1, endIdx, len(t.Players))
		s += styles.StatusDim.Render(indicator)
		s += "\n"
	}

	return s
}

// handleKey processes keyboard input.
func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "up", "k":
		if m.scrollPos > 0 {
			m.scrollPos--
		}
		return m, nil

	case "down", "j":
		if m.tournament != nil {
			visibleRows := m.height - 9
			if visibleRows < 5 {
				visibleRows = 5
			}
			maxScroll := len(m.tournament.Players) - visibleRows
			if maxScroll < 0 {
				maxScroll = 0
			}
			if m.scrollPos < maxScroll {
				m.scrollPos++
			}
		}
		return m, nil

	case "ctrl+u", "pgup":
		visibleRows := m.height - 9
		if visibleRows < 5 {
			visibleRows = 5
		}
		m.scrollPos -= visibleRows / 2
		if m.scrollPos < 0 {
			m.scrollPos = 0
		}
		return m, nil

	case "ctrl+d", "pgdown":
		if m.tournament != nil {
			visibleRows := m.height - 9
			if visibleRows < 5 {
				visibleRows = 5
			}
			maxScroll := len(m.tournament.Players) - visibleRows
			if maxScroll < 0 {
				maxScroll = 0
			}
			m.scrollPos += visibleRows / 2
			if m.scrollPos > maxScroll {
				m.scrollPos = maxScroll
			}
		}
		return m, nil

	case "home", "g":
		m.scrollPos = 0
		return m, nil

	case "end", "G":
		if m.tournament != nil {
			visibleRows := m.height - 9
			maxScroll := len(m.tournament.Players) - visibleRows
			if maxScroll < 0 {
				maxScroll = 0
			}
			m.scrollPos = maxScroll
		}
		return m, nil

	case "r":
		m.loading = true
		m.lastError = ""
		return m, m.fetchData()

	case "?":
		m.showHelp = !m.showHelp
		return m, nil
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
