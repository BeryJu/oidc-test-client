package pkg

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
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

func strToBool(str string) bool {
	strbool := strings.ToLower(str)
	return strbool == "true"
}

func skipTLSVerify() bool {
	tlsVerify := strings.ToLower(Env("OIDC_TLS_VERIFY", "true"))
	return !strToBool(tlsVerify)
}

func createContext(from context.Context) context.Context {
	if skipTLSVerify() {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		httpClient := &http.Client{Transport: tr}
		return oidc.ClientContext(from, httpClient)
	}

	return from
}

func getScopes() []string {
	scopes := []string{oidc.ScopeOpenID, oidc.ScopeOfflineAccess, "profile", "email"}
	if es := os.Getenv("OIDC_SCOPES"); es != "" {
		scopes = strings.Split(es, ",")
	}
	return scopes
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
		Scopes:       getScopes(),
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
		err = userInfo.Claims(&uInfo)
		if err != nil {
			log.WithError(err).Error("failed to get claims from userinfo")
		}
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

		// skipping TLS verification can't be done with
		// the request context so use the clients
		ts := c.config.TokenSource(createContext(r.Context()), oauth2Token)
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
		log.WithError(err).Error("failed to marshal response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = w.Write(data)
	if err != nil {
		log.WithError(err).Error("failed to write response")
	}
}

func (c *OIDCClient) oauthInit(w http.ResponseWriter, r *http.Request) {
	// Get a session. We're ignoring the error resulted from decoding an
	// existing session: Get() always returns a session, even if empty.
	session, _ := c.store.Get(r, "session-name")
	state := base64.RawStdEncoding.EncodeToString(securecookie.GenerateRandomKey(32))
	session.Values["state"] = state
	err := session.Save(r, w)
	if err != nil {
		log.WithError(err).Warning("failed to save session")
		_, err = w.Write([]byte(err.Error()))
		if err != nil {
			log.WithError(err).Warning("failed to write error message during init")
		}
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
	baseUrl, err := url.Parse(Env("OIDC_ROOT_URL", "http://localhost:9009"))
	if err != nil {
		panic(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc(baseUrl.Path+"/implicit/", c.implicit)
	mux.Handle(baseUrl.Path+"/static/", http.StripPrefix("/static/", http.FileServer(http.FS(static))))
	mux.HandleFunc(baseUrl.Path+"/health", c.health)
	// Just to prevent favicon from triggering authorize
	mux.HandleFunc(baseUrl.Path+"/favicon.ico", c.health)
	mux.HandleFunc(baseUrl.Path+"/auth/callback", c.oauthCallback)
	mux.HandleFunc(baseUrl.Path+"/", c.oauthInit)

	listen := Env("OIDC_BIND", "localhost:9009")

	log.Printf("listening on http://%s/", listen)
	log.Fatal(http.ListenAndServe(listen, logRequest(mux)))
}
