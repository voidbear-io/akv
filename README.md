# akv

**akv** is a small, fast command-line tool for [Azure Key Vault](https://azure.microsoft.com/products/key-vault/): work with **secrets**, **keys**, and **certificates** from your terminal, with shell-friendly output and simple vault switching.

- **Project:** [github.com/voidbear-io/akv](https://github.com/voidbear-io/akv)  
- **By:** [Mr VoidBear](https://github.com/voidbear-io) / [voidbear-io](https://github.com/voidbear-io) on GitHub

---

## Quick start

```bash
# One-line install (Linux/macOS) — adjust branch/path if you use a fork
curl -fsSL https://raw.githubusercontent.com/voidbear-io/akv/master/eng/script/install.sh | bash

akv vault add my-vault
akv use my-vault
akv secrets ensure db-password --size 32
akv secrets get db-password
# Use another vault for a single command:
akv --vault my-other-vault get prod-db-password
```

**Windows (PowerShell):**

```powershell
irm https://raw.githubusercontent.com/voidbear-io/akv/master/eng/script/install.ps1 | iex
```

If execution is blocked: `Set-ExecutionPolicy Process Bypass -Force` for the current process only.

---

## Installation

| Method | Command / notes |
|--------|------------------|
| **Install script** | See Quick start (`eng/script/install.sh` or `install.ps1`). Set `AKV_INSTALL_DIR` to change the install location. |
| **GitHub Releases** | Download the archive or binary for your OS/arch from the [Releases](https://github.com/voidbear-io/akv/releases) page. |
| **From source** | `go install github.com/voidbear-io/akv@latest` (requires Go as in `mise.toml` / `go.mod`). |

---

## How it fits together

1. **Register** vaults: `akv vault add <name> [url]`.
2. **Select** a default vault: `akv use <name>` (or set `--vault` / `AKV_VAULT` each time).
3. **Read/write** secrets, keys, and certificates with the commands below.
4. **Update** the CLI when needed: `akv upgrade` (optionally a version or `--pre-release`).

---

## Global usage

```text
akv [global options] <command> [command args]
```

### Global options (root)

| Option | Description |
|--------|-------------|
| `--vault` | Vault **name** or **URL** (also `AKV_VAULT`). Short names become `https://<name>.vault.azure.net`. |
| `--vault-url` | Full vault URL (also `AKV_VAULT_URL`). |
| `--version` | Print the CLI version and exit. |

### Top-level commands (summary)

| Command | What it does |
|---------|----------------|
| `akv completion` | Print shell completion script: `akv completion bash` (or `zsh`, `fish`, `powershell`). |
| `secrets` | Secret subcommands (see below). |
| `keys` | Key subcommands. |
| `certificates` | Certificate subcommands. |
| `config` | Read/write `akv` config file paths. |
| `vault` | Manage named vaults in config. |
| `use <name>` | Set current vault (shortcut for `vault use`). |
| `upgrade [version]` | Self-update the `akv` binary from GitHub Releases. |
| `version` | Print version (also available via `--version`). |
| `ls [pattern]` | List secrets (same as `secrets ls`). |

**Shortcuts** (same as the matching `secrets` subcommand): `get`, `set`, `rm`, `ensure` — e.g. `akv get <name>` = `akv secrets get <name>`.

---

## Commands by area

### Secrets

| Command | Description |
|---------|-------------|
| `akv secrets get <name>...` | Get secret values. Supports `--version`, `--format` (`text`, `json`, `sh`, `bash`, `zsh`, `dotenv`, `azure-devops`, `github-actions`, `cast`, ...). |
| `akv secrets get-data <name>` | Print full secret metadata as JSON. |
| `akv secrets set <name> [value]` | Set a value; use flags like `--generate` where supported. |
| `akv secrets ensure <name>` | Create a secret only if it does not exist. |
| `akv secrets rm <name>` | Delete a secret; `--purge` to remove soft-deleted permanently. |
| `akv secrets purge <name>` | Purge a soft-deleted secret. |
| `akv secrets update <name>` | Update metadata (e.g. tags, `expires`, `not-before`, `--version`). |
| `akv secrets ls [pattern]` | List secret names (glob). |
| `akv secrets import` | Import from JSON. |
| `akv secrets export` | Export to JSON. |
| `akv secrets sync` | Sync from a JSON spec (create/update/delete as described). |
| `akv ls [pattern]` | Alias for `secrets ls`. |

### Keys

| Command | Description |
|---------|-------------|
| `akv keys get <name>` | Key metadata. `--version` optional. |
| `akv keys set <name>` | Create/set key material (see help for options). |
| `akv keys update <name>` | Update key properties. |
| `akv keys rm <name>` | Delete a key. |
| `akv keys purge <name>` | Purge a deleted key. |
| `akv keys ls [pattern]` | List keys. |

### Certificates

| Command | Description |
|---------|-------------|
| `akv certificates get <name>` | Certificate metadata. |
| `akv certificates get-data <name>` | Full record as JSON. |
| `akv certificates set <name>` | Set/upload certificate (see help). |
| `akv certificates update <name>` | Update properties. |
| `akv certificates rm <name>` | Delete. |
| `akv certificates purge <name>` | Purge. |
| `akv certificates ls [pattern]` | List certificates. |
| `akv certificates download <name>` | Download cert (and related secret as applicable). |
| `akv certificates upload <file>` | Upload a certificate file. |

### Vaults and config

| Command | Description |
|---------|-------------|
| `akv vault add <name> [url]` | Add a vault to config. |
| `akv vault rm <name>` | Remove a vault. |
| `akv vault ls [pattern]` | List configured vaults. |
| `akv vault use <name>` | Set current vault. |
| `akv vault show [name]` | Show vault details. |
| `akv use <name>` | Same as `vault use`. |
| `akv config get <path>` | Read a config value by path. |
| `akv config set <path> <value>` | Set a value. |
| `akv config rm <path>` | Remove a path. |

### Meta

| Command | Description |
|---------|-------------|
| `akv version` or `akv --version` | Print `akv` version. |
| `akv upgrade` | Update to the latest release. |
| `akv upgrade <version>` | Update to a specific version tag (without `v` prefix as applicable). |
| `akv upgrade --pre-release` | Include GitHub prereleases when resolving “latest”. |

Run **`akv <command> --help`** for flags and examples for any command.

---

## Environment variables

| Variable | Purpose |
|----------|---------|
| `AKV_VAULT` | Default vault name or URL. |
| `AKV_VAULT_URL` | Default full vault URL. |
| `AKV_INSTALL_DIR` | Target directory for the install script. |
| `CAST_SECRETS` | Output file for `secrets get --format cast`. |
| `AZURE_CLIENT_SECRET` | Client secret (with service principal auth in config). |

---

## Shell completion

- **Bash:** `akv completion bash | sudo tee /etc/bash_completion.d/akv` or `source <(akv completion bash)`.
- **Zsh:** `akv completion zsh > "${fpath[1]}/_akv"` or `source <(akv completion zsh)`.
- **Fish:** `akv completion fish > ~/.config/fish/completions/akv.fish`.
- **PowerShell:** `akv completion powershell | Out-String | Invoke-Expression`.

---

## Authentication

`akv` uses the Azure SDK credential chain, including:

1. Service principal (with `clientId` / `tenantId` in config and `AZURE_CLIENT_SECRET` when required).
2. Managed identity when configured.
3. Other [DefaultAzureCredential](https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/azidentity#NewDefaultAzureCredential) sources (Azure CLI, environment, etc., depending on your environment).

---

## Configuration file

On Linux/macOS, config is typically `~/.config/akv.json`; on Windows, `%APPDATA%\akv.json`. Example:

```json
{
  "currentVault": "my-vault",
  "vaults": {
    "my-vault": {
      "name": "my-vault",
      "url": "https://my-vault.vault.azure.net"
    }
  },
  "auth": {
    "clientId": "your-app-id",
    "tenantId": "your-tenant-id",
    "servicePrincipal": true
  }
}
```

---

## License

See [LICENSE](LICENSE) in this repository (MIT).
