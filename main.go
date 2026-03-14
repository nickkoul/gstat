package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/nickkoul/gstat/internal/model"
)

func main() {
	m := model.New()
	p := tea.NewProgram(m)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running gstat: %v\n", err)
		os.Exit(1)
	}
}
