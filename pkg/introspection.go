package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

func doRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	client := http.DefaultClient
	if c, ok := ctx.Value(oauth2.HTTPClient).(*http.Client); ok {
		client = c
	}
	return client.Do(req.WithContext(ctx))
}

func (c *OIDCClient) oauthTokenIntrospection(tokenSource oauth2.TokenSource) (interface{}, error) {
	var ec struct {
		RevocationEndpoint    string `json:"revocation_endpoint"`
		IntrospectionEndpoint string `json:"introspection_endpoint"`
	}
	err := c.provider.Claims(&ec)
	if err != nil {
		return nil, err
	}

	token, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("oidc: get access token: %v", err)
	}

	form := url.Values{}
	form.Add("token", token.AccessToken)

	req, err := http.NewRequest("POST", ec.IntrospectionEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	token.SetAuthHeader(req)
	resp, err := doRequest(c.ctx, req)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.WithError(err).Warning("failed to close body")
		}
	}()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %s", resp.Status, body)
	}

	var introspection interface{}
	if err := json.Unmarshal(body, &introspection); err != nil {
		return nil, fmt.Errorf("oidc: failed to decode introspection: %v", err)
	}
	return &introspection, nil
}
