package espn

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"
)

const (
	scoreboardURL = "https://site.api.espn.com/apis/site/v2/sports/golf/pga/scoreboard"
	userAgent     = "gstat/1.0"
	httpTimeout   = 10 * time.Second
)

// ESPN uses a shortened ISO 8601 format without seconds.
var espnDateFormats = []string{
	"2006-01-02T15:04Z",
	"2006-01-02T15:04:05Z",
	time.RFC3339,
}

// Client handles communication with the ESPN API.
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new ESPN API client.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: httpTimeout,
		},
	}
}

// FetchLeaderboard fetches the current PGA Tour leaderboard from ESPN.
func (c *Client) FetchLeaderboard() (*Tournament, error) {
	req, err := http.NewRequest("GET", scoreboardURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching leaderboard: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ESPN API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	var scoreboard ScoreboardResponse
	if err := json.Unmarshal(body, &scoreboard); err != nil {
		return nil, fmt.Errorf("parsing JSON response: %w", err)
	}

	return parseTournament(scoreboard)
}

// parseESPNDate tries multiple date formats to parse ESPN's date strings.
func parseESPNDate(s string) time.Time {
	for _, layout := range espnDateFormats {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

// parseTournament converts the raw ESPN response into a simplified Tournament.
func parseTournament(sb ScoreboardResponse) (*Tournament, error) {
	if len(sb.Events) == 0 {
		return nil, fmt.Errorf("no active tournament found")
	}

	event := sb.Events[0]

	t := &Tournament{
		ID:        event.ID,
		Name:      event.Name,
		StartDate: parseESPNDate(event.Date),
		EndDate:   parseESPNDate(event.EndDate),
		Status:    event.Status.Type.State,
	}

	if len(event.Competitions) == 0 {
		return t, nil
	}

	comp := event.Competitions[0]
	t.Round = comp.Status.Period
	t.Detail = comp.Status.Type.Detail

	// Detect whether a cut has been applied by checking if any player
	// has an R3 linescore entry (even a placeholder). This is how we know
	// the cut has been determined.
	cutApplied := false
	for _, c := range comp.Competitors {
		for _, ls := range c.LineScores {
			if ls.Period >= 3 {
				cutApplied = true
				break
			}
		}
		if cutApplied {
			break
		}
	}

	// Parse all players
	for _, c := range comp.Competitors {
		player := parsePlayer(c, t.Round, cutApplied)
		t.Players = append(t.Players, player)
	}

	// Calculate tied display positions while preserving canonical ranks.
	calculateTiedPositions(t.Players)

	return t, nil
}

// parsePlayer converts a raw Competitor into a simplified Player.
func parsePlayer(c Competitor, currentRound int, cutApplied bool) Player {
	playerID := c.Athlete.ID
	if playerID == "" {
		playerID = c.ID
	}

	p := Player{
		ID:              playerID,
		CanonicalRank:   c.Order,
		DisplayPosition: c.Order,
		Name:            c.Athlete.DisplayName,
		ShortName:       c.Athlete.ShortName,
		TotalScore:      c.Score,
	}

	if c.Athlete.Flag != nil {
		p.Country = c.Athlete.Flag.Alt
		p.CountryCode = extractCountryCode(c.Athlete.Flag.Href)
	}

	// Detect WD first -- WD players have a round where value is holes played
	// (e.g., value=4.0 meaning 4 holes) rather than total strokes (60-90 range),
	// and the round has no hole-by-hole data.
	isWD := false
	for _, ls := range c.LineScores {
		if ls.Value > 0 && ls.Value < 30 && ls.LineScores == nil && ls.DisplayValue != "" {
			// Value is way below any plausible stroke count -- this is holes played
			isWD = true
			break
		}
	}

	// Parse round scores (up to 4 rounds)
	for i := 0; i < 4; i++ {
		rs := RoundScore{Round: i + 1}
		if i < len(c.LineScores) {
			ls := c.LineScores[i]
			if ls.Value > 0 && ls.DisplayValue != "" && ls.DisplayValue != "-" {
				// For WD players, the value might be holes-played, not strokes.
				// Real stroke values are >= 50 for any golf round.
				if ls.Value >= 50 {
					rs.Played = true
					rs.Strokes = int(math.Round(ls.Value))
					rs.ToPar = ls.DisplayValue
				} else if !isWD {
					// Normal round with unusual low stroke count (shouldn't happen)
					rs.Played = true
					rs.Strokes = int(math.Round(ls.Value))
					rs.ToPar = ls.DisplayValue
				}
				// For WD players with low value, we skip marking as played
			}
		}
		p.Rounds = append(p.Rounds, rs)
	}

	// Determine "thru"
	p.Thru = determineThru(c)

	// Determine player status
	if isWD {
		p.Status = "WD"
	} else if cutApplied {
		p.Status = determineCutStatus(c)
	}

	return p
}

// calculateTiedPositions assigns tied display positions based on score groupings.
// ESPN's order field gives a unique canonical rank (1,2,3,...) but doesn't indicate ties.
// We detect ties by grouping players with the same score while keeping canonical ranks intact.
func calculateTiedPositions(players []Player) {
	if len(players) == 0 {
		return
	}

	for i := range players {
		players[i].Tied = false
		if players[i].DisplayPosition == 0 {
			players[i].DisplayPosition = players[i].CanonicalRank
		}
	}

	// Group by score (only for active players, not CUT/WD)
	type scoreGroup struct {
		score string
		start int
		count int
	}

	var groups []scoreGroup
	for i, p := range players {
		if p.Status == "CUT" || p.Status == "WD" || p.Status == "DQ" {
			continue
		}
		if len(groups) == 0 || groups[len(groups)-1].score != p.TotalScore {
			groups = append(groups, scoreGroup{score: p.TotalScore, start: i, count: 1})
		} else {
			groups[len(groups)-1].count++
		}
	}

	// Mark tied positions: if multiple players share a score, they're tied
	for _, g := range groups {
		if g.count > 1 {
			for i := g.start; i < g.start+g.count; i++ {
				players[i].Tied = true
				players[i].DisplayPosition = players[g.start].CanonicalRank
			}
		}
	}
}

// determineThru figures out how many holes a player has completed.
func determineThru(c Competitor) string {
	if len(c.LineScores) == 0 {
		return "-"
	}

	// Walk rounds in reverse to find the latest with data
	for i := len(c.LineScores) - 1; i >= 0; i-- {
		ls := c.LineScores[i]

		// If this round has hole-by-hole data, count completed holes
		if len(ls.LineScores) > 0 {
			if len(ls.LineScores) == 18 {
				return "F"
			}
			return fmt.Sprintf("%d", len(ls.LineScores))
		}

		// If the round has a stroke value, it's a completed round
		if ls.Value >= 50 && ls.DisplayValue != "" {
			return "F"
		}
	}

	return "-"
}

// determineCutStatus checks if a player missed the cut.
// The cut is detected by whether a player has an R3+ linescore entry.
func determineCutStatus(c Competitor) string {
	for _, ls := range c.LineScores {
		if ls.Period >= 3 {
			return "" // has R3+ data (even placeholder) = made the cut
		}
	}
	// No R3+ linescore at all = missed the cut
	return "CUT"
}

// extractCountryCode extracts the country code from a flag URL like
// "https://a.espncdn.com/i/teamlogos/countries/500/swe.png"
func extractCountryCode(href string) string {
	if href == "" {
		return ""
	}
	parts := strings.Split(href, "/")
	if len(parts) == 0 {
		return ""
	}
	last := parts[len(parts)-1]
	return strings.TrimSuffix(last, ".png")
}
