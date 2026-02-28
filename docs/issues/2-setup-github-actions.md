---
tags: ["chore"]
type: chore
---

# Setup GitHub Actions

## Context

Set up CI/CD automation for validation and release workflows using GitHub Actions and non-premium GoReleaser features. The release process must support cross-platform binaries, package distribution targets, and generated release announcements.

## Implementation Plan

1. Create baseline CI workflow (`.github/workflows/ci.yml`) triggered on PRs and main pushes:
   - checkout, setup Go, cache modules/build cache,
   - run `go fmt`, `go vet`, lint, unit tests,
   - run `govulncheck` as a required quality gate.
2. Add commented coverage publish stage in CI for future Codecov integration, but keep coverage generation artifact output enabled for manual inspection.
3. Add release workflow (`.github/workflows/release.yml`) triggered by semantic tags and manual dispatch:
   - validate changelog/tag,
   - run full test/lint/vuln pipeline,
   - execute GoReleaser in OSS/non-premium mode.
4. Configure `.goreleaser.yaml` for multi-platform binaries:
   - GOOS: linux, darwin, windows,
   - GOARCH: amd64, arm64,
   - archives/checksums/signing configuration as supported.
5. Configure package publishing targets through GoReleaser where feasible:
   - Homebrew tap formula,
   - Chocolatey package metadata,
   - Linux packages (`deb`, `rpm`, `snap`, `flatpak`, `appimage`) using supported non-premium features and/or post-release packaging jobs.
6. Split Docker publishing into an optional workflow that is disabled until Docker Hub credentials/account setup is complete; keep image build job available for local/CI verification.
7. Add release announcement automation:
   - generate GitHub Release notes from changelog/commits,
   - publish a release summary comment/announcement artifact for reuse in project communication channels.
8. Document required GitHub repository secrets, branch protections, and release process in `docs/` so maintainers can operate CI/CD without tribal knowledge.

## Acceptance Criteria

- CI workflow validates formatting, linting, tests, and vulnerability checks on PRs and main.
- Release workflow can publish tagged cross-platform artifacts through GoReleaser OSS mode.
- Coverage publishing step is present but disabled/commented until Codecov is configured.
- Required secrets and release operator steps are documented for maintainers.
