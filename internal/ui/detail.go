package ui

import (
	"fmt"
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/nickkoul/gstat/internal/espn"
)

const (
	detailLabelWidth   = 4
	detailCellWidth    = 5
	detailWideMinWidth = 138
)

type detailColumn struct {
	label   string
	holes   []int
	summary bool
}

// DetailPanelHeight returns the number of content lines used by the detail panel.
func DetailPanelHeight(player espn.Player, width int, round int) int {
	return len(detailLines(player, width, round, DefaultStyles()))
}

// RenderPlayerDetail renders the inline hole-by-hole scorecard for the selected player.
func RenderPlayerDetail(player espn.Player, width int, round int) string {
	s := DefaultStyles()
	return s.HelpPanel.Width(width).Render(joinLines(detailLines(player, width, round, s)))
}

func detailLines(player espn.Player, width int, round int, s Styles) []string {
	roundScore := detailRound(round, player)
	lines := []string{
		fmt.Sprintf("  %s  %s  %s %s  %s %s",
			s.DetailTitle.Render("Scorecard"),
			s.DetailValue.Render(player.Name),
			s.DetailLabel.Render("TOT"),
			s.DetailValue.Render(player.TotalScore),
			s.DetailLabel.Render("THRU"),
			s.DetailValue.Render(player.Thru),
		),
		renderDetailTabs(round, len(player.Rounds), s),
	}

	if len(roundScore.Holes) == 0 {
		lines = append(lines, s.DetailMuted.Render(fmt.Sprintf("  No hole-by-hole data for Round %d", round)))
		return lines
	}

	holes := detailHoleMap(roundScore)
	if width >= detailWideMinWidth {
		return append(lines, renderScorecardSection(wideScorecardColumns(), holes, s)...)
	}

	lines = append(lines,
		s.DetailMuted.Render("  Front nine"),
	)
	lines = append(lines, renderScorecardSection(frontScorecardColumns(), holes, s)...)
	lines = append(lines,
		s.DetailMuted.Render("  Back nine"),
	)
	lines = append(lines, renderScorecardSection(backScorecardColumns(), holes, s)...)
	return lines
}

func renderDetailTabs(activeRound int, totalRounds int, s Styles) string {
	if totalRounds < 4 {
		totalRounds = 4
	}

	parts := make([]string, 0, totalRounds)
	for i := 1; i <= totalRounds; i++ {
		label := fmt.Sprintf("R%d", i)
		if i == activeRound {
			parts = append(parts, s.DetailTabActive.Render("["+label+"]"))
			continue
		}
		parts = append(parts, s.DetailTab.Render(label))
	}

	return "  " + strings.Join(parts, "  ")
}

func detailRound(round int, player espn.Player) espn.RoundScore {
	if round < 1 || round > len(player.Rounds) {
		return espn.RoundScore{Round: round}
	}
	return player.Rounds[round-1]
}

func detailHoleMap(roundScore espn.RoundScore) map[int]espn.HoleScore {
	holes := make(map[int]espn.HoleScore, len(roundScore.Holes))
	for _, hole := range roundScore.Holes {
		holes[hole.Number] = hole
	}
	return holes
}

func wideScorecardColumns() []detailColumn {
	columns := make([]detailColumn, 0, 21)
	for hole := 1; hole <= 9; hole++ {
		columns = append(columns, detailColumn{label: fmt.Sprintf("%d", hole), holes: []int{hole}})
	}
	columns = append(columns, detailColumn{label: "OUT", holes: holeRange(1, 9), summary: true})
	for hole := 10; hole <= 18; hole++ {
		columns = append(columns, detailColumn{label: fmt.Sprintf("%d", hole), holes: []int{hole}})
	}
	columns = append(columns,
		detailColumn{label: "IN", holes: holeRange(10, 18), summary: true},
		detailColumn{label: "TOT", holes: holeRange(1, 18), summary: true},
	)
	return columns
}

func frontScorecardColumns() []detailColumn {
	columns := make([]detailColumn, 0, 10)
	for hole := 1; hole <= 9; hole++ {
		columns = append(columns, detailColumn{label: fmt.Sprintf("%d", hole), holes: []int{hole}})
	}
	columns = append(columns, detailColumn{label: "OUT", holes: holeRange(1, 9), summary: true})
	return columns
}

func backScorecardColumns() []detailColumn {
	columns := make([]detailColumn, 0, 11)
	for hole := 10; hole <= 18; hole++ {
		columns = append(columns, detailColumn{label: fmt.Sprintf("%d", hole), holes: []int{hole}})
	}
	columns = append(columns,
		detailColumn{label: "IN", holes: holeRange(10, 18), summary: true},
		detailColumn{label: "TOT", holes: holeRange(1, 18), summary: true},
	)
	return columns
}

func holeRange(start, end int) []int {
	holes := make([]int, 0, end-start+1)
	for hole := start; hole <= end; hole++ {
		holes = append(holes, hole)
	}
	return holes
}

func renderScorecardSection(columns []detailColumn, holes map[int]espn.HoleScore, s Styles) []string {
	return []string{
		renderScorecardBorder(columns, s),
		renderScorecardRow("HOLE", columns, s, func(col detailColumn) string {
			return renderDetailTextCell(col.label, col.summary, s)
		}),
		renderScorecardRow("PAR", columns, s, func(col detailColumn) string {
			return renderParCell(col, holes, s)
		}),
		renderScorecardRow("SCR", columns, s, func(col detailColumn) string {
			return renderScoreCell(col, holes, s)
		}),
		renderScorecardRow("RUN", columns, s, func(col detailColumn) string {
			return renderRunningToParCell(col, holes, s)
		}),
		renderScorecardBorder(columns, s),
	}
}

