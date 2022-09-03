package cmd

import (
	"os"

	"beryju.org/oidc-test-client/pkg"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "oidc-test-client",
	Short: "A tool to test various OAuth/OIDC authentication flows",
	Run: func(cmd *cobra.Command, args []string) {
		clientID := os.Getenv("OIDC_CLIENT_ID")
		clientSecret := os.Getenv("OIDC_CLIENT_SECRET")
		provider := os.Getenv("OIDC_PROVIDER")

		client := pkg.NewOIDCClient(clientID, clientSecret, provider)
		client.Run()
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	log.SetLevel(log.DebugLevel)
}
