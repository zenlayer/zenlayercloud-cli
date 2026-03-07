// Package config handles configuration management for zeno.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const (
	// ConfigDirName is the directory name for zenlayer configuration.
	ConfigDirName = ".zenlayer"
	// ConfigFileName is the main configuration file name.
	ConfigFileName = "config.json"
	// CredentialsFileName is the credentials file name.
	CredentialsFileName = "credentials.json"
)

// ProfileConfig holds per-profile settings.
type ProfileConfig struct {
	Language string `json:"language,omitempty"`
	Output   string `json:"output,omitempty"`
}

// Config represents the main configuration structure.
type Config struct {
	CurrentProfile string                   `json:"current_profile"`
	Profiles       map[string]ProfileConfig `json:"profiles"`
}

// globalConfig holds the loaded configuration.
var (
	globalConfig      *Config
	globalCredentials *Credentials
	configMu          sync.RWMutex
	currentProfile    string
)

// getConfigDir returns the configuration directory path.
func getConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ConfigDirName), nil
}

// GetConfigPath returns the full path to the config file.
func GetConfigPath() (string, error) {
	dir, err := getConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ConfigFileName), nil
}

// GetCredentialsPath returns the full path to the credentials file.
func GetCredentialsPath() (string, error) {
	dir, err := getConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, CredentialsFileName), nil
}

// Load reads the configuration files from disk.
func Load() error {
	configMu.Lock()
	defer configMu.Unlock()

	// Initialize defaults
	globalConfig = &Config{
		CurrentProfile: "default",
		Profiles:       make(map[string]ProfileConfig),
	}
	globalCredentials = &Credentials{
		Profiles: make(map[string]CredentialProfile),
	}

	// Load main config
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	if data, err := os.ReadFile(configPath); err == nil {
		if err := json.Unmarshal(data, globalConfig); err != nil {
			return fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	// Load credentials
	credPath, err := GetCredentialsPath()
	if err != nil {
		return err
	}

	if data, err := os.ReadFile(credPath); err == nil {
		if err := json.Unmarshal(data, &globalCredentials.Profiles); err != nil {
			return fmt.Errorf("failed to parse credentials file: %w", err)
		}
	}

	// Set current profile
	currentProfile = globalConfig.CurrentProfile
	if currentProfile == "" {
		currentProfile = "default"
	}

	return nil
}

// Save writes the configuration to disk.
func Save() error {
	configMu.Lock()
	defer configMu.Unlock()

	if globalConfig == nil {
		return fmt.Errorf("config not initialized")
	}

	dir, err := getConfigDir()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Save main config
	configPath := filepath.Join(dir, ConfigFileName)
	configData, err := json.MarshalIndent(globalConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// SaveCredentials writes the credentials to disk with restricted permissions.
func SaveCredentials() error {
	configMu.Lock()
	defer configMu.Unlock()

	if globalCredentials == nil {
		return fmt.Errorf("credentials not initialized")
	}

	dir, err := getConfigDir()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Save credentials with restricted permissions
	credPath := filepath.Join(dir, CredentialsFileName)
	credData, err := json.MarshalIndent(globalCredentials.Profiles, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	if err := os.WriteFile(credPath, credData, 0600); err != nil {
		return fmt.Errorf("failed to write credentials file: %w", err)
	}

	return nil
}

// GetCurrentProfile returns the current profile name.
func GetCurrentProfile() string {
	configMu.RLock()
	defer configMu.RUnlock()

	if currentProfile == "" {
		return "default"
	}
	return currentProfile
}

// SetCurrentProfile sets the current profile name.
func SetCurrentProfile(name string) {
	configMu.Lock()
	defer configMu.Unlock()

	currentProfile = name
	if globalConfig != nil {
		globalConfig.CurrentProfile = name
	}
}

// GetLanguage returns the language setting for the current profile.
func GetLanguage() string {
	configMu.RLock()
	defer configMu.RUnlock()

	if globalConfig == nil {
		return "en"
	}

	profile := currentProfile
	if profile == "" {
		profile = "default"
	}

	if p, ok := globalConfig.Profiles[profile]; ok && p.Language != "" {
		return p.Language
	}
	return "en"
}

// GetOutput returns the output format for the current profile.
func GetOutput() string {
	configMu.RLock()
	defer configMu.RUnlock()

	if globalConfig == nil {
		return "json"
	}

	profile := currentProfile
	if profile == "" {
		profile = "default"
	}

	if p, ok := globalConfig.Profiles[profile]; ok && p.Output != "" {
		return p.Output
	}
	return "json"
}

// SetProfileConfig sets the configuration for a profile.
func SetProfileConfig(name string, language, output string) {
	configMu.Lock()
	defer configMu.Unlock()

	if globalConfig == nil {
		globalConfig = &Config{
			CurrentProfile: "default",
			Profiles:       make(map[string]ProfileConfig),
		}
	}

	globalConfig.Profiles[name] = ProfileConfig{
		Language: language,
		Output:   output,
	}
}

// GetProfileConfig returns the configuration for a specific profile.
func GetProfileConfig(name string) (ProfileConfig, bool) {
	configMu.RLock()
	defer configMu.RUnlock()

	if globalConfig == nil {
		return ProfileConfig{}, false
	}

	p, ok := globalConfig.Profiles[name]
	return p, ok
}

// GetAllProfiles returns all profile names.
func GetAllProfiles() []string {
	configMu.RLock()
	defer configMu.RUnlock()

	if globalConfig == nil {
		return []string{"default"}
	}

	profiles := make([]string, 0, len(globalConfig.Profiles))
	for name := range globalConfig.Profiles {
		profiles = append(profiles, name)
	}

	if len(profiles) == 0 {
		profiles = append(profiles, "default")
	}

	return profiles
}

// Get retrieves a configuration value by key for the current profile.
func Get(key string) string {
	switch key {
	case "profile":
		return GetCurrentProfile()
	case "language":
		return GetLanguage()
	case "output":
		return GetOutput()
	case "access-key-id", "access_key_id":
		return GetAccessKeyID()
	case "access-key-secret", "access_key_secret":
		return GetAccessKeySecret()
	default:
		return ""
	}
}

// Set sets a configuration value by key for the current profile.
func Set(key, value string) error {
	profile := GetCurrentProfile()

	switch key {
	case "profile":
		if !ProfileExists(value) {
			return fmt.Errorf("profile '%s' does not exist, use 'zeno configure' to create it", value)
		}
		SetCurrentProfile(value)
		return Save()
	case "language":
		p, _ := GetProfileConfig(profile)
		SetProfileConfig(profile, value, p.Output)
		return Save()
	case "output":
		p, _ := GetProfileConfig(profile)
		SetProfileConfig(profile, p.Language, value)
		return Save()
	case "access-key-id", "access_key_id":
		SetAccessKeyID(profile, value)
		return SaveCredentials()
	case "access-key-secret", "access_key_secret":
		SetAccessKeySecret(profile, value)
		return SaveCredentials()
	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}
}
