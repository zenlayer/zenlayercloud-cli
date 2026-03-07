package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// resetState resets all global config state to a clean default.
func resetState() {
	configMu.Lock()
	defer configMu.Unlock()
	globalConfig = &Config{
		CurrentProfile: "default",
		Profiles:       make(map[string]ProfileConfig),
	}
	globalCredentials = &Credentials{
		Profiles: make(map[string]CredentialProfile),
	}
	currentProfile = "default"
}

func TestGetConfigPath(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	path, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() error = %v", err)
	}
	want := filepath.Join(tmpDir, ConfigDirName, ConfigFileName)
	if path != want {
		t.Errorf("GetConfigPath() = %q, want %q", path, want)
	}
}

func TestGetCredentialsPath(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	path, err := GetCredentialsPath()
	if err != nil {
		t.Fatalf("GetCredentialsPath() error = %v", err)
	}
	want := filepath.Join(tmpDir, ConfigDirName, CredentialsFileName)
	if path != want {
		t.Errorf("GetCredentialsPath() = %q, want %q", path, want)
	}
}

func TestLoad_NoFiles(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	if err := Load(); err != nil {
		t.Fatalf("Load() error = %v when no config files exist", err)
	}
	if got := GetCurrentProfile(); got != "default" {
		t.Errorf("GetCurrentProfile() = %q after Load with no file, want %q", got, "default")
	}
}

func TestLoad_WithConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	configDir := filepath.Join(tmpDir, ConfigDirName)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatal(err)
	}

	cfg := Config{
		CurrentProfile: "prod",
		Profiles: map[string]ProfileConfig{
			"prod": {Language: "zh", Output: "table"},
		},
	}
	data, _ := json.Marshal(cfg)
	if err := os.WriteFile(filepath.Join(configDir, ConfigFileName), data, 0644); err != nil {
		t.Fatal(err)
	}

	if err := Load(); err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got := GetCurrentProfile(); got != "prod" {
		t.Errorf("GetCurrentProfile() = %q, want %q", got, "prod")
	}
	if got := GetLanguage(); got != "zh" {
		t.Errorf("GetLanguage() = %q, want %q", got, "zh")
	}
	if got := GetOutput(); got != "table" {
		t.Errorf("GetOutput() = %q, want %q", got, "table")
	}
}

func TestLoad_WithCredentialsFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	configDir := filepath.Join(tmpDir, ConfigDirName)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatal(err)
	}

	// Write minimal config file
	cfg := Config{CurrentProfile: "default", Profiles: map[string]ProfileConfig{}}
	data, _ := json.Marshal(cfg)
	if err := os.WriteFile(filepath.Join(configDir, ConfigFileName), data, 0644); err != nil {
		t.Fatal(err)
	}

	// Write credentials file
	creds := map[string]CredentialProfile{
		"default": {AccessKeyID: "test-key", AccessKeySecret: "test-secret"},
	}
	credData, _ := json.Marshal(creds)
	if err := os.WriteFile(filepath.Join(configDir, CredentialsFileName), credData, 0600); err != nil {
		t.Fatal(err)
	}

	if err := Load(); err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got := GetAccessKeyID(); got != "test-key" {
		t.Errorf("GetAccessKeyID() = %q, want %q", got, "test-key")
	}
	if got := GetAccessKeySecret(); got != "test-secret" {
		t.Errorf("GetAccessKeySecret() = %q, want %q", got, "test-secret")
	}
}

func TestLoad_InvalidConfigJSON(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	configDir := filepath.Join(tmpDir, ConfigDirName)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, ConfigFileName), []byte("not-json"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := Load(); err == nil {
		t.Error("Load() expected error for invalid JSON, got nil")
	}
}

func TestSave(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	resetState()

	if err := Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	configPath := filepath.Join(tmpDir, ConfigDirName, ConfigFileName)
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Config file not created: %v", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("Written config is not valid JSON: %v", err)
	}
	if cfg.CurrentProfile != "default" {
		t.Errorf("Saved current_profile = %q, want %q", cfg.CurrentProfile, "default")
	}
}

func TestSave_NilConfig(t *testing.T) {
	configMu.Lock()
	globalConfig = nil
	configMu.Unlock()
	t.Cleanup(resetState)

	if err := Save(); err == nil {
		t.Error("Save() expected error when config is nil, got nil")
	}
}

