# Domain Pitfalls

**Domain:** GoReleaser + Homebrew Tap for Go CLI Distribution
**Researched:** 2026-02-13
**Confidence:** MEDIUM

## Critical Pitfalls

### Pitfall 1: Wrong GitHub Token for Cross-Repository Publishing

**What goes wrong:**
GoReleaser fails with `404 Not Found` when trying to push formula to the tap repository: `PUT https://api.github.com/repos/user/homebrew-tap/contents/Formula/app.rb: 404 Not Found`. The release succeeds but nothing is committed to the Homebrew tap.

**Why it happens:**
The default `GITHUB_TOKEN` provided by GitHub Actions only has permissions for the repository where the workflow runs. Cross-repository publishing requires a separate Personal Access Token (PAT) with `repo` or `contents: write` scope for the tap repository.

**How to avoid:**
1. Create a separate PAT with `repo` scope for the tap repository
2. Add it as a secret (e.g., `HOMEBREW_TOKEN`)
3. Pass it to GoReleaser in the workflow:
   ```yaml
   env:
     HOMEBREW_TOKEN: ${{ secrets.HOMEBREW_TOKEN }}
     GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
   ```
4. Optionally specify it in `.goreleaser.yaml`:
   ```yaml
   repository:
     token: "{{ .Env.HOMEBREW_TOKEN }}"
   ```

**Warning signs:**
- 404 errors during Homebrew tap publish step
- Release completes but tap repository has no new commits
- GitHub Actions logs show "failed to publish artifacts" for Homebrew

**Phase to address:**
Phase 1: Initial Setup - Token configuration must be correct from the start to test the full release pipeline.

---

### Pitfall 2: Personal Access Token Security Risk

**What goes wrong:**
Using a personal user's PAT in CI/CD exposes all of that user's repositories to any malicious CI/CD job. If the PAT is leaked or the repository is compromised, attackers gain access to every repository the user can access, not just the project repos.

**Why it happens:**
PATs are user-scoped, not repository-scoped. They grant access to all repos of the user who created them. Developers often use their personal accounts without realizing the security implications.

**How to avoid:**
1. Create a dedicated "bot" GitHub account for automation
2. Generate a PAT from the bot account
3. Give the bot account minimal permissions - only push access to the tap repository
4. Rotate the PAT periodically
5. Consider GitHub Apps with fine-grained permissions as an alternative (when supported by GoReleaser)

**Warning signs:**
- PAT created from a personal developer account
- PAT with broad scopes (all repos, admin access, etc.)
- No documentation of who owns the PAT or how to rotate it
- PAT stored in plain text or weakly encrypted

**Phase to address:**
Phase 1: Initial Setup - Security decisions made early are hard to change. Create bot account and proper token management from the beginning.

---

### Pitfall 3: Formulas vs Casks Confusion (Deprecated Pattern)

**What goes wrong:**
Using the deprecated `brews` configuration section instead of `homebrew_casks`. The GoReleaser formula itself was disabled on 2025-06-14 "because the cask should be used now instead". Projects continuing to use formulas for pre-compiled binaries violate Homebrew semantics and face deprecation.

**Why it happens:**
Legacy documentation and tutorials still reference the `brews` section. In Homebrew terminology, a "formula" builds from source while a "cask" is pre-compiled. GoReleaser historically created formulas for pre-compiled binaries, which was technically incorrect.

**How to avoid:**
1. Use `homebrew_casks` instead of `brews` in `.goreleaser.yaml`
2. Update tap structure to use Casks/ directory instead of Formula/
3. Understand that since Homebrew v4.0, casks are supported on Linux, making this the correct approach for all platforms
4. Run `goreleaser check` to detect deprecated configuration

**Warning signs:**
- Configuration uses `brews:` section
- Tap repository has Formula/ directory for pre-compiled binaries
- Deprecation warnings when running `goreleaser check`
- Tutorial or example from before mid-2024

**Phase to address:**
Phase 1: Initial Setup - Start with the correct pattern. Migration later is possible but adds unnecessary work.

