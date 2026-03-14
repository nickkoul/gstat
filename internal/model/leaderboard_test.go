package model

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/nickkoul/gstat/internal/espn"
)

func keyWithText(text string) tea.KeyPressMsg {
	var code rune
	if runes := []rune(text); len(runes) == 1 {
		code = runes[0]
	}
	return tea.KeyPressMsg(tea.Key{Text: text, Code: code})
}

func keyWithCode(code rune) tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Code: code})
}

func pressKey(t *testing.T, m Model, msg tea.KeyPressMsg) Model {
	t.Helper()
	updated, _ := m.handleKey(msg)
	next, ok := updated.(Model)
	if !ok {
		t.Fatalf("handleKey returned %T, want model.Model", updated)
	}
	return next
}

func testTournament(names ...string) *espn.Tournament {
	players := make([]espn.Player, 0, len(names))
	for i, name := range names {
		players = append(players, espn.Player{
			ID:              name,
			CanonicalRank:   i + 1,
			DisplayPosition: i + 1,
			Name:            name,
			TotalScore:      "E",
			Thru:            "F",
		})
	}
	return &espn.Tournament{Players: players}
}

func TestFilterPlayersCaseInsensitive(t *testing.T) {
	players := []espn.Player{
		{Name: "Scottie Scheffler"},
		{Name: "Ludvig Aberg"},
		{Name: "Xander Schauffele"},
	}

	filtered := filterPlayers(players, "SCHEF")
	if len(filtered) != 1 {
		t.Fatalf("filtered len = %d, want 1", len(filtered))
	}
	if filtered[0].Name != "Scottie Scheffler" {
		t.Fatalf("filtered[0] = %q, want Scottie Scheffler", filtered[0].Name)
	}
}

func TestHandleKeyStartsSearchMode(t *testing.T) {
	m := New()
	m = pressKey(t, m, keyWithText("/"))

	if !m.searchMode {
		t.Fatal("searchMode = false, want true")
	}
	if m.filterQuery != "" {
		t.Fatalf("filterQuery = %q, want empty", m.filterQuery)
	}
}

func TestHandleSearchKeyUpdatesQueryAndApplies(t *testing.T) {
	m := New()
	m.height = 20
	m.tournament = testTournament("Scottie Scheffler", "Ludvig Aberg")
	m = pressKey(t, m, keyWithText("/"))
	m = pressKey(t, m, keyWithText("s"))
	m = pressKey(t, m, keyWithText("c"))
	m = pressKey(t, m, keyWithCode(tea.KeyEnter))

	if m.searchMode {
		t.Fatal("searchMode = true, want false")
	}
	if m.filterQuery != "sc" {
		t.Fatalf("filterQuery = %q, want sc", m.filterQuery)
	}
	filtered := m.filteredPlayers()
	if len(filtered) != 1 || filtered[0].Name != "Scottie Scheffler" {
		t.Fatalf("filtered players = %+v, want Scottie Scheffler only", filtered)
	}
}

func TestHandleSearchKeyEscapeClearsFilter(t *testing.T) {
	m := New()
	m.height = 20
	m.scrollPos = 4
	m.tournament = testTournament("Scottie Scheffler", "Shane Lowry")
	m.searchMode = true
	m.filterQuery = "sc"

	m = pressKey(t, m, keyWithCode(tea.KeyEscape))

	if m.searchMode {
		t.Fatal("searchMode = true, want false")
	}
	if m.filterQuery != "" {
		t.Fatalf("filterQuery = %q, want empty", m.filterQuery)
	}
	if m.scrollPos != 0 {
		t.Fatalf("scrollPos = %d, want 0", m.scrollPos)
	}
}

func TestSetFilterQueryResetsScroll(t *testing.T) {
	m := New()
	m.height = 20
	m.scrollPos = 8
	m.tournament = testTournament("Scottie Scheffler", "Shane Lowry", "Xander Schauffele")

	m.setFilterQuery("shan")

	if m.scrollPos != 0 {
		t.Fatalf("scrollPos = %d, want 0", m.scrollPos)
	}
	filtered := m.filteredPlayers()
	if len(filtered) != 1 || filtered[0].Name != "Shane Lowry" {
		t.Fatalf("filtered players = %+v, want Shane Lowry only", filtered)
	}
}

func TestDataFetchedPreservesActiveFilter(t *testing.T) {
	m := New()
	m.height = 20
	m.filterQuery = "scott"

	updated, _ := m.Update(DataFetchedMsg{
		Tournament: testTournament("Scottie Scheffler", "Rory McIlroy"),
		FetchedAt:  time.Now(),
	})
	next := updated.(Model)

	if next.filterQuery != "scott" {
		t.Fatalf("filterQuery = %q, want scott", next.filterQuery)
	}
	filtered := next.filteredPlayers()
	if len(filtered) != 1 || filtered[0].Name != "Scottie Scheffler" {
		t.Fatalf("filtered players = %+v, want Scottie Scheffler only", filtered)
	}
}

func TestRenderLeaderboardNoMatches(t *testing.T) {
	m := New()
	m.width = 80
	m.height = 20
	m.tournament = testTournament("Scottie Scheffler", "Rory McIlroy")
	m.filterQuery = "zzz"

	out := m.renderLeaderboard()
	if !strings.Contains(out, "No players match \"zzz\"") {
		t.Fatalf("expected no-match message, got %q", out)
	}
}

func TestHandleKeyTogglesRoundDisplayMode(t *testing.T) {
	m := New()
	if m.roundMode != roundScoreDisplayToPar {
		t.Fatalf("roundMode = %q, want %q", m.roundMode, roundScoreDisplayToPar)
	}

	m = pressKey(t, m, keyWithText("t"))
	if m.roundMode != roundScoreDisplayStrokes {
		t.Fatalf("roundMode = %q, want %q", m.roundMode, roundScoreDisplayStrokes)
	}

	m = pressKey(t, m, keyWithText("t"))
	if m.roundMode != roundScoreDisplayToPar {
		t.Fatalf("roundMode = %q, want %q", m.roundMode, roundScoreDisplayToPar)
	}
}

func TestHandleKeyTogglesHelpPanel(t *testing.T) {
	m := New()
	m = pressKey(t, m, keyWithText("?"))
	if !m.showHelp {
		t.Fatal("showHelp = false, want true")
	}

	m = pressKey(t, m, keyWithText("?"))
	if m.showHelp {
		t.Fatal("showHelp = true, want false")
	}
}

func TestVisibleRowsShrinksWhenHelpPanelOpen(t *testing.T) {
	m := New()
	m.height = 24
	closedRows := m.visibleRows()
	m.showHelp = true
	openRows := m.visibleRows()

	if openRows >= closedRows {
		t.Fatalf("openRows = %d, want less than %d when help is open", openRows, closedRows)
	}
}

func TestVisibleRowsKeepsStatusBarSpace(t *testing.T) {
	m := New()
	m.height = 10

	if got := m.visibleRows(); got != 1 {
		t.Fatalf("visibleRows = %d, want 1 for short terminals", got)
	}
}
