# Agents

## ✅ Always DO

- When you start a session read mise.toml files to know which apps, runtimes and tools
  should be executed with mise.
- Write tests for any code changes if they do not exist.
- Run tests before completing the task.
- Fix any broken tests
- Keep the project structure section up to date. Focus only on directories
  and key files such as go.mod, castfile, etc. 
- If you write scripts that need to stay around, save them to eng/scripts.
- If you write scripts that are temporarily, save them to eng/tmp.
- For ci/cd, store artifacts in the .artifacts folder.  
- Store product requirement documents in the docs/prd folder. 
- Use the gh cli to interact with github and github.com/frostyeti projects.
- Create seperate files for integration tests and use the `// +build integration` directive
  and use the `go test -tags=integration ./...` command to run them
- Use testcontainers for go if they are needed for external dependencies.
- Store e2e tests in the `test/e2e` folder and write the tests so that they build cast
  generate files, directories, etc as needed and then run tests, and remove files/and folders
  when done.
- always avoid dependencies that require CGO.
- run govulncheck and fix any vulnerabilities before completing the task.
- use go mod tidy to keep the go.mod and go.sum files clean and up to date.
- use go fmt to keep the code formatted and consistent.
- use go lint to keep the code clean and idiomatic.
- use go doc comments to document any exported functions, types, etc.
- use go generate to generate any code that can be generated, such as mocks, etc.
- use go releaser to automate releases and versioning.

## Versioning

This project is currently in **alpha** stage. Unless explicitly instructed otherwise, when creating releases:
- Only bump the prerelease version (e.g., `-alpha.0` → `-alpha.1`)
- Do **NOT** increment major, minor, or patch versions (e.g., `0.1.0` → `0.1.0-alpha.1`, NOT `0.2.0-alpha.0`)
- Wait for explicit instruction to bump major, minor, or patch versions

## Project Structure


```text
.
|-- .github/
|   `-- workflows/
|       |-- ci.yml
|       `-- release.yml
|-- .goreleaser.yaml
|-- AGENTS.md
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
|-- cmd/
|   |-- certificates.go
|   |-- certificates_test.go
|   |-- config.go
|   |-- errors.go
|   |-- flag_parsing.go
|   |-- keys.go
|   |-- keys_test.go
|   |-- ls.go
|   |-- root_test.go
|   |-- root.go
|   |-- secrets.go
|   |-- secrets_test.go
|   `-- vault.go
`-- internal/
    |-- auth/
    |   `-- credential.go
    |-- config/
    |   `-- manager.go
    `-- keyvault/
        |-- certificates_service.go
        |-- certificates_service_test.go
        |-- keys_service.go
        |-- keys_service_test.go
        |-- secrets_service.go
        `-- secrets_service_test.go
```
