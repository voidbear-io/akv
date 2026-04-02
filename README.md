# akv

Azure Key Vault command-line interface.

## Quick Start

```bash
curl -fsSL https://raw.githubusercontent.com/frostyeti/akv/master/eng/script/install.sh | bash
akv vault add my-vault
akv use my-vault
akv secrets ensure db-password --size 32 # gets or generates the value and then gets the value
akv secrets get db-password
akv --vault my-other-vault get prod-db-password
```

## Tutorial

1. Install `akv` with the install script or your preferred release asset.
2. Add your vault with `akv vault add <name> [url]`.
3. Select it with `akv use <name>`.
4. Set secrets with `akv secrets set`, or generate one with `--generate`.
5. Read data with `akv secrets get`, `akv secrets get-data`, or `akv secrets ls`.
6. Manage keys with `akv keys ...` and certificates with `akv certificates ...`.
7. Use `akv upgrade` later to self-update.


## Installation

### Script

Linux and macOS:

```bash
curl -fsSL https://raw.githubusercontent.com/frostyeti/akv/master/eng/script/install.sh | bash
```

Windows:

```powershell
Invoke-WebRequest -Uri https://raw.githubusercontent.com/frostyeti/akv/master/eng/script/install.ps1 -OutFile install.ps1
.\install.ps1
```

### Release Assets

Download the matching archive or binary from GitHub Releases.

## Shell Completion

### Bash

```bash
akv completion bash > /etc/bash_completion.d/akv
```

Or for a single shell session:

```bash
source <(akv completion bash)
```

### Zsh

```zsh
akv completion zsh > "${fpath[1]}/_akv"
```

Or for a single session:

```zsh
source <(akv completion zsh)
```

### Fish

```fish
akv completion fish > ~/.config/fish/completions/akv.fish
```

### PowerShell

```powershell
akv completion powershell | Out-String | Invoke-Expression
```


## Commands

### Secrets

- `akv get <name>...` Alias for `akv secrets get`.
- `akv secrets get <name>...` Get secret values.
- `akv secrets get-data <name>` Get the full secret record as JSON.
- `akv ensure <name>` Alias for `akv secrets ensure`.
- `akv set <name> [value]` Alias for `akv secrets set`.
- `akv rm <name>` Alias for `akv secrets rm`.
- `akv secrets set <name> [value]` Set a secret value.
- `akv secrets ensure <name>` Create only if missing.
- `akv secrets rm <name>` Delete a secret.
- `akv secrets purge <name>` Purge a deleted secret.
- `akv secrets update <name>` Update metadata.
- `akv secrets ls [pattern]` List secrets with glob filtering.
- `akv secrets import` Import secrets from JSON.
- `akv secrets export` Export secrets to JSON.
- `akv secrets sync` Sync secrets from JSON.

### Keys

- `akv keys get <name>`
- `akv keys set <name>`
- `akv keys update <name>`
- `akv keys rm <name>`
- `akv keys purge <name>`
- `akv keys ls [pattern]`

### Certificates

- `akv certificates get <name>`
- `akv certificates get-data <name>`
- `akv certificates set <name>`
- `akv certificates update <name>`
- `akv certificates rm <name>`
- `akv certificates purge <name>`
- `akv certificates ls [pattern]`
- `akv certificates download <name>`
- `akv certificates upload <file>`

### Vaults and Config

- `akv vault add <name> [url]`
- `akv vault rm <name>`
- `akv vault ls [pattern]`
- `akv vault use <name>`
- `akv use <name>`
- `akv vault show [name]`
- `akv config get <path>`
- `akv config set <path> <value>`
- `akv config rm <path>`

### Meta

- `akv version`
- `akv upgrade [version]`
- `akv upgrade --pre-release` Include prerelease releases when updating.

## Global Options

- `--vault` - Vault name or URL. Short names expand to `https://<name>.vault.azure.net`.
- `--vault-url` - Full vault URL.
- `--version` - Print the CLI version.

## Environment Variables

- `AKV_VAULT` - Vault name or URL shortcut.
- `AKV_VAULT_URL` - Full vault URL.
- `AKV_INSTALL_DIR` - Install destination for the installer.
- `CAST_SECRETS` - Output file used by `secrets get --format cast`.
- `AZURE_CLIENT_SECRET` - Client secret for service principal auth.

## Upgrade

Use `akv upgrade` to replace the currently installed binary in place.

```bash
akv upgrade
akv upgrade 0.1.0
akv upgrade --pre-release
```

## Authentication

`akv` supports:

1. Service principal auth via config and `AZURE_CLIENT_SECRET`.
2. Managed identity when `clientId` is configured.
3. Default Azure credential chain.

## Configuration

Create `~/.config/akv.json` on Linux/macOS or `~/AppData/Roaming/akv.json` on Windows:

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

## License

MIT License.
