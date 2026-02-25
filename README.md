# wakadash

> Like htop for your coding activity

A beautiful, live-updating terminal dashboard for [WakaTime](https://wakatime.com) and [Wakapi](https://wakapi.dev) coding stats. View your projects, languages, editors, and activity — all in your terminal with stunning visualizations and theme support.

## Features

- **Live Dashboard** — Auto-refreshing stats with configurable interval
- **9 Visualization Panels** — Languages, Projects, Categories, Editors, OS, Machines, Sparkline, Heatmap, Summary
- **6 Theme Presets** — Dracula, Nord, Gruvbox, Monokai, Solarized, Tokyo Night
- **Historical Data** — Browse past weeks via GitHub archive integration
- **Keyboard Navigation** — Toggle panels, switch themes, navigate history
- **Responsive Layout** — 2-column grid (≥80 cols) or vertical stack (40-79 cols)
- **Rate Limit Handling** — Exponential backoff with visual indicator

## Installation

### Homebrew (macOS/Linux)

```bash
brew tap b00y0h/wakadash
brew install wakadash
```

### From Source

```bash
go install github.com/b00y0h/wakadash/cmd/wakadash@latest
```

### Prerequisites

You need a [WakaTime](https://wakatime.com) account (or self-hosted [Wakapi](https://wakapi.dev) instance) with `~/.wakatime.cfg` configured.

## Configuration

wakadash reads `~/.wakatime.cfg` — the same config file used by WakaTime editor plugins.

### Basic Setup

```ini
[settings]
api_url = https://wakatime.com/api/v1
api_key = your-api-key-here
```

For Wakapi users:

```ini
[settings]
api_url = https://your-wakapi-instance.com/api
api_key = your-api-key-here
```

### Theme Selection

```ini
[wakadash]
theme = dracula
```

Available themes: `dracula`, `nord`, `gruvbox`, `monokai`, `solarized`, `tokyo-night`

### Historical Data (Optional)

To browse historical stats beyond 7 days, configure a GitHub archive repository:

```ini
[wakadash]
theme = dracula
history_repo = username/wakatime-archive
```

The `history_repo` should contain archived WakaTime data in the format:
```
data/
  2024-01-15.json
  2024-01-16.json
  ...
```

See [wakasync](https://github.com/b00y0h/wakasync) for automated archiving.

## Usage

```bash
# Launch the dashboard
wakadash

# Show help
wakadash --help

# Show version
wakadash --version
```

## Keyboard Controls

### Panel Toggles

| Key | Action |
|-----|--------|
| `1` | Toggle Languages panel |
| `2` | Toggle Projects panel |
| `3` | Toggle Sparkline panel |
| `4` | Toggle Heatmap panel |
| `5` | Toggle Categories panel |
| `6` | Toggle Editors panel |
| `7` | Toggle OS panel |
| `8` | Toggle Machines panel |
| `9` | Toggle Summary panel |
| `a` | Show all panels |
| `h` | Hide all panels |

### Navigation

| Key | Action |
|-----|--------|
| `←` | Previous week (historical data) |
| `→` | Next week |
| `0` / `Home` | Return to current week |
| `t` | Open theme picker |
| `?` | Toggle help overlay |
| `q` / `Ctrl+C` | Quit |

### Historical Data Navigation

When `history_repo` is configured:
- Press `←` to navigate to previous weeks
- Navigation auto-skips weeks with no data
- `[HISTORICAL]` indicator shows when viewing archived data
- Auto-refresh pauses during historical browsing
- Press `0` to return to live view

## Panels

| Panel | Description |
|-------|-------------|
| **Languages** | Top 10 programming languages with color-coded bars |
| **Projects** | Top 10 projects by time spent |
| **Categories** | Coding, Building, Debugging breakdown |
| **Editors** | Editor/IDE usage distribution |
| **OS** | Operating system breakdown |
| **Machines** | Machine/hostname distribution |
| **Sparkline** | Hourly activity over last 24 hours |
| **Heatmap** | Weekly activity grid (7 days × 24 hours) |
| **Summary** | 30-day totals, daily average, and streak |

## Themes

On first run, wakadash presents a theme picker. You can also:
- Press `t` anytime to change themes
- Set `theme` in `~/.wakatime.cfg` directly

### Theme Previews

- **Dracula** — Purple and pink accents on dark background
- **Nord** — Cool blue tones inspired by Arctic palettes
- **Gruvbox** — Warm retro colors with orange accents
- **Monokai** — Classic syntax highlighting colors
- **Solarized** — Precision colors for readability
- **Tokyo Night** — Modern Japanese cityscape palette

## How It Works

1. **Recent Data (≤7 days)** — Fetched from WakaTime/Wakapi API
2. **Historical Data (>7 days)** — Fetched from GitHub archive (if configured)
3. **Background Prefetch** — Previous week loads silently for instant navigation
4. **Smart Caching** — Empty weeks are cached to avoid redundant fetches

## Troubleshooting

### "No data available"

- Verify your API key is correct in `~/.wakatime.cfg`
- Check that WakaTime plugins are tracking your activity
- Ensure `api_url` matches your provider (WakaTime vs Wakapi)

### Rate limiting

wakadash handles rate limits automatically with exponential backoff. A visual indicator shows when rate-limited.

### Historical data not loading

- Verify `history_repo` format is `username/repo`
- Ensure the repository is public (or you have access)
- Check that archive files exist in `data/YYYY-MM-DD.json` format

## Related Projects

- [wakasync](https://github.com/b00y0h/wakasync) — GitHub Action to archive WakaTime stats daily
- [wakafetch](https://github.com/sahaj-b/wakafetch) — Inspiration for this project

## License

MIT — see [LICENSE](LICENSE)
