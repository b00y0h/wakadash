# Phase 1: Repository Setup - Research

**Researched:** 2026-02-13
**Domain:** GitHub CLI repository management, Homebrew tap infrastructure
**Confidence:** HIGH

## Summary

Phase 1 establishes GitHub infrastructure for automated Homebrew releases using GitHub CLI (`gh`) for repository operations. The core tasks involve forking an upstream repository, creating a dedicated tap repository with proper directory structure, and configuring fine-grained personal access tokens (PATs) with minimal permissions for GitHub Actions workflows.

GitHub CLI v2.83.2 provides comprehensive commands for repository management (`gh repo fork`, `gh repo create`, `gh repo edit`) and secrets management (`gh secret set`). Homebrew taps follow strict naming conventions (`homebrew-*` prefix) and directory structure (Formula/ for command-line tools, Casks/ for GUI apps). Fine-grained PATs offer repository-scoped permissions with Contents read/write being sufficient for most automation tasks.

**Primary recommendation:** Use GitHub CLI for all repository operations with fine-grained PATs scoped to specific repositories. Disable unused features (wiki, projects) to reduce attack surface and simplify repository management.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
**Tap repository naming:**
- Repository name: `homebrew-wakafetch` (dedicated tap for this tool)
- Install command will be: `brew tap b00y0h/wakafetch`
- Standard tap structure: Casks/ folder, LICENSE, README
- License: MIT (matches upstream wakafetch)

**Fork strategy:**
- Fork evolves independently — no syncing with upstream sahaj-b/wakafetch
- Default branch: `main` for releases
- Disable unused features: wiki, projects (keep issues and actions)

**Token permissions scope:**
- Fine-grained PAT with minimal permissions
- Scope: Contents read/write ONLY on homebrew-wakafetch tap repo
- Expiration: 1 year
- Token name: `HOMEBREW_TAP_TOKEN` (matches Actions secret name)

### Claude's Discretion
- Tap README content (minimal installation instructions vs detailed)
- Fork repository description (custom vs keep upstream)
- Whether to document token setup in fork README

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope.
</user_constraints>

## Standard Stack

### Core Tools
| Tool | Version | Purpose | Why Standard |
|------|---------|---------|--------------|
| gh | 2.83.2+ | GitHub CLI | Official GitHub command-line tool with full API coverage |
| git | 2.x+ | Version control | Universal Git operations for repository management |
| Homebrew | 4.x+ | Package manager | macOS/Linux package distribution ecosystem |

### GitHub CLI Commands
| Command | Purpose | When to Use |
|---------|---------|-------------|
| `gh repo fork` | Fork upstream repository | Initial fork creation with remote configuration |
| `gh repo create` | Create new repository | Tap repository initialization |
| `gh repo edit` | Modify repository settings | Disable features, update metadata |
| `gh secret set` | Configure Actions secrets | Token storage for workflow automation |
| `gh auth status` | Verify authentication | Pre-flight checks before operations |

### Authentication
**Fine-grained Personal Access Token** (recommended over classic PATs):
- Repository-scoped permissions
- Time-limited expiration (30-90 days to 1 year)
- Organizational approval workflows (if applicable)
- Granular permission control

**Installation:**
All tools are pre-installed in ClaudeBox environment. For local development:
```bash
# Verify GitHub CLI
gh --version

# Authenticate with GitHub
gh auth login

# Verify authentication
gh auth status
```

## Architecture Patterns

### Recommended Repository Structure

**Fork Repository (b00y0h/wakafetch):**
```
wakafetch/
├── .github/
│   └── workflows/        # GitHub Actions (keep enabled)
├── src/                  # Source code (from upstream)
├── LICENSE              # Keep upstream license
└── README.md            # Update for fork-specific information
```

**Tap Repository (b00y0h/homebrew-wakafetch):**
```
homebrew-wakafetch/
├── Casks/               # GUI application formulas (required by user)
│   └── wakafetch.rb     # Cask formula (future phase)
├── Formula/             # CLI tool formulas (optional for Casks)
├── .github/
│   └── workflows/       # Formula update automation (future phase)
├── LICENSE              # MIT (matches upstream)
└── README.md            # Installation instructions
```

### Pattern 1: Repository Forking with Independent Development

**What:** Fork creates a completely independent copy with no automatic upstream synchronization.

