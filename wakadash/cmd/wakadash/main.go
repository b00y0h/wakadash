// Command wakadash is a live terminal dashboard for WakaTime coding stats.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/b00y0h/wakadash/internal/api"
	"github.com/b00y0h/wakadash/internal/config"
	"github.com/b00y0h/wakadash/internal/theme"
	"github.com/b00y0h/wakadash/internal/tui"
)

// Build-time variables injected by GoReleaser via -ldflags.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	showVersion := flag.Bool("version", false, "Print version information and exit")
	rangeFlag := flag.String("range", "last_7_days",
		"Time range for stats (last_7_days, last_30_days, last_6_months, last_year, all_time)")
	refreshFlag := flag.Int("refresh", 60, "Auto-refresh interval in seconds (0 to disable)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: wakadash [options]\n\n")
		fmt.Fprintf(os.Stderr, "A live terminal dashboard for WakaTime coding stats.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *showVersion {
		fmt.Printf("wakadash %s\n", version)
		fmt.Printf("  commit: %s\n", commit)
		fmt.Printf("  built:  %s\n", date)
		fmt.Printf("  go:     %s\n", runtime.Version())
		os.Exit(0)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "\nEnsure ~/.wakatime.cfg contains api_url and api_key.\n")
		fmt.Fprintf(os.Stderr, "See: https://wakatime.com/settings/account\n")
		os.Exit(1)
	}

	// Ensure [wakadash] section exists in config file (for discoverability)
	if err := config.EnsureWakadashSection(); err != nil {
		// Log warning but continue - this is not critical for startup
		log.Printf("Warning: could not ensure [wakadash] section: %v", err)
	}

	client := api.New(cfg.APIKey, cfg.APIURL)
	refreshInterval := time.Duration(*refreshFlag) * time.Second

	// Check if first run (no theme configured)
	themeName, _ := theme.LoadThemeFromConfig()
	isFirstRun := themeName == ""

	if isFirstRun {
		// Show theme picker on first run
		// isFirstRun=true means Esc/Q are ignored (user MUST select a theme)
		picker := tui.NewThemePicker(true)
		pickerProgram := tea.NewProgram(picker, tea.WithAltScreen())
		_, err := pickerProgram.Run()
		if err != nil {
			log.Fatalf("theme picker error: %v", err)
		}
		// Note: picker.SaveThemeToConfig() already called on Enter (see picker.go)
		// No need to save again here — theme is persisted when user confirms in picker
	}

	m := tui.NewModel(client, *rangeFlag, refreshInterval)

	// tea.WithAltScreen() is the correct approach for full-screen apps.
	// Per research: "Because commands run asynchronously, EnterAltScreen should
	// not be used in Init. Use the WithAltScreen ProgramOption instead."
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
