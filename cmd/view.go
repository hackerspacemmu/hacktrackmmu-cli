package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hackerspace/hacktrackmmu-cli/internal/session"
	"github.com/spf13/cobra"
)

type MemberDetail struct {
	ID                   int           `json:"id"`
	Name                 string        `json:"name"`
	Email                string        `json:"email"`
	Active               bool          `json:"active"`
	Status               string        `json:"status"`
	Comment              string        `json:"comment"`
	RegisterDate         string        `json:"register_date"`
	ContactNumber        string        `json:"contact_number"`
	DiscordTag           string        `json:"discord_tag"`
	ProgressTalkNum      int           `json:"progress_talk_num"`
	SpreadsheetID        string        `json:"spreadsheet_id"`
	StudentID            string        `json:"student_id"`
	DurationActive       string        `json:"duration_active"`
	AvgTimeBetweenTalks  string        `json:"avg_time_between_talks"`
	MeetupsSinceLastTalk int           `json:"meetups_since_last_talk"`
	Projects             []ProjectInfo `json:"projects"`
}

type ProjectInfo struct {
	ID        int          `json:"id"`
	Name      string       `json:"name"`
	Category  string       `json:"category"`
	Completed bool         `json:"completed"`
	Updates   []UpdateInfo `json:"updates"`
}

type UpdateInfo struct {
	ID          int        `json:"id"`
	Category    string     `json:"category"`
	Description string     `json:"description"`
	Meetup      MeetupInfo `json:"meetup"`
}

type MeetupInfo struct {
	ID     int    `json:"id"`
	Date   string `json:"date"`
	Number int    `json:"number"`
}

var viewCmd = &cobra.Command{
	Use:   "view [name]",
	Short: "view a member details",
	Long:  `view a member details`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("Wrong number of arguments. Usage: hacktrackmmu-cli view [name]")
			return
		}

		query := args[0]
		fmt.Printf("Searching for member: %s...\n", query)

		apiURL := cfg.APIUrl + fmt.Sprintf("/members/search?query=%s", url.QueryEscape(query))
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

		var members []MemberDetail
		if err := json.NewDecoder(resp.Body).Decode(&members); err != nil {
			fmt.Printf("Error decoding JSON response: %v\n", err)
			return
		}

		if len(members) == 0 {
			fmt.Printf("No members found matching query: %q\n", query)
			return
		}

		for _, member := range members {
			printMemberDetails(member)
		}
	},
}

func printMemberDetails(member MemberDetail) {
	divider := strings.Repeat("=", 80)
	fmt.Println(divider)
	fmt.Printf("MEMBER DETAILS: %s\n", formatStr(member.Name))
	fmt.Println(divider)

	fmt.Println("General Info")
	fmt.Printf("  Status:        			 %s\n", formatStr(member.Status))
	fmt.Printf("  Progress Talks:            %d\n", member.ProgressTalkNum)
	fmt.Printf("  Duration Active:           %s\n", formatStr(member.DurationActive))
	fmt.Printf("  Avg Time Between Talks:    %s\n", formatStr(member.AvgTimeBetweenTalks))
	fmt.Printf("  Meetups Since Last Talk:   %d\n", member.MeetupsSinceLastTalk)
	fmt.Println()

	fmt.Println("Projects & Talks:")
	if len(member.Projects) == 0 {
		fmt.Println("  No projects registered.")
	} else {
		for _, proj := range member.Projects {
			fmt.Printf("  - %s (category: %s)\n",
				formatStr(proj.Name), formatStr(proj.Category))
			if len(proj.Updates) == 0 {
				fmt.Println("      No updates/talks recorded.")
			} else {
				fmt.Println("      Updates:")
				for _, update := range proj.Updates {
					meetupNum := "Unknown"
					if update.Meetup.Number > 0 {
						meetupNum = fmt.Sprintf("#%d", update.Meetup.Number)
					}
					meetupDate := formatStr(update.Meetup.Date)
					fmt.Printf("        * %s (Meetup %s, %s): %s\n",
						formatCategory(update.Category), meetupNum, meetupDate, formatStr(update.Description))
				}
			}
		}
	}
	fmt.Println()
}

func formatStr(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func formatCategory(c string) string {
	switch c {
	case "idea_talk":
		return "Idea Talk"
	case "progress_talk":
		return "Progress Talk"
	default:
		return formatStr(c)
	}
}

func init() {
	rootCmd.AddCommand(viewCmd)
}
