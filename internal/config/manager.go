// Package config provides configuration management for akv.
// It handles reading, writing, and managing vault configurations.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gobwas/glob"
	"github.com/spf13/viper"
)

const (
	configFileName = "akv"
	configFileExt  = "json"
)

// Config represents the akv configuration structure.
type Config struct {
	CurrentVault string            `json:"currentVault" mapstructure:"currentVault"`
	Vaults       map[string]*Vault `json:"vaults" mapstructure:"vaults"`
	Auth         *AuthConfig       `json:"auth,omitempty" mapstructure:"auth,omitempty"`
}

// Vault represents a single vault configuration.
type Vault struct {
	Name string `json:"name" mapstructure:"name"`
	URL  string `json:"url" mapstructure:"url"`
}

// AuthConfig represents authentication settings.
type AuthConfig struct {
	ClientID         string `json:"clientId" mapstructure:"clientId"`
	TenantID         string `json:"tenantId" mapstructure:"tenantId"`
	ServicePrincipal bool   `json:"servicePrincipal" mapstructure:"servicePrincipal"`
}

// Manager handles configuration operations.
type Manager struct {
	viper      *viper.Viper
	configPath string
}

// getConfigDir returns the appropriate config directory for the current OS.
func getConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	if runtime.GOOS == "windows" {
		return filepath.Join(home, "AppData", "Roaming")
	}
	return filepath.Join(home, ".config")
}

// NewManager creates a new config manager.
func NewManager() (*Manager, error) {
	configDir := getConfigDir()
	if configDir == "" {
		return nil, fmt.Errorf("could not determine config directory")
	}

	v := viper.New()
	v.SetConfigName(configFileName)
	v.SetConfigType(configFileExt)
	v.AddConfigPath(configDir)

	// Set defaults
	v.SetDefault("vaults", map[string]interface{}{})

	// Read existing config
	_ = v.ReadInConfig()

	return &Manager{
		viper:      v,
		configPath: filepath.Join(configDir, configFileName+"."+configFileExt),
	}, nil
}

// normalizePath converts various separators to viper-compatible path.
// Supports dot notation (this.would.be.a.property), colon (this:would:be:a:property),
// and forward slash (this/would/be/a/property).
func normalizePath(path string) string {
	// Replace colons and forward slashes with dots
	path = strings.ReplaceAll(path, ":", ".")
	path = strings.ReplaceAll(path, "/", ".")
	return path
}

// Get retrieves a value from the config using path notation.
// Returns undefined (nil) if the value doesn't exist.
func (m *Manager) Get(path string) interface{} {
	normalizedPath := normalizePath(path)
	return m.viper.Get(normalizedPath)
}

// Set sets a value in the config using path notation.
func (m *Manager) Set(path string, value interface{}) error {
	normalizedPath := normalizePath(path)
	m.viper.Set(normalizedPath, value)
	return m.Save()
}

// Remove removes a value from the config using path notation.
func (m *Manager) Remove(path string) error {
	normalizedPath := normalizePath(path)

	// For nested paths, we need to manually handle removal
	parts := strings.Split(normalizedPath, ".")
	if len(parts) == 1 {
		// Top-level key
		m.viper.Set(normalizedPath, nil)
	} else {
		// Nested key - get parent and remove from it
		parent := m.viper.GetStringMap(strings.Join(parts[:len(parts)-1], "."))
		if parent != nil {
			delete(parent, parts[len(parts)-1])
			m.viper.Set(strings.Join(parts[:len(parts)-1], "."), parent)
		}
	}

	return m.Save()
}

// GetCurrentVault returns the currently selected vault name.
func (m *Manager) GetCurrentVault() string {
	return m.viper.GetString("currentVault")
}

// SetCurrentVault sets the current vault.
func (m *Manager) SetCurrentVault(name string) error {
	m.viper.Set("currentVault", name)
	return m.Save()
}

// GetVault retrieves a vault by name.
func (m *Manager) GetVault(name string) *Vault {
	vaults := m.GetVaults()
	return vaults[name]
}

// GetVaults returns all configured vaults.
func (m *Manager) GetVaults() map[string]*Vault {
	result := make(map[string]*Vault)
	vaults := m.viper.GetStringMap("vaults")

	for name, v := range vaults {
		if vaultMap, ok := v.(map[string]interface{}); ok {
			vault := &Vault{Name: name}
			if url, ok := vaultMap["url"].(string); ok {
				vault.URL = url
			}
			result[name] = vault
		}
	}

	return result
}

// AddVault adds a new vault.
func (m *Manager) AddVault(name, url string) error {
	m.viper.Set("vaults."+name+".name", name)
	m.viper.Set("vaults."+name+".url", url)
	return m.Save()
}

// RemoveVault removes a vault by name.
func (m *Manager) RemoveVault(name string) error {
	vaults := m.viper.GetStringMap("vaults")
	delete(vaults, name)
	m.viper.Set("vaults", vaults)

	// If this was the current vault, clear it
	if m.GetCurrentVault() == name {
		m.viper.Set("currentVault", "")
	}

	return m.Save()
}

// ListVaults returns a list of vault names matching the given glob pattern.
// If pattern is empty, returns all vaults.
func (m *Manager) ListVaults(pattern string) ([]string, error) {
	vaults := m.GetVaults()
	var result []string

	var g glob.Glob
	var err error
	if pattern != "" {
		g, err = glob.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid glob pattern: %w", err)
		}
	}

	for name := range vaults {
		if g == nil || g.Match(name) {
			result = append(result, name)
		}
	}

	return result, nil
}

// GetVaultURL returns the URL for a vault, looking up by name if needed.
func (m *Manager) GetVaultURL(name string) (string, error) {
	if name == "" {
		// Try current vault
		name = m.GetCurrentVault()
		if name == "" {
			return "", fmt.Errorf("no vault specified and no current vault set")
		}
	}

	vault := m.GetVault(name)
	if vault == nil {
		return "", fmt.Errorf("vault %q not found", name)
	}

	return vault.URL, nil
}

// Save writes the config to disk.
func (m *Manager) Save() error {
	// Ensure config directory exists
	configDir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	// Get all settings and save as JSON
	settings := m.viper.AllSettings()
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	return nil
}

// Path returns the config file path.
func (m *Manager) Path() string {
	return m.configPath
}

// All returns all configuration as a map.
func (m *Manager) All() map[string]interface{} {
	return m.viper.AllSettings()
}
