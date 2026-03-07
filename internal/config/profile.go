package config

import "fmt"

// ListCurrentConfig returns all configuration for the current profile (excluding credentials).
func ListCurrentConfig() map[string]string {
	return map[string]string{
		"profile":  GetCurrentProfile(),
		"language": GetLanguage(),
		"output":   GetOutput(),
	}
}

// ValidateOutput validates the output format.
func ValidateOutput(output string) error {
	switch output {
	case "json", "table":
		return nil
	default:
		return fmt.Errorf("invalid output format: %s (must be 'json' or 'table')", output)
	}
}

// ValidateLanguage validates the language setting.
func ValidateLanguage(lang string) error {
	switch lang {
	case "en", "zh":
		return nil
	default:
		return fmt.Errorf("invalid language: %s (must be 'en' or 'zh')", lang)
	}
}

// ProfileExists checks whether a profile exists in configuration or credentials.
func ProfileExists(name string) bool {
	configMu.RLock()
	defer configMu.RUnlock()

	if globalConfig != nil {
		if _, ok := globalConfig.Profiles[name]; ok {
			return true
		}
	}
	if globalCredentials != nil {
		if _, ok := globalCredentials.Profiles[name]; ok {
			return true
		}
	}
	return false
}

// EnsureProfile ensures a profile exists in the configuration.
func EnsureProfile(name string) {
	configMu.Lock()
	defer configMu.Unlock()

	if globalConfig == nil {
		globalConfig = &Config{
			CurrentProfile: "default",
			Profiles:       make(map[string]ProfileConfig),
		}
	}

	if _, ok := globalConfig.Profiles[name]; !ok {
		globalConfig.Profiles[name] = ProfileConfig{
			Language: "en",
			Output:   "json",
		}
	}

	if globalCredentials == nil {
		globalCredentials = &Credentials{
			Profiles: make(map[string]CredentialProfile),
		}
	}

	if _, ok := globalCredentials.Profiles[name]; !ok {
		globalCredentials.Profiles[name] = CredentialProfile{}
	}
}
