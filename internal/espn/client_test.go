package espn

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// --- Helpers ---

func loadFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("failed to load fixture %s: %v", name, err)
	}
	return data
}

// --- Tier 1: Unit Tests ---

func TestParseESPNDate(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string // formatted as "2006-01-02" or empty for zero
	}{
		{name: "shortened ISO (no seconds)", input: "2026-03-12T04:00Z", want: "2026-03-12"},
		{name: "full RFC3339", input: "2026-03-12T04:00:00Z", want: "2026-03-12"},
		{name: "with timezone offset", input: "2026-03-12T04:00:00+00:00", want: "2026-03-12"},
		{name: "empty string", input: "", want: ""},
		{name: "garbage", input: "not-a-date", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseESPNDate(tt.input)
			if tt.want == "" {
				if !got.IsZero() {
					t.Errorf("expected zero time, got %v", got)
				}
				return
			}
			gotStr := got.Format("2006-01-02")
			if gotStr != tt.want {
				t.Errorf("got %s, want %s", gotStr, tt.want)
			}
		})
	}
}

func TestExtractCountryCode(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "normal URL", input: "https://a.espncdn.com/i/teamlogos/countries/500/swe.png", want: "swe"},
		{name: "usa", input: "https://a.espncdn.com/i/teamlogos/countries/500/usa.png", want: "usa"},
		{name: "empty", input: "", want: ""},
		{name: "no .png suffix", input: "https://example.com/flag/swe", want: "swe"},
		{name: "just filename", input: "kor.png", want: "kor"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractCountryCode(tt.input)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCalculateTiedPositions(t *testing.T) {
	tests := []struct {
		name    string
		players []Player
		// Check specific player indices for tie state, canonical rank, and display position.
		checks []struct {
			index             int
			wantTied          bool
			wantCanonicalRank int
			wantDisplayPos    int
		}
	}{
		{
			name:    "empty list",
			players: []Player{},
			checks:  nil,
		},
		{
			name: "solo leader, no ties",
			players: []Player{
				{CanonicalRank: 1, DisplayPosition: 1, TotalScore: "-12"},
				{CanonicalRank: 2, DisplayPosition: 2, TotalScore: "-10"},
				{CanonicalRank: 3, DisplayPosition: 3, TotalScore: "-9"},
			},
			checks: []struct {
				index             int
				wantTied          bool
				wantCanonicalRank int
				wantDisplayPos    int
			}{
				{0, false, 1, 1},
				{1, false, 2, 2},
				{2, false, 3, 3},
			},
		},
		{
			name: "two-way tie",
			players: []Player{
				{CanonicalRank: 1, DisplayPosition: 1, TotalScore: "-12"},
				{CanonicalRank: 2, DisplayPosition: 2, TotalScore: "-10"},
				{CanonicalRank: 3, DisplayPosition: 3, TotalScore: "-10"},
				{CanonicalRank: 4, DisplayPosition: 4, TotalScore: "-8"},
			},
			checks: []struct {
				index             int
				wantTied          bool
				wantCanonicalRank int
				wantDisplayPos    int
			}{
				{0, false, 1, 1},
				{1, true, 2, 2},
				{2, true, 3, 2},
				{3, false, 4, 4},
			},
		},
		{
			name: "three-way tie",
			players: []Player{
				{CanonicalRank: 1, DisplayPosition: 1, TotalScore: "-6"},
				{CanonicalRank: 2, DisplayPosition: 2, TotalScore: "-6"},
				{CanonicalRank: 3, DisplayPosition: 3, TotalScore: "-6"},
				{CanonicalRank: 4, DisplayPosition: 4, TotalScore: "-5"},
			},
			checks: []struct {
				index             int
				wantTied          bool
				wantCanonicalRank int
				wantDisplayPos    int
			}{
				{0, true, 1, 1},
				{1, true, 2, 1},
				{2, true, 3, 1},
				{3, false, 4, 4},
			},
		},
		{
			name: "all players tied",
			players: []Player{
				{CanonicalRank: 1, DisplayPosition: 1, TotalScore: "E"},
				{CanonicalRank: 2, DisplayPosition: 2, TotalScore: "E"},
				{CanonicalRank: 3, DisplayPosition: 3, TotalScore: "E"},
			},
			checks: []struct {
				index             int
				wantTied          bool
				wantCanonicalRank int
				wantDisplayPos    int
			}{
				{0, true, 1, 1},
				{1, true, 2, 1},
				{2, true, 3, 1},
			},
		},
		{
			name: "CUT/WD players excluded from tie groups",
			players: []Player{
				{CanonicalRank: 1, DisplayPosition: 1, TotalScore: "-5"},
				{CanonicalRank: 2, DisplayPosition: 2, TotalScore: "-5"},
				{CanonicalRank: 3, DisplayPosition: 3, TotalScore: "+3", Status: "CUT"},
				{CanonicalRank: 4, DisplayPosition: 4, TotalScore: "+3", Status: "CUT"},
			},
			checks: []struct {
				index             int
				wantTied          bool
				wantCanonicalRank int
				wantDisplayPos    int
			}{
				{0, true, 1, 1},
				{1, true, 2, 1},
				{2, false, 3, 3}, // CUT players not marked tied
				{3, false, 4, 4},
			},
		},
		{
			name: "multiple tie groups",
			players: []Player{
				{CanonicalRank: 1, DisplayPosition: 1, TotalScore: "-12"},
				{CanonicalRank: 2, DisplayPosition: 2, TotalScore: "-10"},
				{CanonicalRank: 3, DisplayPosition: 3, TotalScore: "-10"},
				{CanonicalRank: 4, DisplayPosition: 4, TotalScore: "-8"},
				{CanonicalRank: 5, DisplayPosition: 5, TotalScore: "-8"},
				{CanonicalRank: 6, DisplayPosition: 6, TotalScore: "-8"},
				{CanonicalRank: 7, DisplayPosition: 7, TotalScore: "-7"},
			},
			checks: []struct {
				index             int
				wantTied          bool
				wantCanonicalRank int
				wantDisplayPos    int
			}{
				{0, false, 1, 1},
				{1, true, 2, 2},
				{2, true, 3, 2},
				{3, true, 4, 4},
				{4, true, 5, 4},
				{5, true, 6, 4},
				{6, false, 7, 7},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calculateTiedPositions(tt.players)
			for _, c := range tt.checks {
				p := tt.players[c.index]
				if p.Tied != c.wantTied {
					t.Errorf("player[%d] Tied = %v, want %v", c.index, p.Tied, c.wantTied)
				}
				if p.CanonicalRank != c.wantCanonicalRank {
					t.Errorf("player[%d] CanonicalRank = %d, want %d", c.index, p.CanonicalRank, c.wantCanonicalRank)
				}
				if p.DisplayPosition != c.wantDisplayPos {
					t.Errorf("player[%d] DisplayPosition = %d, want %d", c.index, p.DisplayPosition, c.wantDisplayPos)
				}
			}
		})
	}
}

func TestDetermineCutStatus(t *testing.T) {
	tests := []struct {
		name string
		comp Competitor
		want string
	}{
		{
			name: "made cut - has R3 placeholder",
			comp: Competitor{
				LineScores: []RoundData{
					{Period: 1, Value: 72, DisplayValue: "E"},
					{Period: 2, Value: 74, DisplayValue: "+2"},
					{Period: 3}, // empty placeholder = made cut
				},
			},
			want: "",
		},
		{
			name: "missed cut - no R3 entry",
			comp: Competitor{
				LineScores: []RoundData{
					{Period: 1, Value: 77, DisplayValue: "+5"},
					{Period: 2, Value: 75, DisplayValue: "+3"},
				},
			},
			want: "CUT",
		},
		{
			name: "made cut - has R3 with score data",
			comp: Competitor{
				LineScores: []RoundData{
					{Period: 1, Value: 68, DisplayValue: "-4"},
					{Period: 2, Value: 70, DisplayValue: "-2"},
					{Period: 3, Value: 71, DisplayValue: "-1"},
				},
			},
			want: "",
		},
		{
			name: "only R1 played",
			comp: Competitor{
				LineScores: []RoundData{
					{Period: 1, Value: 72, DisplayValue: "E"},
				},
			},
			want: "CUT",
		},
		{
			name: "no linescores at all",
			comp: Competitor{
				LineScores: []RoundData{},
			},
			want: "CUT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := determineCutStatus(tt.comp)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetermineThru(t *testing.T) {
	tests := []struct {
		name string
		comp Competitor
		want string
	}{
		{
			name: "no linescores",
			comp: Competitor{LineScores: []RoundData{}},
			want: "-",
		},
		{
			name: "completed round - 18 holes",
			comp: Competitor{
				LineScores: []RoundData{
					{
						Period: 1, Value: 72, DisplayValue: "E",
						LineScores: make([]HoleData, 18),
					},
				},
			},
			want: "F",
		},
		{
			name: "mid-round - 12 holes",
			comp: Competitor{
				LineScores: []RoundData{
					{
						Period: 1, Value: 0,
						LineScores: make([]HoleData, 12),
					},
				},
			},
			want: "12",
		},
		{
			name: "completed round - strokes but no hole data",
			comp: Competitor{
				LineScores: []RoundData{
					{Period: 1, Value: 69, DisplayValue: "-3"},
				},
			},
			want: "F",
		},
		{
			name: "R1 complete, R2 in progress at hole 7",
			comp: Competitor{
				LineScores: []RoundData{
					{Period: 1, Value: 69, DisplayValue: "-3", LineScores: make([]HoleData, 18)},
					{Period: 2, LineScores: make([]HoleData, 7)},
				},
			},
			want: "7",
		},
		{
			name: "R1 and R2 complete, R3 placeholder (no data)",
			comp: Competitor{
				LineScores: []RoundData{
					{Period: 1, Value: 69, DisplayValue: "-3"},
					{Period: 2, Value: 71, DisplayValue: "-1"},
					{Period: 3}, // empty placeholder
				},
			},
			want: "F", // falls back to R2 which is complete
		},
		{
			name: "WD player - value is holes played (< 30), not strokes",
			comp: Competitor{
				LineScores: []RoundData{
					{Period: 1, Value: 4, DisplayValue: "E"},
				},
			},
			want: "-", // value < 50, no hole data = not treated as finished
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := determineThru(tt.comp)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParsePlayer(t *testing.T) {
	tests := []struct {
		name       string
		comp       Competitor
		round      int
		cutApplied bool
		wantID     string
		wantName   string
		wantScore  string
		wantStatus string
		wantThru   string
		wantR1     bool // whether R1 is marked as played
		wantRank   int
	}{
		{
			name: "normal player",
			comp: Competitor{
				ID:    "4375972",
				Order: 1,
				Score: "-12",
				Athlete: Athlete{
					DisplayName: "Ludvig Åberg",
					ShortName:   "L. Åberg",
					Flag:        &Flag{Href: "https://a.espncdn.com/i/teamlogos/countries/500/swe.png", Alt: "Sweden"},
				},
				LineScores: []RoundData{
					{Period: 1, Value: 69, DisplayValue: "-3", LineScores: make([]HoleData, 18)},
					{Period: 2, Value: 63, DisplayValue: "-9", LineScores: make([]HoleData, 18)},
					{Period: 3},
				},
			},
			round:      2,
			cutApplied: true,
			wantID:     "4375972",
			wantName:   "Ludvig Åberg",
			wantScore:  "-12",
			wantStatus: "",
			wantThru:   "F",
			wantR1:     true,
			wantRank:   1,
		},
		{
			name: "WD player - value is holes played",
			comp: Competitor{
				ID:    "10140",
				Order: 123,
				Score: "E",
				Athlete: Athlete{
					DisplayName: "Collin Morikawa",
					ShortName:   "C. Morikawa",
				},
				LineScores: []RoundData{
					{Period: 1, Value: 4, DisplayValue: "E"}, // 4 holes played
					{Period: 2, Value: 0, DisplayValue: "-"}, // didn't play R2
				},
			},
			round:      2,
			cutApplied: true,
			wantID:     "10140",
			wantName:   "Collin Morikawa",
			wantScore:  "E",
			wantStatus: "WD",
			wantThru:   "-",
			wantR1:     false, // WD player with value < 30, should not be marked played
			wantRank:   123,
		},
		{
			name: "CUT player - no R3 entry",
			comp: Competitor{
				ID:    "5409",
				Order: 74,
				Score: "+3",
				Athlete: Athlete{
					DisplayName: "Adam Schenk",
					ShortName:   "A. Schenk",
					Flag:        &Flag{Href: "https://a.espncdn.com/i/teamlogos/countries/500/usa.png", Alt: "United States"},
				},
				LineScores: []RoundData{
					{Period: 1, Value: 77, DisplayValue: "+5", LineScores: make([]HoleData, 18)},
					{Period: 2, Value: 70, DisplayValue: "-2", LineScores: make([]HoleData, 18)},
				},
			},
			round:      2,
			cutApplied: true,
			wantID:     "5409",
			wantName:   "Adam Schenk",
			wantScore:  "+3",
			wantStatus: "CUT",
			wantThru:   "F",
			wantR1:     true,
			wantRank:   74,
		},
		{
			name: "player with no flag",
			comp: Competitor{
				ID:    "12345",
				Order: 5,
				Score: "-5",
				Athlete: Athlete{
					DisplayName: "John Doe",
					ShortName:   "J. Doe",
				},
				LineScores: []RoundData{
					{Period: 1, Value: 67, DisplayValue: "-5"},
				},
			},
			round:      1,
			cutApplied: false,
			wantID:     "12345",
			wantName:   "John Doe",
			wantScore:  "-5",
			wantStatus: "",
			wantThru:   "F",
			wantR1:     true,
			wantRank:   5,
		},
		{
			name: "cut not yet applied - player without R3 is not marked CUT",
			comp: Competitor{
				ID:    "67890",
				Order: 74,
				Score: "+3",
				Athlete: Athlete{
					DisplayName: "Someone",
					ShortName:   "S. One",
				},
				LineScores: []RoundData{
					{Period: 1, Value: 77, DisplayValue: "+5"},
					{Period: 2, Value: 70, DisplayValue: "-2"},
				},
			},
			round:      2,
			cutApplied: false, // cut not yet applied
			wantID:     "67890",
			wantName:   "Someone",
			wantScore:  "+3",
			wantStatus: "", // not CUT because cut hasn't been applied
			wantThru:   "F",
			wantR1:     true,
			wantRank:   74,
		},
		{
			name: "prefers nested athlete ID when present",
			comp: Competitor{
				ID:    "competitor-id",
				Order: 9,
				Score: "-1",
				Athlete: Athlete{
					ID:          "athlete-id",
					DisplayName: "Nested ID",
					ShortName:   "N. ID",
				},
				LineScores: []RoundData{{Period: 1, Value: 71, DisplayValue: "-1"}},
			},
			round:      1,
			cutApplied: false,
			wantID:     "athlete-id",
			wantName:   "Nested ID",
			wantScore:  "-1",
			wantStatus: "",
			wantThru:   "F",
			wantR1:     true,
			wantRank:   9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parsePlayer(tt.comp, tt.round, tt.cutApplied)
			if got.ID != tt.wantID {
				t.Errorf("ID = %q, want %q", got.ID, tt.wantID)
			}
			if got.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", got.Name, tt.wantName)
			}
			if got.CanonicalRank != tt.wantRank {
				t.Errorf("CanonicalRank = %d, want %d", got.CanonicalRank, tt.wantRank)
			}
			if got.DisplayPosition != tt.wantRank {
				t.Errorf("DisplayPosition = %d, want %d", got.DisplayPosition, tt.wantRank)
			}
			if got.TotalScore != tt.wantScore {
				t.Errorf("TotalScore = %q, want %q", got.TotalScore, tt.wantScore)
			}
			if got.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", got.Status, tt.wantStatus)
			}
			if got.Thru != tt.wantThru {
				t.Errorf("Thru = %q, want %q", got.Thru, tt.wantThru)
			}
			if len(got.Rounds) < 1 {
				t.Fatal("expected at least 1 round")
			}
			if got.Rounds[0].Played != tt.wantR1 {
				t.Errorf("Rounds[0].Played = %v, want %v", got.Rounds[0].Played, tt.wantR1)
			}
		})
	}
}

// --- Tier 1: Integration Test Against Fixture ---

func TestParseTournamentFixture(t *testing.T) {
	data := loadFixture(t, "scoreboard.json")

	var sb ScoreboardResponse
	if err := json.Unmarshal(data, &sb); err != nil {
		t.Fatalf("failed to unmarshal fixture: %v", err)
	}

	tournament, err := parseTournament(sb)
	if err != nil {
		t.Fatalf("parseTournament failed: %v", err)
	}

	// Tournament metadata
	if tournament.Name != "THE PLAYERS Championship" {
		t.Errorf("Name = %q, want %q", tournament.Name, "THE PLAYERS Championship")
	}
	if tournament.Round != 2 {
		t.Errorf("Round = %d, want 2", tournament.Round)
	}
	if tournament.Status != "in" {
		t.Errorf("Status = %q, want %q", tournament.Status, "in")
	}
	if tournament.StartDate.IsZero() {
		t.Error("StartDate should not be zero")
	}
	if tournament.EndDate.IsZero() {
		t.Error("EndDate should not be zero")
	}
	if tournament.StartDate.Format("2006-01-02") != "2026-03-12" {
		t.Errorf("StartDate = %s, want 2026-03-12", tournament.StartDate.Format("2006-01-02"))
	}

	// Player count
	if len(tournament.Players) != 123 {
		t.Errorf("Players count = %d, want 123", len(tournament.Players))
	}

	// Leader should be Ludvig Åberg at -12
	leader := tournament.Players[0]
	if leader.Name != "Ludvig Åberg" {
		t.Errorf("Leader name = %q, want %q", leader.Name, "Ludvig Åberg")
	}
	if leader.TotalScore != "-12" {
		t.Errorf("Leader score = %q, want %q", leader.TotalScore, "-12")
	}
	if leader.ID != "4375972" {
		t.Errorf("Leader ID = %q, want %q", leader.ID, "4375972")
	}
	if leader.CanonicalRank != 1 {
		t.Errorf("Leader canonical rank = %d, want 1", leader.CanonicalRank)
	}
	if leader.DisplayPosition != 1 {
		t.Errorf("Leader display position = %d, want 1", leader.DisplayPosition)
	}
	if leader.Tied {
		t.Error("Leader should not be tied (solo first)")
	}
	if leader.CountryCode != "swe" {
		t.Errorf("Leader country = %q, want %q", leader.CountryCode, "swe")
	}
	if leader.Thru != "F" {
		t.Errorf("Leader thru = %q, want %q", leader.Thru, "F")
	}
	if !leader.Rounds[0].Played || leader.Rounds[0].Strokes != 69 {
		t.Errorf("Leader R1 = played:%v strokes:%d, want played:true strokes:69",
			leader.Rounds[0].Played, leader.Rounds[0].Strokes)
	}
	if !leader.Rounds[1].Played || leader.Rounds[1].Strokes != 63 {
		t.Errorf("Leader R2 = played:%v strokes:%d, want played:true strokes:63",
			leader.Rounds[1].Played, leader.Rounds[1].Strokes)
	}
	if leader.Rounds[2].Played {
		t.Error("Leader R3 should not be played yet")
	}

	// Tied positions - find players at -8 (should be T4 or T5 depending on fixture)
	var tiedAtMinus8 []Player
	for _, p := range tournament.Players {
		if p.TotalScore == "-8" && p.Status == "" {
			tiedAtMinus8 = append(tiedAtMinus8, p)
		}
	}
	if len(tiedAtMinus8) < 2 {
		t.Fatalf("expected at least 2 players at -8, got %d", len(tiedAtMinus8))
	}
	for _, p := range tiedAtMinus8 {
		if !p.Tied {
			t.Errorf("player %q at -8 should be marked as tied", p.Name)
		}
	}
	// All tied players should share the same display position while keeping canonical rank intact.
	firstPos := tiedAtMinus8[0].DisplayPosition
	firstRank := tiedAtMinus8[0].CanonicalRank
	for _, p := range tiedAtMinus8[1:] {
		if p.DisplayPosition != firstPos {
			t.Errorf("tied player %q has display position %d, want %d (same as first)", p.Name, p.DisplayPosition, firstPos)
		}
		if p.CanonicalRank <= firstRank {
			t.Errorf("tied player %q canonical rank = %d, want greater than %d", p.Name, p.CanonicalRank, firstRank)
		}
	}

	// Cut detection: 50 without R3, minus 1 WD (Morikawa) = 49 CUT
	var cutPlayers, activePlayers, wdPlayers int
	for _, p := range tournament.Players {
		switch p.Status {
		case "CUT":
			cutPlayers++
		case "WD":
			wdPlayers++
		case "":
			activePlayers++
		}
	}
	if cutPlayers != 49 {
		t.Errorf("CUT players = %d, want 49", cutPlayers)
	}
	if wdPlayers != 1 {
		t.Errorf("WD players = %d, want 1", wdPlayers)
	}
	// 73 made cut + active status
	if activePlayers != 73 {
		t.Errorf("Active players = %d, want 73", activePlayers)
	}

	// WD detection - Collin Morikawa
	lastPlayer := tournament.Players[len(tournament.Players)-1]
	if lastPlayer.Name != "Collin Morikawa" {
		t.Errorf("Last player = %q, want Collin Morikawa", lastPlayer.Name)
	}
	if lastPlayer.Status != "WD" {
		t.Errorf("Morikawa status = %q, want WD", lastPlayer.Status)
	}

	// Cut boundary: player 73 made cut, player 74 did not
	player73 := tournament.Players[72]
	player74 := tournament.Players[73]
	if player73.Status == "CUT" {
		t.Errorf("Player 73 (%s) should NOT be CUT (made cut)", player73.Name)
	}
	if player74.Status != "CUT" {
		t.Errorf("Player 74 (%s) should be CUT, got %q", player74.Name, player74.Status)
	}
}

func TestParseTournamentNoEvents(t *testing.T) {
	sb := ScoreboardResponse{
		Events: []Event{},
	}
	_, err := parseTournament(sb)
	if err == nil {
		t.Error("expected error for no events, got nil")
	}
}

func TestParseTournamentNoCompetitions(t *testing.T) {
	sb := ScoreboardResponse{
		Events: []Event{
			{
				ID:   "123",
				Name: "Test Tournament",
				Date: "2026-03-12T04:00Z",
			},
		},
	}
	tournament, err := parseTournament(sb)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tournament.Name != "Test Tournament" {
		t.Errorf("Name = %q, want %q", tournament.Name, "Test Tournament")
	}
	if len(tournament.Players) != 0 {
		t.Errorf("Players = %d, want 0", len(tournament.Players))
	}
}

func TestParseESPNDateIsUTC(t *testing.T) {
	got := parseESPNDate("2026-03-12T04:00Z")
	if got.Location() != time.UTC {
		t.Errorf("expected UTC, got %v", got.Location())
	}
	if got.Hour() != 4 {
		t.Errorf("expected hour 4, got %d", got.Hour())
	}
}
