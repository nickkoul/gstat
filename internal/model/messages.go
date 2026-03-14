package model

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/nickkoul/gstat/internal/espn"
)

// DataFetchedMsg is sent when tournament data is successfully fetched.
type DataFetchedMsg struct {
	Tournament *espn.Tournament
	FetchedAt  time.Time
}

// DataErrorMsg is sent when fetching tournament data fails.
type DataErrorMsg struct {
	Err error
}

// FavoritesLoadedMsg is sent when persisted favorites finish loading.
type FavoritesLoadedMsg struct {
	Favorites map[string]bool
	Err       error
}

// TickMsg is sent on each refresh timer tick.
type TickMsg time.Time

// tickCmd returns a command that sends a TickMsg after the given duration.
func tickCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}
