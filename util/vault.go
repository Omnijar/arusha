package util

import (
	"log"

	vault "github.com/hashicorp/vault/api"
	"github.com/mitchellh/mapstructure"
)

const (
	rootPath = "/secret/arusha"
)

var (
	vaultClient *vault.Client
)

// VaultClient for storing secrets in vault.
type VaultClient struct {
	path   string
	client *vault.Client
}

// Set value for a given key
func (v *VaultClient) Set(key string, value interface{}) {
	var secret = make(map[string]interface{})
	secret["data"] = value

	path := v.path + "/" + key
	_, err := v.client.Logical().Write(path, secret)

	if err != nil {
		log.Println("vault: Failed to write key " + key + ": " + err.Error())
	}
}

// Get the value for a key
func (v *VaultClient) Get(key string, value interface{}) bool {
	path := v.path + "/" + key
	secret, err := v.client.Logical().Read(path)
	if err != nil {
		log.Println("vault: Failed to fetch value for key " + key + ": " + err.Error())
	}

	if secret == nil {
		return false
	}

	if err := mapstructure.Decode(secret.Data["data"], value); err != nil {
		log.Println("vault: Failed to decode value for key " + key + ": " + err.Error())
		return false
	}

	return true
}

// Remove the value corresponding to a key.
func (v *VaultClient) Remove(key string) {
	path := v.path + "/" + key
	if _, err := v.client.Logical().Delete(path); err != nil {
		log.Println("vault: Error removing value for key " + key + ": " + err.Error())
	}
}

// InitializeVaultClient for route handlers.
func InitializeVaultClient() error {
	config := vault.DefaultConfig()
	var err error

	if err = config.ReadEnvironment(); err != nil {
		return err
	}

	vaultClient, err = vault.NewClient(config)
	if err != nil {
		return err
	}

	return nil
}

// GetVaultClient for use in other controllers.
func GetVaultClient(path string) *VaultClient {
	return &VaultClient{
		path:   rootPath + path,
		client: vaultClient,
	}
}
