package config

import (
	"net/url"
	"strings"
)

// Config is used to configure the creation of a client
type Config struct {
	// The ID of the key
	ApplicationKeyID string

	// The secret part of the key
	ApplicationKey string

	// The base URL for authorization API call
	AuthorizationBaseURL *url.URL
}

// FromEnv returns default configuration based on environment variables
func FromEnv(env []string) *Config {
	authURL, _ := url.Parse("https://api.backblazeb2.com/")

	config := &Config{
		AuthorizationBaseURL: authURL,
	}

	// convert to a map so it's easier to work with
	envMap := toMap(env)

	if keyID, ok := envMap["B2_KEY_ID"]; ok {
		config.ApplicationKeyID = keyID
	}

	if keySecret, ok := envMap["B2_KEY_SECRET"]; ok {
		config.ApplicationKey = keySecret
	}

	return config
}

func toMap(env []string) map[string]string {
	m := make(map[string]string, len(env))
	var key, value string

	for _, s := range env {
		parts := strings.SplitN(s, "=", 2)

		key = strings.ToUpper(parts[0])
		value = parts[1]

		m[key] = value
	}

	return m
}
