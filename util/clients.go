package util

import "errors"

// InitializeClients initializes all clients used by this service.
func InitializeClients() error {
	if err := InitializeMailgunClient(); err != nil {
		return errors.New("clients: error initializing mailgun. " + err.Error())
	}

	if err := InitializeVaultClient(); err != nil {
		return errors.New("clients: error initializing vault. " + err.Error())
	}

	if err := VerifyHydraEndpoint(); err != nil {
		return errors.New("clients: error verifying hydra. " + err.Error())
	}

	if err := InitializeKetoClient(); err != nil {
		return errors.New("clients: error initializing keto. " + err.Error())
	}

	return nil
}