**When to use:** When creating a fork that will evolve separately from upstream without contributing back.

**Example:**
```bash
# Fork repository without cloning locally
gh repo fork sahaj-b/wakafetch --fork-name wakafetch --remote=false

# Note: Default behavior renames existing 'origin' to 'upstream'
# Use --remote=false to skip local repository setup
```

**Key characteristic:** Fork is independent — no sync with upstream required. Default branch remains `main` (or whatever upstream uses).

### Pattern 2: Tap Repository Creation with Feature Control

**What:** Create new repository with selective feature enablement.

**When to use:** When creating infrastructure repositories that don't need wiki, projects, or discussions.

**Example:**
```bash
# Create public tap repository with minimal features
gh repo create homebrew-wakafetch \
  --public \
  --description "Homebrew tap for wakafetch" \
  --disable-issues=false \
  --disable-wiki=true \
  --license MIT

# Note: Projects cannot be disabled during creation
# Must use gh repo edit after creation
```

**Post-creation feature adjustment:**
```bash
gh repo edit b00y0h/homebrew-wakafetch --enable-projects=false
```

**Known limitation:** `--enable-projects=false` only disables "Projects (classic)". Modern Projects remain enabled with no CLI/API to disable them — requires web UI manual toggle if strict security required.

### Pattern 3: Fine-Grained Token with Repository-Scoped Permissions

**What:** Fine-grained PAT limited to specific repository with minimal permissions.

**When to use:** For GitHub Actions workflows requiring write access to a single repository.

**Token creation (manual web UI required):**
1. Navigate to GitHub Settings → Developer settings → Personal access tokens → Fine-grained tokens
2. Generate new token with:
   - **Token name:** `HOMEBREW_TAP_TOKEN`
   - **Expiration:** 1 year (or 30-90 days for stricter rotation)
   - **Resource owner:** b00y0h (user account)
   - **Repository access:** Only select repositories → homebrew-wakafetch
   - **Permissions:**
     - **Repository permissions:**
       - Contents: **Read and write** (required for file updates)
       - Metadata: **Read-only** (automatically included)
3. Generate and copy token immediately (shown only once)

**Token storage as repository secret:**
```bash
# Set repository secret for GitHub Actions
# Token value from stdin (secure input)
echo "$COPIED_TOKEN_VALUE" | gh secret set HOMEBREW_TAP_TOKEN \
  --repo b00y0h/wakafetch \
  --app actions

# Verify secret was created (lists names only, not values)
gh secret list --repo b00y0h/wakafetch
```

**Security benefits:**
- Scoped to single repository (homebrew-wakafetch only)
- Minimal permissions (Contents write, no admin/settings/secrets access)
- Time-limited expiration (1 year)
- Revocable without affecting other tokens

### Pattern 4: Repository Feature Management

**What:** Post-creation repository configuration to disable unused features.

**When to use:** After creating fork or new repository to minimize attack surface.

**Example:**
```bash
# Disable wiki and projects on fork
gh repo edit b00y0h/wakafetch \
  --enable-wiki=false \
  --enable-projects=false

# Disable wiki on tap repository
gh repo edit b00y0h/homebrew-wakafetch \
  --enable-wiki=false \
  --enable-projects=false

# Note: Keep issues and actions enabled as specified by user
```

### Anti-Patterns to Avoid

- **Using classic PATs:** Less secure, broader scope, no repository-level granularity
- **Excessive token permissions:** Avoid admin, workflows, or secrets permissions when only Contents write needed
- **Long-lived tokens without rotation:** Set expiration, plan for rotation workflow
- **Hardcoding secrets in workflow files:** Always use GitHub Actions secrets, never commit tokens
- **Enabling unnecessary features:** Wiki and projects increase attack surface and maintenance burden

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| GitHub API authentication | Custom OAuth flow | `gh auth login` | Handles device flow, token storage, scope management |
| Repository forking | Manual fork + git remote setup | `gh repo fork` | Configures remotes correctly, handles naming edge cases |
| Secrets encryption | Custom crypto for secret values | `gh secret set` | Client-side encryption before transmission, matches GitHub's public key |
| Token permission scoping | Repository checks in workflow code | Fine-grained PAT repository selection | Platform-enforced at API level, can't be bypassed |
| Feature flag management | API calls to enable/disable features | `gh repo edit --enable-<feature>=false` | Single command vs. multiple API calls, handles deprecations |

