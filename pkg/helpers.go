package pkg

import (
	"encoding/json"
	"os"

	"golang.org/x/oauth2"
)

func Env(key string, fallback string) string {
	value, exists := os.LookupEnv(key)
	if exists {
		return value
	}
	return fallback
}

type CallbackResponse struct {
	OAuth2Token    *oauth2.Token
	IDTokenClaims  *json.RawMessage // ID Token payload is just JSON.
	RawIDToken     string
	UserInfo       interface{}
	Introspection  interface{}
	Refresh        interface{}
	RefreshIDToken interface{}
}
