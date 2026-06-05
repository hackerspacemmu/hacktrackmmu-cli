package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigDefaults(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("failed to load defaults: %v", err)
	}

	if cfg.LogLevel != "info" {
		t.Errorf("expected default log level 'info', got '%s'", cfg.LogLevel)
	}
}

func TestLoadConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "config.yaml")

	configData := []byte(`log_level: debug
api_url: test-key-123`)

	if err := os.WriteFile(tempFile, configData, 0644); err != nil {
		t.Fatalf("failed to write temp config file: %v", err)
	}

	cfg, err := Load(tempFile)
	if err != nil {
		t.Fatalf("failed to load config file: %v", err)
	}

	if cfg.LogLevel != "debug" {
		t.Errorf("expected log level 'debug', got '%s'", cfg.LogLevel)
	}
}

func TestLoadConfigEnvOverrides(t *testing.T) {
	t.Setenv("HACKTRACK_API_URL", "env-key-xyz")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("failed to load config with environment variables: %v", err)
	}

	if cfg.APIUrl != "env-key-xyz" {
		t.Errorf("expected api key overridden to 'env-key-xyz', got '%s'", cfg.APIUrl)
	}
}
