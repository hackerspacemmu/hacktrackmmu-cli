package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type Session struct {
	Token      string    `json:"token"`
	ValidUntil time.Time `json:"valid_until"`
	IsAdmin    bool      `json:"is_admin"`
}

func getSessionFilePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(configDir, "hacktrackmmu-cli")
	// Ensure directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "session.json"), nil
}

// Load retrieves the session from disk.
func Load() (*Session, error) {
	path, err := getSessionFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}

	return &s, nil
}

// Save writes the session to disk.
func Save(s *Session) error {
	path, err := getSessionFilePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// IsValid checks if the session exists and has not expired.
// Since the Rails backend gives a 1-hour grace period, we count that.
func (s *Session) IsValid() bool {
	if s == nil || s.Token == "" {
		return false
	}
	// The Rails API validates token if: expiry_date + 1.hour > DateTime.now
	actualExpiry := s.ValidUntil.Add(1 * time.Hour)
	return time.Now().Before(actualExpiry)
}
