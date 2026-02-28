---
type: "docs"
tags: ["docs"]
---

# Feature Proposals

## Context

Capture a prioritized backlog of product enhancements that improve usability, migration support, operational safety, and team-scale workflows for `akv`.

## Implementation Plan

1. Add a `profile` command group to improve multi-tenant usage:
   - `profile ls/add/rm/use/show`,
   - per-profile default vault and auth method,
   - safe export/import for team bootstrap.
2. Add secret version management commands:
   - list versions,
   - restore previous version,
   - diff metadata/value hashes without printing sensitive data.
3. Add bulk operations with file-driven workflows:
   - `secrets import` from dotenv/json/yaml,
   - `secrets export` with redaction/scope filters,
   - dry-run mode and transactional error reporting.
4. Add policy/compliance helpers for operational safety:
   - expiration and rotation reminders,
   - tag enforcement checks,
   - report mode for missing mandatory metadata.
5. Add shell completion and interactive UX improvements:
   - generated completion scripts (bash/zsh/fish/powershell),
   - optional interactive selector for vault/profile/resource names.
6. Add audit-friendly logging and trace commands:
   - correlation IDs in logs,
   - optional JSON audit output for SIEM ingestion,
   - command to summarize recent operations and failures.
7. Add migration helpers for users coming from `kpv`/`osv`:
   - compatibility aliases,
   - config import command,
   - migration validator with actionable warnings.

## Acceptance Criteria

- Backlog includes clear feature areas with user-facing outcomes and implementation direction.
- Proposals are scoped enough to be split into individual delivery issues.
- Migration and operational safety are represented alongside developer ergonomics.
- Prioritized list can be used directly for alpha and post-alpha planning.