---

### Pitfall 4: Pre-release Version Overwriting Production Formula

**What goes wrong:**
When tagging pre-release versions (e.g., `v1.0.0-rc1`), GoReleaser updates the main formula in the tap, overwriting the stable production version. Users who run `brew upgrade` get the pre-release version unexpectedly.

**Why it happens:**
By default, GoReleaser publishes every release to the tap. Without `skip_upload: auto`, all versions including pre-releases update the same formula file, making pre-releases public and replacing stable versions.

**How to avoid:**
1. Set `skip_upload: auto` in the homebrew/cask configuration:
   ```yaml
   homebrew_casks:
     - skip_upload: auto
   ```
2. This automatically skips publishing when the tag contains pre-release indicators (rc, beta, alpha, etc.)
3. Alternatively, use separate taps for stable vs pre-release versions
4. For versioned formulas, use `name_template` with version suffixes

**Warning signs:**
- Pre-release versions appearing in the public tap
- User reports of unstable versions after `brew upgrade`
- Formula being updated for every tag including RCs
- No `skip_upload` configuration in `.goreleaser.yaml`

**Phase to address:**
Phase 2: Release Automation - Before creating first pre-release tag. Must be configured before cutting v1.0.0-rc1.

---

### Pitfall 5: Multi-Architecture Archive Conflicts

**What goes wrong:**
Error: "One tap can handle only one archive of an OS/Arch combination. Consider using ids in the brew section." GoReleaser builds both arm64 and amd64 binaries but can't determine which to include in the formula.

**Why it happens:**
GoReleaser builds multiple archives for the same OS (darwin/amd64 and darwin/arm64) but Homebrew formulas/casks expect a single binary per platform. Without explicit configuration, GoReleaser doesn't know which architecture to use.

**How to avoid:**
1. Use universal binaries for macOS when appropriate
2. Or use `ids` in the cask configuration to specify which builds to include:
   ```yaml
   homebrew_casks:
     - ids:
         - darwin_amd64
         - darwin_arm64
   ```
3. Configure architecture-specific URLs and checksums
4. Modern approach: Homebrew handles multiple architectures automatically if properly configured

**Warning signs:**
- Error message about "one archive of an OS/Arch combination"
- Multiple darwin archives in dist/ directory
- GoReleaser release fails at Homebrew publish step
- Confusion about which binary users will get on M1/M2 Macs

**Phase to address:**
Phase 1: Initial Setup - Architecture decisions impact build configuration from the start.

---

### Pitfall 6: Missing `fetch-depth: 0` in GitHub Actions Checkout

**What goes wrong:**
GoReleaser fails to generate changelogs or properly detect version information. The release may fail entirely or produce incomplete metadata. Errors like "failed to generate changelog" or missing version tags.

**Why it happens:**
GitHub Actions checkout step defaults to shallow clone (fetch-depth: 1), which only fetches the latest commit. GoReleaser needs full Git history to generate changelogs and understand semantic versioning.

**How to avoid:**
Always include `fetch-depth: 0` in checkout step:
```yaml
- name: Checkout
  uses: actions/checkout@v4
  with:
    fetch-depth: 0
```

**Warning signs:**
- Missing changelogs in releases
- Version detection errors
- Git history-related failures in GoReleaser
- "shallow clone" warnings

**Phase to address:**
Phase 1: Initial Setup - Required for any GoReleaser GitHub Actions workflow.

---

### Pitfall 7: Deprecated Configuration Fields Not Detected Until Runtime

**What goes wrong:**
Configuration uses deprecated fields like `tap` instead of `repository`, or `plist` instead of `service`. GoReleaser may silently ignore the configuration or fail with obscure YAML parsing errors. The homebrew section may be missing entirely from generated output.

**Why it happens:**
GoReleaser evolves quickly. Fields deprecated in one version are removed in later versions. The configuration file may work today but break on the next GoReleaser upgrade.