**Key insight:** GitHub CLI handles edge cases, API versioning, and authentication complexity. Custom solutions miss non-obvious requirements like client-side secret encryption, device authentication flows, and changing API schemas.

## Common Pitfalls

### Pitfall 1: Token Scope Too Broad
**What goes wrong:** Creating fine-grained PAT with access to all repositories or excessive permissions.

**Why it happens:** Default web UI selections may include "All repositories" or suggest unnecessary permissions.

**How to avoid:**
- Select "Only select repositories" and choose **only** homebrew-wakafetch
- Use minimal permissions: Contents (read/write) and Metadata (read)
- Avoid selecting admin, workflows, or secrets permissions

**Warning signs:**
- Token works on repositories other than homebrew-wakafetch
- Workflows can modify GitHub Actions files or secrets
- Token grants more permissions than required error messages indicate

### Pitfall 2: Projects Feature Not Fully Disabled
**What goes wrong:** Using `gh repo edit --enable-projects=false` but Projects tab still appears.

**Why it happens:** GitHub has two project types: "Projects (classic)" and modern "Projects". CLI only disables classic.

**How to avoid:**
- Accept that CLI cannot disable modern Projects
- If security-critical, manually disable via web UI: Settings → Features → uncheck both Projects options
- Document that Projects tab visibility is acceptable for this use case

**Warning signs:**
- Projects tab visible after running `gh repo edit --enable-projects=false`
- No CLI command successfully hides Projects tab

### Pitfall 3: Secret Value Leakage Through Logs
**What goes wrong:** GitHub Actions logs accidentally print secret values in error messages or debug output.

**Why it happens:**
- Using secrets in command-line arguments (visible via `ps`)
- Printing error messages that include secret context
- Using structured data (JSON/YAML) that prevents GitHub's redaction

**How to avoid:**
- Pass secrets via environment variables, not command-line arguments
- Use `::add-mask::` in workflows to explicitly mark values for redaction
- Avoid embedding secrets in JSON/YAML structures (breaks exact-match redaction)
- Test workflows in fork before using in production

**Warning signs:**
- Error messages showing token fragments
- Debug logs revealing secret values
- Workflow runs displaying unmasked environment variables

### Pitfall 4: Fork Remote Configuration Confusion
**What goes wrong:** After forking, local git remotes point to unexpected repositories or have wrong names.

**Why it happens:** `gh repo fork` default behavior renames existing `origin` to `upstream` when cloning.

**How to avoid:**
- Use `--remote=false` flag if not cloning locally during fork operation
- Use `--clone=false` to skip clone entirely (create fork on GitHub only)
- If cloning: explicitly verify remotes with `git remote -v` after fork

**Warning signs:**
- Pushing to wrong repository (upstream instead of fork)
- Confusion about which remote is "origin" vs "upstream"
- Accidentally contributing to upstream when intending to work on fork

### Pitfall 5: Tap Repository Naming Convention Violation
**What goes wrong:** Creating tap repository without "homebrew-" prefix, breaking `brew tap` shorthand.

**Why it happens:** Unfamiliarity with Homebrew naming conventions.

**How to avoid:**
- Always name tap repositories `homebrew-<name>`
- Verify naming before creation: `b00y0h/homebrew-wakafetch` not `b00y0h/wakafetch-tap`
- Test tap installation: `brew tap b00y0h/wakafetch` (prefix automatically added)

**Warning signs:**
- `brew tap b00y0h/wakafetch` fails with "invalid tap name"
- Must use full URL form: `brew tap b00y0h/wakafetch https://...`

### Pitfall 6: Missing Metadata Permission on Fine-Grained Token
**What goes wrong:** Workflows fail with authentication errors despite Contents write permission.

**Why it happens:** Metadata (read-only) permission is automatically included but may be missed if manually configuring.

**How to avoid:**
- Always include Metadata: Read-only permission with Contents permissions
- Use GitHub's permission helper in token creation UI (suggests required permissions)
- Test token with minimal API call before using in workflow

**Warning signs:**
- 403 Forbidden errors in workflow despite correct token scope
- Error messages mentioning "insufficient permissions" for repository metadata
- Actions logs showing authentication failures on repository info fetch

## Code Examples

Verified patterns from official sources:

