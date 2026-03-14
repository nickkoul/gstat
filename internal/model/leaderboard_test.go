package model

import (
	"errors"
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

type stubFavoritesStore struct {
	loadFavorites map[string]bool
	loadErr       error
	saveErr       error
	saved         []map[string]bool
}

func (s *stubFavoritesStore) Load() (map[string]bool, error) {
	return copyFavorites(s.loadFavorites), s.loadErr
}

func (s *stubFavoritesStore) Save(favorites map[string]bool) error {
	s.saved = append(s.saved, copyFavorites(favorites))
	return s.saveErr
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

func TestDataFetchedDoesNotShowChangeInRoundOne(t *testing.T) {
	m := New()

	updated, _ := m.Update(DataFetchedMsg{
		Tournament: &espn.Tournament{
			Round: 1,
			Players: []espn.Player{
				{ID: "scottie", CanonicalRank: 1, DisplayPosition: 1, Name: "Scottie Scheffler", TotalScore: "-5", Thru: "12", Rounds: []espn.RoundScore{{Round: 1, Played: true, ToPar: "-5"}}},
				{ID: "rory", CanonicalRank: 2, DisplayPosition: 2, Name: "Rory McIlroy", TotalScore: "-4", Thru: "12", Rounds: []espn.RoundScore{{Round: 1, Played: true, ToPar: "-4"}}},
			},
		},
		FetchedAt: time.Now(),
	})
	next := updated.(Model)

	if got := next.changeFor("scottie"); got != playerChangeNone {
		t.Fatalf("change = %q, want none in round one", got)
	}
	if got := next.changeFor("rory"); got != playerChangeNone {
		t.Fatalf("change = %q, want none in round one", got)
	}
}

func TestDataFetchedComputesRoundChangeFromPreviousRound(t *testing.T) {
	m := New()
	tournament := &espn.Tournament{Round: 2, Players: []espn.Player{
		{ID: "scottie", CanonicalRank: 1, DisplayPosition: 1, Name: "Scottie Scheffler", TotalScore: "-9", Thru: "F", Rounds: []espn.RoundScore{{Round: 1, Played: true, ToPar: "-5"}, {Round: 2, Played: true, ToPar: "-4"}}},
		{ID: "rory", CanonicalRank: 2, DisplayPosition: 2, Name: "Rory McIlroy", TotalScore: "-8", Thru: "F", Rounds: []espn.RoundScore{{Round: 1, Played: true, ToPar: "-6"}, {Round: 2, Played: true, ToPar: "-2"}}},
		{ID: "xander", CanonicalRank: 3, DisplayPosition: 3, Name: "Xander Schauffele", TotalScore: "-7", Thru: "F", Rounds: []espn.RoundScore{{Round: 1, Played: true, ToPar: "-4"}, {Round: 2, Played: true, ToPar: "-3"}}},
	}}

	updated, _ := m.Update(DataFetchedMsg{Tournament: tournament, FetchedAt: time.Now()})
	next := updated.(Model)

	if got := next.changeFor("scottie"); got != "+1" {
		t.Fatalf("scottie change = %q, want +1", got)
	}
	if got := next.changeFor("rory"); got != "-1" {
		t.Fatalf("rory change = %q, want -1", got)
	}
	if got := next.changeFor("xander"); got != playerChangeEven {
		t.Fatalf("xander change = %q, want %q", got, playerChangeEven)
	}
}

func TestDataFetchedSuppressesTieGroupNoise(t *testing.T) {
	m := New()
	tournament := &espn.Tournament{Round: 2, Players: []espn.Player{
		{ID: "rory", CanonicalRank: 1, DisplayPosition: 1, Tied: true, Name: "Rory McIlroy", TotalScore: "-8", Thru: "F", Rounds: []espn.RoundScore{{Round: 1, Played: true, ToPar: "-8"}, {Round: 2, Played: true, ToPar: "E"}}},
		{ID: "xander", CanonicalRank: 2, DisplayPosition: 1, Tied: true, Name: "Xander Schauffele", TotalScore: "-8", Thru: "F", Rounds: []espn.RoundScore{{Round: 1, Played: true, ToPar: "-8"}, {Round: 2, Played: true, ToPar: "E"}}},
	}}

	updated, _ := m.Update(DataFetchedMsg{Tournament: tournament, FetchedAt: time.Now()})
	next := updated.(Model)

	if got := next.changeFor("rory"); got != playerChangeEven {
		t.Fatalf("rory change = %q, want E for same tied standing", got)
	}
	if got := next.changeFor("xander"); got != playerChangeEven {
		t.Fatalf("xander change = %q, want E for same tied standing", got)
	}
}

func TestDataFetchedChangeUsesRoundStandingNotVisibleFilters(t *testing.T) {
	m := New()
	m.favoritesOnly = true
	m.filterQuery = "r"
	m.favorites["rory"] = true
	tournament := &espn.Tournament{Round: 2, Players: []espn.Player{
		{ID: "scottie", CanonicalRank: 1, DisplayPosition: 1, Name: "Scottie Scheffler", TotalScore: "-9", Thru: "F", Rounds: []espn.RoundScore{{Round: 1, Played: true, ToPar: "-5"}, {Round: 2, Played: true, ToPar: "-4"}}},
		{ID: "rory", CanonicalRank: 2, DisplayPosition: 2, Name: "Rory McIlroy", TotalScore: "-8", Thru: "F", Rounds: []espn.RoundScore{{Round: 1, Played: true, ToPar: "-6"}, {Round: 2, Played: true, ToPar: "-2"}}},
		{ID: "xander", CanonicalRank: 3, DisplayPosition: 3, Name: "Xander Schauffele", TotalScore: "-7", Thru: "F", Rounds: []espn.RoundScore{{Round: 1, Played: true, ToPar: "-4"}, {Round: 2, Played: true, ToPar: "-3"}}},
	}}

	updated, _ := m.Update(DataFetchedMsg{Tournament: tournament, FetchedAt: time.Now()})
	next := updated.(Model)

	visible := next.visiblePlayers()
	if len(visible) != 1 || visible[0].ID != "rory" {
		t.Fatalf("visible players = %+v, want Rory only", visible)
	}
	if got := next.changeFor("rory"); got != "-1" {
		t.Fatalf("rory change = %q, want -1 even in favorites-only filtered view", got)
	}
	if got := next.changeFor("scottie"); got != "+1" {
		t.Fatalf("scottie change = %q, want +1 based on round standings", got)
	}
}

func TestDataFetchedSelectsFirstVisiblePlayer(t *testing.T) {
	m := New()
	m.height = 20

	updated, _ := m.Update(DataFetchedMsg{
		Tournament: testTournament("Scottie Scheffler", "Rory McIlroy"),
		FetchedAt:  time.Now(),
	})
	next := updated.(Model)

	if next.selectedID != "Scottie Scheffler" {
		t.Fatalf("selectedID = %q, want Scottie Scheffler", next.selectedID)
	}
}

func TestUpdateFavoritesLoadedAppliesPersistedFavorites(t *testing.T) {
	m := New()
	m.height = 20
	m.favoritesOnly = true
	m.tournament = testTournament("Scottie Scheffler", "Rory McIlroy")

	updated, _ := m.Update(FavoritesLoadedMsg{
		Favorites: map[string]bool{"Rory McIlroy": true},
	})
	next := updated.(Model)

	if !next.isFavorite("Rory McIlroy") {
		t.Fatal("Rory McIlroy should be favorited after load")
	}
	visible := next.visiblePlayers()
	if len(visible) != 1 || visible[0].Name != "Rory McIlroy" {
		t.Fatalf("visible players = %+v, want only Rory after load", visible)
	}
	if next.selectedID != "Rory McIlroy" {
		t.Fatalf("selectedID = %q, want Rory McIlroy", next.selectedID)
	}
}

func TestUpdateFavoritesLoadedStoresLoadError(t *testing.T) {
	m := New()

	updated, _ := m.Update(FavoritesLoadedMsg{Err: errors.New("bad json")})
	next := updated.(Model)

	if !strings.Contains(next.favoritesErr, "Favorites load failed: bad json") {
		t.Fatalf("favoritesErr = %q, want load error", next.favoritesErr)
	}
	if len(next.favorites) != 0 {
		t.Fatalf("favorites len = %d, want 0", len(next.favorites))
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

func TestRenderLeaderboardShowsFavoritesOnlyLabel(t *testing.T) {
	m := New()
	m.width = 80
	m.height = 20
	m.favoritesOnly = true
	m.tournament = testTournament("Scottie Scheffler")
	m.favorites["Scottie Scheffler"] = true
	m.syncVisibleState()

	out := m.renderLeaderboard()
	if !strings.Contains(out, "Favorites only") {
		t.Fatalf("expected favorites-only label, got %q", out)
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

func TestHandleKeyMovesSelection(t *testing.T) {
	m := New()
	m.height = 20
	m.tournament = testTournament("Scottie Scheffler", "Rory McIlroy", "Xander Schauffele")
	m.syncVisibleState()

	m = pressKey(t, m, keyWithText("j"))
	if m.selectedID != "Rory McIlroy" {
		t.Fatalf("selectedID after j = %q, want Rory McIlroy", m.selectedID)
	}

	m = pressKey(t, m, keyWithText("k"))
	if m.selectedID != "Scottie Scheffler" {
		t.Fatalf("selectedID after k = %q, want Scottie Scheffler", m.selectedID)
	}
}

func TestHandleKeyToggleFavoriteKeepsVisibleOrder(t *testing.T) {
	m := New()
	m.height = 20
	m.tournament = testTournament("Scottie Scheffler", "Rory McIlroy", "Xander Schauffele")
	m.syncVisibleState()

	m = pressKey(t, m, keyWithText("j"))
	m = pressKey(t, m, keyWithText("f"))

	visible := m.visiblePlayers()
	if len(visible) != 3 {
		t.Fatalf("visible players len = %d, want 3", len(visible))
	}
	if visible[0].Name != "Scottie Scheffler" {
		t.Fatalf("visible[0] = %q, want Scottie Scheffler to stay first", visible[0].Name)
	}
	if !m.isFavorite("Rory McIlroy") {
		t.Fatal("Rory McIlroy should be favorited")
	}
	if m.selectedID != "Rory McIlroy" {
		t.Fatalf("selectedID = %q, want Rory McIlroy", m.selectedID)
	}
}

func TestHandleKeyToggleFavoritePersistsSelection(t *testing.T) {
	store := &stubFavoritesStore{}
	m := New()
	m.height = 20
	m.favoritesStore = store
	m.tournament = testTournament("Scottie Scheffler", "Rory McIlroy")
	m.syncVisibleState()

	m = pressKey(t, m, keyWithText("j"))
	m = pressKey(t, m, keyWithText("f"))

	if len(store.saved) != 1 {
		t.Fatalf("save calls = %d, want 1", len(store.saved))
	}
	if !store.saved[0]["Rory McIlroy"] {
		t.Fatalf("saved favorites = %#v, want Rory McIlroy", store.saved[0])
	}
	if m.favoritesErr != "" {
		t.Fatalf("favoritesErr = %q, want empty on successful save", m.favoritesErr)
	}
}

func TestHandleKeyToggleFavoriteSaveErrorKeepsInMemoryState(t *testing.T) {
	store := &stubFavoritesStore{saveErr: errors.New("disk full")}
	m := New()
	m.height = 20
	m.favoritesStore = store
	m.tournament = testTournament("Scottie Scheffler", "Rory McIlroy")
	m.syncVisibleState()

	m = pressKey(t, m, keyWithText("f"))

	if !m.isFavorite("Scottie Scheffler") {
		t.Fatal("Scottie Scheffler should remain favorited in memory after save failure")
	}
	if len(store.saved) != 1 || !store.saved[0]["Scottie Scheffler"] {
		t.Fatalf("saved favorites = %#v, want attempted Scottie save", store.saved)
	}
	if !strings.Contains(m.favoritesErr, "Favorites save failed: disk full") {
		t.Fatalf("favoritesErr = %q, want save error", m.favoritesErr)
	}
}

func TestHandleKeyToggleFavoriteClearsPreviousSaveErrorOnSuccess(t *testing.T) {
	store := &stubFavoritesStore{}
	m := New()
	m.height = 20
	m.favoritesStore = store
	m.favoritesErr = "Favorites save failed: disk full"
	m.tournament = testTournament("Scottie Scheffler")
	m.syncVisibleState()

	m = pressKey(t, m, keyWithText("f"))

	if m.favoritesErr != "" {
		t.Fatalf("favoritesErr = %q, want empty after successful save", m.favoritesErr)
	}
}

func TestVisiblePlayersNormalModeKeepsCanonicalOrder(t *testing.T) {
	m := New()
	m.tournament = testTournament("Scottie Scheffler", "Rory McIlroy", "Xander Schauffele", "Tommy Fleetwood")
	m.favorites["Rory McIlroy"] = true
	m.favorites["Tommy Fleetwood"] = true

	visible := m.visiblePlayers()
	got := []string{visible[0].Name, visible[1].Name, visible[2].Name, visible[3].Name}
	want := []string{"Scottie Scheffler", "Rory McIlroy", "Xander Schauffele", "Tommy Fleetwood"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("visible order = %v, want %v", got, want)
	}
}

func TestVisiblePlayersFavoritesRespectFilterInNormalMode(t *testing.T) {
	m := New()
	m.tournament = testTournament("Scottie Scheffler", "Shane Lowry", "Xander Schauffele")
	m.favorites["Shane Lowry"] = true
	m.filterQuery = "s"

	visible := m.visiblePlayers()
	got := []string{visible[0].Name, visible[1].Name, visible[2].Name}
	want := []string{"Scottie Scheffler", "Shane Lowry", "Xander Schauffele"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("visible order = %v, want %v", got, want)
	}
}

func TestVisiblePlayersFavoritesOnlyFiltersAndKeepsTournamentRank(t *testing.T) {
	m := New()
	m.favoritesOnly = true
	m.tournament = &espn.Tournament{Players: []espn.Player{
		{ID: "scottie", CanonicalRank: 1, DisplayPosition: 1, Name: "Scottie Scheffler", TotalScore: "-10", Thru: "F"},
		{ID: "rory", CanonicalRank: 5, DisplayPosition: 5, Tied: true, Name: "Rory McIlroy", TotalScore: "-8", Thru: "F"},
		{ID: "xander", CanonicalRank: 8, DisplayPosition: 8, Tied: true, Name: "Xander Schauffele", TotalScore: "-8", Thru: "F"},
		{ID: "tommy", CanonicalRank: 12, DisplayPosition: 12, Name: "Tommy Fleetwood", TotalScore: "-4", Thru: "F"},
	}}
	m.favorites["rory"] = true
	m.favorites["xander"] = true
	m.favorites["tommy"] = true

	visible := m.visiblePlayers()
	if len(visible) != 3 {
		t.Fatalf("visible players len = %d, want 3", len(visible))
	}
	if visible[0].Name != "Rory McIlroy" || visible[0].DisplayPosition != 5 || !visible[0].Tied {
		t.Fatalf("visible[0] = %+v, want Rory as T5", visible[0])
	}
	if visible[1].Name != "Xander Schauffele" || visible[1].DisplayPosition != 8 || !visible[1].Tied {
		t.Fatalf("visible[1] = %+v, want Xander as T8", visible[1])
	}
	if visible[2].Name != "Tommy Fleetwood" || visible[2].DisplayPosition != 12 {
		t.Fatalf("visible[2] = %+v, want Tommy as 12", visible[2])
	}
	if visible[0].CanonicalRank != 5 || visible[1].CanonicalRank != 8 || visible[2].CanonicalRank != 12 {
		t.Fatalf("canonical ranks = %d,%d,%d, want 5,8,12", visible[0].CanonicalRank, visible[1].CanonicalRank, visible[2].CanonicalRank)
	}
}

func TestHandleKeyTogglesFavoritesOnlyMode(t *testing.T) {
	m := New()
	m.height = 20
	m.tournament = testTournament("Scottie Scheffler", "Rory McIlroy", "Xander Schauffele")
	m.favorites["Rory McIlroy"] = true
	m.syncVisibleState()

	m = pressKey(t, m, keyWithText("F"))
	if !m.favoritesOnly {
		t.Fatal("favoritesOnly = false, want true")
	}
	visible := m.visiblePlayers()
	if len(visible) != 1 || visible[0].Name != "Rory McIlroy" {
		t.Fatalf("visible players = %+v, want only Rory", visible)
	}

	m = pressKey(t, m, keyWithText("F"))
	if m.favoritesOnly {
		t.Fatal("favoritesOnly = true, want false")
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
