package cmd

import "errors"

var ErrVaultURLRequired = errors.New("vault URL is required (set --vault-url or AKV_VAULT_URL)")
