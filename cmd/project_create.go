package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"charm.land/huh/v2"
	"github.com/hackerspace/hacktrackmmu-cli/internal/session"
	"github.com/spf13/cobra"
)

type Member struct {
	Members []MemberDetail `json:"members"`
}

type MemberDetails struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type CreateProjectPayload struct {
	Project ProjectPostDetails `json:"project"`
}

type ProjectPostDetails struct {
	Name      string `json:"name"`
	Category  string `json:"category"`
	MemberIDs []int  `json:"member_ids"`
	Completed bool   `json:"completed"`
}

var projectCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new project",
	Long:  `Create a new project by providing its details interactively or via flags.`,
	Run: func(cmd *cobra.Command, args []string) {
		apiURL := cfg.APIUrl + "/dashboard/create_project"
		client := &http.Client{Timeout: 10 * time.Second}

		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			return
		}

		if s, err := session.Load(); err == nil && s != nil {
			req.Header.Set("Authorization", "Bearer "+s.Token)
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

		var members Member
		if err := json.NewDecoder(resp.Body).Decode(&members); err != nil {
			fmt.Printf("Error decoding JSON response: %v\n", err)
			return
		}

		fmt.Println(colorize("================================================================================", ansiCyan))
		fmt.Println(colorize("                           CREATE A NEW PROJECT", ansiBoldCyan))
		fmt.Println(colorize("================================================================================", ansiCyan))

		var projectName string
		var selectedMemberIDs []int
		var projectCategory string

		memberOptions := make([]huh.Option[int], len(members.Members))
		for i, m := range members.Members {
			memberOptions[i] = huh.NewOption(m.Name, m.ID)
		}

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Project Name").
					Value(&projectName).
					Validate(func(str string) error {
						if len(strings.TrimSpace(str)) == 0 {
							return fmt.Errorf("project name cannot be empty")
						}
						return nil
					}),
				huh.NewMultiSelect[int]().
					Title("Members").
					Options(memberOptions...).
					Value(&selectedMemberIDs).
					Filterable(true),
				huh.NewSelect[string]().
					Title("Category").
					Options(
						huh.NewOption("Project", "project"),
						huh.NewOption("Group Project", "group_project"),
					).
					Value(&projectCategory),
			),
		)

		err = form.Run()
		if err != nil {
			fmt.Printf("Form error: %v\n", err)
			return
		}

		fmt.Printf("\nCreating project %s with selected members...\n", colorize(projectName, ansiBoldGreen))

		payload := CreateProjectPayload{
			Project: ProjectPostDetails{
				Name:      projectName,
				Category:  projectCategory,
				MemberIDs: selectedMemberIDs,
				Completed: false,
			},
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			fmt.Printf("Error marshaling payload: %v\n", err)
			return
		}

		postURL := cfg.APIUrl + "/projects"
		postReq, err := http.NewRequest("POST", postURL, bytes.NewBuffer(payloadBytes))
		if err != nil {
			fmt.Printf("Error creating POST request: %v\n", err)
			return
		}

		postReq.Header.Set("Content-Type", "application/json")
		if s, err := session.Load(); err == nil && s != nil {
			postReq.Header.Set("Authorization", "Bearer "+s.Token)
		}

		postResp, err := client.Do(postReq)
		if err != nil {
			fmt.Printf("Error sending POST request: %v\n", err)
			return
		}
		defer postResp.Body.Close()

		body, _ := io.ReadAll(postResp.Body)
		if postResp.StatusCode == http.StatusCreated || postResp.StatusCode == http.StatusOK {
			var successMsg string
			var apiResp struct {
				Message string `json:"message"`
			}
			if err := json.Unmarshal(body, &apiResp); err == nil && apiResp.Message != "" {
				successMsg = apiResp.Message
			} else {
				successMsg = "Project successfully created."
			}
			fmt.Printf(colorize("\nSuccess: %s\n", ansiBoldGreen), successMsg)
		} else {
			var errMsg string
			var apiResp struct {
				Message string `json:"message"`
			}
			if err := json.Unmarshal(body, &apiResp); err == nil && apiResp.Message != "" {
				errMsg = apiResp.Message
			} else {
				errMsg = string(body)
			}
			fmt.Printf(colorize("\nError: Failed to create project (Status %d): %s\n", ansiRed), postResp.StatusCode, errMsg)
		}
	},
}

func init() {
	projectCmd.AddCommand(projectCreateCmd)
}
