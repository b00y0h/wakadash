# Changelog

All notable changes to wakadash will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0] - 2026-02-25

### Added

- Weekly summary browser with dedicated `w` keybinding
- WeeklyBrowserModel for browsing weekly coding summaries

### Fixed

- Test suite updated to use DataSource instead of archive.Fetcher

## [1.0.0] - 2026-02-25

### Added

- Historical data browsing with week-based navigation (`[`/`]` keys)
- GitHub archive fetcher for historical WakaTime data
- Hybrid DataSource with automatic routing between live API and archive
- Background prefetch for smoother navigation through history
- End-of-history detection with banner indicator
- Auto-skip empty weeks when navigating
- Week range display in status bar
- Paused auto-refresh indicator when viewing historical data
- Configuration validation with `[wakadash]` section and `history_repo` field
- Auto-create config section template on startup
- Bordered panels with centered titles
- Two-tone horizontal bar charts matching wakafetch style

### Changed

- Restructured repository to root-level code layout
- Stats computed from summaries endpoint instead of broken stats endpoint
- Bar charts stretch to fill terminal width

### Fixed

- Config parser reads `api_url` and `api_key` from `[settings]` section
- Time/percent labels on Languages and Projects charts
- Bar chart row spacing and visual alignment

## [0.3.0] - 2026-02-23

### Added

- Theme system with 6 built-in presets
- Interactive theme picker with mini dashboard preview
- Theme persistence via config file
- Stats panels for languages, projects, editors, and operating systems
- Summary panel with streak calculation
- Responsive grid layout with panel visibility toggles (number keys)
- Case-insensitive theme lookup

### Fixed

- Division by zero protection in stats panels
- Enhanced terminal size error with dimensions

## [0.2.0] - 2026-02-19

### Added

- Full-screen TUI dashboard using Bubbletea/Bubbles/Lipgloss
- Horizontal bar charts for languages and projects with GitHub Linguist color palette
- Sparkline for hourly activity visualization
- Heatmap for weekly activity visualization
- Auto-refresh with configurable interval
- Help overlay with keybinding reference
- Exponential backoff for API rate limiting
- Panel resize handling
- Homebrew tap publishing via GoReleaser

## [0.1.0] - 2026-02-19

### Added

- Initial release
- Go module with CLI entry point (`--help`, `--version`)
- WakaTime API client with config and type packages
- GitHub Actions CI/CD with GoReleaser v2
- Build workflow for push/PR verification

<!-- GoReleaser will populate releases here -->

[1.1.0]: https://github.com/b00y0h/wakafetch-brew/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/b00y0h/wakafetch-brew/compare/v0.3.0...v1.0.0
[0.3.0]: https://github.com/b00y0h/wakafetch-brew/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/b00y0h/wakafetch-brew/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/b00y0h/wakafetch-brew/releases/tag/v0.1.0
