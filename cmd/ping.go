/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "ping hacktrack api server",
	Long: `ping hacktrack api server before running other requests.
recommended to use when server in sleep and having a cold start.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("ping called")

		apiURL := "https://hacktrackmmu.herokuapp.com/healthcheck"
		client := &http.Client{Timeout: 10 * time.Second}

		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			return
		}

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error making request: %v\n", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			fmt.Printf("Error: Received status code %d from API: %s\n", resp.StatusCode, string(body))
			return
		}

		fmt.Println("ping successfull")
	},
}

func init() {
	rootCmd.AddCommand(pingCmd)
}
