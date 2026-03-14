package espn

import "time"

// TestParseDate is a helper for tests to parse ESPN date strings.
func TestParseDate(s string) time.Time {
	for _, layout := range []string{
		"2006-01-02T15:04Z",
		"2006-01-02T15:04:05Z",
		time.RFC3339,
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}
