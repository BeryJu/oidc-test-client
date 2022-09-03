package cmd

import (
	"fmt"
	"net/http"

	"github.com/cli/oauth/device"
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

		code, err := device.RequestCode(httpClient, deviceUrl, clientId, scopes)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Copy code: %s\n", code.UserCode)
		fmt.Printf("then open: %s\n", code.VerificationURI)

		accessToken, err := device.PollToken(httpClient, codeUrl, clientId, code)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Access token: %s\n", accessToken.Token)
	},
}

func init() {
	deviceCmd.PersistentFlags().StringVarP(&clientId, "client-id", "c", "", "Client ID")
	deviceCmd.PersistentFlags().StringVarP(&deviceUrl, "device-url", "d", "", "Device URL")
	deviceCmd.PersistentFlags().StringVarP(&codeUrl, "code-url", "u", "", "Code URL")
	deviceCmd.PersistentFlags().StringSliceVarP(&scopes, "scopes", "s", []string{}, "Scopes")
	rootCmd.AddCommand(deviceCmd)
}