**How to avoid:**
1. Run `goreleaser check` regularly to detect deprecated fields
2. Review deprecation notices at https://goreleaser.com/deprecations/
3. Key deprecations to watch:
   - `tap` → `repository` (deprecated v1.19.0, removed v2.0)
   - `plist` → `service` (deprecated by Homebrew)
   - `brews` → `homebrew_casks` (deprecated v2.10)
4. Subscribe to GoReleaser release notes
5. Test configuration with `goreleaser release --snapshot --clean` locally

**Warning signs:**
- YAML unmarshal errors mentioning field names
- Warnings in `goreleaser check` output
- Homebrew configuration being silently ignored
- Empty dist/config.yml missing expected sections
- Using examples from tutorials older than 1 year

**Phase to address:**
Phase 2: Release Automation - Check during initial configuration and establish process for ongoing validation.

---

## Moderate Pitfalls

### Pitfall 8: Insufficient GitHub Actions Workflow Permissions

**What goes wrong:**
GitHub Actions workflow fails with permission errors: "Resource not accessible by integration" or fails to create releases, upload assets, or close milestones.

**Why it happens:**
GitHub Actions has moved to more restrictive default permissions. Workflows need explicit permission grants for specific operations.

**How to avoid:**
Configure required permissions in workflow:
```yaml
permissions:
  contents: write    # Required for releases and Homebrew
  packages: write    # If pushing Docker images
  issues: write      # If closing milestones
  id-token: write    # If using Cosign with OIDC
```

**Warning signs:**
- "Resource not accessible" errors in GitHub Actions
- Workflow has no `permissions:` block
- Release created but assets not uploaded
- Milestones not being closed automatically

**Phase to address:**
Phase 1: Initial Setup - Configure when setting up GitHub Actions workflow.

---

### Pitfall 9: Formula Naming Conventions Violations

**What goes wrong:**
Homebrew audit failures or formula rejections due to incorrect naming. Class names don't match file names. Formulas with spaces in paths fail to install.

**Why it happens:**
Homebrew has strict naming conventions: filenames must be lowercase, class names must be CamelCase equivalents. Package names should match how the project markets itself, not variations.

**How to avoid:**
1. Name formula like the project markets it: `pkgconf` not `pkgconfig`, `sdl_mixer` not `sdl-mixer`
2. Filenames: all lowercase (gnu-go.rb, sdl_mixer.rb)
3. Class names: strict CamelCase (GnuGo, SdlMixer)
4. No spaces in directory paths (causes build script failures)
5. Run `brew audit --strict --online` before publishing
6. For new formulas: `brew audit --new-formula <name>`

**Warning signs:**
- `brew audit` failures
- "Class name doesn't match file name" errors
- Formula installations failing on certain systems
- Formula naming different from GitHub repository name

**Phase to address:**
Phase 1: Initial Setup - Choose correct name from the start; renaming later requires coordinated changes.

---

### Pitfall 10: Test Block Missing or Inadequate

**What goes wrong:**
Formula published without proper installation verification. Users report installation failures or binaries that don't work as expected. Formula fails Homebrew audit for missing tests.

**Why it happens:**
Developers focus on the install block but forget that Homebrew requires a test block to verify successful installation. Basic "assert version" tests are common but insufficient.

**How to avoid:**
1. Always include a test block in formula/cask
2. Test that the binary exists and is executable
3. Test that `--version` or `--help` works
4. Test core functionality if possible:
   ```ruby
   test do
     system "#{bin}/wakafetch", "--version"
     assert_match "wakafetch", shell_output("#{bin}/wakafetch --help")
   end
   ```
5. Test on both Intel and ARM Macs before publishing
6. Use `brew install --build-from-source` to test locally

**Warning signs:**
- No test block in generated formula
- Test only checks `assert true`
- Formula installs but binary doesn't run
- Different behavior between `brew install` and manual installation

**Phase to address:**
Phase 2: Release Automation - Add testing as part of GoReleaser configuration validation.

---

### Pitfall 11: Version String Handling in Formulas

