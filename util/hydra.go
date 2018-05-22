package util

import (
	"context"
	"errors"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/ory/hydra/sdk/go/hydra"
	hydraAPI "github.com/ory/hydra/sdk/go/hydra/swagger"
	"gitlab.com/omnijar/arusha/config"
	"golang.org/x/oauth2"
)

const (
	// EnvHydraPublicURL for hydra's publicly accessible endpoint.
	EnvHydraPublicURL = "HYDRA_PUBLIC_URL"
	// EnvHydraPrivateURL for setting hydra's internal URL.
	EnvHydraPrivateURL = "HYDRA_PRIVATE_URL"
	// SessionPeriodSeconds to remember a login.
	// FIXME: Move this to command.
	SessionPeriodSeconds = 8 * 3600
)

var (
	hydraPrivateEndpoint string
	hydraPublicEndpoint  string
	hydraClient          *hydra.CodeGenSDK
	hydraOAuth2Client    *hydraAPI.OAuth2Client
	oauth2Config         *oauth2.Config
)

// VerifyHydraEndpoint for communicating with hydra.
func VerifyHydraEndpoint() error {
	endpoint := os.Getenv(EnvHydraPrivateURL)
	_, err := url.Parse(endpoint)
	if err != nil || endpoint == "" {
		return errors.New(EnvHydraPrivateURL + " variable not configured or invalid")
	}

	hydraPrivateEndpoint = strings.TrimRight(endpoint, "/")

	endpoint = os.Getenv(EnvHydraPublicURL)
	_, err = url.Parse(endpoint)
	if err != nil || endpoint == "" {
		return errors.New(EnvHydraPublicURL + " variable not configured or invalid")
	}

	hydraPublicEndpoint = strings.TrimRight(endpoint, "/")

	return err
}

// InitializeRootHydraClient for use by controller.
func InitializeRootHydraClient(scopes []string) error {
	log.Printf("hydra: creating root client for self with scopes %v", scopes)

	// FIXME: This always creates a new client.

	api := hydraAPI.NewOAuth2ApiWithBasePath(hydraPrivateEndpoint)
	clientConfig := hydraAPI.OAuth2Client{
		Id:            "arusha-root",
		ClientSecret:  "",
		ResponseTypes: []string{"code", "id_token"},
		Scope:         strings.Join(scopes, " "),
		GrantTypes:    []string{"authorization_code", "client_credentials"},
		RedirectUris:  []string{config.Default.ArushaClientCallbackURL},
		ClientName:    "arusha",
		Public:        false,
	}

	// Delete existing client
	_, _ = api.DeleteOAuth2Client("arusha-root")

	var err error
	hydraOAuth2Client, _, err = api.CreateOAuth2Client(clientConfig)
	if err != nil {
		return errors.New("hydra: error creating client. " + err.Error())
	}

	hydraClient, err = hydra.NewSDK(&hydra.Configuration{
		EndpointURL:  hydraPrivateEndpoint,
		ClientID:     hydraOAuth2Client.Id,
		ClientSecret: hydraOAuth2Client.ClientSecret,
		Scopes:       scopes,
	})

	if err != nil {
		return errors.New("hydra: error initializing client. " + err.Error())
	}

	log.Printf("hydra: created client (id: %s) with scopes %v", hydraOAuth2Client.Id, scopes)

	oauth2Config = &oauth2.Config{
		ClientID:     hydraOAuth2Client.Id,
		ClientSecret: hydraOAuth2Client.ClientSecret,
		Endpoint: oauth2.Endpoint{
			TokenURL: hydraPrivateEndpoint + "/oauth2/token",
			AuthURL:  hydraPublicEndpoint + "/oauth2/auth",
		},
		RedirectURL: config.Default.ArushaClientCallbackURL,
		Scopes:      scopes,
	}

	return nil
}

// GetAuthURL for a session.
func GetAuthURL() (*string, error) {
	if oauth2Config == nil {
		return nil, ErrorOAuthNotInitialized
	}

	state := RandomAlphaNumeric(24)
	nonce := RandomAlphaNumeric(24)
	u := oauth2Config.AuthCodeURL(state) + "&nonce=" + nonce

	return &u, nil
}

// HydraRedirectResponse has the subject associated with this session and the redirect URL (if any)
// for performing any redirects.
type HydraRedirectResponse struct {
	Subject     string `json:"subject"`
	RedirectURL string `json:"redirectURL"`
}

// GetLoginRequest for the given challenge ID. This fetches the request and reuses
func GetLoginRequest(challenge string) (*HydraRedirectResponse, error) {
	if hydraClient == nil {
		return nil, ErrorOAuthNotInitialized
	}

	loginRequest, _, err := hydraClient.OAuth2Api.GetLoginRequest(challenge)
	if err != nil {
		log.Printf("hydra: error getting login request. " + err.Error())
		return nil, ErrorOAuthFetch
	}

	if loginRequest.Skip {
		completion, _, err := hydraClient.OAuth2Api.AcceptLoginRequest(challenge, hydraAPI.AcceptLoginRequest{
			Subject: loginRequest.Subject,
		})

		if err != nil {
			log.Printf("hydra: error accepting login request. " + err.Error())
			return nil, ErrorOAuthFetch
		}

		return &HydraRedirectResponse{
			Subject:     loginRequest.Subject,
			RedirectURL: completion.RedirectTo,
		}, nil
	}

	return nil, nil
}

