package main

import (
	"os"

	"github.com/BeryJu/oidc-test-client/pkg"
	"github.com/cenkalti/backoff/v4"

	log "github.com/sirupsen/logrus"
)

var (
	clientID     = os.Getenv("OIDC_CLIENT_ID")
	clientSecret = os.Getenv("OIDC_CLIENT_SECRET")
	provider     = os.Getenv("OIDC_PROVIDER")
)

func main() {
	log.SetLevel(log.DebugLevel)
	attempt := 0
	maxAttempts := 20

	operation := func() error {
		log.WithField("attempt", attempt).Debug("Attempting to start...")
		client, err := pkg.NewOIDCClient(clientID, clientSecret, provider)
		if err != nil {
			attempt++
			return err
		}
		client.Run()
		return nil
	}

	err := backoff.Retry(operation, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), uint64(maxAttempts)))
	if err != nil {
		log.Fatal(err)
	}
}
