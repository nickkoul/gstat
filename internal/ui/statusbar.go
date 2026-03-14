package ui

import (
	"fmt"
	"time"
)

// RenderStatusBar renders the bottom status bar with last update time and keybinds.
func RenderStatusBar(lastUpdate time.Time, nextRefresh time.Duration, width int, errMsg string, filterQuery string, searchMode bool) string {
	s := DefaultStyles()

	var left string
	if errMsg != "" {
		left = s.Error.Render(fmt.Sprintf(" %s", errMsg))
	} else if !lastUpdate.IsZero() {
		timeStr := lastUpdate.Format("3:04:05 PM")
		left = fmt.Sprintf(" %s %s",
			s.StatusDim.Render("Updated"),
			s.StatusValue.Render(timeStr),
		)
		if nextRefresh > 0 {
			secs := int(nextRefresh.Seconds())
			left += fmt.Sprintf("  %s %s",
				s.StatusDim.Render("Next"),
				s.StatusValue.Render(fmt.Sprintf("%ds", secs)),
			)
		}
	} else {
		left = s.StatusDim.Render(" Fetching data...")
	}

	if filterQuery != "" || searchMode {
		query := filterQuery
		if query == "" {
			query = ""
		}
		left += fmt.Sprintf("  %s %s",
			s.StatusDim.Render("Filter"),
			s.StatusValue.Render(fmt.Sprintf("/%s", query)),
		)
		if searchMode {
			left += fmt.Sprintf(" %s", s.StatusDim.Render("(search)"))
		}
	}

	// Keybind hints
	var right string
	if searchMode {
		right = fmt.Sprintf("%s %s  %s %s  %s %s ",
			s.StatusKey.Render("enter"),
			s.StatusDim.Render("apply"),
			s.StatusKey.Render("esc"),
			s.StatusDim.Render("clear"),
			s.StatusKey.Render("^c"),
			s.StatusDim.Render("quit"),
		)
	} else {
		right = fmt.Sprintf("%s %s  %s %s  %s %s  %s %s  %s %s ",
			s.StatusKey.Render("/"),
			s.StatusDim.Render("search"),
			s.StatusKey.Render("j/k"),
			s.StatusDim.Render("scroll"),
			s.StatusKey.Render("^d/^u"),
			s.StatusDim.Render("jump"),
			s.StatusKey.Render("r"),
			s.StatusDim.Render("refresh"),
			s.StatusKey.Render("q"),
			s.StatusDim.Render("quit"),
		)
	}

	// Calculate gap between left and right
	leftWidth := lipglossWidth(left)
	rightWidth := lipglossWidth(right)
	gap := width - leftWidth - rightWidth
	if gap < 1 {
		gap = 1
	}

	bar := left + fmt.Sprintf("%*s", gap, "") + right
	return s.StatusBar.Width(width).Render(bar)
}
