/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"charm.land/huh/v2"
	"github.com/hackerspace/hacktrackmmu-cli/internal/session"
	"github.com/spf13/cobra"
)

type CreateUpdateFormData struct {
	Members []UpdateMember `json:"members"`
	Dates   []UpdateDate   `json:"dates"`
}

type UpdateMember struct {
	ID       int             `json:"id"`
	Name     string          `json:"name"`
	Projects []UpdateProject `json:"projects"`
}

type UpdateProject struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type UpdateDate struct {
	ID   int    `json:"id"`
	Date string `json:"date"`
}

type CreateUpdatePayload struct {
	Update UpdatePostDetails `json:"update"`
}

type UpdatePostDetails struct {
	MeetupID    int    `json:"meetup_id"`
	ProjectID   int    `json:"project_id"`
	MemberID    int    `json:"member_id"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

var updateCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new update for a member",
	Long:  `Create a new update by interactively selecting a member, project, meetup date, category, and entering a description.`,
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		apiURL := cfg.APIUrl + "/dashboard/create_update"
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

		var formData CreateUpdateFormData
		if err := json.NewDecoder(resp.Body).Decode(&formData); err != nil {
			fmt.Printf("Error decoding JSON response: %v\n", err)
			return
		}

		if len(formData.Members) == 0 {
			fmt.Println("No members available to create an update.")
			return
		}

		if len(formData.Dates) == 0 {
			fmt.Println("No meetup dates available to create an update.")
			return
		}

		fmt.Println(colorize("================================================================================", ansiCyan))
		fmt.Println(colorize("                         CREATE A NEW UPDATE", ansiBoldCyan))
		fmt.Println(colorize("================================================================================", ansiCyan))

		var selectedMemberID int

		memberOptions := make([]huh.Option[int], len(formData.Members))
		for i, m := range formData.Members {
			memberOptions[i] = huh.NewOption(m.Name, m.ID)
		}

		memberForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[int]().
					Title("Select Member").
					Options(memberOptions...).
					Value(&selectedMemberID),
			),
		)

		if err := memberForm.Run(); err != nil {
			fmt.Printf("Form error: %v\n", err)
			return
		}

		var selectedMember *UpdateMember
		for i := range formData.Members {
			if formData.Members[i].ID == selectedMemberID {
				selectedMember = &formData.Members[i]
				break
			}
		}

		if selectedMember == nil {
			fmt.Println("Error: selected member not found.")
			return
		}

		if len(selectedMember.Projects) == 0 {
			fmt.Printf("Error: %s has no projects assigned.\n", selectedMember.Name)
			return
		}

		var selectedProjectID int
		var selectedMeetupID int
		var selectedCategory string
		var description string

		projectOptions := make([]huh.Option[int], len(selectedMember.Projects))
		for i, p := range selectedMember.Projects {
			projectOptions[i] = huh.NewOption(p.Name, p.ID)
		}

		dateOptions := make([]huh.Option[int], len(formData.Dates))
		for i, d := range formData.Dates {
			dateOptions[i] = huh.NewOption(d.Date, d.ID)
		}

		detailsForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[int]().
					Title("Select Project").
					Options(projectOptions...).
					Value(&selectedProjectID),
				huh.NewSelect[int]().
					Title("Select Meetup Date").
					Options(dateOptions...).
					Value(&selectedMeetupID),
				huh.NewSelect[string]().
					Title("Category").
					Options(
						huh.NewOption("Idea Talk", "idea_talk"),
						huh.NewOption("Progress Talk", "progress_talk"),
					).
					Value(&selectedCategory),
				huh.NewText().
					Title("Description").
					Value(&description).
					Validate(func(str string) error {
						if len(str) == 0 {
							return fmt.Errorf("description cannot be empty")
						}
						return nil
					}),
			),
		)

		if err := detailsForm.Run(); err != nil {
			fmt.Printf("Form error: %v\n", err)
			return
		}

		fmt.Printf("\nCreating update for %s...\n", colorize(selectedMember.Name, ansiBoldGreen))

		payload := CreateUpdatePayload{
			Update: UpdatePostDetails{
				MeetupID:    selectedMeetupID,
				ProjectID:   selectedProjectID,
				MemberID:    selectedMemberID,
				Category:    selectedCategory,
				Description: description,
			},
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			fmt.Printf("Error marshaling payload: %v\n", err)
			return
		}

		postURL := cfg.APIUrl + "/updates"
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
				successMsg = "Update successfully created."
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
			fmt.Printf(colorize("\nError: Failed to create update (Status %d): %s\n", ansiRed), postResp.StatusCode, errMsg)
		}
	},
}

func init() {
	updateCmd.AddCommand(updateCreateCmd)
}
