/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hackerspace/hacktrackmmu-cli/internal/session"
	"github.com/spf13/cobra"
)

type MeetupDetail struct {
	Meetup MeetupFormDetails `json:"meetup"`
	// MeetupNumber 	int					`json:"meetup_number"`
	Hosts Host `json:"hosts"`
}

type MeetupFormDetails struct {
	ID              int    `json:"id"`
	Date            string `json:"date"`
	Number          int    `json:"number"`
	HackathonNumber int    `json:"hackathon_number"`
	Category        string `json:"category"`
	HostID          int    `json:"host_id"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

type Host struct {
	YetTohost  []HostDetails `json:"Yet To Host"`
	HaveHosted []HostDetails `json:"Have Hosted"`
}

type HostDetails struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// this method is needed because the api is rendering the json in a weird way. exp: [name, id]
func (h *HostDetails) UnmarshalJSON(data []byte) error {
	var arr []interface{}
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}
	if len(arr) < 2 {
		return fmt.Errorf("invalid host details format: expected [name, id]")
	}
	name, ok := arr[0].(string)
	if !ok {
		return fmt.Errorf("invalid host details format: name must be a string")
	}
	idVal, ok := arr[1].(float64)
	if !ok {
		return fmt.Errorf("invalid host details format: id must be a number")
	}
	h.Name = name
	h.ID = int(idVal)
	return nil
}

type CreateMeetupPayload struct {
	Meetup MeetupPostDetails `json:"meetup"`
}

type MeetupPostDetails struct {
	Date            string `json:"date"`
	Category        string `json:"category"`
	Number          *int   `json:"number,omitempty"`
	HackathonNumber *int   `json:"hackathon_number,omitempty"`
	HostID          int    `json:"host_id"`
}

var meetupCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new meetup interactively",
	Long:  `Create a new meetup by interactively choosing the category, date, meetup number, and host.`,
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		apiURL := cfg.APIUrl + "/dashboard/create_meetup"
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

		var meetup MeetupDetail
		if err := json.NewDecoder(resp.Body).Decode(&meetup); err != nil {
			fmt.Printf("Error decoding JSON response: %v\n", err)
			return
		}

		reader := bufio.NewReader(os.Stdin)

		fmt.Println(colorize("================================================================================", ansiCyan))
		fmt.Println(colorize("                           CREATE A NEW MEETUP", ansiBoldCyan))
		fmt.Println(colorize("================================================================================", ansiCyan))

		fmt.Println(colorize("\nSelect Meetup Category:", ansiBoldCyan))
		fmt.Println("  [1] Regular Meetup")
		fmt.Println("  [2] Hackathon")

		var category string
		for {
			choice, err := promptString(reader, colorize("Enter choice (1-2)", ansiBoldYellow), "1")
			if err != nil {
				fmt.Printf("Error reading input: %v\n", err)
				return
			}
			switch choice {
			case "1":
				category = "regular_meetup"
			case "2":
				category = "hackathon"
			default:
				fmt.Print(colorize("Invalid choice. Please enter 1 or 2.\n", ansiRed))
				continue
			}
			break
		}

		var meetupDate string
		for {
			dateStr, err := promptString(reader, colorize("Enter Date (YYYY-MM-DD)", ansiBoldYellow), time.Now().Format("2006-01-02"))
			if err != nil {
				fmt.Printf("Error reading input: %v\n", err)
				return
			}
			if _, err := time.Parse("2006-01-02", dateStr); err != nil {
				fmt.Print(colorize("Invalid date format. Please use YYYY-MM-DD.\n", ansiRed))
				continue
			}
			meetupDate = dateStr
			break
		}

		var meetupNumber int
		var hackathonNumber int
		var useNumber, useHackathonNumber bool

		if category == "regular_meetup" {
			defaultNum := meetup.Meetup.Number
			for {
				numStr, err := promptString(reader, colorize("Enter Meetup Number", ansiBoldYellow), strconv.Itoa(defaultNum))
				if err != nil {
					fmt.Printf("Error reading input: %v\n", err)
					return
				}
				num, err := strconv.Atoi(numStr)
				if err != nil || num <= 0 {
					fmt.Print(colorize("Invalid number. Please enter a positive integer.\n", ansiRed))
					continue
				}
				meetupNumber = num
				useNumber = true
				break
			}
		} else if category == "hackathon" {
			defaultNum := meetup.Meetup.HackathonNumber
			for {
				numStr, err := promptString(reader, colorize("Enter Hackathon Number", ansiBoldYellow), strconv.Itoa(defaultNum))
				if err != nil {
					fmt.Printf("Error reading input: %v\n", err)
					return
				}
				num, err := strconv.Atoi(numStr)
				if err != nil || num <= 0 {
					fmt.Print(colorize("Invalid number. Please enter a positive integer.\n", ansiRed))
					continue
				}
				hackathonNumber = num
				useHackathonNumber = true
				break
			}
		}

		fmt.Println(colorize("\nSelect Host:", ansiBoldCyan))
		var hostChoices []HostDetails

		fmt.Println(colorize("--- Yet To Host ---", ansiBoldWhite))
		for _, h := range meetup.Hosts.YetTohost {
			hostChoices = append(hostChoices, h)
			fmt.Printf("  [%2d] %s\n", len(hostChoices), h.Name)
		}

		fmt.Println(colorize("\n--- Have Hosted ---", ansiBoldWhite))
		for _, h := range meetup.Hosts.HaveHosted {
			hostChoices = append(hostChoices, h)
			fmt.Printf("  [%2d] %s\n", len(hostChoices), h.Name)
		}

		var selectedHost HostDetails
		for {
			choiceStr, err := promptString(reader, colorize(fmt.Sprintf("Enter host number (1-%d)", len(hostChoices)), ansiBoldYellow), "")
			if err != nil {
				fmt.Printf("Error reading input: %v\n", err)
				return
			}
			if choiceStr == "" {
				fmt.Print(colorize("A host must be selected.\n", ansiRed))
				continue
			}
			choiceIdx, err := strconv.Atoi(choiceStr)
			if err != nil || choiceIdx < 1 || choiceIdx > len(hostChoices) {
				fmt.Print(colorize(fmt.Sprintf("Invalid choice. Please enter a number between 1 and %d.\n", len(hostChoices)), ansiRed))
				continue
			}
			selectedHost = hostChoices[choiceIdx-1]
			break
		}

		fmt.Printf("\nCreating meetup with host: %s...\n", colorize(selectedHost.Name, ansiBoldGreen))

		payload := CreateMeetupPayload{
			Meetup: MeetupPostDetails{
				Date:     meetupDate,
				Category: category,
				HostID:   selectedHost.ID,
			},
		}
		if useNumber {
			payload.Meetup.Number = &meetupNumber
		}
		if useHackathonNumber {
			payload.Meetup.HackathonNumber = &hackathonNumber
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			fmt.Printf("Error marshaling payload: %v\n", err)
			return
		}

		postURL := cfg.APIUrl + "/meetups"
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
				successMsg = "Meetup successfully created."
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
			fmt.Printf(colorize("\nError: Failed to create meetup (Status %d): %s\n", ansiRed), postResp.StatusCode, errMsg)
		}
	},
}

func promptString(reader *bufio.Reader, prompt string, defaultVal string) (string, error) {
	if defaultVal != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultVal)
	} else {
		fmt.Printf("%s: ", prompt)
	}
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultVal, nil
	}
	return input, nil
}

func init() {
	meetupCmd.AddCommand(meetupCreateCmd)
}
