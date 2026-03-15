package ui

import (
	"charm.land/lipgloss/v2"
)

// Color palette - modern, golf-inspired
var (
	// Base colors
	colorWhite    = lipgloss.Color("#FFFFFF")
	colorGray     = lipgloss.Color("#6B7280")
	colorDarkGray = lipgloss.Color("#374151")
	colorDimGray  = lipgloss.Color("#4B5563")
	colorBg       = lipgloss.Color("#111827")
	colorBgAlt    = lipgloss.Color("#1F2937")
	colorBorder   = lipgloss.Color("#374151")

	// Accent colors
	colorGreen  = lipgloss.Color("#10B981") // under par
	colorRed    = lipgloss.Color("#EF4444") // over par
	colorYellow = lipgloss.Color("#F59E0B") // even par
	colorBlue   = lipgloss.Color("#3B82F6") // highlights
	colorCyan   = lipgloss.Color("#06B6D4") // tournament info
	colorPurple = lipgloss.Color("#8B5CF6") // eagle or better
	colorAmber  = lipgloss.Color("#D97706") // bogey+

	// Score colors
	colorEagle  = lipgloss.Color("#A78BFA") // purple for eagle
	colorBirdie = lipgloss.Color("#34D399") // green for birdie
	colorPar    = lipgloss.Color("#9CA3AF") // gray for par
	colorBogey  = lipgloss.Color("#FBBF24") // yellow for bogey
	colorDouble = lipgloss.Color("#F87171") // red for double+
)

// Styles holds all the Lip Gloss styles used in the UI.
type Styles struct {
	// Header
	TournamentName lipgloss.Style
	RoundStatus    lipgloss.Style
	DateRange      lipgloss.Style
	HeaderBar      lipgloss.Style

	// Table
	TableHeader  lipgloss.Style
	TableRow     lipgloss.Style
	TableRowAlt  lipgloss.Style
	TableDivider lipgloss.Style
	SelectedRow  lipgloss.Style

	// Columns
	Marker         lipgloss.Style
	Position       lipgloss.Style
	ChangeUp       lipgloss.Style
	ChangeDown     lipgloss.Style
	ChangeNeutral  lipgloss.Style
	UpdateScore    lipgloss.Style
	UpdateStanding lipgloss.Style
	UpdateBoth     lipgloss.Style
	PlayerName     lipgloss.Style
	FavoritePlayer lipgloss.Style
	Country        lipgloss.Style
	ScoreUnder     lipgloss.Style
	ScoreOver      lipgloss.Style
	ScoreEven      lipgloss.Style
	RoundScore     lipgloss.Style
	Thru           lipgloss.Style

	// Status
	StatusCut lipgloss.Style
	StatusWD  lipgloss.Style
	CutLine   lipgloss.Style

	// Status bar
	StatusBar       lipgloss.Style
	StatusKey       lipgloss.Style
	StatusValue     lipgloss.Style
	StatusDim       lipgloss.Style
	HelpPanel       lipgloss.Style
	HelpTitle       lipgloss.Style
	DetailTitle     lipgloss.Style
	DetailTab       lipgloss.Style
	DetailTabActive lipgloss.Style
	DetailLabel     lipgloss.Style
	DetailValue     lipgloss.Style
	DetailMuted     lipgloss.Style
	DetailEagle     lipgloss.Style
	DetailBirdie    lipgloss.Style
	DetailPar       lipgloss.Style
	DetailBogey     lipgloss.Style
	DetailDouble    lipgloss.Style

	// General
	App     lipgloss.Style
	Error   lipgloss.Style
	Loading lipgloss.Style
}

