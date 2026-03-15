package espn

import "time"

// ScoreboardResponse is the top-level response from the ESPN golf scoreboard API.
type ScoreboardResponse struct {
	Leagues []League `json:"leagues"`
	Season  Season   `json:"season"`
	Day     Day      `json:"day"`
	Events  []Event  `json:"events"`
}

type Season struct {
	Type int `json:"type"`
	Year int `json:"year"`
}

type Day struct {
	Date string `json:"date"`
}

// League contains PGA Tour metadata and the season calendar.
type League struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Abbreviation string          `json:"abbreviation"`
	Slug         string          `json:"slug"`
	Season       LeagueSeason    `json:"season"`
	Calendar     []CalendarEntry `json:"calendar"`
}

type LeagueSeason struct {
	Year      int    `json:"year"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
}

type CalendarEntry struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
}

// Event represents a single tournament.
type Event struct {
	ID           string        `json:"id"`
	UID          string        `json:"uid"`
	Name         string        `json:"name"`
	ShortName    string        `json:"shortName"`
	Date         string        `json:"date"`
	EndDate      string        `json:"endDate"`
	Season       EventSeason   `json:"season"`
	Competitions []Competition `json:"competitions"`
	Status       EventStatus   `json:"status"`
	Links        []Link        `json:"links"`
}

type EventSeason struct {
	Year int    `json:"year"`
	Type int    `json:"type"`
	Slug string `json:"slug"`
}

type EventStatus struct {
	Type StatusType `json:"type"`
}

type Link struct {
	Href string `json:"href"`
	Text string `json:"text"`
}

// Competition holds the actual leaderboard data.
type Competition struct {
	ID          string       `json:"id"`
	Date        string       `json:"date"`
	StartDate   string       `json:"startDate"`
	EndDate     string       `json:"endDate"`
	Status      CompStatus   `json:"status"`
	Competitors []Competitor `json:"competitors"`
}

type CompStatus struct {
	Period int        `json:"period"`
	Type   StatusType `json:"type"`
}

type StatusType struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	State       string `json:"state"`
	Completed   bool   `json:"completed"`
	Detail      string `json:"detail"`
	ShortDetail string `json:"shortDetail"`
}

// Competitor is a player on the leaderboard.
type Competitor struct {
	ID         string      `json:"id"`
	UID        string      `json:"uid"`
	Type       string      `json:"type"`
	Order      int         `json:"order"`
	Score      string      `json:"score"`
	Athlete    Athlete     `json:"athlete"`
	LineScores []RoundData `json:"linescores"`
	Status     *CompStatus `json:"status,omitempty"`
}

type Athlete struct {
	ID          string `json:"id"`
	FullName    string `json:"fullName"`
	DisplayName string `json:"displayName"`
	ShortName   string `json:"shortName"`
	Flag        *Flag  `json:"flag,omitempty"`
}

type Flag struct {
	Href string   `json:"href"`
	Alt  string   `json:"alt"`
	Rel  []string `json:"rel"`
}

// RoundData holds per-round scoring data.
type RoundData struct {
	Value        float64          `json:"value"`
	DisplayValue string           `json:"displayValue"`
	Period       int              `json:"period"`
	LineScores   []HoleData       `json:"linescores,omitempty"`
	Statistics   *RoundStatistics `json:"statistics,omitempty"`
}

// HoleData holds per-hole scoring data within a round.
type HoleData struct {
	Value        float64    `json:"value"`
	DisplayValue string     `json:"displayValue"`
	Period       int        `json:"period"` // hole number (1-18)
	ScoreType    *ScoreType `json:"scoreType,omitempty"`
}

type ScoreType struct {
	DisplayValue string `json:"displayValue"`
}

type RoundStatistics struct {
	Categories []StatCategory `json:"categories"`
}

type StatCategory struct {
	Stats []Stat `json:"stats"`
}

type Stat struct {
	Value        float64 `json:"value"`
	DisplayValue string  `json:"displayValue"`
}

// Tournament is a simplified view of an event for display purposes.
type Tournament struct {
	ID        string
	Name      string
	StartDate time.Time
	EndDate   time.Time
	Round     int
	Status    string // "pre", "in", "post"
	Detail    string // e.g. "Round 2 - In Progress"
	Players   []Player
}

// Player is a simplified view of a competitor for display purposes.
type Player struct {
	ID              string
	CanonicalRank   int
	DisplayPosition int
	Tied            bool // true if tied with another player at this display position
	Name            string
	ShortName       string
	Country         string
	CountryCode     string
	TotalScore      string // e.g. "-12", "+3", "E"
	Rounds          []RoundScore
	Thru            string // e.g. "F", "12", "-"
	Status          string // "", "CUT", "WD", "MDF", "DQ"
}

// RoundScore is a simplified per-round score.
type RoundScore struct {
	Round   int
	Strokes int    // 0 if not played
	ToPar   string // e.g. "-3", "+1", "E", ""
	Played  bool
	Holes   []HoleScore
}

// HoleScore is a simplified per-hole score within a round.
type HoleScore struct {
	Number    int
	Par       int
	Strokes   int
	ScoreType string // eagle, birdie, par, bogey, double+, or ""
	Played    bool
}