### Fork Upstream Repository
```bash
# Fork repository to personal account
# --fork-name: Custom name for fork (defaults to upstream name)
# --remote=false: Don't configure local git remotes (GitHub only)
gh repo fork sahaj-b/wakafetch \
  --fork-name wakafetch \
  --remote=false

# Verify fork created
gh repo view b00y0h/wakafetch
```
**Source:** [GitHub CLI Manual - gh repo fork](https://cli.github.com/manual/gh_repo_fork)

### Create Tap Repository
```bash
# Create public repository with selective features
# --disable-wiki: Disable wiki feature
# --license: Add license file (MIT matches upstream)
gh repo create homebrew-wakafetch \
  --public \
  --description "Homebrew tap for wakafetch - a modern system information tool" \
  --disable-wiki=true \
  --license MIT

# Disable projects feature (post-creation)
gh repo edit homebrew-wakafetch --enable-projects=false
```
**Source:** [GitHub CLI Manual - gh repo create](https://cli.github.com/manual/gh_repo_create)

### Configure Fork Settings
```bash
# Disable unused features on fork
gh repo edit wakafetch \
  --enable-wiki=false \
  --enable-projects=false

# Keep issues and actions enabled (no action needed - enabled by default)
```
**Source:** [GitHub CLI Manual - gh repo edit](https://cli.github.com/manual/gh_repo_edit)

### Store Token as Repository Secret
```bash
# Set repository secret for GitHub Actions
# Token value from environment variable (secure)
echo "$HOMEBREW_TAP_TOKEN" | gh secret set HOMEBREW_TAP_TOKEN \
  --repo b00y0h/wakafetch \
  --app actions

# Alternative: Interactive input (prompts for secret value)
gh secret set HOMEBREW_TAP_TOKEN --repo b00y0h/wakafetch

# Verify secret exists (shows names only, not values)
gh secret list --repo b00y0h/wakafetch
```
**Source:** [GitHub CLI Manual - gh secret set](https://cli.github.com/manual/gh_secret_set)

### Authentication Pre-Flight Check
```bash
# Verify GitHub CLI authentication status
gh auth status

# Expected output shows:
# - Logged in to github.com
# - Authentication scopes (repo, read:org, etc.)
# - Token expiration date

# If not authenticated:
gh auth login --git-protocol https --web
```
**Source:** [GitHub CLI Manual - gh auth](https://cli.github.com/manual/gh_auth_status)

### Initialize Tap Directory Structure
```bash
# Create minimal tap structure locally (after cloning)
# Required for Homebrew tap recognition
mkdir -p Casks
touch LICENSE README.md

# Add MIT license content
cat > LICENSE << 'EOF'
MIT License

Copyright (c) 2026 b00y0h

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
EOF

# Add basic README
cat > README.md << 'EOF'
# Homebrew Tap for wakafetch

## Installation

```bash
brew tap b00y0h/wakafetch
brew install --cask wakafetch
```

## About

This tap provides the wakafetch cask for Homebrew installation.

## License

MIT
EOF
```
**Source:** [Homebrew Documentation - How to Create and Maintain a Tap](https://docs.brew.sh/How-to-Create-and-Maintain-a-Tap)

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Classic PATs with full repo scope | Fine-grained PATs with per-repo permissions | 2022-2023 | Better security, organizational approval, reduced blast radius |
| Manual git + GitHub API | GitHub CLI (`gh`) for operations | Stable since 2020 | Simplified workflows, better error handling, automatic authentication |
| Formula/ directory only | Formula/ (CLI tools) and Casks/ (GUI apps) | Pre-2015 standard | Clear separation of package types, proper installation locations |
| Tap requires Formula/ mandatory | Casks/ only allowed (Formula/ optional) | ~2020 | Taps can provide GUI apps without CLI tool formulas |
| Projects feature toggle (single) | Projects (classic) vs modern Projects | 2023-2024 | CLI can only disable classic, web UI required for modern |

**Deprecated/outdated:**
- **Classic Personal Access Tokens:** Still supported but GitHub recommends fine-grained PATs for new use cases (better security, granular permissions)
- **Homebrew bottle builds in tap repos:** Modern approach uses GitHub Actions automation with pre-built bottles uploaded to releases (this project will implement in later phases)
- **Manual secret encryption:** `gh secret set` handles client-side encryption automatically (older guides show manual crypto operations)

## Open Questions

1. **Token Rotation Workflow**
   - What we know: Fine-grained PATs expire after set period (1 year in user decision)
   - What's unclear: Best practice for rotating token in GitHub Actions without workflow downtime
   - Recommendation: Document token expiration date in project notes, set calendar reminder 1 week before expiration to regenerate and update secret

2. **Projects Feature Web UI Requirement**
   - What we know: `gh repo edit --enable-projects=false` only disables Projects (classic), not modern Projects
   - What's unclear: Whether modern Projects actually increase security risk or just UI clutter
   - Recommendation: Accept that Projects tab will be visible via CLI approach. If security audit requires full disablement, document manual web UI step in runbook

3. **Fork Repository Description**
   - What we know: Marked as "Claude's discretion" in CONTEXT.md
   - What's unclear: Whether to keep upstream description or customize for fork purpose
   - Recommendation: Customize description to: "Personal fork of wakafetch for automated Homebrew releases" (clarifies fork purpose, helps future users understand scope)

4. **Tap README Detail Level**
   - What we know: Marked as "Claude's discretion" in CONTEXT.md
   - What's unclear: Whether to include only installation commands or add troubleshooting, contributing guidelines
   - Recommendation: Start minimal (installation instructions + license + brief description), expand based on user issues in future phases

## Sources

### Primary (HIGH confidence)
- [GitHub CLI Manual - gh repo fork](https://cli.github.com/manual/gh_repo_fork) - Fork command options and behavior
- [GitHub CLI Manual - gh repo create](https://cli.github.com/manual/gh_repo_create) - Repository creation flags
- [GitHub CLI Manual - gh repo edit](https://cli.github.com/manual/gh_repo_edit) - Feature management and settings
- [GitHub CLI Manual - gh secret set](https://cli.github.com/manual/gh_secret_set) - Secret storage commands
- [GitHub Docs - Managing Personal Access Tokens](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens) - Fine-grained PAT setup and permissions
- [GitHub Docs - Permissions Required for Fine-Grained PATs](https://docs.github.com/en/rest/authentication/permissions-required-for-fine-grained-personal-access-tokens) - Permission types and scopes
- [Homebrew Documentation - How to Create and Maintain a Tap](https://docs.brew.sh/How-to-Create-and-Maintain-a-Tap) - Tap structure and requirements
- [Homebrew Documentation - Taps](https://docs.brew.sh/Taps) - Tap naming and conventions

### Secondary (MEDIUM confidence)
- [GitHub CLI Issue #6652 - Projects Disable Limitation](https://github.com/cli/cli/issues/6652) - Known limitation with Projects (classic) vs modern Projects
- [Medium - Fine-Grained GitHub Tokens](https://medium.com/@patrickduch93/fine-grained-github-tokens-dont-expose-all-your-repositories-b49d4c8581c5) - Best practices for token scoping (Jan 2026)
- [Arctiq - GitHub Actions Security Pitfalls](https://arctiq.com/blog/top-10-github-actions-security-pitfalls-the-ultimate-guide-to-bulletproof-workflows) - Common security mistakes and mitigations (2026)
- [OneUpTime - GitHub Actions Secrets Management](https://oneuptime.com/blog/post/2026-01-25-github-actions-manage-secrets/view) - Current best practices (Jan 2026)

### Tertiary (LOW confidence)
- Community discussions on GitHub about tap structure and naming conventions - Reinforces official documentation but not primary source

## Metadata

**Confidence breakdown:**
- Standard stack: **HIGH** - GitHub CLI is official tool with stable API, version verified in environment
- Architecture patterns: **HIGH** - All patterns verified with official documentation and CLI manual pages
- Pitfalls: **MEDIUM** - Based on known issues, official bug reports, and 2025-2026 security guides (some extrapolation of edge cases)
- Token permissions: **HIGH** - Official GitHub documentation and manual testing in environments

**Research date:** 2026-02-13
**Valid until:** 2026-03-15 (30 days - GitHub CLI stable, Homebrew conventions unlikely to change)

**Notes:**
- GitHub CLI v2.83.2 current as of Dec 2025
- Fine-grained PATs feature stable since 2023
- Homebrew tap conventions unchanged since ~2020
- Projects feature limitation (classic vs modern) unlikely to be resolved soon based on GitHub issue tracker
