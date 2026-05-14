package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/allenbijo/QuickSSH-terminal/internal/store"
	"github.com/allenbijo/QuickSSH-terminal/internal/tui"
)

// Populated by goreleaser ldflags at release time.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	for _, a := range os.Args[1:] {
		switch a {
		case "-v", "--version", "version":
			fmt.Printf("qsh %s (commit %s, built %s)\n", version, commit, date)
			return
		case "-h", "--help", "help":
			fmt.Println("qsh — terminal UI for SSH port-forward tunnels (QuickSSH Terminal)")
			fmt.Println("Usage: qsh             launch the TUI")
			fmt.Println("       qsh --version   print version")
			return
		}
	}

	s, err := store.Open()
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to open config store:", err)
		os.Exit(1)
	}

	model := tui.NewModel(s)
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "tui error:", err)
		os.Exit(1)
	}
}
