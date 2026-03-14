package ui

import (
	"fmt"
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/nickkoul/gstat/internal/espn"
)

// Column widths
const (
	colMarker = 2
	colPos    = 5
	colName   = 24
	colCtry   = 5
	colScore  = 6
	colRound  = 5
	colThru   = 5
)

// countryEmoji maps common country codes to flag emojis.
// Falls back to the uppercase code if not found.
var countryEmoji = map[string]string{
	"usa": "USA", "swe": "SWE", "nir": "NIR", "eng": "ENG",
	"kor": "KOR", "jpn": "JPN", "aus": "AUS", "can": "CAN",
	"rsa": "RSA", "esp": "ESP", "nor": "NOR", "irl": "IRL",
	"ger": "GER", "fra": "FRA", "sco": "SCO", "wal": "WAL",
	"col": "COL", "chi": "CHI", "arg": "ARG", "tha": "THA",
	"chn": "CHN", "twn": "TPE", "ind": "IND", "ita": "ITA",
	"aut": "AUT", "bel": "BEL", "den": "DEN", "fin": "FIN",
	"mex": "MEX", "ven": "VEN", "par": "PAR", "zim": "ZIM",
	"fij": "FIJ", "nzl": "NZL", "ber": "BER", "pur": "PUR",
}

// RenderTableHeader renders the column headers for the leaderboard.
func RenderTableHeader(width int, totalRounds int) string {
	s := DefaultStyles()

	marker := padRight("", colMarker)
	pos := padRight("POS", colPos)
	name := padRight("PLAYER", colName)
	ctry := padRight("CTRY", colCtry)
	score := padLeft("TOT", colScore)

	var rounds string
	for i := 1; i <= totalRounds; i++ {
		rounds += padLeft(fmt.Sprintf("R%d", i), colRound)
	}

	thru := padLeft("THRU", colThru)

	header := fmt.Sprintf("%s %s %s %s %s%s %s",
		marker, pos, name, ctry, score, rounds, thru)

	// Trim or pad to width
	headerStyled := s.TableHeader.Width(width).Render(header)
	return headerStyled
}

// RenderPlayerRow renders a single player row in the leaderboard.
func RenderPlayerRow(p espn.Player, index int, width int, totalRounds int, cutLine int, showRoundScoresToPar bool, selected bool, favorite bool) string {
	s := DefaultStyles()

	marker := s.Marker.Render(padRight(formatMarker(selected, favorite), colMarker))

	// Position display with tie indicator
	posStr := formatPosition(p)
	pos := padRight(posStr, colPos)
	posStyled := s.Position.Render(pos)

	// Player name - dim if cut/wd
	nameStr := truncate(p.Name, colName-1)
	nameStr = padRight(nameStr, colName)
	var nameStyled string
	switch p.Status {
	case "CUT":
		nameStyled = s.StatusCut.Render(nameStr)
	case "WD":
		nameStyled = s.StatusWD.Render(nameStr)
	default:
		if favorite {
			nameStyled = s.FavoritePlayer.Render(nameStr)
		} else {
			nameStyled = s.PlayerName.Render(nameStr)
		}
	}

	// Country
	ctryStr := formatCountry(p.CountryCode)
	ctryStr = padRight(ctryStr, colCtry)
	ctryStyled := s.Country.Render(ctryStr)

	// Total score (color-coded)
	scoreStyled := renderScore(p.TotalScore, s)

	// Round scores
	var roundsStyled string
	for i := 0; i < totalRounds; i++ {
		if i < len(p.Rounds) {
			roundsStyled += renderRoundScore(p.Rounds[i], showRoundScoresToPar, s)
		} else {
			roundsStyled += s.RoundScore.Render(padLeft("-", colRound))
		}
	}

	// Thru
	thruStr := padLeft(p.Thru, colThru)
	var thruStyled string
	if p.Status == "CUT" {
		thruStyled = s.StatusCut.Render(thruStr)
	} else if p.Status == "WD" {
		thruStyled = s.StatusWD.Render(thruStr)
	} else {
		thruStyled = s.Thru.Render(thruStr)
	}

	row := fmt.Sprintf("%s %s %s %s %s%s %s",
		marker, posStyled, nameStyled, ctryStyled, scoreStyled, roundsStyled, thruStyled)
	if selected {
		return s.SelectedRow.Width(width).Render(row)
	}

	return row
}

// RenderCutLine renders the cut line separator.
func RenderCutLine(width int) string {
	s := DefaultStyles()
	label := " CUT "
	lineLen := (width - len(label)) / 2
	if lineLen < 3 {
		lineLen = 3
	}
	line := strings.Repeat("─", lineLen) + label + strings.Repeat("─", lineLen)
	return s.CutLine.Render(line)
}

// renderScore renders the total score with appropriate color.
func renderRelativeScore(score string, width int, s Styles) string {
	scoreStr := padLeft(score, width)

	if score == "" || score == "-" {
		return s.RoundScore.Render(scoreStr)
	}

	if strings.HasPrefix(score, "-") {
		return s.ScoreUnder.Render(scoreStr)
	}
	if strings.HasPrefix(score, "+") {
		return s.ScoreOver.Render(scoreStr)
	}
	// "E" for even
	return s.ScoreEven.Render(scoreStr)
}

func renderScore(score string, s Styles) string {
	return renderRelativeScore(score, colScore, s)
}

func renderRoundScore(round espn.RoundScore, showToPar bool, s Styles) string {
	if !round.Played {
		return s.RoundScore.Render(padLeft("-", colRound))
	}

	if showToPar {
		if round.ToPar == "" {
			return s.RoundScore.Render(padLeft("-", colRound))
		}
		return renderRelativeScore(round.ToPar, colRound, s)
	}

	return s.RoundScore.Render(padLeft(fmt.Sprintf("%d", round.Strokes), colRound))
}

// formatPosition formats the position with tie indicator.
func formatPosition(p espn.Player) string {
	if p.Status == "CUT" || p.Status == "WD" || p.Status == "DQ" {
		return p.Status
	}
	if p.Tied {
		return fmt.Sprintf("T%d", p.DisplayPosition)
	}
	return fmt.Sprintf("%d", p.DisplayPosition)
}

func formatMarker(selected bool, favorite bool) string {
	switch {
	case selected && favorite:
		return ">*"
	case selected:
		return ">"
	case favorite:
		return "*"
	default:
		return ""
	}
}

// formatCountry converts country code to display string.
func formatCountry(code string) string {
	if code == "" {
		return "---"
	}
	code = strings.ToLower(code)
	if emoji, ok := countryEmoji[code]; ok {
		return emoji
	}
	return strings.ToUpper(code)
}

// Helper functions
// These use runewidth.StringWidth for display-width-aware padding,
// so that multi-byte unicode characters (ø, ä, å, etc.) don't break
// column alignment.

func padRight(s string, width int) string {
	sw := runewidth.StringWidth(s)
	if sw >= width {
		return runewidth.Truncate(s, width, "")
	}
	return s + strings.Repeat(" ", width-sw)
}

func padLeft(s string, width int) string {
	sw := runewidth.StringWidth(s)
	if sw >= width {
		return runewidth.Truncate(s, width, "")
	}
	return strings.Repeat(" ", width-sw) + s
}

func truncate(s string, maxLen int) string {
	sw := runewidth.StringWidth(s)
	if sw <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return runewidth.Truncate(s, maxLen, "")
	}
	return runewidth.Truncate(s, maxLen, "…")
}