**What goes wrong:**
Versioned formulas have incorrect naming: `wakafetch@1.2.3.rb` instead of `[email protected]`. Formula class names don't follow the `AT` convention. Users can't install specific major/minor versions.

**Why it happens:**
Homebrew's versioning conventions are non-obvious. The `@` symbol in filenames must be translated to `AT` in class names. Homebrew versions differ in major/minor, not patch versions (security policy).

**How to avoid:**
1. Versioned formulas should differ in major/minor only: `[email protected]`, not `wakafetch@1.2.3`
2. File naming: `wakafetch@1.2.rb`
3. Class naming: `WakafetchAT12` (not `WakafetchAt12` or `Wakafetch@12`)
4. Use GoReleaser's `name_template` for versioned formulas
5. Understand Homebrew wants users to get security updates (patch versions)
6. Tag both versioned and unversioned: `foo.rb` + `[email protected]`

**Warning signs:**
- Formula filenames include patch version numbers
- Class name uses `@` symbol or lowercase `at`
- Users report can't install specific versions
- `brew audit` warnings about versioning

**Phase to address:**
Phase 3: Versioned Formula Support - Only needed if supporting multiple major versions simultaneously.

---

### Pitfall 12: Configuration File Path Handling

**What goes wrong:**
After reinstalling formula, user configuration files are replaced with defaults. User loses their settings. Conflicts between different formulas' config files.

**Why it happens:**
Using `prefix.install` to place config files in the main install directory means reinstalls overwrite them. Config files need special handling to persist across installations.

**How to avoid:**
1. Use Homebrew's `etc` directory for configuration:
   ```ruby
   etc.install "config.yaml" => "wakafetch/config.yaml"
   ```
2. Ensure config files are uniquely named (include app name)
3. Don't overwrite existing config files in install block
4. Document config file location in formula description
5. Consider using `etc.install` only if config doesn't exist

**Warning signs:**
- User complaints about lost configuration
- Config files in `{prefix}/` or `bin/`
- Multiple formulas sharing generic config names
- No mention of config directory in formula

**Phase to address:**
Phase 2: Release Automation - If application uses configuration files. Otherwise skip.

---

## Technical Debt Patterns

Shortcuts that seem reasonable but create long-term problems.

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Using personal PAT instead of bot account | Faster initial setup, no new account needed | Security risk, harder to rotate, tied to individual | Never - always use bot account |
| Skipping `brew audit` during development | Faster iteration | Formula rejected later, users get broken installs | Only during initial local testing |
| Using `brews` instead of `homebrew_casks` | Works with old tutorials | Deprecated pattern, migration needed | Never - start with casks |
| No test block in formula | Simpler configuration | Can't verify installation success, audit failures | Never - tests are required |
| Single architecture (amd64 only) | Simpler build configuration | M1/M2 users can't install | Only if targeting Linux-only or legacy systems |
| Hardcoded version strings | No templating syntax needed | Manual updates required, prone to errors | Never - use GoReleaser templates |
| Not using `skip_upload: auto` | All versions published | Pre-releases pollute tap, confuse users | During initial MVP if no pre-releases planned |

## Integration Gotchas

Common mistakes when connecting to external services.

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| GitHub Actions | Using default `GITHUB_TOKEN` for cross-repo | Create separate PAT with `repo` scope |
| GoReleaser | Not running `goreleaser check` before release | Run check in CI and pre-commit |
| Homebrew Audit | Publishing without local audit | Run `brew audit --strict --online` locally |
| Multiple Repos | Committing formula to wrong branch | Configure default branch in GoReleaser config |
| Semantic Versioning | Tags without `v` prefix | Use `v1.2.3` format (GoReleaser default) |
| Token Environment | Typo in env var name (e.g., `HOMEBREW_TOKEN` vs expected name) | Verify exact variable names in GoReleaser config |

## Performance Traps