// DefaultStyles returns the default style set.
func DefaultStyles() Styles {
	return Styles{
		// Header
		TournamentName: lipgloss.NewStyle().
			Bold(true).
			Foreground(colorWhite),

		RoundStatus: lipgloss.NewStyle().
			Foreground(colorCyan),

		DateRange: lipgloss.NewStyle().
			Foreground(colorGray),

		HeaderBar: lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(colorBorder).
			MarginBottom(1),

		// Table
		TableHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(colorGray).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(colorBorder),

		TableRow: lipgloss.NewStyle().
			Foreground(colorWhite),

		TableRowAlt: lipgloss.NewStyle().
			Foreground(colorWhite),

		TableDivider: lipgloss.NewStyle().
			Foreground(colorBorder),

		SelectedRow: lipgloss.NewStyle().
			Background(colorBgAlt),

		// Columns — width/alignment handled by padRight/padLeft in table.go
		// to correctly handle multi-byte unicode characters (ø, ä, å, etc.)
		Marker: lipgloss.NewStyle().
			Foreground(colorBlue).
			Bold(true),

		Position: lipgloss.NewStyle().
			Foreground(colorDimGray),

		ChangeUp: lipgloss.NewStyle().
			Foreground(colorGreen).
			Bold(true),

		ChangeDown: lipgloss.NewStyle().
			Foreground(colorRed).
			Bold(true),

		ChangeNeutral: lipgloss.NewStyle().
			Foreground(colorDimGray),

		UpdateScore: lipgloss.NewStyle().
			Foreground(colorCyan).
			Bold(true),

		UpdateStanding: lipgloss.NewStyle().
			Foreground(colorBlue).
			Bold(true),

		UpdateBoth: lipgloss.NewStyle().
			Foreground(colorYellow).
			Bold(true),

		PlayerName: lipgloss.NewStyle().
			Foreground(colorWhite),

		FavoritePlayer: lipgloss.NewStyle().
			Foreground(colorWhite).
			Bold(true),

		Country: lipgloss.NewStyle().
			Foreground(colorGray),

		ScoreUnder: lipgloss.NewStyle().
			Foreground(colorGreen).
			Bold(true),

		ScoreOver: lipgloss.NewStyle().
			Foreground(colorRed).
			Bold(true),

		ScoreEven: lipgloss.NewStyle().
			Foreground(colorYellow).
			Bold(true),

		RoundScore: lipgloss.NewStyle().
			Foreground(colorGray),

		Thru: lipgloss.NewStyle().
			Foreground(colorDimGray),

		// Status
		StatusCut: lipgloss.NewStyle().
			Foreground(colorRed).
			Faint(true),

		StatusWD: lipgloss.NewStyle().
			Foreground(colorAmber).
			Faint(true),

		CutLine: lipgloss.NewStyle().
			Foreground(colorBorder).
			Faint(true),

		// Status bar
		StatusBar: lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(colorBorder).
			Foreground(colorGray).
			MarginTop(1),

		StatusKey: lipgloss.NewStyle().
			Foreground(colorBlue).
			Bold(true),

		StatusValue: lipgloss.NewStyle().
			Foreground(colorGray),

		StatusDim: lipgloss.NewStyle().
			Foreground(colorDimGray),

		HelpPanel: lipgloss.NewStyle().
			Foreground(colorGray),

		HelpTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(colorWhite),

		DetailTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(colorWhite),

		DetailTab: lipgloss.NewStyle().
			Foreground(colorGray),

		DetailTabActive: lipgloss.NewStyle().
			Foreground(colorBlue).
			Bold(true),

		DetailLabel: lipgloss.NewStyle().
			Foreground(colorDimGray),

		DetailValue: lipgloss.NewStyle().
			Foreground(colorWhite),

		DetailMuted: lipgloss.NewStyle().
			Foreground(colorGray),

		DetailEagle: lipgloss.NewStyle().
			Foreground(colorPurple).
			Bold(true),

		DetailBirdie: lipgloss.NewStyle().
			Foreground(colorBirdie).
			Bold(true),

		DetailPar: lipgloss.NewStyle().
			Foreground(colorPar),

		DetailBogey: lipgloss.NewStyle().
			Foreground(colorBogey).
			Bold(true),

		DetailDouble: lipgloss.NewStyle().
			Foreground(colorDouble).
			Bold(true),

		// General
		App: lipgloss.NewStyle(),

		Error: lipgloss.NewStyle().
			Foreground(colorRed).
			Bold(true),

		Loading: lipgloss.NewStyle().
			Foreground(colorCyan),
	}
}
