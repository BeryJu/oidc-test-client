package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/cli/oauth"
	"github.com/spf13/cobra"
)

var (
	deviceUrl string
	codeUrl   string
	clientId  string
	scopes    []string
)

// deviceCmd represents the device command
var deviceCmd = &cobra.Command{
	Use:   "device",
	Short: "Test OAuth device flow",
	Run: func(cmd *cobra.Command, args []string) {
		httpClient := http.DefaultClient

		flow := &oauth.Flow{
			Host: &oauth.Host{
				DeviceCodeURL: deviceUrl,
				TokenURL:      codeUrl,
			},
			ClientID:   clientId,
			Scopes:     scopes,
			HTTPClient: httpClient,
		}

		accessToken, err := flow.DeviceFlow()
		if err != nil {
			panic(err)
		}

		fmt.Printf("Access token: %s\n", accessToken.Token)
	},
}

func init() {
	deviceCmd.PersistentFlags().StringVarP(&clientId, "client-id", "c", os.Getenv("OIDC_CLIENT_ID"), "Client ID")
	deviceCmd.PersistentFlags().StringVarP(&deviceUrl, "device-url", "d", "", "Device URL")
	deviceCmd.PersistentFlags().StringVarP(&codeUrl, "code-url", "u", "", "Code URL")
	deviceCmd.PersistentFlags().StringSliceVarP(&scopes, "scopes", "s", []string{}, "Scopes")
	rootCmd.AddCommand(deviceCmd)
}
