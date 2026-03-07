package config

import (
	"testing"
)

func TestGetSetAccessKeyID(t *testing.T) {
	tests := []struct {
		name    string
		profile string
		keyID   string
	}{
		{"set for default profile", "default", "my-key-id"},
		{"set for custom profile", "prod", "prod-key-id"},
		{"overwrite existing key", "default", "new-key-id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetState()
			t.Cleanup(resetState)

			SetCurrentProfile(tt.profile)
			SetAccessKeyID(tt.profile, tt.keyID)

			if got := GetAccessKeyID(); got != tt.keyID {
				t.Errorf("GetAccessKeyID() = %q, want %q", got, tt.keyID)
			}
		})
	}
}

func TestGetAccessKeyID_NilCredentials(t *testing.T) {
	configMu.Lock()
	globalCredentials = nil
	configMu.Unlock()
	t.Cleanup(resetState)

	if got := GetAccessKeyID(); got != "" {
		t.Errorf("GetAccessKeyID() = %q when nil, want empty string", got)
	}
}

func TestGetAccessKeyID_MissingProfile(t *testing.T) {
	resetState()
	t.Cleanup(resetState)

	SetCurrentProfile("nonexistent")
	if got := GetAccessKeyID(); got != "" {
		t.Errorf("GetAccessKeyID() = %q for missing profile, want empty string", got)
	}
}

func TestGetSetAccessKeySecret(t *testing.T) {
	tests := []struct {
		name    string
		profile string
		secret  string
	}{
		{"set for default profile", "default", "my-secret"},
		{"set for custom profile", "prod", "prod-secret"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetState()
			t.Cleanup(resetState)

			SetCurrentProfile(tt.profile)
			SetAccessKeySecret(tt.profile, tt.secret)

			if got := GetAccessKeySecret(); got != tt.secret {
				t.Errorf("GetAccessKeySecret() = %q, want %q", got, tt.secret)
			}
		})
	}
}

func TestGetAccessKeySecret_NilCredentials(t *testing.T) {
	configMu.Lock()
	globalCredentials = nil
	configMu.Unlock()
	t.Cleanup(resetState)

	if got := GetAccessKeySecret(); got != "" {
		t.Errorf("GetAccessKeySecret() = %q when nil, want empty string", got)
	}
}

func TestSetAccessKeyID_NilCredentials(t *testing.T) {
	configMu.Lock()
	globalCredentials = nil
	configMu.Unlock()
	t.Cleanup(resetState)

	SetAccessKeyID("default", "key")

	if got := GetAccessKeyID(); got != "key" {
		t.Errorf("GetAccessKeyID() = %q after SetAccessKeyID with nil creds, want %q", got, "key")
	}
}

func TestSetAccessKeySecret_NilCredentials(t *testing.T) {
	configMu.Lock()
	globalCredentials = nil
	configMu.Unlock()
	t.Cleanup(resetState)

	SetAccessKeySecret("default", "secret")

	if got := GetAccessKeySecret(); got != "secret" {
		t.Errorf("GetAccessKeySecret() = %q after SetAccessKeySecret with nil creds, want %q", got, "secret")
	}
}

func TestSetCredentials(t *testing.T) {
	resetState()
	t.Cleanup(resetState)

	SetCredentials("prod", "prod-key", "prod-secret")
	SetCurrentProfile("prod")

	if got := GetAccessKeyID(); got != "prod-key" {
		t.Errorf("GetAccessKeyID() = %q, want %q", got, "prod-key")
	}
	if got := GetAccessKeySecret(); got != "prod-secret" {
		t.Errorf("GetAccessKeySecret() = %q, want %q", got, "prod-secret")
	}
}

func TestSetCredentials_NilCredentials(t *testing.T) {
	configMu.Lock()
	globalCredentials = nil
	configMu.Unlock()
	t.Cleanup(resetState)

	SetCredentials("default", "key", "secret")

	cred, ok := GetCredentials("default")
	if !ok {
		t.Fatal("GetCredentials() returned false after SetCredentials")
	}
	if cred.AccessKeyID != "key" {
		t.Errorf("cred.AccessKeyID = %q, want %q", cred.AccessKeyID, "key")
	}
	if cred.AccessKeySecret != "secret" {
		t.Errorf("cred.AccessKeySecret = %q, want %q", cred.AccessKeySecret, "secret")
	}
}

func TestGetCredentials(t *testing.T) {
	t.Run("existing profile", func(t *testing.T) {
		resetState()
		t.Cleanup(resetState)

		SetCredentials("dev", "dev-key", "dev-secret")

		cred, ok := GetCredentials("dev")
		if !ok {
			t.Fatal("GetCredentials() returned false for existing profile")
		}
		if cred.AccessKeyID != "dev-key" {
			t.Errorf("cred.AccessKeyID = %q, want %q", cred.AccessKeyID, "dev-key")
		}
		if cred.AccessKeySecret != "dev-secret" {
			t.Errorf("cred.AccessKeySecret = %q, want %q", cred.AccessKeySecret, "dev-secret")
		}
	})

	t.Run("missing profile", func(t *testing.T) {
		resetState()
		t.Cleanup(resetState)

		_, ok := GetCredentials("nonexistent")
		if ok {
			t.Error("GetCredentials() returned true for nonexistent profile, want false")
		}
	})

	t.Run("nil credentials", func(t *testing.T) {
		configMu.Lock()
		globalCredentials = nil
		configMu.Unlock()
		t.Cleanup(resetState)

		_, ok := GetCredentials("any")
		if ok {
			t.Error("GetCredentials() returned true when credentials is nil, want false")
		}
	})
}

func TestSetAccessKeyID_PreservesSecret(t *testing.T) {
	resetState()
	t.Cleanup(resetState)

	SetCredentials("default", "old-key", "my-secret")
	SetAccessKeyID("default", "new-key")

	cred, _ := GetCredentials("default")
	if cred.AccessKeyID != "new-key" {
		t.Errorf("cred.AccessKeyID = %q, want %q", cred.AccessKeyID, "new-key")
	}
	if cred.AccessKeySecret != "my-secret" {
		t.Errorf("cred.AccessKeySecret = %q (should be preserved), want %q", cred.AccessKeySecret, "my-secret")
	}
}

func TestSetAccessKeySecret_PreservesKeyID(t *testing.T) {
	resetState()
	t.Cleanup(resetState)

	SetCredentials("default", "my-key", "old-secret")
	SetAccessKeySecret("default", "new-secret")

	cred, _ := GetCredentials("default")
	if cred.AccessKeySecret != "new-secret" {
		t.Errorf("cred.AccessKeySecret = %q, want %q", cred.AccessKeySecret, "new-secret")
	}
	if cred.AccessKeyID != "my-key" {
		t.Errorf("cred.AccessKeyID = %q (should be preserved), want %q", cred.AccessKeyID, "my-key")
	}
}
