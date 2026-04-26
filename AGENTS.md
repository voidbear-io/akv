# Agents

## Always DO

- When you start a session read `mise.toml` files to know which apps, runtimes and tools should be executed with mise.
- Write tests for any code changes if they do not exist.
- Run tests before completing the task.
- Fix any broken tests.
- **Before committing:** run `golangci-lint run ./...`, fix all reported issues, then run `go test ./...` (or the full test command used in CI) and fix any failing tests. Do not commit with lint or test failures.
- Keep the project structure section up to date. Focus only on directories and key files such as `go.mod`, `castfile`, etc.
- If you write scripts that need to stay around, save them to `eng/scripts`.
- If you write scripts that are temporary, save them to `eng/tmp`.
- For ci/cd, store artifacts in the `.artifacts` folder.
- Store product requirement documents in the `docs/prd` folder.
- Use the `gh` CLI to interact with GitHub and [github.com/voidbear-io](https://github.com/voidbear-io) projects.
- Create separate files for integration tests and use the `//go:build integration` directive and use the `go test -tags=integration ./...` command to run them.
- Use testcontainers for Go if they are needed for external dependencies.
- Store e2e tests in the `test/e2e` folder and write the tests so that they build, generate files, directories, etc. as needed and then run tests, and remove files and folders when done.
- Always avoid dependencies that require CGO.
- Run `govulncheck` and fix any vulnerabilities before completing the task.
- Use `go mod tidy` to keep the `go.mod` and `go.sum` files clean and up to date.
- Use `go fmt` to keep the code formatted and consistent.
- Use `golangci-lint` (see CI) to keep the code clean and idiomatic.
- Use Go doc comments to document any exported functions, types, etc.
- Use `go generate` to generate any code that can be generated, such as mocks, etc.
- Use GoReleaser to automate releases and versioning.

## Git commit messages

Standardize on [Conventional Commits](https://www.conventionalcommits.org/) as summarized in
[Standardizing Git Commit Messages: tags like `wip` and beyond](https://dev.to/robinncode/standardizing-git-commit-messages-the-role-of-tags-like-wip-and-beyond-lpe).

- **Format:** `type(optional-scope): short description` (imperative mood, no trailing period in the subject).
- **Types:** `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`, `perf`, `ci` (and similar from the article).
- **`wip`:** acceptable only for local or short-lived branches; do not merge `wip` commits to `main`—squash or reword them first.
- Prefer consistency and machine-readable history; use hooks or CI to validate messages when available.

## Versioning

The CLI **source** default version is **0.0.0-alpha.0** in `cmd/version.go` (release builds may override via GoReleaser `ldflags`).

This project is in **alpha**. Unless explicitly instructed otherwise, when creating releases:

- Only bump the prerelease segment (e.g. `0.0.0-alpha.0` → `0.0.0-alpha.1`).
- Do **not** increment major, minor, or patch without explicit instruction.
- Wait for explicit instruction to bump major, minor, or patch version numbers.

`github.com/voidbear-io/*` **module** dependencies should use version **0.0.0-alpha.0** (or newer prerelease tags that match project policy) unless instructed otherwise.

## Project Structure

```text
.
|-- .github/
|   `-- workflows/
|       |-- ci.yml
|       `-- release.yml
|-- .vscode/
|   `-- settings.json   # castfile YAML schema (Cast)
|-- .goreleaser.yaml
|-- AGENTS.md
|-- castfile            # Cast tasks: build, test, lint, format, vuln, test-e2e
|-- docs/
|   `-- issues/
|       |-- 0-setup-project.md
|       |-- 1-implement.tests.md
|       |-- 2-setup-github-actions.md
|       |-- 3-create-docs.md
|       `-- 4-features.md
|-- go.mod
|-- go.sum
|-- LICENSE
|-- main.go
|-- mise.toml
|-- README.md
|-- cmd/             # Cobra CLI: root, secrets*, keys*, certificates*, vault, config, use, upgrade, version, ls, etc.
`-- internal/
    |-- auth/
    |   `-- credential.go
    |-- config/
    |   `-- manager.go
    `-- keyvault/    # Azure Key Vault service adapters and tests
```
