package ui

import (
	"fmt"
	"strings"

	"github.com/nickkoul/gstat/internal/espn"
)

// RenderHeader renders the tournament header bar.
func RenderHeader(t *espn.Tournament, width int) string {
	s := DefaultStyles()

	if t == nil {
		return s.Loading.Render("  Loading tournament data...")
	}

	// Tournament name
	name := s.TournamentName.Render(t.Name)

	// Round status
	status := s.RoundStatus.Render(t.Detail)

	// Date range
	dateFormat := "Jan 2, 2006"
	dates := fmt.Sprintf("%s - %s",
		t.StartDate.Format(dateFormat),
		t.EndDate.Format(dateFormat),
	)
	dateStr := s.DateRange.Render(dates)

	// Build the header
	// Line 1: Tournament name + status (right-aligned)
	nameWidth := lipglossWidth(name)
	statusWidth := lipglossWidth(status)
	gap := width - nameWidth - statusWidth - 4 // 4 for padding
	if gap < 1 {
		gap = 1
	}
	line1 := fmt.Sprintf("  %s%s%s", name, strings.Repeat(" ", gap), status)

	// Line 2: Date range
	line2 := fmt.Sprintf("  %s", dateStr)

	header := line1 + "\n" + line2
	return s.HeaderBar.Width(width).Render(header)
}

// lipglossWidth returns the visible width of a rendered string,
// accounting for ANSI escape sequences.
func lipglossWidth(s string) int {
	// Strip ANSI escape sequences to get the actual visible width.
	// A simple approach: count non-escape characters.
	inEscape := false
	width := 0
	for _, r := range s {
		if r == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				inEscape = false
			}
			continue
		}
		width++
	}
	return width
}
