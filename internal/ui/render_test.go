package ui_test

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/nickkoul/gstat/internal/espn"
	"github.com/nickkoul/gstat/internal/ui"
)

// stripANSI removes ANSI escape sequences so we can assert on visible content.
func stripANSI(s string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	return re.ReplaceAllString(s, "")
}

func mustContain(t *testing.T, output, label, substr string) {
	t.Helper()
	plain := stripANSI(output)
	if !strings.Contains(plain, substr) {
		t.Errorf("%s: expected %q in output, got:\n%s", label, substr, plain)
	}
}

func mustNotContain(t *testing.T, output, label, substr string) {
	t.Helper()
	plain := stripANSI(output)
	if strings.Contains(plain, substr) {
		t.Errorf("%s: did NOT expect %q in output, got:\n%s", label, substr, plain)
	}
}

// --- Helpers ---

func makeRounds(r1, r2, r3, r4 int) []espn.RoundScore {
	rounds := make([]espn.RoundScore, 4)
	strokes := []int{r1, r2, r3, r4}
	for i, s := range strokes {
		rounds[i] = espn.RoundScore{Round: i + 1}
		if s > 0 {
			rounds[i].Played = true
			rounds[i].Strokes = s
		}
	}
	return rounds
}

// --- Header Tests ---

func TestRenderHeader(t *testing.T) {
	tournament := &espn.Tournament{
		Name:      "THE PLAYERS Championship",
		Detail:    "Round 2 - Play Complete",
		StartDate: espn.TestParseDate("2026-03-12T04:00Z"),
		EndDate:   espn.TestParseDate("2026-03-15T04:00Z"),
	}

	out := ui.RenderHeader(tournament, 80)

	mustContain(t, out, "tournament name", "THE PLAYERS Championship")
	mustContain(t, out, "round status", "Round 2 - Play Complete")
	mustContain(t, out, "start date", "Mar 12, 2026")
	mustContain(t, out, "end date", "Mar 15, 2026")
}

func TestRenderHeaderNilTournament(t *testing.T) {
	out := ui.RenderHeader(nil, 80)

	if out == "" {
		t.Error("header should not be empty for nil tournament")
	}
	mustContain(t, out, "loading message", "Loading")
}

func TestRenderHeaderNarrowWidth(t *testing.T) {
	tournament := &espn.Tournament{
		Name:   "THE PLAYERS Championship",
		Detail: "Round 2 - Play Complete",
	}

	// Should not panic at narrow widths
	out := ui.RenderHeader(tournament, 40)
	if out == "" {
		t.Error("header should not be empty at narrow width")
	}
}

// --- Table Header Tests ---

func TestRenderTableHeader(t *testing.T) {
	out := ui.RenderTableHeader(80, 4)

	mustContain(t, out, "POS column", "POS")
	mustContain(t, out, "PLAYER column", "PLAYER")
	mustContain(t, out, "CTRY column", "CTRY")
	mustContain(t, out, "TOT column", "TOT")
	mustContain(t, out, "R1 column", "R1")
	mustContain(t, out, "R2 column", "R2")
	mustContain(t, out, "R3 column", "R3")
	mustContain(t, out, "R4 column", "R4")
	mustContain(t, out, "THRU column", "THRU")
}

func TestRenderTableHeaderTwoRounds(t *testing.T) {
	out := ui.RenderTableHeader(80, 2)

	mustContain(t, out, "R1 column", "R1")
	mustContain(t, out, "R2 column", "R2")
	mustNotContain(t, out, "no R3", "R3")
	mustNotContain(t, out, "no R4", "R4")
}

// --- Player Row Tests ---

func TestRenderPlayerRowLeader(t *testing.T) {
	player := espn.Player{
		Position: 1, Name: "Ludvig Åberg", CountryCode: "swe",
		TotalScore: "-12", Thru: "F",
		Rounds: makeRounds(69, 63, 0, 0),
	}

	out := ui.RenderPlayerRow(player, 0, 80, 4, -1)

	mustContain(t, out, "position", "1")
	mustContain(t, out, "name", "Ludvig")
	mustContain(t, out, "country", "SWE")
	mustContain(t, out, "score", "-12")
	mustContain(t, out, "R1 strokes", "69")
	mustContain(t, out, "R2 strokes", "63")
	mustContain(t, out, "thru", "F")
}

func TestRenderPlayerRowTied(t *testing.T) {
	player := espn.Player{
		Position: 4, Tied: true, Name: "Corey Conners", CountryCode: "can",
		TotalScore: "-8", Thru: "F",
		Rounds: makeRounds(69, 67, 0, 0),
	}

	out := ui.RenderPlayerRow(player, 0, 80, 4, -1)

	mustContain(t, out, "tied position", "T4")
}

