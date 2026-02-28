// Package auth provides Azure authentication utilities for akv.
// It supports multiple authentication methods including service principals,
// managed identities, and the default Azure credential chain.
package auth

import (
	"fmt"
	"os"
	"runtime"

	"github.com/99designs/keyring"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/spf13/viper"
)

// AuthConfig represents the authentication configuration from the config file.
type AuthConfig struct {
	ClientID         string `json:"clientId" mapstructure:"clientId"`
	TenantID         string `json:"tenantId" mapstructure:"tenantId"`
	ServicePrincipal bool   `json:"servicePrincipal" mapstructure:"servicePrincipal"`
}

const (
	configFileName  = "akv"
	configFileExt   = "json"
	keyringService  = "akv"
	keyringKey      = "azure-client-secret"
	envClientSecret = "AZURE_CLIENT_SECRET"
)

// yellowWarning prints a yellow warning message to stderr.
func yellowWarning(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "\x1b[33m[warning]\x1b[0m "+format+"\n", a...)
}

// getConfigDir returns the appropriate config directory for the current OS.
func getConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	if runtime.GOOS == "windows" {
		return home + "/AppData/Roaming"
	}
	return home + "/.config"
}

// loadAuthConfig loads the authentication configuration from the config file.
func loadAuthConfig() (*AuthConfig, error) {
	configDir := getConfigDir()
	if configDir == "" {
		return nil, fmt.Errorf("could not determine config directory")
	}

	viper.SetConfigName(configFileName)
	viper.SetConfigType(configFileExt)
	viper.AddConfigPath(configDir)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, nil
		}
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg AuthConfig
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &cfg, nil
}

// openKeyring opens the OS keyring for retrieving the client secret.
func openKeyring() (keyring.Keyring, error) {
	kr, err := keyring.Open(keyring.Config{
		ServiceName: keyringService,
		AllowedBackends: []keyring.BackendType{
			keyring.KeychainBackend,
			keyring.WinCredBackend,
			keyring.SecretServiceBackend,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("open keyring: %w", err)
	}
	return kr, nil
}

// getClientSecret retrieves the client secret from environment variable or keyring.
func getClientSecret() (string, error) {
	// First check environment variable
	if secret := os.Getenv(envClientSecret); secret != "" {
		return secret, nil
	}

	// Then check keyring
	kr, err := openKeyring()
	if err != nil {
		return "", err
	}

	item, err := kr.Get(keyringKey)
	if err != nil {
		if err == keyring.ErrKeyNotFound {
			return "", nil
		}
		return "", fmt.Errorf("get secret from keyring: %w", err)
	}

	return string(item.Data), nil
}

// NewCredential creates an Azure credential based on configuration.
// It attempts the following in order:
// 1. If servicePrincipal=true with clientId and tenantId:
//   - Get clientSecret from AZURE_CLIENT_SECRET env var or OS keyring
//   - If clientSecret found, use ClientSecretCredential
//   - Otherwise, print warning and fall back to DefaultAzureCredential
//
// 2. If only clientId exists (no tenantId):
//   - Use ManagedIdentityCredential with the specified clientId
//
// 3. Otherwise:
//   - Use DefaultAzureCredential
func NewCredential() (azcore.TokenCredential, error) {
	cfg, err := loadAuthConfig()
	if err != nil {
		return nil, fmt.Errorf("load auth config: %w", err)
	}

	// No config file or empty config, use default
	if cfg == nil {
		return azidentity.NewDefaultAzureCredential(nil)
	}

	// Service Principal authentication
	if cfg.ServicePrincipal && cfg.ClientID != "" && cfg.TenantID != "" {
		clientSecret, err := getClientSecret()
		if err != nil {
			return nil, fmt.Errorf("get client secret: %w", err)
		}

		if clientSecret != "" {
			cred, err := azidentity.NewClientSecretCredential(cfg.TenantID, cfg.ClientID, clientSecret, nil)
			if err != nil {
				return nil, fmt.Errorf("create client secret credential: %w", err)
			}
			return cred, nil
		}

		// No client secret found, warn and fall back
		yellowWarning("No client secret found in environment variable %s or OS keychain, falling back to default Azure credential", envClientSecret)
		return azidentity.NewDefaultAzureCredential(nil)
	}

	// Managed Identity with specific client ID
	if cfg.ClientID != "" && cfg.TenantID == "" {
		cred, err := azidentity.NewManagedIdentityCredential(&azidentity.ManagedIdentityCredentialOptions{
			ID: azidentity.ClientID(cfg.ClientID),
		})
		if err != nil {
			return nil, fmt.Errorf("create managed identity credential: %w", err)
		}
		return cred, nil
	}

	// Default credential chain
	return azidentity.NewDefaultAzureCredential(nil)
}
