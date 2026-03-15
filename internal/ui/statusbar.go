package ui

import (
	"fmt"
	"time"
)

type statusSegment struct {
	rendered string
	width    int
}

// RenderStatusBar renders the bottom status bar with last update time and keybinds.
func RenderStatusBar(lastUpdate time.Time, nextRefresh time.Duration, width int, errMsg string, filterQuery string, searchMode bool, showHelp bool, roundMode string, favoritesOnly bool, showDetail bool) string {
	s := DefaultStyles()

	var leftSegments []statusSegment
	if errMsg != "" {
		leftSegments = append(leftSegments, newStatusSegment(s.Error.Render(fmt.Sprintf(" %s", errMsg))))
	} else if !lastUpdate.IsZero() {
		timeStr := lastUpdate.Format("3:04:05 PM")
		leftSegments = append(leftSegments, newStatusSegment(fmt.Sprintf(" %s %s",
			s.StatusDim.Render("Updated"),
			s.StatusValue.Render(timeStr),
		)))
		if nextRefresh > 0 {
			secs := int(nextRefresh.Seconds())
			leftSegments = append(leftSegments, newStatusSegment(fmt.Sprintf("%s %s",
				s.StatusDim.Render("Next"),
				s.StatusValue.Render(fmt.Sprintf("%ds", secs)),
			)))
		}
	} else {
		leftSegments = append(leftSegments, newStatusSegment(s.StatusDim.Render(" Fetching data...")))
	}

	if filterQuery != "" || searchMode {
		leftSegments = append(leftSegments, newStatusSegment(fmt.Sprintf("%s %s",
			s.StatusDim.Render("Filter"),
			s.StatusValue.Render(fmt.Sprintf("/%s", filterQuery)),
		)))
		if searchMode {
			leftSegments = append(leftSegments, newStatusSegment(s.StatusDim.Render("(search)")))
		}
	}

	if roundMode != "" {
		leftSegments = append(leftSegments, newStatusSegment(fmt.Sprintf("%s %s",
			s.StatusDim.Render("Rounds"),
			s.StatusValue.Render(roundMode),
		)))
	}

	if favoritesOnly {
		leftSegments = append(leftSegments, newStatusSegment(fmt.Sprintf("%s %s",
			s.StatusDim.Render("View"),
			s.StatusValue.Render("favorites"),
		)))
	}

	// Keybind hints
	helpLabel := "show hints"
	if showHelp {
		helpLabel = "hide hints"
	}

	rightCandidates := buildRightCandidates(s, searchMode, helpLabel, showDetail)
	fullLeftWidth := totalStatusWidth(leftSegments)
	right := rightCandidates[len(rightCandidates)-1]
	for _, candidate := range rightCandidates {
		if width-candidate.width-1 >= fullLeftWidth {
			right = candidate
			break
		}
	}

	maxLeftWidth := width - right.width - 1
	left := joinStatusSegments(leftSegments, maxLeftWidth)
	leftWidth := lipglossWidth(left)
	gap := width - leftWidth - right.width
	if gap < 1 {
		gap = 1
	}

	bar := left + fmt.Sprintf("%*s", gap, "") + right.rendered
	return s.StatusBar.Width(width).Render(bar)
}

func newStatusSegment(rendered string) statusSegment {
	return statusSegment{rendered: rendered, width: lipglossWidth(rendered)}
}

func joinStatusSegments(segments []statusSegment, maxWidth int) string {
	if len(segments) == 0 || maxWidth <= 0 {
		return ""
	}

	joined := segments[0].rendered
	used := segments[0].width
	for _, segment := range segments[1:] {
		segmentWidth := 2 + segment.width
		if used+segmentWidth > maxWidth {
			break
		}
		joined += "  " + segment.rendered
		used += segmentWidth
	}

	return joined
}

func totalStatusWidth(segments []statusSegment) int {
	if len(segments) == 0 {
		return 0
	}

	width := segments[0].width
	for _, segment := range segments[1:] {
		width += 2 + segment.width
	}
	return width
}

