package version

import (
	"strings"
	"testing"
)

func TestGetInfo(t *testing.T) {
	Version = "1.2.3"

	info := GetInfo()

	if info.Version != "1.2.3" {
		t.Errorf("expected Version to be 1.2.3, got %s", info.Version)
	}
}

func TestInfoString(t *testing.T) {
	Version = "1.2.3"

	info := GetInfo()
	str := info.String()

	if !strings.Contains(str, "1.2.3") {
		t.Errorf("expected string to contain version, got %s", str)
	}
}
