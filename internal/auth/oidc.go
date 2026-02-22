package auth

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type OIDCConfig struct {
	Provider     *oidc.Provider
	OAuth2Config oauth2.Config
	Verifier     *oidc.IDTokenVerifier
}

func NewOIDC(ctx context.Context, issuer, clientID, clientSecret, redirectURL string) (*OIDCConfig, error) {
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, fmt.Errorf("creating OIDC provider: %w", err)
	}

	oauth2Config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: clientID})

	return &OIDCConfig{
		Provider:     provider,
		OAuth2Config: oauth2Config,
		Verifier:     verifier,
	}, nil
}
