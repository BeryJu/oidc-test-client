package main

import (
	"os"

	"beryju.org/oidc-test-client/pkg"

	log "github.com/sirupsen/logrus"
)

var (
	clientID     = os.Getenv("OIDC_CLIENT_ID")
	clientSecret = os.Getenv("OIDC_CLIENT_SECRET")
	provider     = os.Getenv("OIDC_PROVIDER")
)

func main() {
	log.SetLevel(log.DebugLevel)

	client := pkg.NewOIDCClient(clientID, clientSecret, provider)
	client.Run()
}