func renderScorecardBorder(columns []detailColumn, s Styles) string {
	parts := []string{"+" + strings.Repeat("-", detailLabelWidth) + "+"}
	for range columns {
		parts = append(parts, strings.Repeat("-", detailCellWidth)+"+")
	}
	return "  " + s.DetailMuted.Render(strings.Join(parts, ""))
}

func renderScorecardRow(label string, columns []detailColumn, s Styles, render func(detailColumn) string) string {
	parts := []string{"|" + s.DetailLabel.Render(padCenter(label, detailLabelWidth)) + "|"}
	for _, col := range columns {
		parts = append(parts, render(col)+"|")
	}
	return "  " + strings.Join(parts, "")
}

func renderDetailTextCell(text string, summary bool, s Styles) string {
	text = padCenter(text, detailCellWidth)
	if summary {
		return s.DetailTitle.Render(text)
	}
	return s.DetailValue.Render(text)
}

func renderParCell(col detailColumn, holes map[int]espn.HoleScore, s Styles) string {
	if value, ok := parValue(col, holes); ok {
		return renderDetailTextCell(fmt.Sprintf("%d", value), col.summary, s)
	}
	return s.DetailMuted.Render(padCenter("-", detailCellWidth))
}

func renderScoreCell(col detailColumn, holes map[int]espn.HoleScore, s Styles) string {
	if col.summary {
		if value, ok := strokeValue(col, holes); ok {
			return s.DetailTitle.Render(padCenter(fmt.Sprintf("%d", value), detailCellWidth))
		}
		return s.DetailMuted.Render(padCenter("-", detailCellWidth))
	}

	hole, ok := singleHole(col, holes)
	if !ok || !hole.Played {
		return s.DetailMuted.Render(padCenter("-", detailCellWidth))
	}

	text := formatScoreMarker(hole)
	scoreStyle := s.DetailValue
	switch hole.ScoreType {
	case "eagle":
		scoreStyle = s.DetailEagle
	case "birdie":
		scoreStyle = s.DetailBirdie
	case "bogey":
		scoreStyle = s.DetailBogey
	case "double+":
		scoreStyle = s.DetailDouble
	case "par":
		scoreStyle = s.DetailPar
	}
	return scoreStyle.Render(padCenter(text, detailCellWidth))
}

func renderRunningToParCell(col detailColumn, holes map[int]espn.HoleScore, s Styles) string {
	if !col.summary {
		hole, ok := singleHole(col, holes)
		if !ok || !hole.Played {
			return s.DetailMuted.Render(padCenter("-", detailCellWidth))
		}
	}

	if value, ok := runningRelativeValue(col, holes); ok {
		return renderRelativeScore(formatRelativeValue(value), detailCellWidth, s)
	}
	return s.DetailMuted.Render(padCenter("-", detailCellWidth))
}

func formatScoreMarker(hole espn.HoleScore) string {
	strokes := fmt.Sprintf("%d", hole.Strokes)
	switch hole.ScoreType {
	case "eagle":
		return "((" + strokes + "))"
	case "birdie":
		return "(" + strokes + ")"
	case "bogey":
		return "[" + strokes + "]"
	case "double+":
		return "[[" + strokes + "]]"
	default:
		return strokes
	}
}

func singleHole(col detailColumn, holes map[int]espn.HoleScore) (espn.HoleScore, bool) {
	if len(col.holes) != 1 {
		return espn.HoleScore{}, false
	}
	hole, ok := holes[col.holes[0]]
	return hole, ok
}

func parValue(col detailColumn, holes map[int]espn.HoleScore) (int, bool) {
	total := 0
	found := false
	for _, holeNum := range col.holes {
		hole, ok := holes[holeNum]
		if !ok || hole.Par <= 0 {
			continue
		}
		total += hole.Par
		found = true
	}
	return total, found
}

func strokeValue(col detailColumn, holes map[int]espn.HoleScore) (int, bool) {
	total := 0
	found := false
	for _, holeNum := range col.holes {
		hole, ok := holes[holeNum]
		if !ok || !hole.Played || hole.Strokes <= 0 {
			continue
		}
		total += hole.Strokes
		found = true
	}
	return total, found
}

func relativeValue(col detailColumn, holes map[int]espn.HoleScore) (int, bool) {
	total := 0
	found := false
	for _, holeNum := range col.holes {
		hole, ok := holes[holeNum]
		if !ok || !hole.Played || hole.Strokes <= 0 || hole.Par <= 0 {
			continue
		}
		total += hole.Strokes - hole.Par
		found = true
	}
	return total, found
}

func runningRelativeValue(col detailColumn, holes map[int]espn.HoleScore) (int, bool) {
	if len(col.holes) == 0 {
		return 0, false
	}

	end := col.holes[len(col.holes)-1]
	total := 0
	found := false
	for holeNum := 1; holeNum <= end; holeNum++ {
		hole, ok := holes[holeNum]
		if !ok || !hole.Played || hole.Strokes <= 0 || hole.Par <= 0 {
			continue
		}
		total += hole.Strokes - hole.Par
		found = true
	}
	return total, found
}

func formatRelativeValue(value int) string {
	switch {
	case value == 0:
		return "E"
	case value > 0:
		return fmt.Sprintf("+%d", value)
	default:
		return fmt.Sprintf("%d", value)
	}
}

func padCenter(text string, width int) string {
	tw := runewidth.StringWidth(text)
	if tw >= width {
		return runewidth.Truncate(text, width, "")
	}
	left := (width - tw) / 2
	right := width - tw - left
	return strings.Repeat(" ", left) + text + strings.Repeat(" ", right)
}
