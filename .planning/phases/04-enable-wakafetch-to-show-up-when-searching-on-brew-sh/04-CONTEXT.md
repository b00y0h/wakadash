# Phase 4: Repository Setup - Context

**Gathered:** 2026-02-18
**Status:** Ready for planning

<domain>
## Phase Boundary

Fresh standalone repository with ported wakafetch code and working release automation. Users can clone, build, and run `wakadash --help`. Version tags trigger GitHub Actions to produce release artifacts.

</domain>

<decisions>
## Implementation Decisions

### Repository structure
- Single-binary repository — just wakadash, no multi-tool setup
- Go module: `github.com/b00y0h/wakadash`
- Standard documentation set: README.md, LICENSE, CONTRIBUTING.md, CHANGELOG.md

### Code porting strategy
- Clean rewrite of WakaTime API client — don't just copy wakafetch code
- Dashboard-only scope — only port what the TUI needs, no extra features
- Config: Read directly from `~/.wakatime.cfg` (WakaTime's standard location)
- No wakafetch config migration needed — wakadash doesn't have its own config

### CLI interface design
- Default action: Launch dashboard immediately (no subcommand required)
- No one-shot mode — wakadash is purely a TUI dashboard tool
- Version flag: `--version` shows version + git commit + build date + Go version

### Claude's Discretion
- Directory layout (flat vs cmd/internal/pkg structure)
- Package naming (generic vs branded)
- Help text verbosity

</decisions>

<specifics>
## Specific Ideas

- "Like htop for your coding activity" — dashboard launches immediately on run
- Uses WakaTime's standard config location, no additional config files
- Focused scope — lean codebase that does one thing well

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 04-repository-setup*
*Context gathered: 2026-02-18*
