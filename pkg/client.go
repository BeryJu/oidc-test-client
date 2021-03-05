package pkg

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	oidc "github.com/coreos/go-oidc"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	log "github.com/sirupsen/logrus"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

type OIDCClient struct {
	provider    *oidc.Provider
	providerURL string
	verifier    *oidc.IDTokenVerifier
	rootURL     string

	config *oauth2.Config

	store *sessions.CookieStore

	doRefreshChecks       bool
	doIntrospectionChecks bool
	doUserInfoChecks      bool

	ctx context.Context
}

func NewOIDCClient(clientID string, clientSecret string, providerURL string) *OIDCClient {
	ctx := context.Background()

	provider, err := oidc.NewProvider(ctx, providerURL)
	if err != nil {
		log.Fatal(err)
	}

	rootURL := Env("OIDC_ROOT_URL", "http://localhost:9009")

	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  fmt.Sprintf("%s/auth/callback", rootURL),
		Scopes:       []string{oidc.ScopeOpenID, oidc.ScopeOfflineAccess, "profile", "email"},
	}

	client := OIDCClient{
		rootURL:     rootURL,
		config:      config,
		ctx:         ctx,
		provider:    provider,
		providerURL: providerURL,
		store:       sessions.NewCookieStore(securecookie.GenerateRandomKey(32)),
		verifier: provider.Verifier(&oidc.Config{
			ClientID: clientID,
		}),
		doRefreshChecks:       strings.ToLower(Env("OIDC_DO_REFRESH", "true")) == "true",
		doIntrospectionChecks: strings.ToLower(Env("OIDC_DO_INTROSPECTION", "true")) == "true",
		doUserInfoChecks:      strings.ToLower(Env("OIDC_DO_USER_INFO", "true")) == "true",
	}
	return &client
}

func (c *OIDCClient) oauthCallback(w http.ResponseWriter, r *http.Request) {
	session, _ := c.store.Get(r, "session-name")

	if r.URL.Query().Get("state") != session.Values["state"] {
		log.Error("state did not match")
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	oauth2Token, err := c.config.Exchange(c.ctx, r.URL.Query().Get("code"))
	if err != nil {
		log.WithError(err).Error("Failed to exchange token")
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	tokenSource := oauth2.StaticTokenSource(oauth2Token)

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		log.Error("No id_token field in oauth2 token.")
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	idToken, err := c.verifier.Verify(c.ctx, rawIDToken)
	if err != nil {
		log.WithError(err).Error("Failed to verify ID Token")
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	resp := CallbackResponse{
		OAuth2Token:   oauth2Token,
		IDTokenClaims: new(json.RawMessage),
	}

	if err := idToken.Claims(&resp.IDTokenClaims); err != nil {
		log.WithError(err).Error("Failed to get claims from ID Token")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// UserInfo Checks
	if c.doUserInfoChecks {
		userInfo, err := c.provider.UserInfo(c.ctx, tokenSource)
		if err != nil {
			log.WithError(err).Error("Failed to get userinfo")
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		var uInfo interface{}
		userInfo.Claims(&uInfo)
		resp.UserInfo = uInfo
	}

	// Introspection checks
	if c.doIntrospectionChecks {
		introspection, err := c.oauthTokenIntrospection(tokenSource)
		if err != nil {
			log.WithError(err).Error("Failed to do token introspection")
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		resp.Introspection = introspection
	}

	// check refresh token
	if c.doRefreshChecks {
		// force token expiry
		oauth2Token.Expiry = time.Now()
		ts := c.config.TokenSource(r.Context(), oauth2Token)
		refresh, err := ts.Token()
		if err != nil {
			log.WithError(err).Warning("Failed to refresh token")
		}

		refreshRawIDToken, ok := refresh.Extra("id_token").(string)
		if !ok {
			log.Warning("No id_token field in refresh oauth2 token.")
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		refreshIDToken, err := c.verifier.Verify(c.ctx, refreshRawIDToken)
		if err != nil {
			log.WithError(err).Warning("Failed to verify ID Token in refresh token")
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		resp.Refresh = refresh
		resp.RefreshIDToken = refreshIDToken
	}

	data, err := json.MarshalIndent(resp, "", "    ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func (c *OIDCClient) oauthInit(w http.ResponseWriter, r *http.Request) {
	// Get a session. We're ignoring the error resulted from decoding an
	// existing session: Get() always returns a session, even if empty.
	session, _ := c.store.Get(r, "session-name")
	state := base64.RawStdEncoding.EncodeToString(securecookie.GenerateRandomKey(32))
	session.Values["state"] = state
	err := session.Save(r, w)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	http.Redirect(w, r, c.config.AuthCodeURL(state), http.StatusFound)
}

func (c *OIDCClient) health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	fmt.Fprint(w, "hello :)")
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.WithField("remoteAddr", r.RemoteAddr).WithField("method", r.Method).Info(r.URL)
		handler.ServeHTTP(w, r)
	})
}

func (c *OIDCClient) Run() {
	mux := http.NewServeMux()
	mux.HandleFunc("/implicit/", c.implicit)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(static))))
	mux.HandleFunc("/health", c.health)
	// Just to prevent favicon from triggering authorize
	mux.HandleFunc("/favicon.ico", c.health)
	mux.HandleFunc("/auth/callback", c.oauthCallback)
	mux.HandleFunc("/", c.oauthInit)

	listen := Env("OIDC_BIND", "localhost:9009")

	log.Printf("listening on http://%s/", listen)
	log.Fatal(http.ListenAndServe(listen, logRequest(mux)))
}
