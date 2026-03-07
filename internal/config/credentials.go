package config

// CredentialProfile holds credentials for a single profile.
type CredentialProfile struct {
	AccessKeyID     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
}

// Credentials holds all credential profiles.
type Credentials struct {
	Profiles map[string]CredentialProfile
}

// GetAccessKeyID returns the access key ID for the current profile.
func GetAccessKeyID() string {
	configMu.RLock()
	defer configMu.RUnlock()

	if globalCredentials == nil {
		return ""
	}

	profile := currentProfile
	if profile == "" {
		profile = "default"
	}

	if cred, ok := globalCredentials.Profiles[profile]; ok {
		return cred.AccessKeyID
	}
	return ""
}

// GetAccessKeySecret returns the access key secret for the current profile.
func GetAccessKeySecret() string {
	configMu.RLock()
	defer configMu.RUnlock()

	if globalCredentials == nil {
		return ""
	}

	profile := currentProfile
	if profile == "" {
		profile = "default"
	}

	if cred, ok := globalCredentials.Profiles[profile]; ok {
		return cred.AccessKeySecret
	}
	return ""
}

// SetAccessKeyID sets the access key ID for a profile.
func SetAccessKeyID(profile, keyID string) {
	configMu.Lock()
	defer configMu.Unlock()

	if globalCredentials == nil {
		globalCredentials = &Credentials{
			Profiles: make(map[string]CredentialProfile),
		}
	}

	cred := globalCredentials.Profiles[profile]
	cred.AccessKeyID = keyID
	globalCredentials.Profiles[profile] = cred
}

// SetAccessKeySecret sets the access key secret for a profile.
func SetAccessKeySecret(profile, secret string) {
	configMu.Lock()
	defer configMu.Unlock()

	if globalCredentials == nil {
		globalCredentials = &Credentials{
			Profiles: make(map[string]CredentialProfile),
		}
	}

	cred := globalCredentials.Profiles[profile]
	cred.AccessKeySecret = secret
	globalCredentials.Profiles[profile] = cred
}

// SetCredentials sets both access key ID and secret for a profile.
func SetCredentials(profile, keyID, secret string) {
	configMu.Lock()
	defer configMu.Unlock()

	if globalCredentials == nil {
		globalCredentials = &Credentials{
			Profiles: make(map[string]CredentialProfile),
		}
	}

	globalCredentials.Profiles[profile] = CredentialProfile{
		AccessKeyID:     keyID,
		AccessKeySecret: secret,
	}
}

// GetCredentials returns the credentials for a specific profile.
func GetCredentials(profile string) (CredentialProfile, bool) {
	configMu.RLock()
	defer configMu.RUnlock()

	if globalCredentials == nil {
		return CredentialProfile{}, false
	}

	cred, ok := globalCredentials.Profiles[profile]
	return cred, ok
}
