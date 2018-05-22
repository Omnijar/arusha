package config

import (
	"errors"
	"net/url"
	"os"
	"time"
)

const (
	// EnvEmailVerificationURL env variable for the URL used for verifying emails.
	EnvEmailVerificationURL = "ARUSHA_EMAIL_VERIFY_URL"
	// EnvTokenVerificationURL env variable for the URL used for verifying tokens.
	EnvTokenVerificationURL = "ARUSHA_TOKEN_VERIFY_URL"
	// EnvArushaClusterURL env variable for setting Arusha's private URL.
	EnvArushaClusterURL = "ARUSHA_CLUSTER_URL"
	// EnvArushaClientCallbackURL is the callback URL for the client after verifying auth.
	EnvArushaClientCallbackURL = "ARUSHA_CLIENT_CALLBACK_URL"
)

var (
	// Default configuration for the service.
	Default = &Config{}
)

// Config stores the configuration state for Arusha.
type Config struct {
	BuildVersion            string `yaml:"-"`
	BuildHash               string `yaml:"-"`
	BuildTime               string `yaml:"-"`
	EmailVerificationURL    string
	TokenVerificationURL    string
	ArushaClusterURL        string
	ArushaClientCallbackURL string
}

// Initialize the configuration of the service.
func Initialize() error {
	Default.BuildVersion = "dev-master"
	Default.BuildTime = time.Now().String()
	Default.BuildHash = "undefined"

	v := os.Getenv(EnvEmailVerificationURL)
	if _, err := url.Parse(v); err != nil || v == "" {
		return errors.New(EnvEmailVerificationURL + " variable not configured or invalid")
	}

	Default.EmailVerificationURL = v

	v = os.Getenv(EnvTokenVerificationURL)
	if _, err := url.Parse(v); err != nil || v == "" {
		return errors.New(EnvTokenVerificationURL + " variable not configured or invalid")
	}

	Default.TokenVerificationURL = v

	v = os.Getenv(EnvArushaClusterURL)
	if _, err := url.Parse(v); err != nil || v == "" {
		return errors.New(EnvArushaClusterURL + " variable not configured or invalid")
	}

	Default.ArushaClusterURL = v

	v = os.Getenv(EnvArushaClientCallbackURL)
	if _, err := url.Parse(v); err != nil || v == "" {
		return errors.New(EnvArushaClientCallbackURL + " variable not configured or invalid")
	}

	Default.ArushaClientCallbackURL = v

	return nil
}