func TestRenderPlayerRowNotTied(t *testing.T) {
	player := espn.Player{
		Position: 3, Tied: false, Name: "Someone", CountryCode: "usa",
		TotalScore: "-9", Thru: "F",
		Rounds: makeRounds(68, 67, 0, 0),
	}

	out := ui.RenderPlayerRow(player, 0, 80, 4, -1)

	// Should show "3" not "T3"
	mustNotContain(t, out, "no T prefix", "T3")
	mustContain(t, out, "position", "3")
}

func TestRenderPlayerRowCUT(t *testing.T) {
	player := espn.Player{
		Position: 74, Name: "Adam Schenk", CountryCode: "usa",
		TotalScore: "+3", Thru: "F", Status: "CUT",
		Rounds: makeRounds(77, 70, 0, 0),
	}

	out := ui.RenderPlayerRow(player, 0, 80, 4, 0)

	mustContain(t, out, "CUT status", "CUT")
	mustContain(t, out, "over par score", "+3")
	mustContain(t, out, "name", "Adam Schenk")
}

func TestRenderPlayerRowWD(t *testing.T) {
	player := espn.Player{
		Position: 123, Name: "Collin Morikawa", CountryCode: "usa",
		TotalScore: "E", Thru: "-", Status: "WD",
		Rounds: makeRounds(0, 0, 0, 0),
	}

	out := ui.RenderPlayerRow(player, 0, 80, 4, -1)

	mustContain(t, out, "WD status", "WD")
	mustContain(t, out, "name", "Collin Morikawa")
}

func TestRenderPlayerRowEvenPar(t *testing.T) {
	player := espn.Player{
		Position: 10, Name: "Even Player", CountryCode: "eng",
		TotalScore: "E", Thru: "F",
		Rounds: makeRounds(72, 72, 0, 0),
	}

	out := ui.RenderPlayerRow(player, 0, 80, 4, -1)

	mustContain(t, out, "even par", "E")
}

func TestRenderPlayerRowOverPar(t *testing.T) {
	player := espn.Player{
		Position: 50, Name: "Over Player", CountryCode: "usa",
		TotalScore: "+5", Thru: "F",
		Rounds: makeRounds(77, 72, 0, 0),
	}

	out := ui.RenderPlayerRow(player, 0, 80, 4, -1)

	mustContain(t, out, "over par score", "+5")
}

func TestRenderPlayerRowLongNameTruncation(t *testing.T) {
	player := espn.Player{
		Position: 1, Name: "Superlongfirstname Superlonglastname", CountryCode: "usa",
		TotalScore: "-5", Thru: "F",
		Rounds: makeRounds(67, 0, 0, 0),
	}

	out := ui.RenderPlayerRow(player, 0, 80, 4, -1)

	// Should not panic, and should contain some truncated version
	if out == "" {
		t.Error("should not be empty for long name")
	}
	// Full name shouldn't appear since it's too long for the column
	plain := stripANSI(out)
	if strings.Contains(plain, "Superlonglastname") {
		t.Error("expected name to be truncated")
	}
}

func TestRenderPlayerRowUnplayedRounds(t *testing.T) {
	player := espn.Player{
		Position: 1, Name: "Test Player", CountryCode: "usa",
		TotalScore: "-3", Thru: "F",
		Rounds: makeRounds(69, 0, 0, 0), // only R1 played
	}

	out := ui.RenderPlayerRow(player, 0, 80, 4, -1)

	mustContain(t, out, "R1 strokes", "69")
	// Unplayed rounds should show "-"
	// Count dashes - there should be at least 3 (for R2, R3, R4)
	plain := stripANSI(out)
	dashCount := strings.Count(plain, "-")
	if dashCount < 3 {
		t.Errorf("expected at least 3 dashes for unplayed rounds, got %d in: %s", dashCount, plain)
	}
}

func TestRenderPlayerRowEmptyCountry(t *testing.T) {
	player := espn.Player{
		Position: 1, Name: "No Country", CountryCode: "",
		TotalScore: "-5", Thru: "F",
		Rounds: makeRounds(67, 0, 0, 0),
	}

	out := ui.RenderPlayerRow(player, 0, 80, 4, -1)

	// Should show "---" for empty country
	mustContain(t, out, "empty country", "---")
}

// --- Unicode Column Alignment Tests ---

