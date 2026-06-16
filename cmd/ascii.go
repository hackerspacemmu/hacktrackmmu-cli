package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/hackerspace/hacktrackmmu-cli/internal/config"
)

const (
	colorBlue   = "\x1b[38;2;0;85;212m"
	colorRed    = "\x1b[38;2;255;42;42m"
	colorWhite  = "\x1b[38;2;255;255;255m\x1b[1m"
	colorYellow = "\x1b[38;2;255;193;7m\x1b[1m"
	colorReset  = "\x1b[0m"
)

func visualLen(s string) int {
	inEscape := false
	length := 0
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}
		length++
	}
	return length
}

func padRight(s string, width int) string {
	vl := visualLen(s)
	if vl >= width {
		return s
	}
	return s + strings.Repeat(" ", width-vl)
}

func fetchSingleCount(ctx context.Context, client *http.Client, urlStr string) (int, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return 0, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var cr map[string]interface{}
	if err := json.Unmarshal(body, &cr); err == nil {
		if totalVal, ok := cr["total"]; ok {
			if f, ok := totalVal.(float64); ok {
				return int(f), nil
			}
		}
		if countVal, ok := cr["count"]; ok {
			if f, ok := countVal.(float64); ok {
				return int(f), nil
			}
		}
	}

	var val int
	if _, err := fmt.Sscanf(strings.TrimSpace(string(body)), "%d", &val); err == nil {
		return val, nil
	}

	return 0, fmt.Errorf("invalid response format")
}

func fetchStats(cfg *config.Config) (daysActive int, members int, meetups int, updates int, success bool) {
	if cfg == nil || cfg.APIUrl == "" {
		return 0, 0, 0, 0, false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	var mu sync.Mutex
	results := make(map[string]int)
	errors := make(map[string]error)

	client := &http.Client{}

	// Members
	wg.Add(1)
	go func() {
		defer wg.Done()
		val, err := fetchSingleCount(ctx, client, cfg.APIUrl+"/members/count")
		mu.Lock()
		if err != nil {
			errors["members"] = err
		} else {
			results["members"] = val
		}
		mu.Unlock()
	}()

	// Meetups
	wg.Add(1)
	go func() {
		defer wg.Done()
		val, err := fetchSingleCount(ctx, client, cfg.APIUrl+"/meetup/count")
		if err != nil {
			val, err = fetchSingleCount(ctx, client, cfg.APIUrl+"/meetups/count")
		}
		mu.Lock()
		if err != nil {
			errors["meetups"] = err
		} else {
			results["meetups"] = val
		}
		mu.Unlock()
	}()

	// Updates
	wg.Add(1)
	go func() {
		defer wg.Done()
		val, err := fetchSingleCount(ctx, client, cfg.APIUrl+"/updates/count")
		mu.Lock()
		if err != nil {
			errors["updates"] = err
		} else {
			results["updates"] = val
		}
		mu.Unlock()
	}()

	wg.Wait()

	startDate := time.Date(2011, time.June, 9, 0, 0, 0, 0, time.UTC)
	now := time.Now()
	daysActive = daysBetween(startDate, now)

	if len(errors) > 0 {
		return 0, 0, 0, 0, false
	}

	return daysActive, results["members"], results["meetups"], results["updates"], true
}

func daysBetween(start, end time.Time) int {
	y1, m1, d1 := start.Date()
	y2, m2, d2 := end.Date()

	startUTC := time.Date(y1, m1, d1, 0, 0, 0, 0, time.UTC)
	endUTC := time.Date(y2, m2, d2, 0, 0, 0, 0, time.UTC)

	diff := endUTC.Sub(startUTC)
	days := int(diff.Hours() / 24)

	return days
}

func PrintAscii() {
	// daysActive, members, meetups, updates, success := fetchStats(cfg)

	// daysActiveStr := "N/A"
	// membersStr := "N/A"
	// meetupsStr := "N/A"
	// updatesStr := "N/A"
	// if success {
	// 	daysActiveStr = fmt.Sprintf("%d", daysActive)
	// 	membersStr = fmt.Sprintf("%d", members)
	// 	meetupsStr = fmt.Sprintf("%d", meetups)
	// 	updatesStr = fmt.Sprintf("%d", updates)
	// }

	lines := []string{
		"",
		"                 " + colorBlue + "###################" + colorReset,
		"                " + colorBlue + "###################" + colorReset,
		"                " + colorBlue + "#######" + colorReset,
		"     " + colorWhite + "@@@   @@@" + colorReset + "  " + colorBlue + "#####     #" + colorReset + "  " + colorWhite + "@@@    @@@" + colorReset,
		"     " + colorWhite + "@@@   @@@" + colorReset + "  " + colorBlue + "###.      #" + colorReset + "  " + colorWhite + "@@@@  %@@@" + colorReset,
		"     " + colorWhite + "@@@@@@@@@" + colorReset + "  " + colorBlue + "###" + colorReset + " " + colorRed + "***" + colorReset + "  " + colorBlue + "##" + colorReset + "  " + colorWhite + "@@@@  @@@@" + colorReset,
		"     " + colorWhite + "@@@@@@@@@" + colorReset + "  " + colorBlue + "##" + colorReset + "  " + colorRed + "***" + colorReset + " " + colorBlue + "###" + colorReset + "  " + colorWhite + "@@@@@@@@@@" + colorReset,
		"     " + colorWhite + "@@@   @@@" + colorReset + "  " + colorBlue + "#      =###" + colorReset + "  " + colorWhite + "@@ @@@@ @@" + colorReset,
		"     " + colorWhite + "@@@   @@@" + colorReset + "  " + colorBlue + "#     #####" + colorReset + "  " + colorWhite + "@@  @@  @@" + colorReset,
		"                    " + colorBlue + "#######" + colorReset,
		"       " + colorBlue + ":###################" + colorReset,
		"       " + colorBlue + "###################" + colorReset,
		"",
	}

	// lines[4] = padRight(lines[4], 50) + colorYellow + "Statistics: " + colorReset
	// lines[5] = padRight(lines[5], 50) + colorWhite + "--------------" + colorReset
	// lines[6] = padRight(lines[6], 50) + colorWhite + "Members: " + colorReset + colorYellow + membersStr + colorReset
	// lines[7] = padRight(lines[7], 50) + colorWhite + "Meetups: " + colorReset + colorYellow + meetupsStr + colorReset
	// lines[8] = padRight(lines[8], 50) + colorWhite + "Updates: " + colorReset + colorYellow + updatesStr + colorReset
	// lines[9] = padRight(lines[9], 50) + colorWhite + "Days Active: " + colorReset + colorYellow + daysActiveStr + colorReset

	for _, line := range lines {
		fmt.Println(line)
	}
}
