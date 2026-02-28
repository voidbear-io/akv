# akv

Azure Key Vault command-line interface

## Overview

`akv` is a CLI tool for managing Azure Key Vault resources including secrets, keys, and certificates.

## Features

- **Secrets**: Get, set, delete, and update secrets
- **Keys**: Manage cryptographic keys
- **Certificates**: Create and manage certificates
- **Cross-platform**: Works on Linux, macOS, and Windows

## Installation

### Binary Downloads

Download the appropriate archive for your platform from the [GitHub Releases](https://github.com/frostyeti/akv/releases) page.

### Linux Packages

- **Debian/Ubuntu**: Download the `.deb` package and install with `sudo dpkg -i akv-*.deb`
- **RHEL/CentOS/Fedora**: Download the `.rpm` package and install with `sudo rpm -i akv-*.rpm`
- **Arch Linux**: Download the `.pkg.tar.zst` package and install with `sudo pacman -U akv-*.pkg.tar.zst`

### Windows

- Download the `.zip` file and extract it to a directory in your PATH
- Or use the standalone `.exe` file

## Usage

```bash
# Get a secret
akv secret get --vault-url https://myvault.vault.azure.net mysecret

# Set a secret
akv secret set --vault-url https://myvault.vault.azure.net mysecret "myvalue"

# Delete a secret
akv secret rm --vault-url https://myvault.vault.azure.net mysecret
```

## Authentication

`akv` supports multiple authentication methods:

1. **Service Principal** (recommended for CI/CD):
   - Configure `~/.config/akv.json` with `clientId`, `tenantId`, and `servicePrincipal: true`
   - Store client secret in `AZURE_CLIENT_SECRET` environment variable or OS keychain

2. **Managed Identity**: Set only `clientId` in config file

3. **Default Azure Credential**: Environment variables, Azure CLI, etc.

## Configuration

Create `~/.config/akv.json` (Linux/macOS) or `~/AppData/Roaming/akv.json` (Windows):

```json
{
  "clientId": "your-app-id",
  "tenantId": "your-tenant-id",
  "servicePrincipal": true
}
```

## License

MIT License - see LICENSE file for details