func TestRenderPlayerRowUnicodeAlignment(t *testing.T) {
	// Players with multi-byte unicode characters (ø, ä, å, etc.) should
	// have the same column alignment as players with ASCII-only names.
	ascii := espn.Player{
		Position: 1, Name: "Adam Scott", CountryCode: "aus",
		TotalScore: "-5", Thru: "F",
		Rounds: makeRounds(67, 0, 0, 0),
	}
	unicode1 := espn.Player{
		Position: 2, Name: "Rasmus Højgaard", CountryCode: "den",
		TotalScore: "-4", Thru: "F",
		Rounds: makeRounds(68, 0, 0, 0),
	}
	unicode2 := espn.Player{
		Position: 3, Name: "Sami Välimäki", CountryCode: "fin",
		TotalScore: "-3", Thru: "F",
		Rounds: makeRounds(69, 0, 0, 0),
	}
	unicode3 := espn.Player{
		Position: 4, Name: "Ludvig Åberg", CountryCode: "swe",
		TotalScore: "-2", Thru: "F",
		Rounds: makeRounds(70, 0, 0, 0),
	}

	asciiRow := stripANSI(ui.RenderPlayerRow(ascii, 0, 80, 4, -1))
	row1 := stripANSI(ui.RenderPlayerRow(unicode1, 1, 80, 4, -1))
	row2 := stripANSI(ui.RenderPlayerRow(unicode2, 2, 80, 4, -1))
	row3 := stripANSI(ui.RenderPlayerRow(unicode3, 3, 80, 4, -1))

	// Find the display column (not byte offset) of the country code.
	// We need to measure display width up to the match, not byte offset,
	// because multi-byte unicode chars take 1 display column but 2+ bytes.
	asciiCtryCol := displayColumn(asciiRow, "AUS")
	col1 := displayColumn(row1, "DEN")
	col2 := displayColumn(row2, "FIN")
	col3 := displayColumn(row3, "SWE")

	if asciiCtryCol < 0 || col1 < 0 || col2 < 0 || col3 < 0 {
		t.Fatalf("could not find country codes in output:\n  ascii: %q\n  row1:  %q\n  row2:  %q\n  row3:  %q",
			asciiRow, row1, row2, row3)
	}

	if col1 != asciiCtryCol {
		t.Errorf("Højgaard country display column at %d, want %d (same as ASCII row)\n  ascii: %q\n  row1:  %q",
			col1, asciiCtryCol, asciiRow, row1)
	}
	if col2 != asciiCtryCol {
		t.Errorf("Välimäki country display column at %d, want %d (same as ASCII row)\n  ascii: %q\n  row2:  %q",
			col2, asciiCtryCol, asciiRow, row2)
	}
	if col3 != asciiCtryCol {
		t.Errorf("Åberg country display column at %d, want %d (same as ASCII row)\n  ascii: %q\n  row3:  %q",
			col3, asciiCtryCol, asciiRow, row3)
	}
}

// displayColumn returns the display column position (not byte offset) of
// substr within s. Returns -1 if not found. This correctly handles
// multi-byte unicode characters that take 1 display column but 2+ bytes.
func displayColumn(s, substr string) int {
	idx := strings.Index(s, substr)
	if idx < 0 {
		return -1
	}
	// Count runes (display columns) before the match.
	// All characters in golf player names are single-width.
	prefix := s[:idx]
	return len([]rune(prefix))
}

// --- Cut Line Tests ---

func TestRenderCutLine(t *testing.T) {
	out := ui.RenderCutLine(80)

	mustContain(t, out, "CUT label", "CUT")

	// Should contain line characters
	plain := stripANSI(out)
	if !strings.Contains(plain, "─") {
		t.Error("cut line should contain line drawing characters")
	}
}

// --- Status Bar Tests ---

func TestRenderStatusBar(t *testing.T) {
	now := time.Now()
	out := ui.RenderStatusBar(now, 25*time.Second, 80, "", "", false)

	mustContain(t, out, "refresh countdown", "25s")
	mustContain(t, out, "quit hint", "quit")
	mustContain(t, out, "refresh hint", "refresh")
	mustContain(t, out, "search hint", "search")
}

func TestRenderStatusBarWithError(t *testing.T) {
	now := time.Now()
	out := ui.RenderStatusBar(now, 10*time.Second, 80, "connection refused", "", false)

	mustContain(t, out, "error message", "connection refused")
}

func TestRenderStatusBarZeroTime(t *testing.T) {
	out := ui.RenderStatusBar(time.Time{}, 0, 80, "", "", false)

	mustContain(t, out, "fetching message", "Fetching")
}

func TestRenderStatusBarSearchMode(t *testing.T) {
	now := time.Now()
	out := ui.RenderStatusBar(now, 10*time.Second, 80, "", "schef", true)

	mustContain(t, out, "filter query", "/schef")
	mustContain(t, out, "search mode", "search")
	mustContain(t, out, "enter hint", "apply")
	mustContain(t, out, "escape hint", "clear")
}