Patterns that work at small scale but fail as usage grows.

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Large binary size | Slow installation, user complaints | Strip binaries, use compression | >100MB uncompressed |
| No architecture-specific builds | Rosetta translation overhead on M1/M2 | Build native arm64 binaries | Users notice slow startup |
| Downloading from slow/unreliable URLs | Installation timeouts, failures | Use GitHub releases (fast CDN) | >10k installations/month |
| Formula in distant tap | `brew update` takes long | Use canonical tap location | >50 formulas in tap |

## Security Mistakes

Domain-specific security issues beyond general web security.

| Mistake | Risk | Prevention |
|---------|------|------------|
| Personal PAT in repository secrets | Leak exposes all user repos | Use bot account with minimal permissions |
| PAT with excessive scopes | Unnecessary attack surface | Only grant `contents: write` on tap repo |
| No PAT rotation policy | Compromised tokens stay valid | Rotate every 90 days, document in runbook |
| Unsigned macOS binaries | Gatekeeper blocks, users bypass security | Sign and notarize (requires Apple Developer account) |
| Using `xattr` to bypass Gatekeeper in docs | Users disable security protections | Properly sign binaries instead |
| Formula downloads over HTTP | MITM attacks possible | Always use HTTPS URLs |
| No checksum verification | Modified binaries installed | GoReleaser handles this - verify it's enabled |

## UX Pitfalls

Common user experience mistakes in this domain.

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| Pre-release in main tap | Users get unstable versions | Use `skip_upload: auto` |
| No `--version` flag | Can't verify installation | Always implement `--version` |
| Missing or unclear test block | Users don't know if install worked | Test actual functionality, not just binary existence |
| No install verification | Silent failures, confusion | Test block should exit non-zero on failure |
| Formula description missing | Users don't know what they're installing | Write clear description in GoReleaser config |
| No homepage URL | Users can't find documentation | Include `homepage` in config |
| Breaking changes without version bump | Unexpected behavior after upgrade | Follow semantic versioning strictly |

## "Looks Done But Isn't" Checklist

Things that appear complete but are missing critical pieces.

- [ ] **Cross-repo publishing:** Often missing separate HOMEBREW_TOKEN — verify with actual release
- [ ] **Audit passing:** Often untested — run `brew audit --strict` before first release
- [ ] **ARM64 support:** Often missing for macOS — test on M1/M2 Mac
- [ ] **Test block:** Often too simple (just checks binary exists) — verify functional test
- [ ] **Pre-release handling:** Often missing `skip_upload: auto` — verify RC doesn't overwrite stable
- [ ] **Workflow permissions:** Often missing required scopes — check `permissions:` block
- [ ] **Fetch depth:** Often missing `fetch-depth: 0` — verify changelog generation works
- [ ] **Config deprecations:** Often using old fields — run `goreleaser check`
- [ ] **Bot account:** Often using personal PAT — verify dedicated account exists
- [ ] **Formula name:** Often doesn't match project marketing — verify with team

## Recovery Strategies

When pitfalls occur despite prevention, how to recover.

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Wrong token used | LOW | Add correct token to secrets, re-run workflow |
| Personal PAT exposed | MEDIUM | Create bot account, rotate tokens, update secrets, re-run |
| Formula naming wrong | MEDIUM | Rename file/class, deprecate old name, update docs, create symlink |
| Pre-release overwrote stable | LOW | Tag stable version, re-release, communicate to users |
| Missing architecture | MEDIUM | Update build config, create new release with new tag |
| Deprecated config used | LOW | Update .goreleaser.yaml, run check, re-release |
| No test block | LOW | Add test block, push to tap repo, no new release needed |
| Wrong permissions | LOW | Update workflow permissions, re-run workflow |
| Shallow clone | LOW | Add fetch-depth: 0, re-run workflow |
| Formula in wrong tap | HIGH | Create correct tap, deprecate old tap, communicate to users, wait for migration |
| Unsigned binary causing Gatekeeper issues | HIGH | Get Apple Developer account ($99/year), configure signing, re-release |
| Config file path wrong | MEDIUM | Update install block, new release, document migration for users |

