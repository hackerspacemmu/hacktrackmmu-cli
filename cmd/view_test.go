package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintMemberDetails(t *testing.T) {
	member := MemberDetail{
		ID:                   1,
		Name:                 "John Doe",
		Status:               "Active",
		ProgressTalkNum:      3,
		DurationActive:       "6 months",
		AvgTimeBetweenTalks:  "2 weeks",
		MeetupsSinceLastTalk: 1,
		Projects: []ProjectInfo{
			{
				ID:        10,
				Name:      "Project Alpha",
				Category:  "Software",
				Completed: false,
				Updates: []UpdateInfo{
					{
						ID:          101,
						Category:    "progress_talk",
						Description: "Initial demo of project Alpha",
						Meetup: MeetupInfo{
							ID:     5,
							Number: 42,
							Date:   "2026-05-15",
						},
					},
					{
						ID:          102,
						Category:    "progress_talk",
						Description: "This is a very long update description that should definitely be wrapped because it exceeds the standard terminal width of eighty characters.",
						Meetup: MeetupInfo{
							ID:     6,
							Number: 43,
							Date:   "2026-06-01",
						},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	printMemberDetails(&buf, member)

	output := buf.String()

	expectedSubstrings := []string{
		"MEMBER DETAILS: John Doe",
		"Status:",
		"Active",
		"Progress Talks:            3",
		"Duration Active:           6 months",
		"Avg Time Between Talks:    2 weeks",
		"Meetups Since Last Talk:   1",
		"Projects & Talks",
		"- Project Alpha (category: Software)",
		"Progress Talk",
		"Meetup #42",
		"2026-05-15",
		"Initial demo of project Alpha",
		"This is a very long update",
		"\n          description that should definitely be wrapped",
	}

	for _, expected := range expectedSubstrings {
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain %q, but it did not. Output:\n%s", expected, output)
		}
	}
}