func buildRightCandidates(s Styles, searchMode bool, helpLabel string, showDetail bool) []statusSegment {
	if searchMode {
		return []statusSegment{
			newStatusSegment(fmt.Sprintf("%s %s  %s %s  %s %s  %s %s ",
				s.StatusKey.Render("?"),
				s.StatusDim.Render(helpLabel),
				s.StatusKey.Render("enter"),
				s.StatusDim.Render("apply"),
				s.StatusKey.Render("esc"),
				s.StatusDim.Render("clear"),
				s.StatusKey.Render("^c"),
				s.StatusDim.Render("quit"),
			)),
			newStatusSegment(fmt.Sprintf("%s %s  %s %s  %s %s ",
				s.StatusKey.Render("?"),
				s.StatusDim.Render(helpLabel),
				s.StatusKey.Render("enter"),
				s.StatusDim.Render("apply"),
				s.StatusKey.Render("esc"),
				s.StatusDim.Render("clear"),
			)),
			newStatusSegment(fmt.Sprintf("%s %s ",
				s.StatusKey.Render("?"),
				s.StatusDim.Render(helpLabel),
			)),
		}
	}

	if showDetail {
		return []statusSegment{
			newStatusSegment(fmt.Sprintf("%s %s  %s %s  %s %s  %s %s  %s %s ",
				s.StatusKey.Render("?"),
				s.StatusDim.Render(helpLabel),
				s.StatusKey.Render("tab"),
				s.StatusDim.Render("next round"),
				s.StatusKey.Render("esc"),
				s.StatusDim.Render("close"),
				s.StatusKey.Render("r"),
				s.StatusDim.Render("refresh"),
				s.StatusKey.Render("q"),
				s.StatusDim.Render("quit"),
			)),
			newStatusSegment(fmt.Sprintf("%s %s  %s %s  %s %s ",
				s.StatusKey.Render("?"),
				s.StatusDim.Render(helpLabel),
				s.StatusKey.Render("tab"),
				s.StatusDim.Render("next round"),
				s.StatusKey.Render("esc"),
				s.StatusDim.Render("close"),
			)),
			newStatusSegment(fmt.Sprintf("%s %s ",
				s.StatusKey.Render("?"),
				s.StatusDim.Render(helpLabel),
			)),
		}
	}

	return []statusSegment{
		newStatusSegment(fmt.Sprintf("%s %s  %s %s  %s %s  %s %s  %s %s  %s %s  %s %s  %s %s ",
			s.StatusKey.Render("?"),
			s.StatusDim.Render(helpLabel),
			s.StatusKey.Render("enter"),
			s.StatusDim.Render("scorecard"),
			s.StatusKey.Render("/"),
			s.StatusDim.Render("search"),
			s.StatusKey.Render("f"),
			s.StatusDim.Render("favorite"),
			s.StatusKey.Render("F"),
			s.StatusDim.Render("favorites only"),
			s.StatusKey.Render("t"),
			s.StatusDim.Render("rounds"),
			s.StatusKey.Render("r"),
			s.StatusDim.Render("refresh"),
			s.StatusKey.Render("q"),
			s.StatusDim.Render("quit"),
		)),
		newStatusSegment(fmt.Sprintf("%s %s  %s %s  %s %s  %s %s ",
			s.StatusKey.Render("?"),
			s.StatusDim.Render(helpLabel),
			s.StatusKey.Render("enter"),
			s.StatusDim.Render("scorecard"),
			s.StatusKey.Render("F"),
			s.StatusDim.Render("favorites only"),
			s.StatusKey.Render("q"),
			s.StatusDim.Render("quit"),
		)),
		newStatusSegment(fmt.Sprintf("%s %s  %s %s ",
			s.StatusKey.Render("?"),
			s.StatusDim.Render(helpLabel),
			s.StatusKey.Render("/"),
			s.StatusDim.Render("search"),
		)),
	}
}
