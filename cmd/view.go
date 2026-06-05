package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
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

		var buf strings.Builder
		for _, member := range members {
			printMemberDetails(&buf, member)
		}
		pageOutput(buf.String())
	},
}

const (
	ansiReset      = "\033[0m"
	ansiBold       = "\033[1m"
	ansiRed        = "\033[31m"
	ansiGreen      = "\033[32m"
	ansiYellow     = "\033[33m"
	ansiBlue       = "\033[34m"
	ansiMagenta    = "\033[35m"
	ansiCyan       = "\033[36m"
	ansiWhite      = "\033[37m"
	ansiBoldCyan   = "\033[1;36m"
	ansiBoldGreen  = "\033[1;32m"
	ansiBoldYellow = "\033[1;33m"
	ansiBoldBlue   = "\033[1;34m"
	ansiBoldWhite  = "\033[1;37m"
)

func colorize(text string, ansiCode string) string {
	if !isTerminal() {
		return text
	}
	return ansiCode + text + ansiReset
}

func printMemberDetails(w io.Writer, member MemberDetail) {
	divider := strings.Repeat("=", 80)
	fmt.Fprintln(w, colorize(divider, ansiCyan))
	fmt.Fprintf(w, "%s: %s\n", colorize("MEMBER DETAILS", ansiBoldWhite), colorize(formatStr(member.Name), ansiBoldCyan))
	fmt.Fprintln(w, colorize(divider, ansiCyan))

	fmt.Fprintln(w, colorize("General Info", ansiBold))

	statusStr := formatStr(member.Status)
	statusColor := ansiWhite
	if strings.ToLower(statusStr) == "active" {
		statusColor = ansiBoldGreen
	} else if strings.ToLower(statusStr) == "inactive" {
		statusColor = ansiRed
	} else {
		statusColor = ansiYellow
	}

	fmt.Fprintf(w, "  Status:                    %s\n", colorize(statusStr, statusColor))
	fmt.Fprintf(w, "  Progress Talks:            %s\n", colorize(fmt.Sprintf("%d", member.ProgressTalkNum), ansiBold))
	fmt.Fprintf(w, "  Duration Active:           %s\n", colorize(formatStr(member.DurationActive), ansiBold))
	fmt.Fprintf(w, "  Avg Time Between Talks:    %s\n", colorize(formatStr(member.AvgTimeBetweenTalks), ansiBold))
	fmt.Fprintf(w, "  Meetups Since Last Talk:   %s\n", colorize(fmt.Sprintf("%d", member.MeetupsSinceLastTalk), ansiBold))
	fmt.Fprintln(w)

	fmt.Fprintln(w, colorize(divider, ansiCyan))
	fmt.Fprintln(w, colorize("Projects & Talks", ansiBold))
	fmt.Fprintln(w, colorize(divider, ansiCyan))
	if len(member.Projects) == 0 {
		fmt.Fprintln(w, "  No projects registered.")
	} else {
		for _, proj := range member.Projects {
			projName := colorize(formatStr(proj.Name), ansiBoldYellow)
			projCategory := colorize(formatStr(proj.Category), ansiCyan)
			fmt.Fprintf(w, "  - %s (category: %s)\n", projName, projCategory)
			if len(proj.Updates) == 0 {
				fmt.Fprintln(w, "      No updates/talks recorded.")
			} else {
				fmt.Fprintln(w, "      Updates:")
				for _, update := range proj.Updates {
					meetupNum := "Unknown"
					if update.Meetup.Number > 0 {
						meetupNum = fmt.Sprintf("#%d", update.Meetup.Number)
					}
					meetupDate := formatStr(update.Meetup.Date)

					catStr := formatCategory(update.Category)
					var catColor string
					if strings.Contains(strings.ToLower(catStr), "progress") {
						catColor = ansiBoldGreen
					} else if strings.Contains(strings.ToLower(catStr), "idea") {
						catColor = ansiBoldCyan
					} else {
						catColor = ansiBoldWhite
					}

					coloredCat := colorize(catStr, catColor)
					coloredMeetupNum := colorize(meetupNum, ansiYellow)
					coloredMeetupDate := colorize(meetupDate, ansiWhite)

					prefixLen := 25 + len(catStr) + len(meetupNum) + len(meetupDate)
					subsequentIndent := "          " // 10 spaces
					wrappedDesc := wrapText(formatStr(update.Description), 80, prefixLen, subsequentIndent)

					fmt.Fprintf(w, "        * %s (Meetup %s, %s): %s\n",
						coloredCat, coloredMeetupNum, coloredMeetupDate, wrappedDesc)
				}
			}
			fmt.Fprintln(w)
		}
	}
	fmt.Fprintln(w)
}

func pageOutput(output string) {
	if !isTerminal() {
		fmt.Print(output)
		return
	}

	pager := os.Getenv("PAGER")
	if pager == "" {
		pager = "less"
	}

	var cmd *exec.Cmd
	if pager == "less" {
		cmd = exec.Command("less", "-R", "-F", "-X")
	} else {
		args := strings.Fields(pager)
		if len(args) > 1 {
			cmd = exec.Command(args[0], args[1:]...)
		} else if len(args) == 1 {
			cmd = exec.Command(args[0])
		} else {
			cmd = exec.Command("less", "-R", "-F", "-X")
		}
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		fmt.Print(output)
		return
	}

	if err := cmd.Start(); err != nil {
		fmt.Print(output)
		return
	}

	_, _ = io.WriteString(stdinPipe, output)
	stdinPipe.Close()

	_ = cmd.Wait()
}

func isTerminal() bool {
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
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

func wrapText(text string, limit int, firstLinePrefixLen int, subsequentIndent string) string {
	words := strings.Fields(strings.TrimSpace(text))
	if len(words) == 0 {
		return ""
	}

	var result strings.Builder
	currentLineLen := firstLinePrefixLen

	for i, word := range words {
		wordLen := len(word)

		if i == 0 {
			result.WriteString(word)
			currentLineLen += wordLen
		} else {
			if currentLineLen+1+wordLen > limit {
				result.WriteString("\n" + subsequentIndent + word)
				currentLineLen = len(subsequentIndent) + wordLen
			} else {
				result.WriteString(" " + word)
				currentLineLen += 1 + wordLen
			}
		}
	}

	return result.String()
}

func init() {
	rootCmd.AddCommand(viewCmd)
}
