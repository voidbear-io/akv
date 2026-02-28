---
type: task
tags: ["task"]
---

# Setup Project

## Context

Build an `akv` CLI that mirrors the developer experience of `kpv` and `osv`, but targets Azure Key Vault secrets, keys, and certificates. Root commands such as `get`, `set`, `rm`, and `ensure` should behave as aliases for `secrets` subcommands. Configuration should support multiple known vaults, profile switching, and secure credential input from env vars, files, or keyring.

## Implementation Plan

1. Review `kpv` and `osv` command trees and UX patterns, then lock a matching command surface for `akv` (`get`, `set`, `rm`, `ensure` as aliases for `secrets` subcommands).
2. Bootstrap a Go CLI project (`cmd/akv`, `internal/...`) using Cobra, and add shared command wiring for `secrets`, `keys`, and `certificates` with consistent output and exit codes.
3. Implement configuration management with Viper:
   - config file discovery and profile support,
   - well-known vault aliases,
   - default vault and credential selection,
   - secure handling of sensitive config fields.
4. Implement Azure authentication strategies in priority order:
   - environment variables,
   - secret file input,
   - OS keyring lookup,
   - fallback to Azure SDK default credentials when applicable.
5. Implement `secrets` operations against Azure Key Vault (get/set/rm/ensure), including metadata support, idempotency behavior, and user-friendly error mapping for common Azure failures.
6. Implement Azure-native command groups:
   - `keys` CRUD + backup/restore where supported,
   - `certificates` CRUD + import/export where supported,
   - `purge` commands for soft-deleted resources when vault policy permits.
7. Add cross-cutting concerns: structured logging levels, output formats (table/json), retry/backoff for transient API failures, and clear auth/vault troubleshooting messages.
8. Write a README covering install/build, auth setup, config examples, command cookbook, and migration notes for users familiar with `kpv`/`osv`.
9. Add delivery checks before merge: unit tests for command/config layers, integration tests for Azure SDK adapters (with mocks/fakes), linting, vulnerability scan, and release build verification.

## Acceptance Criteria

- `akv` provides `secrets`, `keys`, and `certificates` command groups, plus root aliases for secret operations.
- Users can choose and persist vault/auth profiles using config-backed defaults.
- Credentials can be loaded via env vars, file input, and OS keyring without exposing sensitive output.
- README documents setup, authentication, common workflows, and troubleshooting.
