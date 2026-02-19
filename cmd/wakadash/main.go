// Command wakadash is a live terminal dashboard for WakaTime coding stats.
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/b00y0h/wakadash/internal/config"
)

// Build-time variables injected by GoReleaser via -ldflags.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	showVersion := flag.Bool("version", false, "Print version information and exit")

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

	// Phase 5 will replace this stub with the full TUI dashboard.
	fmt.Printf("wakadash: dashboard launching... (Phase 5)\n")
	fmt.Printf("  API URL: %s\n", cfg.APIURL)
}