// AcceptLoginRequest for a given challenge with the given subject (which is the user's ID).
func AcceptLoginRequest(challenge, subject string) (*HydraRedirectResponse, error) {
	redirect, err := GetLoginRequest(challenge)
	if redirect != nil || err != nil {
		return redirect, err
	}

	completion, _, err := hydraClient.OAuth2Api.AcceptLoginRequest(challenge, hydraAPI.AcceptLoginRequest{
		Subject:     subject,
		Remember:    true,
		RememberFor: SessionPeriodSeconds,
	})

	if err != nil {
		log.Printf("hydra: error accepting login request. " + err.Error())
		return nil, ErrorOAuthFetch
	}

	return &HydraRedirectResponse{
		Subject:     subject,
		RedirectURL: completion.RedirectTo,
	}, nil
}

// RejectLoginRequest for a given challenge (i.e., if the login fails).
func RejectLoginRequest(challenge, reason string) (*HydraRedirectResponse, error) {
	completion, _, err := hydraClient.OAuth2Api.RejectLoginRequest(challenge, hydraAPI.RejectRequest{
		ErrorDescription: reason,
	})

	if err != nil {
		log.Printf("hydra: error rejecting login request. " + err.Error())
		return nil, ErrorOAuthFetch
	}

	return &HydraRedirectResponse{
		Subject:     "",
		RedirectURL: completion.RedirectTo,
	}, nil
}

// BlindlyAcceptConsentRequest for the given challenge.
func BlindlyAcceptConsentRequest(challenge string) (*HydraRedirectResponse, error) {
	if hydraClient == nil {
		return nil, ErrorOAuthNotInitialized
	}

	consentRequest, _, err := hydraClient.OAuth2Api.GetConsentRequest(challenge)
	if err != nil {
		log.Printf("hydra: error getting consent request. " + err.Error())
		return nil, ErrorOAuthFetch
	}

	if consentRequest.Skip {
		completion, _, err := hydraClient.OAuth2Api.AcceptConsentRequest(challenge, hydraAPI.AcceptConsentRequest{
			GrantScope: oauth2Config.Scopes,
		})

		if err != nil {
			log.Printf("hydra: error accepting consent request. " + err.Error())
			return nil, ErrorOAuthFetch
		}

		return &HydraRedirectResponse{
			Subject:     consentRequest.Subject,
			RedirectURL: completion.RedirectTo,
		}, nil
	}

	completion, _, err := hydraClient.OAuth2Api.AcceptConsentRequest(challenge, hydraAPI.AcceptConsentRequest{
		GrantScope:  oauth2Config.Scopes,
		Remember:    true,
		RememberFor: 0,
	})

	if err != nil {
		log.Printf("hydra: error accepting consent request. " + err.Error())
		return nil, ErrorOAuthFetch
	}

	return &HydraRedirectResponse{
		Subject:     consentRequest.Subject,
		RedirectURL: completion.RedirectTo,
	}, nil
}

// HydraSessionToken has the access and refresh tokens obtained from the auth code.
type HydraSessionToken struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// GetToken (access and refresh tokens) for the given authorization code.
func GetToken(code string) (*HydraSessionToken, error) {
	if hydraClient == nil {
		return nil, ErrorOAuthNotInitialized
	}

	ctx := context.Background()
	tokenData, err := oauth2Config.Exchange(ctx, code)
	if err != nil {
		log.Printf("hydra: error exchanging code for token: %s", err)
		return nil, errors.New("error getting auth token")
	}

	return &HydraSessionToken{
		AccessToken:  tokenData.AccessToken,
		RefreshToken: tokenData.RefreshToken,
	}, nil
}

// RevokeToken to invalidate refresh and access tokens. Invalidating one will invalidate the other too.
func RevokeToken(token string) error {
	if hydraClient == nil {
		return ErrorOAuthNotInitialized
	}

	_, err := hydraClient.OAuth2Api.RevokeOAuth2Token(token)
	if err != nil {
		log.Printf("hydra: error revoking token: %s", err)
		return errors.New("error revoking token")
	}

	return nil
}

// AuthorizeToken to identify the subject.
func AuthorizeToken(token string) (*string, error) {
	if token == "" {
		return nil, ErrorInvalidToken
	}

	if hydraClient == nil {
		return nil, ErrorOAuthNotInitialized
	}

	data, _, err := hydraClient.OAuth2Api.IntrospectOAuth2Token(token, strings.Join(oauth2Config.Scopes, " "))
	if !data.Active {
		return nil, ErrorInvalidToken
	}

	if err != nil {
		log.Printf("hydra: error introspecting token: %s", err)
		return nil, errors.New("error checking token")
	}

	return &data.Sub, nil
}