## Pitfall-to-Phase Mapping

How roadmap phases should address these pitfalls.

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| Wrong GitHub token | Phase 1: Initial Setup | Test release to tap repo succeeds |
| Personal PAT security risk | Phase 1: Initial Setup | Verify bot account exists and owns PAT |
| Formulas vs Casks confusion | Phase 1: Initial Setup | Verify `homebrew_casks` in config, Casks/ directory in tap |
| Pre-release overwriting | Phase 2: Release Automation | Tag RC version, verify stable formula unchanged |
| Multi-arch conflicts | Phase 1: Initial Setup | Build for arm64 and amd64, verify both in dist/ |
| Missing fetch-depth | Phase 1: Initial Setup | Verify changelog generated correctly |
| Deprecated config fields | Phase 2: Release Automation | `goreleaser check` passes with no warnings |
| Insufficient permissions | Phase 1: Initial Setup | All workflow steps succeed including milestone close |
| Formula naming violations | Phase 1: Initial Setup | `brew audit --strict` passes |
| Missing test block | Phase 2: Release Automation | Test block executes during `brew install --build-from-source` |
| Version string handling | Phase 3: Versioned Support | Install both `wakafetch` and `[email protected]` successfully |
| Config file handling | Phase 2: Release Automation | Reinstall formula, verify config preserved |

## Sources

### Official Documentation
- [GoReleaser Homebrew Taps](https://goreleaser.com/customization/homebrew/)
- [GoReleaser Homebrew Casks](https://goreleaser.com/customization/homebrew_casks/)
- [GoReleaser GitHub Actions](https://goreleaser.com/ci/actions/)
- [GoReleaser Deprecation Notices](https://goreleaser.com/deprecations/)
- [Homebrew Formula Cookbook](https://docs.brew.sh/Formula-Cookbook)
- [Homebrew Versions Documentation](https://docs.brew.sh/Versions)

### GitHub Issues & Discussions
- [Personal Access Token Security Warning - goreleaser/goreleaser#2026](https://github.com/goreleaser/goreleaser/issues/2026)
- [Homebrew Tokens in GitHub - goreleaser Discussion #4926](https://github.com/orgs/goreleaser/discussions/4926)
- [Brew packages should be casks - goreleaser Discussion #5563](https://github.com/orgs/goreleaser/discussions/5563)
- [Brew updating with PR enabled fails - goreleaser/goreleaser#4283](https://github.com/goreleaser/goreleaser/issues/4283)
- [CI broken due to 404 from Homebrew repo - goreleaser/goreleaser#4634](https://github.com/goreleaser/goreleaser/issues/4634)
- [Brew Formula Class Name with Version - goreleaser/goreleaser#3116](https://github.com/goreleaser/goreleaser/issues/3116)
- [Brew audit expectations - Homebrew Discussion #6138](https://github.com/orgs/Homebrew/discussions/6138)

### Tutorials & Blog Posts
- [How to release to Homebrew with GoReleaser, GitHub Actions and Semantic Release - Billy Hadlow](https://billyhadlow.com/blog/how-to-release-to-homebrew/)
- [Creating Homebrew Formulas with GoReleaser - Bindplane](https://bindplane.com/blog/creating-homebrew-formulas-with-goreleaser)
- [Creating Homebrew Formulas With GoReleaser - DZone](https://dzone.com/articles/creating-homebrew-formulas-with-goreleaser)
- [Homebrew Security Best Practices - guessi's blog](https://guessi.github.io/posts/2025/homeberw-tips-security/)
- [Security and the Homebrew contribution model - Workbrew Blog](https://workbrew.com/blog/security-and-the-homebrew-contribution-model)

### Source Code
- [goreleaser/internal/pipe/brew/brew.go](https://github.com/goreleaser/goreleaser/blob/main/internal/pipe/brew/brew.go)

---

*Research confidence: MEDIUM - Based on official documentation, multiple GitHub issues, and community resources. Some findings from web search require validation against current GoReleaser versions.*
