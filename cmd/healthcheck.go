package cmd

import (
	"fmt"
	"net/http"
	"os"

	"beryju.io/oidc-test-client/pkg"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var healthcheckCmd = &cobra.Command{
	Use: "healthcheck",
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(check())
	},
}

func check() int {
	root := pkg.Env("OIDC_ROOT_URL", "http://localhost:9009")
	url := fmt.Sprintf("%s/health", root)
	res, err := http.DefaultClient.Head(url)
	if err != nil {
		log.WithError(err).Warning("failed to send healthcheck request")
		return 1
	}
	if res.StatusCode >= 400 {
		log.WithField("status", res.StatusCode).Warning("unhealthy status code")
		return 1
	}
	log.Debug("successfully checked health")
	return 0
}

func init() {
	rootCmd.AddCommand(healthcheckCmd)
}
