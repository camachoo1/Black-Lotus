package common

import (
	"fmt"
	"net/url"
	"os"
)

// GetAuthorizationURL returns the URL to redirect users for OAuth login
func GetAuthorizationURL(provider string, redirectURI string, state string) string {
	switch provider {
	case "github":
		return fmt.Sprintf(
			"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=user:email&state=%s",
			os.Getenv("GITHUB_CLIENT_ID"),
			url.QueryEscape(redirectURI),
			url.QueryEscape(state),
		)
	case "google":
		return fmt.Sprintf(
			"https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=email%%20profile&state=%s",
			os.Getenv("GOOGLE_CLIENT_ID"),
			url.QueryEscape(redirectURI),
			url.QueryEscape(state),
		)
	default:
		return ""
	}
}
