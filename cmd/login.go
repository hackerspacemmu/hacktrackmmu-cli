package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hackerspace/hacktrackmmu-cli/internal/session"
	"github.com/spf13/cobra"
)

type LoginRequest struct {
	Session LoginSession `json:"session"`
}

type LoginSession struct {
	Password   string `json:"password"`
	RememberMe bool   `json:"remember_me"`
}

type LoginResponse struct {
	Message    string    `json:"message"`
	Token      string    `json:"token"`
	IsAdmin    bool      `json:"isAdmin"`
	ValidUntil time.Time `json:"valid_until"`
}

var loginCmd = &cobra.Command{
	Use:   "login [password]",
	Short: "Authenticate with the Hacktrack MMU server",
	Long:  `Authenticate with the Hacktrack MMU server using your password.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("Usage: ht login [password]")
			return
		}

		cfg := GetConfig()
		password := args[0]
		fmt.Println("Logging in...")

		loginReq := LoginRequest{
			Session: LoginSession{
				Password:   password,
				RememberMe: true, // Keep token valid for 30 days
			},
		}

		jsonData, err := json.Marshal(loginReq)
		if err != nil {
			fmt.Printf("Error marshalling request payload: %v\n", err)
			return
		}

		apiURL := cfg.APIUrl + "/login"
		fmt.Println(apiURL)
		client := &http.Client{Timeout: 10 * time.Second}

		resp, err := client.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("Error making request to server: %v\n", err)
			return
		}
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Error reading response body: %v\n", err)
			return
		}

		if resp.StatusCode != http.StatusOK {
			// Handle error message
			var errResp struct {
				Message string `json:"message"`
			}
			if err := json.Unmarshal(bodyBytes, &errResp); err == nil && errResp.Message != "" {
				fmt.Printf("Login failed (status %d): %s\n", resp.StatusCode, errResp.Message)
			} else {
				fmt.Printf("Login failed with status: %d\n", resp.StatusCode)
			}
			return
		}

		var loginResp LoginResponse
		if err := json.Unmarshal(bodyBytes, &loginResp); err != nil {
			fmt.Printf("Error decoding login response: %v\n", err)
			return
		}

		// Save session
		s := &session.Session{
			Token:      loginResp.Token,
			ValidUntil: loginResp.ValidUntil,
			IsAdmin:    loginResp.IsAdmin,
		}

		if err := session.Save(s); err != nil {
			fmt.Printf("Failed to save session to disk: %v\n", err)
			return
		}

		fmt.Println("Success!")
		fmt.Printf("Message: %s\n", loginResp.Message)
		fmt.Printf("Valid until: %s\n", loginResp.ValidUntil.Format(time.RFC1123))
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
