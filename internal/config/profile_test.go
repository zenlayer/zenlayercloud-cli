package config

import (
	"testing"
)

func TestListCurrentConfig(t *testing.T) {
	resetState()
	t.Cleanup(resetState)

	SetCurrentProfile("testprofile")
	SetProfileConfig("testprofile", "zh", "table")

	got := ListCurrentConfig()

	if got["profile"] != "testprofile" {
		t.Errorf("ListCurrentConfig()[\"profile\"] = %q, want %q", got["profile"], "testprofile")
	}
	if got["language"] != "zh" {
		t.Errorf("ListCurrentConfig()[\"language\"] = %q, want %q", got["language"], "zh")
	}
	if got["output"] != "table" {
		t.Errorf("ListCurrentConfig()[\"output\"] = %q, want %q", got["output"], "table")
	}
}

func TestListCurrentConfig_ContainsExpectedKeys(t *testing.T) {
	resetState()
	t.Cleanup(resetState)

	got := ListCurrentConfig()

	requiredKeys := []string{"profile", "language", "output"}
	for _, key := range requiredKeys {
		if _, ok := got[key]; !ok {
			t.Errorf("ListCurrentConfig() missing required key %q", key)
		}
	}
}

func TestValidateOutput(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		wantErr bool
	}{
		{"json is valid", "json", false},
		{"table is valid", "table", false},
		{"csv is invalid", "csv", true},
		{"xml is invalid", "xml", true},
		{"empty string is invalid", "", true},
		{"JSON uppercase is invalid", "JSON", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOutput(tt.output)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateOutput(%q) error = %v, wantErr %v", tt.output, err, tt.wantErr)
			}
		})
	}
}

func TestValidateLanguage(t *testing.T) {
	tests := []struct {
		name    string
		lang    string
		wantErr bool
	}{
		{"en is valid", "en", false},
		{"zh is valid", "zh", false},
		{"fr is invalid", "fr", true},
		{"de is invalid", "de", true},
		{"empty string is invalid", "", true},
		{"EN uppercase is invalid", "EN", true},
		{"ZH uppercase is invalid", "ZH", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLanguage(tt.lang)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateLanguage(%q) error = %v, wantErr %v", tt.lang, err, tt.wantErr)
			}
		})
	}
}

func TestProfileExists(t *testing.T) {
	t.Run("returns true for profile in config", func(t *testing.T) {
		resetState()
		t.Cleanup(resetState)
		SetProfileConfig("prod", "en", "json")
		if !ProfileExists("prod") {
			t.Error("ProfileExists('prod') = false, want true")
		}
	})

	t.Run("returns true for profile in credentials only", func(t *testing.T) {
		resetState()
		t.Cleanup(resetState)
		SetCredentials("credonly", "key", "secret")
		if !ProfileExists("credonly") {
			t.Error("ProfileExists('credonly') = false, want true")
		}
	})

	t.Run("returns false for nonexistent profile", func(t *testing.T) {
		resetState()
		t.Cleanup(resetState)
		if ProfileExists("nonexistent") {
			t.Error("ProfileExists('nonexistent') = true, want false")
		}
	})

	t.Run("returns false when config is nil", func(t *testing.T) {
		configMu.Lock()
		globalConfig = nil
		globalCredentials = nil
		configMu.Unlock()
		t.Cleanup(resetState)
		if ProfileExists("any") {
			t.Error("ProfileExists('any') = true when config nil, want false")
		}
	})
}

func TestEnsureProfile_NewProfile(t *testing.T) {
	resetState()
	t.Cleanup(resetState)

	EnsureProfile("newprofile")

	p, ok := GetProfileConfig("newprofile")
	if !ok {
		t.Fatal("GetProfileConfig() returned false after EnsureProfile")
	}
	if p.Language != "en" {
		t.Errorf("EnsureProfile set Language = %q, want %q", p.Language, "en")
	}
	if p.Output != "json" {
		t.Errorf("EnsureProfile set Output = %q, want %q", p.Output, "json")
	}

	_, ok = GetCredentials("newprofile")
	if !ok {
		t.Error("EnsureProfile did not create credential profile for new profile")
	}
}

func TestEnsureProfile_ExistingProfileNotOverwritten(t *testing.T) {
	resetState()
	t.Cleanup(resetState)

	SetProfileConfig("existing", "zh", "table")
	EnsureProfile("existing")

	p, ok := GetProfileConfig("existing")
	if !ok {
		t.Fatal("GetProfileConfig() returned false for existing profile")
	}
	if p.Language != "zh" {
		t.Errorf("EnsureProfile overwrote Language: got %q, want %q", p.Language, "zh")
	}
	if p.Output != "table" {
		t.Errorf("EnsureProfile overwrote Output: got %q, want %q", p.Output, "table")
	}
}

func TestEnsureProfile_NilGlobalConfig(t *testing.T) {
	configMu.Lock()
	globalConfig = nil
	globalCredentials = nil
	configMu.Unlock()
	t.Cleanup(resetState)

	EnsureProfile("testprofile")

	p, ok := GetProfileConfig("testprofile")
	if !ok {
		t.Fatal("GetProfileConfig() returned false after EnsureProfile with nil config")
	}
	if p.Language != "en" {
		t.Errorf("Profile Language = %q, want %q", p.Language, "en")
	}
}

func TestEnsureProfile_ExistingCredentialsNotOverwritten(t *testing.T) {
	resetState()
	t.Cleanup(resetState)

	SetCredentials("existing", "existing-key", "existing-secret")
	EnsureProfile("existing")

	cred, ok := GetCredentials("existing")
	if !ok {
		t.Fatal("GetCredentials() returned false for existing profile")
	}
	if cred.AccessKeyID != "existing-key" {
		t.Errorf("EnsureProfile overwrote AccessKeyID: got %q, want %q", cred.AccessKeyID, "existing-key")
	}
}
