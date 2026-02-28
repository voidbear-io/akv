---
type: task
tags: ["test"]
---

# Implement Tests

## Context

Define and implement a reliable automated testing strategy for the `akv` CLI, balancing fast feedback for local development with optional integration coverage for Azure-backed behavior.

## Implementation Plan

1. Define the test strategy by layer:
   - fast unit tests for command parsing, validation, and config loading,
   - service tests for Azure Key Vault adapters with mocked SDK clients,
   - integration tests (tagged) for live Azure scenarios in isolated test vaults.
2. Establish test fixtures and helpers for deterministic input/output:
   - temp config files and env var isolation,
   - fake keyring provider,
   - standard error/assertion helpers for CLI output.
3. Add command-level tests for `secrets`, `keys`, and `certificates` paths, including alias behavior (`get` == `secrets get`) and expected exit codes.
4. Add behavior tests for secret sourcing priority (env > file > keyring) and vault/profile resolution logic.
5. Create integration test suite behind `integration` build tag using dedicated Azure test resources and strict cleanup for created/deleted assets.
6. Wire test execution in local and CI workflows:
   - `go test ./...` for unit/service tests,
   - optional `go test -tags=integration ./...` gated by secrets and environment setup.
7. Track quality gates: minimum coverage target for core packages, flaky-test policy, and test matrix by supported Go versions.

## Acceptance Criteria

- Unit and service tests run with `go test ./...` and cover core command/config logic.
- Integration tests exist in separate files and run only with the `integration` build tag.
- Test helpers isolate env/config/keyring state and avoid leaking secrets.
- CI can run standard tests by default and integration tests when gated credentials are present.
