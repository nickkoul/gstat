# Testing Guidelines

## Philosophy

Tests are organized into tiers based on value and complexity:

| Tier | What | Where | Priority |
|------|------|-------|----------|
| **1. Business Logic** | Data parsing, tied positions, cut/WD detection, date parsing | `internal/espn/*_test.go` | Highest -- these are where subtle bugs live |
| **2. UI Rendering** | Render functions produce correct content | `internal/ui/*_test.go` | Medium -- pure functions, easy to test |
| **3. Integration** | Full parsing pipeline against real API fixtures | `internal/espn/*_test.go` with `testdata/` | Medium -- catches regressions across the pipeline |

### What we don't test

- **Bubble Tea model lifecycle** -- requires TTY simulation, high effort, low value for a TUI this simple
- **Live ESPN API calls** -- the API is undocumented and can change; tests must never hit the network
- **Exact ANSI output** -- escape codes are terminal-dependent; assert on content, not formatting

## Running Tests

```bash
# Run all tests
go test ./...

# Run with verbose output (useful for visual render tests)
go test -v ./...

# Run a specific package
go test ./internal/espn/
go test ./internal/ui/

# Run a specific test
go test ./internal/espn/ -run TestCalculateTiedPositions

# Run with race detector
go test -race ./...

# Run with coverage
go test -cover ./...
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
```

## Conventions

### File placement

Test files live next to the code they test:

```
internal/espn/
├── client.go
├── client_test.go       ← tests for client.go
├── types.go
├── helpers.go
└── testdata/
    └── scoreboard.json  ← real ESPN API response fixture
```

### Package declaration

Use the **same package** (not `_test` suffix) for business logic tests. This lets us test unexported functions directly:

```go
package espn  // NOT package espn_test
```

Use the **external test package** (`_test` suffix) for UI tests, since those test the public API:

```go
package ui_test
```

### Table-driven tests

All business logic tests use table-driven format:

```go
func TestSomething(t *testing.T) {
    tests := []struct {
        name string
        // inputs
        want // expected output
    }{
        {name: "descriptive case name", ...},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := functionUnderTest(tt.input)
            if got != tt.want {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

### No network calls

Tests must never make HTTP requests. Use fixture files for integration tests:

```go
func loadFixture(t *testing.T, name string) []byte {
    t.Helper()
    data, err := os.ReadFile(filepath.Join("testdata", name))
    if err != nil {
        t.Fatalf("failed to load fixture %s: %v", name, err)
    }
    return data
}
```

### ANSI-safe assertions for UI tests

UI render functions return strings with ANSI escape codes. Use a strip function to assert on visible content:

```go
func stripANSI(s string) string {
    re := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
    return re.ReplaceAllString(s, "")
}

// Then assert:
got := stripANSI(ui.RenderPlayerRow(...))
if !strings.Contains(got, "T4") {
    t.Errorf("expected tied position T4 in output: %s", got)
}
```

## Test Fixtures

### Location

`internal/espn/testdata/`

### Capturing a new fixture

Save a real ESPN API response for offline testing:

```bash
curl -s 'https://site.api.espn.com/apis/site/v2/sports/golf/pga/scoreboard' | \
  python3 -m json.tool > internal/espn/testdata/scoreboard.json
```

Name fixtures descriptively if multiple are needed:

- `scoreboard.json` -- standard response with tournament in progress
- `scoreboard_no_event.json` -- response when no tournament is active
- `scoreboard_mid_round.json` -- response during an active round with partial holes

### What fixtures validate

Running `parseTournament()` against a fixture tests the entire parsing pipeline:
- JSON unmarshalling into typed structs
- Date parsing
- Player parsing (all 100+ players)
- Tied position calculation
- Cut detection
- WD detection
- Thru calculation

If the ESPN API format changes, fixture tests will catch the mismatch immediately.

## What to Test When Adding Features

| Change | Tests to add/update |
|--------|-------------------|
| New ESPN data field | Add struct field to `types.go`, add parsing in `client.go`, add unit test for the parsing logic |
| New keybinding | No test needed (Bubble Tea model, Tier 3 -- skip) |
| New UI column | Add render test asserting the column header and cell content appear |
| New business logic (e.g., favorites) | Table-driven unit tests for all edge cases |
| Bug fix | Add a regression test reproducing the bug before fixing it |
