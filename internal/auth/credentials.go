package auth

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Build-time에 ldflags로 주입되는 OAuth2 client credentials.
// 소스 빌드 시에는 비어 있으며, GoReleaser가 빌드할 때 주입됩니다.
var (
	oauthClientID     string
	oauthClientSecret string
)

func hasEmbeddedCredentials() bool {
	return oauthClientID != "" && oauthClientSecret != ""
}

func embeddedOAuthConfig(scopes []string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     oauthClientID,
		ClientSecret: oauthClientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       scopes,
	}
}
