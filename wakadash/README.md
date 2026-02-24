# wakadash

> Like htop for your coding activity

A live terminal dashboard for WakaTime coding stats. See your projects, languages, editors, and time breakdown — live in your terminal.

## Install

### Homebrew (macOS/Linux)

```bash
brew tap b00y0h/wakadash
brew install wakadash
```

### From source

```bash
go install github.com/b00y0h/wakadash/cmd/wakadash@latest
```

### Prerequisites

You need a [WakaTime](https://wakatime.com) account (or self-hosted [Wakapi](https://wakapi.dev) instance) and `~/.wakatime.cfg` configured with your API key.

Your `~/.wakatime.cfg` should contain:

```ini
[settings]
api_url = https://wakatime.com/api
api_key = your-api-key-here
```

## Usage

```bash
# Launch the dashboard
wakadash

# Show help
wakadash --help

# Show version
wakadash --version
```

## Configuration

wakadash reads `~/.wakatime.cfg` automatically. This is the same config file used by the WakaTime editor plugins, so if you already use WakaTime, you're all set.

## Inspiration

Inspired by [wakafetch](https://github.com/sahaj-b/wakafetch) by sahaj-b — a clean fetch-style tool for WakaTime stats.

## License

MIT — see [LICENSE](LICENSE)