func TestSaveCredentials(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	resetState()

	SetCredentials("default", "my-key", "my-secret")

	if err := SaveCredentials(); err != nil {
		t.Fatalf("SaveCredentials() error = %v", err)
	}

	credPath := filepath.Join(tmpDir, ConfigDirName, CredentialsFileName)
	data, err := os.ReadFile(credPath)
	if err != nil {
		t.Fatalf("Credentials file not created: %v", err)
	}

	var creds map[string]CredentialProfile
	if err := json.Unmarshal(data, &creds); err != nil {
		t.Fatalf("Written credentials is not valid JSON: %v", err)
	}
	if creds["default"].AccessKeyID != "my-key" {
		t.Errorf("Saved access_key_id = %q, want %q", creds["default"].AccessKeyID, "my-key")
	}

	// Verify restricted file permissions (0600)
	info, err := os.Stat(credPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("Credentials file permissions = %o, want 0600", info.Mode().Perm())
	}
}

func TestSaveCredentials_NilCredentials(t *testing.T) {
	configMu.Lock()
	globalCredentials = nil
	configMu.Unlock()
	t.Cleanup(resetState)

	if err := SaveCredentials(); err == nil {
		t.Error("SaveCredentials() expected error when credentials is nil, got nil")
	}
}

func TestGetSetCurrentProfile(t *testing.T) {
	resetState()
	t.Cleanup(resetState)

	tests := []struct {
		name    string
		profile string
		want    string
	}{
		{"set to prod", "prod", "prod"},
		{"set to staging", "staging", "staging"},
		{"set to default", "default", "default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetCurrentProfile(tt.profile)
			if got := GetCurrentProfile(); got != tt.want {
				t.Errorf("GetCurrentProfile() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetCurrentProfile_EmptyReturnsDefault(t *testing.T) {
	configMu.Lock()
	currentProfile = ""
	configMu.Unlock()
	t.Cleanup(resetState)

	if got := GetCurrentProfile(); got != "default" {
		t.Errorf("GetCurrentProfile() = %q when empty, want %q", got, "default")
	}
}

func TestGetLanguage(t *testing.T) {
	tests := []struct {
		name           string
		profileName    string
		profileLang    string
		currentProfile string
		want           string
	}{
		{
			name:           "returns language for current profile",
			profileName:    "test",
			profileLang:    "zh",
			currentProfile: "test",
			want:           "zh",
		},
		{
			name:           "returns default 'en' when no profile language set",
			profileName:    "test",
			profileLang:    "",
			currentProfile: "test",
			want:           "en",
		},
		{
			name:           "returns 'en' when profile doesn't exist",
			profileName:    "other",
			profileLang:    "zh",
			currentProfile: "nonexistent",
			want:           "en",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetState()
			t.Cleanup(resetState)
			SetProfileConfig(tt.profileName, tt.profileLang, "")
			SetCurrentProfile(tt.currentProfile)
			if got := GetLanguage(); got != tt.want {
				t.Errorf("GetLanguage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetLanguage_NilConfig(t *testing.T) {
	configMu.Lock()
	globalConfig = nil
	configMu.Unlock()
	t.Cleanup(resetState)

	if got := GetLanguage(); got != "en" {
		t.Errorf("GetLanguage() = %q when config nil, want %q", got, "en")
	}
}

func TestGetOutput(t *testing.T) {
	tests := []struct {
		name           string
		profileName    string
		profileOutput  string
		currentProfile string
		want           string
	}{
		{
			name:           "returns output for current profile",
			profileName:    "test",
			profileOutput:  "table",
			currentProfile: "test",
			want:           "table",
		},
		{
			name:           "returns default 'json' when no profile output set",
			profileName:    "test",
			profileOutput:  "",
			currentProfile: "test",
			want:           "json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetState()
			t.Cleanup(resetState)
			SetProfileConfig(tt.profileName, "", tt.profileOutput)
			SetCurrentProfile(tt.currentProfile)
			if got := GetOutput(); got != tt.want {
				t.Errorf("GetOutput() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetOutput_NilConfig(t *testing.T) {
	configMu.Lock()
	globalConfig = nil
	configMu.Unlock()
	t.Cleanup(resetState)

	if got := GetOutput(); got != "json" {
		t.Errorf("GetOutput() = %q when config nil, want %q", got, "json")
	}
}

func TestSetProfileConfig(t *testing.T) {
	resetState()
	t.Cleanup(resetState)

	SetProfileConfig("myprofile", "zh", "table")

	p, ok := GetProfileConfig("myprofile")
	if !ok {
		t.Fatal("GetProfileConfig() returned false, expected profile to exist")
	}
	if p.Language != "zh" {
		t.Errorf("Profile.Language = %q, want %q", p.Language, "zh")
	}
	if p.Output != "table" {
		t.Errorf("Profile.Output = %q, want %q", p.Output, "table")
	}
}

func TestSetProfileConfig_NilGlobalConfig(t *testing.T) {
	configMu.Lock()
	globalConfig = nil
	configMu.Unlock()
	t.Cleanup(resetState)

	SetProfileConfig("test", "en", "json")

	p, ok := GetProfileConfig("test")
	if !ok {
		t.Fatal("GetProfileConfig() returned false after SetProfileConfig with nil config")
	}
	if p.Language != "en" {
		t.Errorf("Profile.Language = %q, want %q", p.Language, "en")
	}
}

func TestGetProfileConfig_Missing(t *testing.T) {
	resetState()
	t.Cleanup(resetState)

	_, ok := GetProfileConfig("nonexistent")
	if ok {
		t.Error("GetProfileConfig() returned true for nonexistent profile, want false")
	}
}

func TestGetProfileConfig_NilConfig(t *testing.T) {
	configMu.Lock()
	globalConfig = nil
	configMu.Unlock()
	t.Cleanup(resetState)

	_, ok := GetProfileConfig("any")
	if ok {
		t.Error("GetProfileConfig() returned true when config is nil, want false")
	}
}

func TestGetAllProfiles(t *testing.T) {
	t.Run("returns default when no profiles exist", func(t *testing.T) {
		resetState()
		t.Cleanup(resetState)
		profiles := GetAllProfiles()
		if len(profiles) != 1 || profiles[0] != "default" {
			t.Errorf("GetAllProfiles() = %v, want [default]", profiles)
		}
	})

	t.Run("returns all configured profiles", func(t *testing.T) {
		resetState()
		t.Cleanup(resetState)
		SetProfileConfig("prod", "en", "json")
		SetProfileConfig("dev", "zh", "table")
		profiles := GetAllProfiles()
		if len(profiles) != 2 {
			t.Errorf("GetAllProfiles() returned %d profiles, want 2", len(profiles))
		}
	})

	t.Run("returns default when config is nil", func(t *testing.T) {
		configMu.Lock()
		globalConfig = nil
		configMu.Unlock()
		t.Cleanup(resetState)
		profiles := GetAllProfiles()
		if len(profiles) != 1 || profiles[0] != "default" {
			t.Errorf("GetAllProfiles() = %v when nil, want [default]", profiles)
		}
	})
}

func TestGet(t *testing.T) {
	resetState()
	t.Cleanup(resetState)

	SetCurrentProfile("myprofile")
	SetProfileConfig("myprofile", "zh", "table")
	SetCredentials("myprofile", "keyid", "keysecret")

	tests := []struct {
		key  string
		want string
	}{
		{"profile", "myprofile"},
		{"language", "zh"},
		{"output", "table"},
		{"access-key-id", "keyid"},
		{"access_key_id", "keyid"},
		{"access-key-secret", "keysecret"},
		{"access_key_secret", "keysecret"},
		{"unknown-key", ""},
	}

	for _, tt := range tests {
		t.Run("key="+tt.key, func(t *testing.T) {
			if got := Get(tt.key); got != tt.want {
				t.Errorf("Get(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestSet(t *testing.T) {
	t.Run("set language", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)
		resetState()
		t.Cleanup(resetState)

		if err := Set("language", "zh"); err != nil {
			t.Fatalf("Set('language', 'zh') error = %v", err)
		}
		if got := GetLanguage(); got != "zh" {
			t.Errorf("GetLanguage() = %q after Set, want %q", got, "zh")
		}
	})

	t.Run("set output", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)
		resetState()
		t.Cleanup(resetState)

		if err := Set("output", "table"); err != nil {
			t.Fatalf("Set('output', 'table') error = %v", err)
		}
		if got := GetOutput(); got != "table" {
			t.Errorf("GetOutput() = %q after Set, want %q", got, "table")
		}
	})

	t.Run("set profile", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)
		resetState()
		t.Cleanup(resetState)

		// Profile must exist before it can be set as current
		SetProfileConfig("staging", "en", "json")

		if err := Set("profile", "staging"); err != nil {
			t.Fatalf("Set('profile', 'staging') error = %v", err)
		}
		if got := GetCurrentProfile(); got != "staging" {
			t.Errorf("GetCurrentProfile() = %q after Set, want %q", got, "staging")
		}
	})

	t.Run("set profile nonexistent returns error", func(t *testing.T) {
		resetState()
		t.Cleanup(resetState)

		if err := Set("profile", "nonexistent"); err == nil {
			t.Error("Set('profile', 'nonexistent') expected error for non-existent profile, got nil")
		}
	})

	t.Run("set access-key-id", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)
		resetState()
		t.Cleanup(resetState)

		if err := Set("access-key-id", "newkey"); err != nil {
			t.Fatalf("Set('access-key-id', ...) error = %v", err)
		}
		if got := GetAccessKeyID(); got != "newkey" {
			t.Errorf("GetAccessKeyID() = %q after Set, want %q", got, "newkey")
		}
	})

	t.Run("set unknown key returns error", func(t *testing.T) {
		resetState()
		t.Cleanup(resetState)
		if err := Set("unknown-key", "value"); err == nil {
			t.Error("Set() expected error for unknown key, got nil")
		}
	})
}
